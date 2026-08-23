# product_maker — 量化雙軌 ➔ 質化收斂

內部探索工具：兩條量化管線（Open Data、GitHub Issues）產出分數化 Insight，匯聚後寫入 `tasks.md` 與本 repo Issues；可選 xAI 評難度／風險、破冰草稿，並為 top 任務開 discovery branch。

## 架構

```text
量化 A  Open Data (PCC + SAM.gov)  →  opportunity_alert
量化 B  GitHub Issues GraphQL       →  pain_point
              ↓ Convergence
         tasks.md（去重 + xAI difficulty/risk → priority）
         質化 CRM (Digest + signal issues) + xAI outreach
              ↓
         discovery/<week>/<hash> branches（xAI solution paths）
```

## 快速開始

```bash
go test ./...
go run ./cmd/discovery run --config configs/pipeline.yaml --dry-run
go run ./cmd/discovery tasks --briefing out/briefing.json
go run ./cmd/discovery branches --tasks tasks.md --dry-run
```

## CLI

| 指令 | 說明 |
| --- | --- |
| `discovery run` | 量化 A/B → 匯聚 → `tasks.md` → CRM → outreach |
| `discovery opendata` | 只跑量化 A |
| `discovery issues` | 只跑量化 B |
| `discovery converge` | 從 `out/*.json` 匯聚 |
| `discovery tasks` | 從 briefing 寫／合併 `tasks.md`（可呼叫 xAI axes） |
| `discovery branches` | 讀 `tasks.md` top-N，xAI solution path，建 git branch |
| `discovery synthesize` | 對 briefing 產生 outreach 草稿 |
| `discovery jtbd --transcript file` | 本機 JTBD 分析（不進排程） |

## 環境變數

| 變數 | 必填 | 說明 |
| --- | --- | --- |
| `GITHUB_TOKEN` | GHA 自動提供 | GraphQL、CRM、push branches |
| `GITHUB_REPOSITORY` | GHA 自動提供 | CRM 目標 repo |
| `SAM_API_KEY` | 建議 | [SAM.gov](https://sam.gov) Public API Key（[`docs/sam-api-key.md`](docs/sam-api-key.md)） |
| `XAI_API_KEY` | 建議 | axes／outreach／branches／JTBD；缺 key 時 axes 用啟發式、branches skip |
| `GITHUB_STEP_SUMMARY` | GHA 自動提供 | Job summary markdown |

## GitHub Actions

- **CI** (`.github/workflows/ci.yml`)：PR / push 跑 `go test ./...`
- **Discovery** (`.github/workflows/discovery.yml`)：每週一 00:00 UTC + 手動觸發  
  順序：`run` → upload artifacts → **commit `tasks.md`** → **`discovery branches`**

### Secrets

- `SAM_API_KEY` — 見 [`docs/sam-api-key.md`](docs/sam-api-key.md)
- `XAI_API_KEY` — [console.x.ai](https://console.x.ai)

## 設定檔

- [`configs/pipeline.yaml`](configs/pipeline.yaml) — 含 `tasks.axes_top_n` / `tasks.branch_top_n`
- [`configs/repos.yaml`](configs/repos.yaml) — GitHub 掃描 repo

## 輸出

| 檔案／產物 | 內容 |
| --- | --- |
| `out/opendata.json` | 量化 A |
| `out/issues.json` | 量化 B |
| `out/briefing.json` | 匯聚 |
| `out/digest.md` | 週報 |
| [`tasks.md`](tasks.md) | 去重 backlog（priority = score_norm × reward / (difficulty × risk)） |
| `discovery/<week>/<hash>` | brief branch：`discovery/briefs/<hash>.md` |

## 授權與資料源

- **PCC v1**：[pcc.mlwmlw.org](https://pcc.mlwmlw.org/api)；商業用途請改官方 ZIP。
- **SAM.gov**：Opportunities API v2；個人帳號約 10 req/day。

## 明確不做（v1）

- 自動私訊 GitHub 使用者
- 訪談逐字稿進週排程
- 自動開 PR／merge discovery branches
- Notion / Airtable CRM
