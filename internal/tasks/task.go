package tasks

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BabyGrootCICD/product_maker/internal/converge"
	"github.com/BabyGrootCICD/product_maker/internal/domain"
)

const (
	AxesSourceXAI       = "xai"
	AxesSourceHeuristic = "heuristic"
)

type Task struct {
	Fingerprint string
	Title       string
	Pipeline    string
	URL         string
	Summary     string
	Score       float64
	ScoreNorm   float64
	Reward      float64
	Difficulty  int
	Risk        int
	AxesSource  string
	Rationale   string
	Priority    float64
	Seen        string
	Stale       bool
	Metrics     map[string]float64
}

var (
	fpRe       = regexp.MustCompile(`<!--\s*fingerprint:([^\s]+)\s*-->`)
	headingRe  = regexp.MustCompile(`(?m)^###\s+\d+\.\s+\[([^\]]+)\]\s+(.+)$`)
	fieldRe    = regexp.MustCompile(`(?m)^-\s+\*\*([^*]+)\*\*:\s*(.+)$`)
	notesSplit = regexp.MustCompile(`(?m)^## Notes\s*$`)
)

func FromBriefing(b domain.Briefing, week string) []Task {
	if week == "" {
		week = ISOWeek(time.Now())
	}

	var out []Task
	matchedOD := map[string]bool{}
	matchedIS := map[string]bool{}

	for _, o := range b.Overlaps {
		matchedOD[o.OpenData.Fingerprint] = true
		matchedIS[o.Issue.Fingerprint] = true
		out = append(out, Task{
			Fingerprint: "overlap:" + o.OpenData.Fingerprint + "+" + o.Issue.Fingerprint,
			Title:       o.Issue.Title,
			Pipeline:    "overlap",
			URL:         o.Issue.URL,
			Summary:     fmt.Sprintf("%s | %s (%s)", o.OpenData.Summary, o.Issue.Summary, o.Reason),
			Score:       o.OpenData.Score + o.Issue.Score,
			Reward:      3,
			Seen:        week,
			Metrics:     mergeMetrics(o.OpenData.Metrics, o.Issue.Metrics),
		})
	}

	unmatchedOD, unmatchedIS := converge.UnmatchedTop(b, len(b.RankedOpenData)+len(b.RankedIssues))
	for _, od := range unmatchedOD {
		if matchedOD[od.Fingerprint] {
			continue
		}
		out = append(out, taskFromInsight(od, 1, week))
	}
	for _, is := range unmatchedIS {
		if matchedIS[is.Fingerprint] {
			continue
		}
		out = append(out, taskFromInsight(is, 2, week))
	}
	for _, od := range b.RankedOpenData {
		if matchedOD[od.Fingerprint] || containsFP(out, od.Fingerprint) {
			continue
		}
		out = append(out, taskFromInsight(od, 1, week))
	}
	for _, is := range b.RankedIssues {
		if matchedIS[is.Fingerprint] || containsFP(out, is.Fingerprint) {
			continue
		}
		out = append(out, taskFromInsight(is, 2, week))
	}

	ApplyHeuristicAxes(out)
	NormalizeScores(out)
	Prioritize(out)
	return out
}

func taskFromInsight(ins domain.Insight, reward float64, week string) Task {
	return Task{
		Fingerprint: ins.Fingerprint,
		Title:       ins.Title,
		Pipeline:    string(ins.Pipeline),
		URL:         ins.URL,
		Summary:     ins.Summary,
		Score:       ins.Score,
		Reward:      reward,
		Seen:        week,
		Metrics:     ins.Metrics,
	}
}

func ApplyHeuristicAxes(list []Task) {
	for i := range list {
		t := &list[i]
		t.Difficulty = heuristicDifficulty(t.Pipeline)
		t.Risk = heuristicRisk(t.Metrics)
		t.AxesSource = AxesSourceHeuristic
		if t.Rationale == "" {
			t.Rationale = "heuristic axes"
		}
	}
}

func heuristicDifficulty(pipeline string) int {
	switch pipeline {
	case "overlap":
		return 1
	case string(domain.PipelineIssues):
		return 2
	default:
		return 3
	}
}

func heuristicRisk(metrics map[string]float64) int {
	if metrics != nil {
		if c, ok := metrics["comments"]; ok && c > 200 {
			return 2
		}
	}
	return 1
}

func NormalizeScores(list []Task) {
	if len(list) == 0 {
		return
	}
	minS, maxS := list[0].Score, list[0].Score
	for _, t := range list {
		if t.Score < minS {
			minS = t.Score
		}
		if t.Score > maxS {
			maxS = t.Score
		}
	}
	span := maxS - minS
	for i := range list {
		if span <= 0 {
			list[i].ScoreNorm = 1.0
			continue
		}
		list[i].ScoreNorm = 0.1 + 0.9*((list[i].Score-minS)/span)
	}
}

func Prioritize(list []Task) {
	for i := range list {
		t := &list[i]
		d := float64(t.Difficulty)
		r := float64(t.Risk)
		if d < 1 {
			d = 1
		}
		if r < 1 {
			r = 1
		}
		sn := t.ScoreNorm
		if sn <= 0 {
			sn = 0.1
		}
		t.Priority = sn * t.Reward / (d * r)
		if math.IsNaN(t.Priority) || math.IsInf(t.Priority, 0) {
			t.Priority = 0
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Stale != list[j].Stale {
			return !list[i].Stale && list[j].Stale
		}
		if list[i].Priority != list[j].Priority {
			return list[i].Priority > list[j].Priority
		}
		return list[i].Score > list[j].Score
	})
}

func Merge(existing, incoming []Task, week string) []Task {
	byFP := map[string]*Task{}
	var order []string
	for i := range existing {
		t := existing[i]
		t.Stale = true
		cp := t
		byFP[t.Fingerprint] = &cp
		order = append(order, t.Fingerprint)
	}
	for _, in := range incoming {
		if prev, ok := byFP[in.Fingerprint]; ok {
			prev.Title = in.Title
			prev.Pipeline = in.Pipeline
			prev.URL = in.URL
			prev.Summary = in.Summary
			prev.Score = in.Score
			prev.ScoreNorm = in.ScoreNorm
			prev.Reward = in.Reward
			prev.Difficulty = in.Difficulty
			prev.Risk = in.Risk
			prev.AxesSource = in.AxesSource
			prev.Rationale = in.Rationale
			prev.Priority = in.Priority
			prev.Seen = week
			prev.Stale = false
			prev.Metrics = in.Metrics
			continue
		}
		cp := in
		cp.Seen = week
		cp.Stale = false
		byFP[in.Fingerprint] = &cp
		order = append(order, in.Fingerprint)
	}
	var out []Task
	for _, fp := range order {
		out = append(out, *byFP[fp])
	}
	Prioritize(out)
	return out
}

func SelectTop(list []Task, n int, skipStale bool) []Task {
	var filtered []Task
	for _, t := range list {
		if skipStale && t.Stale {
			continue
		}
		filtered = append(filtered, t)
	}
	if n > len(filtered) {
		n = len(filtered)
	}
	if n < 0 {
		n = 0
	}
	return filtered[:n]
}

func Render(list []Task, week, notes string) string {
	var b strings.Builder
	b.WriteString("# Discovery Tasks\n\n")
	b.WriteString(fmt.Sprintf("> Last updated: %s · source: discovery pipeline\n\n", week))
	b.WriteString("## Ranked backlog\n\n")
	if len(list) == 0 {
		b.WriteString("_No tasks this week._\n")
	}
	for i, t := range list {
		staleMark := ""
		if t.Stale {
			staleMark = " · stale"
		}
		b.WriteString(fmt.Sprintf("### %d. [%s] %s%s\n", i+1, t.Pipeline, t.Title, staleMark))
		b.WriteString(fmt.Sprintf("<!-- fingerprint:%s -->\n", t.Fingerprint))
		b.WriteString(fmt.Sprintf("- **priority**: %.2f\n", t.Priority))
		b.WriteString(fmt.Sprintf("- **score**: %.1f\n", t.Score))
		b.WriteString(fmt.Sprintf("- **axes**: reward=%.0f difficulty=%d risk=%d source=%s\n",
			t.Reward, t.Difficulty, t.Risk, t.AxesSource))
		if t.Rationale != "" {
			b.WriteString(fmt.Sprintf("- **rationale**: %s\n", sanitizeOneLine(t.Rationale)))
		}
		b.WriteString(fmt.Sprintf("- **pipeline**: %s\n", t.Pipeline))
		url := t.URL
		if url == "" {
			url = "_n/a_"
		}
		b.WriteString(fmt.Sprintf("- **url**: %s\n", url))
		b.WriteString(fmt.Sprintf("- **summary**: %s\n", sanitizeOneLine(t.Summary)))
		b.WriteString(fmt.Sprintf("- **seen**: %s\n", t.Seen))
		if t.Stale {
			b.WriteString("- **stale**: true\n")
		}
		b.WriteString("\n")
	}
	if notes != "" {
		b.WriteString("## Notes\n\n")
		b.WriteString(strings.TrimSpace(notes))
		b.WriteString("\n")
	}
	return b.String()
}

func Parse(md string) (list []Task, notes string) {
	parts := notesSplit.Split(md, 2)
	body := parts[0]
	if len(parts) == 2 {
		notes = strings.TrimSpace(parts[1])
	}
	for _, block := range splitTaskBlocks(body) {
		t, ok := parseBlock(block)
		if ok {
			list = append(list, t)
		}
	}
	return list, notes
}

func splitTaskBlocks(md string) []string {
	idxs := headingRe.FindAllStringIndex(md, -1)
	if len(idxs) == 0 {
		return nil
	}
	var blocks []string
	for i, loc := range idxs {
		start := loc[0]
		end := len(md)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		blocks = append(blocks, md[start:end])
	}
	return blocks
}

func parseBlock(block string) (Task, bool) {
	hm := headingRe.FindStringSubmatch(block)
	if hm == nil {
		return Task{}, false
	}
	fm := fpRe.FindStringSubmatch(block)
	if fm == nil {
		return Task{}, false
	}
	t := Task{
		Fingerprint: fm[1],
		Pipeline:    hm[1],
		Title:       strings.TrimSuffix(strings.TrimSpace(hm[2]), " · stale"),
		Metrics:     map[string]float64{},
	}
	for _, m := range fieldRe.FindAllStringSubmatch(block, -1) {
		key := strings.ToLower(strings.TrimSpace(m[1]))
		val := strings.TrimSpace(m[2])
		switch key {
		case "priority":
			t.Priority, _ = strconv.ParseFloat(val, 64)
		case "score":
			t.Score, _ = strconv.ParseFloat(val, 64)
		case "axes":
			parseAxes(val, &t)
		case "rationale":
			t.Rationale = val
		case "pipeline":
			t.Pipeline = val
		case "url":
			if val != "_n/a_" {
				t.URL = val
			}
		case "summary":
			t.Summary = val
		case "seen":
			t.Seen = val
		case "stale":
			t.Stale = val == "true"
		}
	}
	if strings.Contains(hm[2], "· stale") {
		t.Stale = true
	}
	if t.Reward == 0 {
		t.Reward = 1
	}
	if t.Difficulty == 0 {
		t.Difficulty = heuristicDifficulty(t.Pipeline)
	}
	if t.Risk == 0 {
		t.Risk = 1
	}
	if t.AxesSource == "" {
		t.AxesSource = AxesSourceHeuristic
	}
	return t, true
}

func parseAxes(val string, t *Task) {
	for _, part := range strings.Fields(val) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "reward":
			t.Reward, _ = strconv.ParseFloat(kv[1], 64)
		case "difficulty":
			t.Difficulty, _ = strconv.Atoi(kv[1])
		case "risk":
			t.Risk, _ = strconv.Atoi(kv[1])
		case "source":
			t.AxesSource = kv[1]
		}
	}
}

func containsFP(list []Task, fp string) bool {
	for _, t := range list {
		if t.Fingerprint == fp {
			return true
		}
	}
	return false
}

func mergeMetrics(a, b map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func sanitizeOneLine(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
}

func ClampAxes(difficulty, risk int) (int, int) {
	if difficulty < 1 {
		difficulty = 1
	}
	if difficulty > 5 {
		difficulty = 5
	}
	if risk < 1 {
		risk = 1
	}
	if risk > 5 {
		risk = 5
	}
	return difficulty, risk
}

func ISOWeek(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}
