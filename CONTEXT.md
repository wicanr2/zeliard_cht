# 目前狀態

更新日期：2026-10-18，Asia/Taipei。

## 目前狀態表

| 項目 | 現況 | 證據與限制 |
|---|---|---|
| 使用者目標 | dosgolem、繁中、HD、工具列手把綁定 | 已授權開工；規則見 [AGENTS.md](AGENTS.md) |
| GitHub | 已建立並推送 [wicanr2/zeliard_cht](https://github.com/wicanr2/zeliard_cht) | 主機 gh 登入成功，回讀 `visibility=private`；首章中文功能提交 `390c199` 已推送，HD 既有成果保留 |
| 本地 Git | 已初始化 `main`，設定 `origin` | 環境 placeholder 已不存在；使用 GitHub noreply 信箱作本儲存庫提交身份 |
| 遊戲包 | 已安全解包與逐檔盤點 | 30 個壓縮項目，含目錄；29 個檔案，清單在本地 `workplace/source-manifest.json` |
| 中文手冊 | RAR 已列目錄，17 張 JPG | 已人工讀第 1、5–8 頁；另含封面、地圖等掃描，17 張不等於 17 頁 |
| dosgolem | 正常新局、國王初訪兩段與回城鎮已驗 | 原版負責對話、翻頁與 1000 賞金；兩段原文均走到結尾。存讀檔與戰鬥未驗 |
| 架構 | 使用者已選定直接執行原版再覆繪 | 參考 psychic-war，不另建玩法重製引擎；見 [架構規格](docs/spec/001-runtime-overlay.md) |
| 前端 | `workplace/bin/zeliard-preview` 已建置，GUI 雛形可啟動 | 中文工具列、手把綁定設定；音效輸出尚未接入 |
| 中文 | 工具列、首章 30 行與國王初訪兩段 prototype | 國王 241 個原文字形事件、兩段譯文與正常結束已驗；正式字級、術語及其餘文字待完成 |
| HD | 手繪奇幻邊框與國王兩姿態 prototype | 肖像內部 52,452 像素改變，保留黃色外框；同狀態切換區外 0、完整還原。其餘場景／角色及正式倍率待完成 |
| 手把 | 裝置選擇、重綁定、保存、死區及斷線釋放雛形 | 純輸入測試通過；尚無實體 USB／藍牙驗收 |
| 文件 | AGENTS、README 與 LICENSE 已建立 | 故事背景已核對中文手冊；自有內容採 RRSAL-1.0，原版資料不入庫 |

## 原始輸入

| 檔案 | bytes | SHA-256 |
|---|---:|---|
| `神劍封魔.jsdos.zip` | 2092797 | `f7fc8bd3ea02e56ef01f170b9a00faa921634dc96a6f9bf6a6a736514e0bb6ac` |
| `珍084-神劍封魔.rar` | 8557678 | `cb251314e065f2fef3e4eb007fa332a267a501aa155ebe35722245214b80e472` |

ZIP 包含 `ZELIARD.EXE`、`GAME.BIN`、`ZELRES1.SAR` 至 `ZELRES3.SAR`、多個顯示／音效模組、`STICK.BIN`、`RESOURCE.CFG`、使用者存檔及 `Zeliard_Manual.pdf`。檔名只作盤點，不代表用途已完全解明。

下載來源：`https://b2.zczc.cz/file/dos-bin/%E7%A5%9E%E5%89%91%E5%B0%81%E9%AD%94.jsdos.zip`。

## 有效決策與待確認

- 已確認：參考 `softworld_san` 工具列、參考 `psychic-war` 覆繪、以 dosgolem 直接執行原版、繁中＋HD、藍牙／USB 手把綁定、私人 GitHub、README 手冊故事背景。HD 畫風已選定比較圖 B「手繪奇幻」。
- 待確認：HD 正式倍率／預設／構圖、正式前端版面與手把預設動作配置、正式翻譯 catalog 規格、平台與發行方式。現有 JSON 只供 prototype 使用；正式翻譯來源依本地化契約採 UTF-8 TSV，鍵與版面尚未審查。
- 可繼續：來源盤點、手冊核對、原版啟動探針、能力缺口定位、可撤回顯示與輸入 prototype。
- 不沿用參考專案的 MSC 編譯器結論、Big5 編碼、固定畫布、AI 強化或 Esc 工具列快捷鍵。

## 下一步

依 [WORKLIST.md](WORKLIST.md) 推進城鎮商店／存讀檔的正常路徑、其餘片頭與對話來源，再接 HD 場景／角色資源。首章與國王初訪掛鉤已有最小充分證據，不追無關 helper。正式字級、術語及素材仍待校訂。證據入口為 [RESEARCH-LOG.md](RESEARCH-LOG.md)，驗證狀態以 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md) 的實際結果為準。

## 工具與本地證據入口

- dosgolem 來源 HEAD：`3e17c1eb3f3d58935b1e8a59b978b8c899f4a653`，來源工作樹有其他工作的未提交修改。
- 本專案固定快照：`workplace/dosgolem-source/`，882 個檔案；逐檔清單雜湊 `2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`。HEAD 不能單獨識別這份工具。
- 探針：`workplace/probe-startup-mcga/`、`probe-startup-replay/`、`probe-space-40m/`；重播比較為 `workplace/startup-replay-comparison.json`。
- 即時前端片頭基線：`workplace/probe-intro-clock3000/`。3000 cycles／毫秒、10,729,940 指令、無輸入；與診斷 GUI 原版畫格逐點一致。
- 單句與手把頁的既有 GUI 收據：`workplace/gui-smoke-v6/receipt.json`。新的首章捲動與國王正常操作見下方最新入口；未將舊 GUI 的二進位身份冒充本輪建置。
- 正常新局與右移：`workplace/gui-town-v2/receipt.json`，以 X11 實際視窗按鍵操作，未讀存檔或注入狀態。原版 66m 右移重播為 `workplace/town-replay-comparison.json`。
- 放開驗證：`workplace/gui-town-v2/movement-region-comparison.json`，城堡靜態錨點在右移前後不同、放開後一致；原版持鍵集合為空。不把此比較稱為整張畫面的同狀態對拍。
- 英文字面盤點：`workplace/literal-text-search.json`；片頭與商店選項分別在 `ZELRES1.SAR`、`ZELRES2.SAR`，僅證明字面位置，不證明控制碼契約。
- 片頭文字鏈：`workplace/intro-memory-anchors.json`、`ida-intro-v4/intro-evidence.json` 與 `probe-intro-consumer/`。首章的來源與逐列顯示已接入唯讀 prototype；其他章節及場景的特效契約未解。
- 本地初始 Git 審查：`workplace/git-initial-audit.json`；24 個自有程式／文件，完整授權範本保留，29 個原版檔案不變，無原版素材入庫。
- HD 樣圖：`workplace/hd-style-preview.png`；使用者只選定 B 的畫風，未定稿構圖。新生成邊框、完整提示詞與來源為 `workplace/hd-border-prototype.png`、`workplace/hd-border-generation.json`，均只留本機研究區。
- HD 實際切換：`workplace/gui-hd-frame-v1/receipt.json`、`hd-region-comparison.json`。兩側 172,017 像素改變，中間區域 0；切回原貌後完整畫格相同。它只驗證裝飾邊框，不代表全遊戲 HD。
- 位址索引：[docs/source-index.json](docs/source-index.json)，IDA 匯出自動合併原始定位、證據等級與警示；首章增補匯出在 `workplace/ida-intro-scroll-indexed/`。動態定位收據為 `workplace/probe-intro-display/`。
- 前輪 HD 與首章稽核：`workplace/current-audit-hd.json`、`current-audit-scroll.json`，保留各自來源與二進位身份。本輪為 `workplace/current-audit-dialogue-hd.json`。
- 同狀態區域比較：`workplace/gui-smoke-v4/intro-region-comparison.json`。原版色號完全相同，覆繪區內 3,655 個像素不同，區外 0。
- 初始化工具身份：`workplace/gui-smoke-v4/provenance.json`，記錄當時 13 個來源檔、二進位、dosgolem 快照及映像 ID。該份舊收據沒有 Git 身份，現行程式以本儲存庫 HEAD 為準。
- 最新前端二進位 SHA-256：`507b368ee2d41516bc02b87012b9bb4607229dd7687e11f2a6bc10cfa699427d`。建置基於 `9c4590d` 加本輪修改來源，未稱正式版本。
- 首章中文最新重跑：`workplace/intro-scroll-v3/receipt.json`。30 個來源行、430 個完成列事件；10,729,940／14m／30m／65m／110m／160m 指令抽樣，完整 1 MiB 記憶體、畫面色號、色盤、CS:IP 與虛擬時間均與無掛鉤原版相同。覆繪區外差異 0；結尾捲出與換章移除已驗。
- 首章實際 X 視窗：`workplace/gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`。自然片頭擷取時已觀察 7 個來源行、62 個完成列。專名暫留原文，正式術語與字級尚未定稿。
- 早期國王正常操作：`workplace/probe-door-up-70m/`、`probe-king-0-84m/` 及 `gui-king-v2/receipt.json`。探針使用既有六次 Space，再於 60m 按 Right 2m、64m 按 Up 250k。GUI 以畫面核對進入國王房間，首段文字與原版收據相符；空白鍵後 GOLD 更新至 1000。不把賞金確認稱為對話全部翻頁完成。
- 國王對話目前證據：`workplace/dialogue-v3/receipt.json`，130m 正常新局、領賞與返回城鎮。九個抽樣點的完整 1 MiB 記憶體、色號、色盤、CS:IP 及 cycles 均與無覆繪原版相同；241 字形、兩段話、50 個捲動列，覆繪區外 0。
- 國王實際 X 視窗：`workplace/gui-dialogue-hd-v5/receipt.json`。兩段中文全部顯示、按原版待續標記翻頁、1000 賞金及返回城鎮通過；最後持鍵為空、無覆繪失配。國王與回城鎮的原貌→HD→原貌分別驗完整還原。
- 國王 HD：`workplace/hd-king-prototype.png`、`hd-king-speaking-prototype.png`，完整 imagegen 提示詞組與來源在 `hd-king-generation.json`。只留本機研究區；兩個已觀察姿態，不稱為所有動畫完成。
- 對話位址索引：[docs/dialogue-source-index.json](docs/dialogue-source-index.json)，固定 84m 國王段傾印 SHA-256 `06b74a1ea26f25d7e81de50ecb9f9ece1e1851d9429b469c59cf6a80bbb1d156`。最新 IDA 匯出 `workplace/ida-dialogue-indexed-v3/`，不與片頭段混用。
- 本輪八項具名 Go 測試與 go vet 通過，無測試 skip；`bitmap` 套件沒有獨立測試。測試紀錄為 `workplace/test-dialogue-hd-v3.jsonl`，現行稽核為 `workplace/current-audit-dialogue-hd.json`。
- 完整中文、其餘場景／角色 HD、音效輸出、存讀檔／戰鬥與實體手把仍未完成或未驗。
