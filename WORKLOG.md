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

- 初始審查通過：24 個程式／文件檔案、完整 RRSAL 範本、29 個原版檔案不變。初始提交 `93b4dbd` 已推送 `origin/main`，沒有原版檔案或本地收據入庫。

## 2026-10-18：手繪 HD 雙側邊框試作

- 使用者看過畫風比較圖後選定 B「手繪奇幻」。此決定只涵蓋畫風，正式構圖、角色比例、倍率與預設主題仍未定稿。
- 使用內建 imagegen 產生比較圖與透明手繪石雕雙欄圖。完整原圖、提示詞、來源及雜湊存本地 `workplace/`，不把未定稿試作圖當作正式發行素材或加入 Git。
- 新增 `internal/prototype/hd/`，核對固定原版基線與圖像身份，僅覆蓋原生左右各 48×200 的裝飾區。圖面直接合成在顯示解析度；左右任一失配、色盤變化、缺圖或損壞都保留原貌。
- 前端增加可選 HD 輸入與工具列切換。起始原貌、HD 試作與退回狀態均有明示標籤。所有功能留在 preview／prototype，沒有修改原版玩法、記憶體或存檔格式。
- 建置、`go vet` 與五項 Go 測試通過，沒有 skip。真實城鎮與生成圖的測試包含邊界、唯讀輸入、失配、色盤變化、不完整畫格及缺圖。
- `gui-hd-frame-v1` 走正常冷開機與城鎮操作，再經工具列暫停，同一畫格下切原貌→HD→原貌。兩側 172,017 像素改變，中間區外 0，還原畫格完全相同；實際視窗已目視核對。
- 新二進位的原貌、手把頁與片頭單句在 `gui-smoke-v6` 重跑通過。生成圖、本地原版與字型保持唯讀輸入；沒有以合成手把事件宣稱 USB／藍牙驗收。
- 中文研究：IDA 與 dosgolem 暫存器將 `sub_332C` 的片頭寫入區定位為 `(0,32,320,120)`，含背景及遮罩，沒有直接整塊覆蓋。細節見 RESEARCH-LOG，完整中文尚未完成。
- 新增受版控的 `docs/source-index.json`，匯出工具自動合併原始位址、語意、等級與證據，未證實項帶警示。最新 JSON 身份與索引雜湊於最終稽核核對。
- IDA 的第一輪索引讀取碰到 image 預設 ASCII；改成明確 UTF-8，從相同唯讀輸入重建，匯出成功。這是工具編碼問題，不是原版缺陷。沒有追查其他無關 helper 或硬體時序。
- 沿用 Go／GUI／IDA 映像，所有工作負載在 Docker、UID/GID 1000:1000 下執行；新增 Go 子套件承載顯示試作，沒有新增研究根目錄或正式交付目錄。

### HD 重跑入口

沿用前節唯讀原版／字型、可寫專案及 Go 模組快取掛載。額外的測試環境為 `ZELIARD_HD_REFERENCE=/project/workplace/probe-newgame-60m/screen.bin`、`ZELIARD_HD_ART=/project/workplace/hd-border-prototype.png`。缺少本地輸入時 HD 測試 skip，不算驗收。

GUI 容器內執行 `python3 tools/gui_smoke.py --gameplay --hd-reference /project/workplace/probe-newgame-60m/screen.bin --hd-art /project/workplace/hd-border-prototype.png --output /project/workplace/<新目錄>`。輸出原貌、HD、還原的實際視窗、原始色號、metadata 及 `hd-region-comparison.json`。

正式場景／角色 HD、完整中文與換頁、對話／存讀檔／戰鬥、音效輸出及實體手把仍待後續。沒有 Release、正式打包或把儲存庫改為公開。

### 本輪最終稽核

- `workplace/current-audit-hd.json`：27 個自有來源檔，沒有原版、圖片、手冊、記憶體或本地收據入 Git。原始 ZIP／RAR、29 個解包檔案與 882 個 dosgolem 快照檔案雜湊全部不變。
- 42 個本地文件連結、Python 語法、JSON、IDA 索引與輸入身份核對通過；正式範本 LICENSE 保留，實際忽略規則涵蓋原始包、試作圖、收據與 go.work。
- 最新二進位為 `1d4b472fb36fda44a59452040830205150add0890743613617a8a9f0f763ea0f`，Go metadata 記錄基底 `93b4dbd` 加未提交的本輪功能。收據另記來源檔指紋，不把此建置稱為正式版本或乾淨 tag 產物。
- `find /project -user root -o -type d -name '*.md'` 沒有異常。依工作樹掛載檢查 Docker，執行中及已停止的相關容器均為空；沒有新映像或需要移除的被取代映像。
- HD 功能提交 `7401bab8517e1c3dbbfeb045c7abc233fce45bd0` 已推送 `origin/main`。主機 gh API 回讀遠端 `main` 的 SHA 與本地一致，工作樹乾淨。此後只整理交接文件，沒有修改已驗證的程式或素材。

## 2026-10-18：首章 30 行中文與國王正常操作

- 路由命中復古 remake 技能、規格閘門、本地化顯示隔離、正常玩家路徑及文件職責。改首頁摘要前另讀 README 標準；索引重匯出沿用 IDA 9.4 技能與既有 py312-v1，不新增 image。
- 新增首章捲動覆繪 prototype 及 UTF-8 TSV 試譯。來源限定本地 SAR 版本，原文不入 Git；唯讀觀察來源與逐列完成事件。正式版面、字級及專名尚待校訂，其餘章節保留原文。
- 六項 Go 測試、go vet 及建置通過，沒有 skip。首章測試自然執行到 160m 指令，六個抽樣點的完整 1 MiB 記憶體、色號、色盤、CS:IP 與虛擬 cycles 均與無掛鉤原版相同。遮罩只改顯示，區外差異 0。
- 修正首輪在 FF 結束碼過早移除中文的問題，保留舊收據。v2 確認結尾中文繼續捲出、換章移除。沒有改原版控制碼或時序。
- 原版探針正常新局進城堡，等待國王首段文字。GUI 腳本新增 `--king`，以實際地名及對話文字核對進門結果；Space 後金額更新至 1000。首輪固定時間仍在門左側，分類為腳本時機問題，改用小步移動後乾淨重跑通過。
- `gui-intro-scroll-v2`、`gui-king-v2` 分別驗首章多行中文及國王正常操作，均使用新二進位，原版唯讀、scratch 獨立。未把國王賞金確認稱為全對話結束，也未宣稱國王已中文化。
- 新 Go 檔與 TSV 索引掛入 AGENTS；研究與收據沿用既有 `workplace/`。無新交付根目錄、Release 或正式封包。完整中文、場景／角色 HD、音效與實體手把仍待後續。

### 重跑入口

沿用前節 Docker 掛載及資源限制。首章測試環境增加 `ZELIARD_ORIG=/orig`、`ZELIARD_SCROLL_OUTPUT=/project/workplace/<新目錄>`，原版及字型唯讀；執行 `go test -p 2 ./internal/prototype/... -count=1 -v -timeout 95s`，其餘真實覆繪與 HD 環境沿用前節。缺少資料的 skip 不算通過。

GUI 分別執行 `python3 tools/gui_smoke.py --intro-scroll --output /project/workplace/<新目錄>` 與 `python3 tools/gui_smoke.py --gameplay --king --output /project/workplace/<新目錄>`。腳本有界持有並回收 Xvfb、前端與子程序；只用 X11 正常按鍵操作。

現行來源、輸入、二進位、IDA 索引與收據身份記在 `workplace/current-audit-scroll.json`。提交及 Docker 清理結果於收尾追加。

### 本輪收尾

- 首章中文功能提交 `390c199a08c78c12023f7256565cf97fa75fc7f4` 已推送 `origin/main`，主機 gh 回讀儲存庫仍為 private。提交只含自有程式、文件與試譯，沒有原版資料或本地收據。
- 現行稽核核對 30 個自有來源檔、29 個原版檔案、882 個執行器快照檔案及 46 個文件連結。IDA 9.4 匯出與現行位址索引、輸入 SHA-256 相符；六項測試與兩條新 GUI 流程通過，未 skip。
- 最新二進位 SHA-256 為 `b99597ac7e7e7500cc34f56f8a7d25280174628ef182a2b50578d151b76d0a0b`。建置 metadata 為 `2959b8a` 加修改來源；實際 Go 來源在本輪功能提交中固定，不冒稱正式版。
- `find /project -user root -o -type d -name '*.md'` 無異常；依工作樹掛載檢查 Docker，沒有相關執行中或停止容器。沒有新 image，也未清理其他專案資源。

## 2026-10-18：國王初訪中文與手繪肖像

- 上輪判定為進展：首章中文與正常進門已有來源及提交。本輪沿用目前工作樹，未重新開啟已完成的啟動／首章研究。
- 路由命中復古逆向、本地化顯示隔離、正常玩家路徑、IDA 9.4、GUI 還原及文件職責。依各入口只讀必要契約；更新首頁摘要前另讀 README 標準。未使用平行代理，未修改共用 dosgolem。
- 正常初訪段傾印與原始 SAR 錨點定位兩段文字。新增 UTF-8 TSV、唯讀逐字／捲動／標記覆繪、DRAFT 規格、對話位址索引及 IDA 匯出工具。新增檔案同輪掛入 AGENTS，不新增研究根目錄。
- 實際視窗暴露 CH 帶回其他計數的情況。先記錄來源與暫存器，再查 IDA 的 `xor ch,ch`，把縱座標判定修正為 CL。v1／v3 的 240 字形與三次段落啟用未計通過；v5 為 241 字形、兩段、無失配，正常返回城鎮。
- v2 未辨認到城鎮時停在黑畫面，沒有足夠進度證據，不列為產品缺陷或通過。新增可選 `-observe` 記錄執行進度、焦點與失配；GUI 軟體繪圖執行緒限制為 2，保持資源有界。沒有改原版速度參數來通過測試。
- 使用內建 imagegen 製作國王閉口及說話圖。完整提示詞、來源、雜湊與本機試作分類存 `workplace/`；圖面由工具預設位置唯讀複製，不搬運憑證、不改原版圖檔。兩個姿態完整核對後接入，保留黃色外框；未知姿態、淡入及缺圖退回原貌。
- 八項具名 Go 測試、go vet 與建置通過。無測試 skip，`bitmap` 套件沒有獨立測試。紀錄為 `workplace/test-dialogue-hd-v3.jsonl`；首章重跑 `intro-scroll-v3`，國王完整流程 `dialogue-v3`，均與無覆繪原版記憶體及時序相同。
- 正常 X 視窗 `gui-dialogue-hd-v5` 通過兩段中文、1000 賞金、回城鎮與國王／城鎮 HD 還原。v4 最後的城鎮比較是腳本沿用 HD 開關的錯誤假設；明確切原貌後以同一容器映像重跑，未改產品來迎合比較。
- 中文手冊第 3–4、7–8 頁再次對照原圖。手冊仍作操作與術語來源，本輪沒有補造角色官方譯名。提取掃描只留 `workplace/manual-terms-v1/`，不入 Git。
- 自有程式、文件及譯文可提交；原版、掃描、生成圖、記憶體與收據仍留本地。沒有發行版、Release、正式封包或公開方式變更。

### 重跑入口

沿用前節 Docker 映像、UID/GID、唯讀原版／字型／快照與資源上限。完整具名測試環境增加 `ZELIARD_DIALOGUE_OUTPUT`，以及 `ZELIARD_KING_A`、`ZELIARD_KING_B`、`ZELIARD_KING_CLOSED`、`ZELIARD_KING_OPEN`，分別指向本地兩個原版姿態與兩張手繪 PNG。執行 `go test -p 2 ./internal/prototype/... -count=1 -json -timeout 165s`，缺本地資料的測試 skip 不算通過。

GUI 容器內執行 `python3 tools/gui_smoke.py --gameplay --king --dialogue --king-hd --hd-reference /project/workplace/probe-newgame-60m/screen.bin --hd-art /project/workplace/hd-border-prototype.png --output /project/workplace/<新目錄>`。Xvfb／前端由腳本有界持有，只有實際 X 按鍵及工具列操作；啟動前後核對二進位身份，避免測試中途被替換。

本輪現行稽核為 `workplace/current-audit-dialogue-hd.json`。提交、遠端 SHA 與 Docker 清理結果於收尾追加。

### 本輪收尾

- 國王中文與手繪肖像功能提交 `18728aa` 已推送 `origin/main`，主機 gh 回讀儲存庫仍為 private。沒有原版檔案、掃描、生成圖或本地收據入 Git。
- 現行稽核核對 37 個自有來源檔、29 個原版檔案、882 個 dosgolem 快照檔案及 55 個文件連結。原始 ZIP／RAR、LICENSE、IDA 輸入與索引身份均不變或符合既有記錄。
- 最新二進位為 `507b368ee2d41516bc02b87012b9bb4607229dd7687e11f2a6bc10cfa699427d`；metadata 記錄基底 `9c4590d` 加修改來源。正式 Go 來源在功能提交中固定，未宣稱乾淨 tag 或發行版。
- 八項具名 Go 測試與完整國王 X 視窗流程通過。容器依工作樹掛載檢查為空，`find /project -user root -o -type d -name '*.md'` 無異常；沿用既有 image，沒有新增或全域清理映像。

## 2026-10-18：長老中文、正常存檔與冷重開讀取

- 路由載入復古逆向、本地化顯示隔離、正常玩家路徑、IDA 9.4、文件職責與 README 標準。正常試玩入口明文要求測試代理，本輪依該入口分派唯一試玩代理；主流程處理來源、覆繪與配對驗證。沒有使用額外代理。
- 讀手冊第 10、13 頁定位長老存檔入口，正常探索城鎮與來源映射。研究與試玩輸出均在 workplace，不新增研究根目錄、不修改共用 dosgolem。
- 新增模組／來源字串雙重核對的 ASCIIZ 選單覆繪、18 筆試譯及長老四列正常路徑測試。對話支援長老初訪與存檔來源，修正中文行首句尾標點。新來源與測試同輪掛入 AGENTS；原始文本留在本地遊戲資料。
- 十項具名測試通過：`test-sage-full-v1.jsonl` 九項、`test-save-overlay-v1.jsonl` 一項；無具名測試 skip。建置及 go vet 通過。首次前端接入用了已存在的 labels 欄位名稱，編譯檢查發現後改為 menuLabels，再乾淨建置。這是實作命名錯誤，沒有改原版規則。
- 新存檔配對十九點完整原版狀態相同，兩側各自建立相同新檔，再退出與冷讀回。實際 X 視窗另完成 XVIEW1 保存、退出碼 0、冷重開 F7 讀回位置／金額；名稱輸入框與 Yes/No 保留英文，既有 UI 切換會安全丟棄舊中文，診斷原樣保存。
- 新二進位 SHA-256 `a899b4b8fa5887be801e20147d4ac6668649c4a3d7943d74400a51b8b5656390`，在 GUI 執行期間未替換。此次沒有增加 HD 素材或開啟新的 HD 比較，先前邊框／國王的收據仍保留原身份。
- 窄 IDA 匯出擴充共用 ASCIIZ 繪字函式與分級索引，從原固定國王段重建；長老來源另保存輸入雜湊與位址空間，未混淆兩份段傾印。
- 待完成：名稱輸入／讀檔 UI、Yes/No、其他商店與遊戲文字、其餘 HD 圖面、音效輸出、戰鬥及實體手把。最新稽核、GitHub 提交與容器清理結果在收尾追加。
- 收尾稽核通過：42 個自有來源檔、29 個原版檔案、882 個執行器快照檔案、65 個本地文件連結；原始包、LICENSE、IDA 索引與輸入身份相符。`workplace/current-audit-sage-save.json` 保存當輪身份，沒有原版素材加入 Git。
- 功能提交 `1f40ff9326485f5ab612b85c8eb9a05254e38b2a` 已推送私人儲存庫；主機 gh 回讀 main 相符，visibility 仍為 private。所有本專案探針／GUI／IDA 容器已回收，擁有權自檢無異常，沒有修改其他專案資源。

## 2026-10-18：存讀提示中文與速度輸入保留

- 路由沿用復古逆向、本地化顯示隔離、正常試玩、IDA 9.4 與文件職責。依正常試玩入口的明文要求，沿用唯一試玩代理驗 GUI；沒有新增代理。
- 選單覆繪追加 SAR1 與 STICK 固定來源、FF／CR 多行繪字觀察，試譯從 18 增至 25 筆。名稱提示、是／否、退出、F7 及速度提示可顯示中文，檔名與 Re-Start 保留原值；共用 dosgolem 未修改。
- 正常 GUI 首次未在 90 次步行觀察內走到長老，保留 `save-ui-gui-v1/failure.json`。這是當時 CPU 競爭及腳本預算問題；調整有界觀察後，`save-ui-gui-v2` 通過 UIVIEW2 保存、正常退出及冷讀回位置／1000 金幣。
- 短 GUI 發現 F9 輸入 3 後中文退回英文。根因是最長行寬涵蓋第二行的原版數字，修正為逐行文字前綴核對與清除。`save-ui-modal-v1` 保留失敗，沒有改原版鍵盤或速度規則。
- 修正後 `save-ui-v3` 配對 22 點完整原版狀態相同，文字區外 0；各側獨立建立相同 256-byte 存檔、退出及冷讀回。新增像素範圍檢查同時保證數字區保留。固定鍵序在 `tools/save-load.commands`。
- 最新短 GUI `save-ui-modal-v2` 通過 F7 完整兩行、N 取消、F9／3 中文及數字保留。Enter 後實際右移兩秒，角色位移且提示消失，放開後持鍵空。最後退出碼 0，原版資料不變。
- 十項具名測試通過：`test-ui-rest-v2.jsonl` 九項，加 `test-save-ui-v3.jsonl` 一項；無具名 skip。go vet、建置與 git diff --check 通過。最新二進位 SHA-256 `fb93dc06c9d384a0e9901a6129971e7622595e5e3421211ffe25048e2ebb335d`，GUI 測試期間未替換。
- 完整存讀 GUI 使用前一個 `a678bbbb` 建置，最新建置只重驗受改動的短 GUI 及完整配對，不把舊收據改名冒充新建置。所有圖像、原版傾印、生成素材、字型及收據維持本機 workplace，不加入 Git。
- IDA 9.4 在原固定 84m 國王段重建 FF 多行函式，輸出 `ida-dialogue-indexed-v5`，索引保留原始定位、bytes 與證據分級。新增本地證據入口同輪掛入 CONTEXT／研究紀錄。
- 文件追加腳本一度未傳入容器 stdin，命令正常結束卻沒有寫入。以 Git 差異核對發現後改用檔案編輯工具，回讀內容驗證。這是工作腳本問題，沒有重跑已通過的遊戲測試。
- 限制：彈窗下舊對話仍可能退回英文，其他片頭／商店／遊戲文字與其餘 HD 待完成，音效輸出、戰鬥與實體手把未驗。本輪沒有新增 HD 素材、正式版號或發行包。
