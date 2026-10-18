# 工作歷程

## 2026-10-18：初始化與開工

- 目標：建立 AGENTS，參考 softworld_san，以 dosgolem 推進繁中、HD 與工具列手把；使用者接續授權建立私人 GitHub 儲存庫與 README。
- 載入：知識路由、reverse-engineer-retro-game-remake 技能、文件職責、dosgolem README／CLAUDE、參考專案規則及 HD／工具列入口。共同決策契約已讀，未將參考專案視覺取捨當成本專案授權。
- 輸入：原始 ZIP CRC 通過並計算 SHA-256；RAR 目錄確認 17 張掃描。人工讀手冊第 1、7–8 頁，README 背景據此改寫。
- 變更：建立 AGENTS、README、CONTEXT、WORKLIST、RESEARCH-LOG 與 VERIFICATION-MATRIX，索引集中在 AGENTS。
- 工作區理由：現有根目錄只有原始包與唯讀環境目錄；開工後新增 `workplace/` 保存可重生的解包、probe 與工具中間產物，不另造研究目錄。原始 ZIP／RAR 保留原處、唯讀使用。
- `tools/` 保存需版控的盤點與收據腳本，與不入庫的 `workplace/` 分開；入口掛入 AGENTS 文件索引。
- 使用者選定直接執行原版的架構，新增 psychic-war 為覆繪參考。`docs/spec/` 保存架構契約，`cmd/` 與 `internal/prototype/` 保存可撤回前端，`text/` 保存其可見文字；均有 AGENTS 入口，不建立第二個研究工作區。
- GitHub 阻塞：依使用者授權請求在主機執行 `/home/anr2/.local/bin/gh auth status`，工具政策拒絕提升權限。這是執行環境拒絕，不能稱為憑證失效。尚未建立或推送私人儲存庫。
- 本地 Git：目前 `.git/` 是唯讀環境目錄，先前狀態檢查回報非 Git 儲存庫；不覆寫環境目錄。
- Docker：本輪沿用既有 Python 與 dkk-inventory 映像；一次性容器使用 `--rm`、UID/GID 1000:1000 與資源限制。原始輸入唯讀。收尾清理與開工驗證結果待本輪追加。

## 2026-10-18：前端雛形與首句覆繪

- 決策：使用者明確選定直接執行原版，再指定 psychic-war 的覆繪方式。架構寫入 `docs/spec/001-runtime-overlay.md`；HD 畫風與正式版面仍未定案。
- 變更：加入 `cmd/zeliard-preview`、中文工具列、裝置選擇、按鈕／軸綁定、死區、獨立設定保存、失焦／斷線釋放，以及可選片頭單句覆繪。所有功能仍限 prototype；未接 HD 圖面或音效裝置。
- 驗證：`go vet` 通過；三項輸入與設定測試、真實原版畫格覆繪測試通過。GUI v4 的三條流程通過，實際視窗中文字已目視確認。區域比較見 RESEARCH-LOG。
- 環境修正：Go 在容器預設開出的平行工作超過 PID 上限，限制 `GOMAXPROCS=2` 與 `-p` 後乾淨重跑成功。沒有以跳過 vet 解決。
- 擷取修正：初版實際 X 視窗擷取早於事件處理，改為等待 Draw metadata，再擷取視窗。v2、v3 的單句匹配失敗保留；固定時鐘的原版基線重生後，v4 成功。
- `xvfb-run` 的額外診斷卡在啟動等待，已停止並刪除該容器。正式 GUI 驗證改用 `tools/gui_smoke.py` 的有界 Xvfb 擁有程序，`finally` 回收前端與 Xvfb。
- GitHub：一般權限的 `gh auth status` 也未成功，但受限環境結果不能證明主機憑證失效。未登出、未搬運憑證，未建立遠端或進行 commit／push。
- 可寫研究快照再次核對 882 檔與既有 digest 相同；29 個遊戲檔案盤點可重生。沒有修改共用 dosgolem。

### 重跑入口

全部工作負載在 Docker 內執行；下列容器內命令需搭配本輪已驗證的掛載與環境。

| 項目 | 映像及必要掛載 | 容器內命令 |
|---|---|---|
| 素材與工具快照 | Python 3.13；專案 `/project` 可寫、dosgolem `/source` 唯讀 | `python /project/tools/prepare.py --source-head <來源 HEAD>` |
| 建置探針 | dosgolem-go-ebiten:1.24-r1；快照為 cwd、模組 cache 唯讀、GOCACHE 可寫 | `go build -p 2 -mod=readonly -o /workplace/dosgolem-probe ./cmd/probe` |
| 即時片頭基線 | Python 3.13；原版 `/orig`、探針 `/probe` 唯讀，專案可寫 | `python /project/tools/probe_receipt.py --steps 10729940 --cycles 3000 --output /project/workplace/<新的基線目錄>` |
| 前端 | dosgolem-go-ebiten:1.24-r1；cwd `/project`，GOMAXPROCS=2，GOPROXY=off，模組 cache 唯讀 | `go build -p 2 -mod=readonly -o workplace/bin/zeliard-preview ./cmd/zeliard-preview` |
| 靜態與輸入測試 | 同前端；另掛字型並設定下列三個測試變數 | `go vet -p 1 -mod=readonly ./...`；`go test -p 1 -mod=readonly -v ./internal/prototype/...` |
| GUI | psychicwar-go-ebiten:latest；專案 `/project` 可寫、原版 `/orig` 與字型 `/fonts` 唯讀 | `python3 /project/tools/gui_smoke.py --output /project/workplace/<新的 GUI 收據目錄>` |

測試變數：`ZELIARD_REFERENCE=/project/workplace/probe-intro-clock3000/screen.bin`、`ZELIARD_FONT=/fonts/unifont.hex.gz`、`ZELIARD_INTRO_TEXT=/project/text/intro-prototype.json`。未掛載時覆繪測試會 skip，不算完成。

本輪編譯共用 cache：主機 `/home/anr2/cht/dosgolem/workplace/gomodcache` 唯讀掛載為 `/gomodcache`，`GOMODCACHE=/gomodcache`、`GOCACHE=/project/workplace/gocache`、`GOTOOLCHAIN=local`。原始 dosgolem 快照另唯讀掛載，Go workspace 由 prepare 建立。

### 本輪收尾

- 原始 ZIP、RAR 與 29 個解包檔案的 SHA-256 全部重查通過，來源沒有變動。
- 文件連結、Python 語法與前端二進位對 GUI 收據的身份核對通過；`provenance.json` 收錄 13 個來源檔及工具身份。
- 容器內執行 `find /project -user root -o -type d -name '*.md'`，沒有異常輸出。
- `zeliard-*` 的執行中與已停止容器均已清空。沒有建置新映像或清理其他專案映像。
- 本輪 `/tmp/zeliard-manual-preview-20261018` 已刪除。原始 RAR、研究快照與收據保留本機，均不作可公開素材。
- 本輪沒有 commit、push 或建立 GitHub 儲存庫；主機存取政策阻塞仍存在。完整中文、HD 美術、正常新局／城鎮／戰鬥、音效輸出與實體手把留在 WORKLIST，未標完成。

## 2026-10-18：城鎮切片與片頭文字鏈

- 使用者授權使用 host gh。主機提升權限的 `/home/anr2/.local/bin/gh auth status` 仍被 granular 執行政策拒絕；一般受限呼叫回報登入失敗，不能當作主機憑證失效的結論。未登出、搬運憑證或建立遠端。授權寫入 AGENTS 與 CONTEXT，不再重問。
- 載入正常玩家路徑、文件職責、RRSAL 契約、逆向證據 reference，以及 `use-ida-pro-9-4` 與指定工具文件。沒有為本輪另建 Docker 映像。
- 原版進度：冷開機、片頭跳過、城鎮、右移的 66m 指令序列可重播。收據、固定時鐘、按鍵及 SHA-256 見 RESEARCH-LOG。
- 前端進度：`tools/gui_smoke.py --gameplay` 以 X11 視窗按鍵進城鎮，辨識狀態列後右移並放開。`gui-town-v2` 通過，城堡錨點比較與空持鍵集合支持放開後停止。Z03 完成範圍限此切片，新增 Z09 承接正常對話與存讀檔。
- 驗證腳本修正：第一輪牆鐘腳本在片頭轉場送鍵，沒有驗到城鎮。改為畫面辨識，不改遊戲流程。前端新增可選 `-capture-request`，只控制 Draw 擷取時機；metadata 增加累計 cycles 與持鍵集合。`gui-town-v1` 保留為未完成的探索，不計通過。
- 前端重建、`go vet` 與四項 Go 測試通過，未 skip。新二進位重跑 `gui-smoke-v5` 的原貌、手把頁及單句覆繪三條流程均通過；它們不取代實體手把驗收。
- 中文研究：ASCII 字面找到原始 SAR 偏移。三個錨點對上實際段傾印；IDA 加動態暫存器證據確認首、次行進入 `sub_32C9`。正式畫面定位、控制碼與特效仍待核對，沒有把研究線索直接接入正式翻譯路徑。
- 工具變更：`probe_receipt.py` 增加有界段傾印與暫存器觀察，檢查段傾印檔存在及大小。新增 `ida_intro_evidence.py` 匯出原始函式、位址、運算元與 bytes，已加入 AGENTS 索引；所有含原版記憶體與反組譯的輸出只留忽略的 `workplace/`。
- IDA 工具驗證：優先候選 `locked-v1` 本機不存在，沿用既有 `py312-v1`，UID/GID 1000:1000。最初 Binary loader 預設 32 位元，明確設成 16 位元後從唯讀輸入重建；下一輪發現 `tag_remove` 應由 `ida_lines` 提供，依 image 內 API 修正並重新建庫。v3、v4 JSON 的 schema、版本、位元數、輸入 SHA-256 均核對成功。這兩項屬工具腳本問題，不是遊戲缺陷。
- 研究目錄均放既有 `workplace/`：`probe-town-*`、`gui-town-*`、`probe-intro-*` 與 `ida-intro-*` 分別保存原版重播、GUI、段傾印及一次性資料庫。沒有新增同義研究根目錄或正式交付目錄。

### 本輪重跑入口

沿用前節映像、掛載與 Go 環境；新增的容器內命令如下，輸出目錄每次使用新的名稱。

| 項目 | 命令與條件 |
|---|---|
| 右移原版序列 | `python3 tools/probe_receipt.py --probe /project/workplace/dosgolem-probe --steps 66000000 --cycles 3000 --hold 'Space@12000000+50000,Space@20000000+50000,Space@28000000+50000,Space@36000000+50000,Space@44000000+50000,Space@52000000+50000,Right@60000000+3000000' --output /project/workplace/<新目錄>` |
| 正常前端切片 | `python3 tools/gui_smoke.py --gameplay --output /project/workplace/<新目錄>`，使用 psychicwar GUI 映像，原版與字型唯讀 |
| 片頭段傾印 | 探針收據腳本加 `--steps 10729940 --cycles 3000 --segment 01B1:0000:65536` |
| 片頭暫存器觀察 | 探針收據腳本加 `--steps 10729940 --cycles 3000 --registers-at 01B1:6366,01B1:32C9,01B1:32D8` |
| IDA 匯出 | `ZELIARD_IDA_OUTPUT=<輸出 JSON>`；`idat -A -pmetapc -TBinary -b0 -o<新輸出 DB> -L<log> -S/project/tools/ida_intro_evidence.py /input/memory-01B1_0000.bin`，`/input` 唯讀 |

本輪沒有 commit／push、Release 或正式打包。GitHub 與本地 Git 仍受環境限制。完整中文、HD、音效、對話、存讀檔、戰鬥及實體手把保持未完成或未驗。

### 本輪收尾核對

- 原始 ZIP、RAR 與 29 個解包檔案的 SHA-256 重查通過。37 個本地文件連結、Python 語法、GUI 二進位身份與 IDA 輸入身份核對通過，收據在 `workplace/current-audit.json`。
- 本輪 IDA image 實際 ID：`sha256:4ac62de83339c215bab10e455cee3a22d9c6efed9fd0d8ed0f068327b83a06ab`。未讀取、匯出或搬運授權內容。
- 容器內執行 `find /project -user root -o -type d -name '*.md'` 無異常；抽查 `.i64`、GUI 收據與二進位均為 1000:1000。
- `docker ps -a --filter name=zeliard-` 為空；一次性前端及 Xvfb 都已回收。未建立新 image，未碰其他專案容器或 image。
- `git status` 仍回報非 Git 儲存庫。沒有覆寫環境的唯讀 `.git/`，沒有 commit 或 push。

## 2026-10-18：主機存取恢復與私人儲存庫

- 執行環境改為可存取主機及網路後，主機 gh 登入成功。先前受限環境的失敗不代表憑證損壞，沒有重新登入或搬運憑證。
- 建立 `https://github.com/wicanr2/zeliard_cht`，回讀確認 `visibility=PRIVATE`。原本的 `.git` placeholder 已不存在，初始化 `main` 並設定 `origin`。
- Git 提交身份沿用 `Chun-Yu Wang`，此儲存庫信箱設為 GitHub 帳號 ID 對應的 `7655317+wicanr2@users.noreply.github.com`，不改全域設定。
- 從完整 RRSAL-1.0 範本建立 LICENSE，只替換占位符。README 增加授權摘要與條款連結；遊戲、字型、手冊、記憶體傾印與收據不加入 Git。
- 首次提交前審查實際 Git 清單、忽略規則與原始檔案雜湊，結果與提交紀錄於本節追加。
