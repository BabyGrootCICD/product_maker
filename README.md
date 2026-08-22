# product_maker — 量化雙軌 ➔ 質化收斂

內部探索工具：兩條量化管線（Open Data、GitHub Issues）產出分數化 Insight，匯聚後寫入本 repo Issues 作為質化訪談佇列。

## 架構

```text
量化 A  Open Data (PCC + SAM.gov)  →  opportunity_alert
量化 B  GitHub Issues GraphQL       →  pain_point
              ↓ Convergence (theme / keyword overlap)
         質化 CRM (Digest + signal issues) + 可選 xAI outreach
```

## 快速開始

```bash
go test ./...
go run ./cmd/discovery run --config configs/pipeline.yaml --dry-run
```

## CLI

| 指令 | 說明 |
| --- | --- |
| `discovery run` | 完整管線：量化 A/B → 匯聚 → CRM → 可選 LLM |
| `discovery opendata` | 只跑量化 A |
| `discovery issues` | 只跑量化 B |
| `discovery converge` | 從 `out/*.json` 匯聚 |
| `discovery synthesize` | 對 briefing 產生 outreach 草稿 |
| `discovery jtbd --transcript file` | 本機 JTBD 分析（不進排程） |

## 環境變數

| 變數 | 必填 | 說明 |
| --- | --- | --- |
| `GITHUB_TOKEN` | GHA 自動提供 | GraphQL 讀取 + Issues CRM 寫入 |
| `GITHUB_REPOSITORY` | GHA 自動提供 | CRM 目標 repo |
| `SAM_API_KEY` | 建議 | [SAM.gov](https://sam.gov) Public API Key |
| `XAI_API_KEY` | 可選 | xAI outreach / JTBD；缺 key 時量化仍成功 |
| `GITHUB_STEP_SUMMARY` | GHA 自動提供 | Job summary markdown |

## GitHub Actions

- **CI** (`.github/workflows/ci.yml`)：PR / push 跑 `go test ./...`
- **Discovery** (`.github/workflows/discovery.yml`)：每週一 00:00 UTC（台北 08:00）+ 手動觸發

### Secrets 設定

在 repo **Settings → Secrets and variables → Actions** 新增：

- `SAM_API_KEY` — 從 SAM.gov Account Details 申請 Public API Key
- `XAI_API_KEY` — 從 [console.x.ai](https://console.x.ai) 建立（可選）

## 設定檔

- [`configs/pipeline.yaml`](configs/pipeline.yaml) — 閾值、NAICS、LLM 模型
- [`configs/repos.yaml`](configs/repos.yaml) — GitHub 掃描 repo 與 theme

## 輸出 artifacts

| 檔案 | 內容 |
| --- | --- |
| `out/opendata.json` | 量化 A insights |
| `out/issues.json` | 量化 B insights |
| `out/briefing.json` | 匯聚結果含 overlaps |
| `out/digest.md` | 週報 markdown |

## 授權與資料源

- **PCC v1**：使用 [pcc.mlwmlw.org](https://pcc.mlwmlw.org/api)（g0v 社群鏡像）；商業用途請改官方開放資料 ZIP。
- **SAM.gov**：官方 Opportunities API v2；個人帳號約 10 req/day，管線已限制每週 request 預算。

## 明確不做（v1）

- 自動私訊 GitHub 使用者
- 訪談逐字稿進週排程
- Notion / Airtable CRM
