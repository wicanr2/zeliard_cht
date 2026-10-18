# 目前狀態

更新日期：2026-10-18，Asia/Taipei。

## 目前狀態表

| 項目 | 現況 | 證據與限制 |
|---|---|---|
| 使用者目標 | dosgolem、繁中、HD、工具列手把綁定 | 已授權開工；規則見 [AGENTS.md](AGENTS.md) |
| GitHub | 已建立 [wicanr2/zeliard_cht](https://github.com/wicanr2/zeliard_cht) | 主機 gh 登入成功，建立後回讀 `visibility=PRIVATE`；首次提交審查中 |
| 本地 Git | 已初始化 `main`，設定 `origin` | 環境 placeholder 已不存在；使用 GitHub noreply 信箱作本儲存庫提交身份 |
| 遊戲包 | 已安全解包與逐檔盤點 | 30 個壓縮項目，含目錄；29 個檔案，清單在本地 `workplace/source-manifest.json` |
| 中文手冊 | RAR 已列目錄，17 張 JPG | 已人工讀第 1、5–8 頁；另含封面、地圖等掃描，17 張不等於 17 頁 |
| dosgolem | 已從冷開機進入城鎮並右移 | 固定 6600 萬指令重播的色號、色盤、PNG 與末端軌跡一致；對話、存讀檔與戰鬥未驗 |
| 架構 | 使用者已選定直接執行原版再覆繪 | 參考 psychic-war，不另建玩法重製引擎；見 [架構規格](docs/spec/001-runtime-overlay.md) |
| 前端 | `workplace/bin/zeliard-preview` 已建置，GUI 雛形可啟動 | 中文工具列、手把綁定設定；音效輸出尚未接入 |
| 中文 | 工具列與片頭單句 prototype | 片頭原文字面位於 `ZELRES1.SAR` 的檔案偏移 `0x109F`；正式繪字與控制碼尚待定位，單句覆繪區外像素差異 0 |
| HD | 尚未製作或接入美術 | 保留原版 → HD → 中文的合成契約；畫風／倍率待確認 |
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

- 已確認：參考 `softworld_san` 工具列、參考 `psychic-war` 覆繪、以 dosgolem 直接執行原版、繁中＋HD、藍牙／USB 手把綁定、私人 GitHub、README 手冊故事背景。
- 待確認：HD 樣式／倍率／預設、正式前端版面與手把預設動作配置、正式翻譯 catalog 格式、平台與發行方式。現有 JSON 只供 prototype 使用。
- 可繼續：來源盤點、手冊核對、原版啟動探針、能力缺口定位、可撤回顯示與輸入 prototype。
- 不沿用參考專案的 MSC 編譯器結論、Big5 編碼、固定畫布、AI 強化或 Esc 工具列快捷鍵。

## 下一步

依 [WORKLIST.md](WORKLIST.md) 追查片頭單行緩衝的畫面定位與控制碼，再推進正常對話、存讀檔及 HD 資源識別。首次提交只包含自有程式、文件及譯文。證據入口為 [RESEARCH-LOG.md](RESEARCH-LOG.md)，驗證狀態只以 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md) 的實際結果為準。

## 工具與本地證據入口

- dosgolem 來源 HEAD：`3e17c1eb3f3d58935b1e8a59b978b8c899f4a653`，來源工作樹有其他工作的未提交修改。
- 本專案固定快照：`workplace/dosgolem-source/`，882 個檔案；逐檔清單雜湊 `2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`。HEAD 不能單獨識別這份工具。
- 探針：`workplace/probe-startup-mcga/`、`probe-startup-replay/`、`probe-space-40m/`；重播比較為 `workplace/startup-replay-comparison.json`。
- 即時前端片頭基線：`workplace/probe-intro-clock3000/`。3000 cycles／毫秒、10,729,940 指令、無輸入；與診斷 GUI 原版畫格逐點一致。
- GUI 最新收據：`workplace/gui-smoke-v5/receipt.json`；原貌、滑鼠開啟手把頁、片頭單句覆繪三條流程通過。實際視窗為 `original-window.png`、`controller-window.png`、`intro-zh-window.png`。
- 正常新局與右移：`workplace/gui-town-v2/receipt.json`，以 X11 實際視窗按鍵操作，未讀存檔或注入狀態。原版 66m 右移重播為 `workplace/town-replay-comparison.json`。
- 放開驗證：`workplace/gui-town-v2/movement-region-comparison.json`，城堡靜態錨點在右移前後不同、放開後一致；原版持鍵集合為空。不把此比較稱為整張畫面的同狀態對拍。
- 英文字面盤點：`workplace/literal-text-search.json`；片頭與商店選項分別在 `ZELRES1.SAR`、`ZELRES2.SAR`，僅證明字面位置，不證明控制碼契約。
- 片頭文字鏈：`workplace/intro-memory-anchors.json`、`ida-intro-v4/intro-evidence.json` 與 `probe-intro-consumer/`。IDA `sub_6358` 的間接呼叫實際進入 `sub_32C9`，讀取首行及第二行；完整畫面定位與特效契約仍未解。
- 本輪稽核：`workplace/current-audit.json`，原始包、29 個來源檔、37 個本地文件連結、Python 語法、GUI 二進位身份及 IDA 輸入身份通過；沒有 root-owned 異常或殘留專案容器。
- 同狀態區域比較：`workplace/gui-smoke-v4/intro-region-comparison.json`。原版色號完全相同，覆繪區內 3,655 個像素不同，區外 0。
- 來源與工具身份補充：`workplace/gui-smoke-v4/provenance.json`，記錄 13 個前端／文字／工具來源檔雜湊、二進位雜湊、dosgolem 快照及映像 ID。尚無本地 Git commit。
- 最新前端二進位 SHA-256：`8c0b24f4d2449f8cfcee61acfebf40e5753b0f71a0573d3ea11cc31c09c38adf`。新加擷取訊號檔只供 GUI 驗證，未改遊戲輸入或規則。
- 完整中文、HD 美術、音效輸出、原版對話／存讀檔／戰鬥與實體手把均未完成或未驗，不以城鎮操作替代。
