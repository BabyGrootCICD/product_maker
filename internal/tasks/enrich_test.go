package tasks

import (
	"context"
	"testing"

	"github.com/BabyGrootCICD/product_maker/internal/qual/xai"
)

type fakeRater struct {
	d, r int
}

func (f fakeRater) RateTaskAxes(ctx context.Context, title, summary, pipeline, url string, metrics map[string]float64) (xai.TaskAxes, error) {
	return xai.TaskAxes{Difficulty: f.d, Risk: f.r, Rationale: "from xai"}, nil
}

func TestEnrichAxesOverridesHeuristic(t *testing.T) {
	list := []Task{
		{Fingerprint: "a", Score: 10, ScoreNorm: 1, Reward: 2, Difficulty: 1, Risk: 1, AxesSource: AxesSourceHeuristic},
	}
	EnrichAxes(context.Background(), list, fakeRater{d: 5, r: 4}, 15)
	if list[0].Difficulty != 5 || list[0].Risk != 4 || list[0].AxesSource != AxesSourceXAI {
		t.Fatalf("%+v", list[0])
	}
}
