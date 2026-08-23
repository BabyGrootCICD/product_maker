# 申請 SAM.gov Public API Key（`SAM_API_KEY`）

本專案量化 A（Open Data）會呼叫 [SAM.gov Opportunities API v2](https://open.gsa.gov/api/get-opportunities-public-api/) 掃描美國聯邦採購公告。沒有 key 時管線會寫入 skip：`SAM.gov skipped: missing SAM_API_KEY`，其餘階段仍會繼續。

官方說明：每位使用者須在 SAM.gov **Account Details** 頁申請 **Public API Key**（個人帳號 key）。

---

## 你會得到什麼

| 項目 | 說明 |
| --- | --- |
| 名稱 | Public API Key（文件裡有時也稱 Individual Account API Key） |
| 用途 | Query parameter：`api_key=...` |
| 生產端點 | `https://api.sam.gov/opportunities/v2/search` |
| 基本帳號配額 | 常見約 **10 requests / day**（Entity 註冊後可更高） |
| 本專案用量 | 每週 job 最多約 8 次（見 `configs/pipeline.yaml` → `samgov.max_requests`） |

---

## 步驟 1：註冊並登入 SAM.gov

1. 開啟 [https://sam.gov](https://sam.gov)。
2. 右上角選擇 **Sign In** / **Create an account**。
3. 用個人信箱完成註冊與 email 驗證。
4. 登入後進入你的 Workspace / 個人首頁。

> 不需要先完成 Entity registration 才能申請 Public API Key；基本個人帳號即可（配額較低）。

---

## 步驟 2：打開 Account Details

1. 登入後點右上角帳號選單（頭像 / 名字）。
2. 進入 **Account Details**（有時在 Profile / Workspace → Account）。
3. 捲到 **Public API Key** 區塊。

官方文件也寫明：key 是在 Account Details 頁申請／檢視。參考：[Get Opportunities Public API](https://open.gsa.gov/api/get-opportunities-public-api/)。

---

## 步驟 3：Request API Key

1. 在 **Public API Key** 區塊按 **Request API Key**。
2. 依畫面提示輸入帳號密碼（部分環境會要求再驗證一次才能顯示 key）。
3. 產生後，key **會立刻顯示**；離開該頁後，之後再看通常需要重新輸入密碼才能顯示。

**請立刻複製並安全保存。** 不要貼到公開 issue、PR、commit 或 chat。

---

## 步驟 4：本地驗證（可選）

把 `YOUR_KEY` 換成剛申請的 key，確認能拿到 200：

```bash
curl -sS -o /tmp/sam.json -w "%{http_code}\n" \
  "https://api.sam.gov/opportunities/v2/search?api_key=YOUR_KEY&postedFrom=01/01/2026&postedTo=01/07/2026&ncode=541512&limit=10&offset=0"

# 預期輸出：200
# 內容應含 opportunitiesData 陣列
python3 -c "import json; d=json.load(open('/tmp/sam.json')); print(len(d.get('opportunitiesData', [])))"
```

常見錯誤：

| HTTP | 可能原因 |
| --- | --- |
| 401 / 403 | key 錯誤、未啟用、或權限不足 |
| 429 | 當日配額用盡（基本帳號約 10 次／日） |
| 400 | `postedFrom` / `postedTo` 格式錯誤（須 `MM/dd/yyyy`）或缺少必填參數 |

---

## 步驟 5：寫入 GitHub Actions Secret

本專案 workflow [`.github/workflows/discovery.yml`](../.github/workflows/discovery.yml) 讀取 secret 名稱：`SAM_API_KEY`。

1. 開啟 GitHub repo → **Settings**。
2. 左側 **Secrets and variables** → **Actions**。
3. **New repository secret**：
   - Name：`SAM_API_KEY`（必須完全一致）
   - Secret：貼上剛才複製的 Public API Key
4. 儲存。

本地 dry-run 也可：

```bash
export SAM_API_KEY='your-key-here'
go run ./cmd/discovery run --config configs/pipeline.yaml --dry-run
```

---

## 步驟 6：手動跑一次 Discovery pipeline

1. GitHub → **Actions** → **Discovery Pipeline**。
2. **Run workflow**（`workflow_dispatch`）。
3. 跑完後看 Job Summary：
   - `SAM.gov skipped: missing SAM_API_KEY` 應消失。
   - `Open data insights` 可能仍為 0（若該週 NAICS 未過閾值），但 Skips 不應再含 SAM。

---

## 配額與本專案設計

- 個人 Public API Key：**約 10 req/day**。
- 本管線每週最多打 **4 個 NAICS × 1 次**（`max_requests: 8` 為上限）。
- 日期窗：`lookback_days: 14`（可在 `configs/pipeline.yaml` 調整）。
- SAM **沒有**與台灣 PCC「流標率」對等欄位；本專案用 `Sources Sought` / `Special Notice` 占比當 `pain_proxy`，不要當成真實流標率。

若需要更高配額，需走 Entity / System Account（流程較長，見 GSA / SAM System Account 指南）。探索期個人 key 通常夠用。

---

## 安全注意

- 只放在 GitHub Actions Secrets 或本機環境變數。
- 不要 commit 到 `configs/`、`.env`（若有）、或 workflow YAML 明文。
- 輪替／外洩時：回 Account Details 重新產生（或依 SAM 介面 revoke）並更新 secret。

---

## 相關連結

- [SAM.gov](https://sam.gov)
- [Get Opportunities Public API 文件](https://open.gsa.gov/api/get-opportunities-public-api/)
- 本專案設定：[`configs/pipeline.yaml`](../configs/pipeline.yaml) → `opendata.samgov`
- 本專案 README：[`README.md`](../README.md)
