# 目前狀態

更新日期：2026-10-18，Asia/Taipei。

## 目前狀態表

| 項目 | 現況 | 證據與限制 |
|---|---|---|
| 使用者目標 | dosgolem、繁中、HD、工具列手把綁定 | 已授權開工；規則見 [AGENTS.md](AGENTS.md) |
| GitHub | 已建立並推送 [wicanr2/zeliard_cht](https://github.com/wicanr2/zeliard_cht) | 回讀 `visibility=private`；長老 HD 與武器店中文提交 `da235e6` 已推送，既有成果保留 |
| 本地 Git | 已初始化 `main`，設定 `origin` | 環境 placeholder 已不存在；使用 GitHub noreply 信箱作本儲存庫提交身份 |
| 遊戲包 | 已安全解包與逐檔盤點 | 30 個壓縮項目，含目錄；29 個檔案，清單在本地 `workplace/source-manifest.json` |
| 中文手冊 | RAR 已列目錄，17 張 JPG | 已人工讀第 1、5–8、10、13 頁；另含封面、地圖等掃描，17 張不等於 17 頁 |
| dosgolem | 正常新局、國王領賞、長老存檔與冷讀檔已驗 | 原版、中文覆繪配對及實際 X 視窗均通過；戰鬥未驗 |
| 架構 | 使用者已選定直接執行原版再覆繪 | 參考 psychic-war，不另建玩法重製引擎；見 [架構規格](docs/spec/001-runtime-overlay.md) |
| 前端 | `workplace/bin/zeliard-preview` 已建置，GUI 雛形可啟動 | 中文工具列、手把綁定設定；音效輸出尚未接入 |
| 中文 | 工具列、首章、國王、長老、存讀提示與武器店 prototype | 武器店五列、初訪／交易後／告別及兩個商品名已接入；動態報價、裝備頁、彈窗下舊對話與其餘文字待完成 |
| HD | 手繪奇幻邊框、國王及長老 prototype | 長老切換內部 50,926 像素改變，區外 0、完整還原，四個動畫像素不變；其餘場景／角色及正式倍率待完成 |
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

依 [WORKLIST.md](WORKLIST.md) 先接武器店動態報價與裝備頁中文，再推進其他商店、片頭／對話及 HD 場景／角色。已有正常交易及報價字形追蹤，不重新猜測商品繪字入口；名稱準備→暫存列→VGA 搬移鏈已證實。彈窗下舊對話仍需清框／還原證據。正常存讀檔已完成，不重新開啟輸入時序研究。正式字級、術語及素材待校訂。證據入口為 [RESEARCH-LOG.md](RESEARCH-LOG.md)，驗證狀態見 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md)。

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
- 前輪 HD 與首章稽核：`workplace/current-audit-hd.json`、`current-audit-scroll.json`，保留各自來源與二進位身份。先前提示輪為 `workplace/current-audit-ui.json`，目前為 `workplace/current-audit-weapon-sage.json`。
- 同狀態區域比較：`workplace/gui-smoke-v4/intro-region-comparison.json`。原版色號完全相同，覆繪區內 3,655 個像素不同，區外 0。
- 初始化工具身份：`workplace/gui-smoke-v4/provenance.json`，記錄當時 13 個來源檔、二進位、dosgolem 快照及映像 ID。該份舊收據沒有 Git 身份，現行程式以本儲存庫 HEAD 為準。
- 最新前端二進位 SHA-256：`5ad8556d7ebc674aa10f0bcf38b4c7743469cb7b44b2f89d17c002a2dffbf2ce`。建置基於 `c789462` 加本輪修改來源，未稱正式版本；主入口與 `zeliard-preview-weapon` 相同。
- 首章中文最新重跑：`workplace/intro-scroll-v3/receipt.json`。30 個來源行、430 個完成列事件；10,729,940／14m／30m／65m／110m／160m 指令抽樣，完整 1 MiB 記憶體、畫面色號、色盤、CS:IP 與虛擬時間均與無掛鉤原版相同。覆繪區外差異 0；結尾捲出與換章移除已驗。
- 首章實際 X 視窗：`workplace/gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`。自然片頭擷取時已觀察 7 個來源行、62 個完成列。專名暫留原文，正式術語與字級尚未定稿。
- 早期國王正常操作：`workplace/probe-door-up-70m/`、`probe-king-0-84m/` 及 `gui-king-v2/receipt.json`。探針使用既有六次 Space，再於 60m 按 Right 2m、64m 按 Up 250k。GUI 以畫面核對進入國王房間，首段文字與原版收據相符；空白鍵後 GOLD 更新至 1000。不把賞金確認稱為對話全部翻頁完成。
- 國王對話目前證據：`workplace/dialogue-v3/receipt.json`，130m 正常新局、領賞與返回城鎮。九個抽樣點的完整 1 MiB 記憶體、色號、色盤、CS:IP 及 cycles 均與無覆繪原版相同；241 字形、兩段話、50 個捲動列，覆繪區外 0。
- 國王實際 X 視窗：`workplace/gui-dialogue-hd-v5/receipt.json`。兩段中文全部顯示、按原版待續標記翻頁、1000 賞金及返回城鎮通過；最後持鍵為空、無覆繪失配。國王與回城鎮的原貌→HD→原貌分別驗完整還原。
- 國王 HD：`workplace/hd-king-prototype.png`、`hd-king-speaking-prototype.png`，完整 imagegen 提示詞組與來源在 `hd-king-generation.json`。只留本機研究區；兩個已觀察姿態，不稱為所有動畫完成。
- 對話位址索引：[docs/dialogue-source-index.json](docs/dialogue-source-index.json)，固定 84m 國王段傾印 SHA-256 `06b74a1ea26f25d7e81de50ecb9f9ece1e1851d9429b469c59cf6a80bbb1d156`。最新 IDA 匯出 `workplace/ida-dialogue-indexed-v5/`，新增 FF 多行路徑，不與片頭段混用。
- 本輪十二項具名 Go 測試、建置與 go vet 通過，無具名測試 skip；`bitmap` 套件沒有獨立測試。紀錄為 `workplace/test-weapon-rest-v1.jsonl` 十一項及 `test-weapon-v2.jsonl` 一項。存讀檔回歸為 `save-weapon-v1/receipt.json`，22 點完整狀態及區域比較通過。現行稽核為 `workplace/current-audit-weapon-sage.json`。
- 完整中文、其餘場景／角色 HD、音效輸出、戰鬥與實體手把仍未完成或未驗。

- 長老來源映射：`workplace/sage-source-mapping.json`，正常 283.9m 段傾印 SHA-256 `22076a51f2a0f3d10115840090b8331107efe79d31776ceea90775b7c7236a00`。SAR payload `0x1BCC6`、6973 bytes 對應 `01B1:A000`，九個文字／選單來源錨點均相符。
- 長老初訪與選單：`workplace/labels-sage-v3/receipt.json`。國王 241 字形加長老初訪 114 字形，四列選單；兩點完整原版狀態相同，對話／標籤區外差異 0。
- 中文存讀檔配對：`workplace/save-overlay-v1/receipt.json`。十九個正常路徑抽樣點完整 1 MiB 記憶體、色號、色盤、CS:IP、steps、cycles 均相同；兩側各自建立 256-byte SAVETEST.usr，內容相同，再退出並冷讀回。固定鍵序入口為 [tools/save-load.commands](tools/save-load.commands)。
- 原版獨立試玩：`workplace/save-path-verified/audit.json`，29 個原始檔不變、SAVETEST.usr SHA-256 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`；正常 Ctrl+Q／Y 退出碼 0，冷重開 F7／Y／輸名／Enter，回到 Muralla Town 長老門口及 1000 金幣。
- 完整存讀實際 X 視窗：`workplace/save-ui-gui-v2/receipt.json`、`visual-review.json`。新建 UIVIEW2.usr、正常退出、冷讀回 Muralla Town／1000 金幣；名稱提示、是／否與退出中文已目視。此流程使用逐行修正前的 `a678bbbb` 建置，未開 HD，不冒充最新建置的完整 GUI 重跑。
- UI 切換退回：`save-path-gui-v1/observe-before-restart.json` 保留一次 kind=2、first_difference=148 的舊對話畫格失效；原版名稱輸入框改變後安全移除中文，流程正常，沒有把它列為存檔失敗或宣稱零失配。

- 提示來源：`workplace/ui-source-mapping.json`，輸入 `save-path-verified/name-before-keyup.segment` SHA-256 `7fdeaf5ba642d54686cfa6c49c5ac41f11477e03dc05746dcbfb3d477754c3a5`；SAR1、SAR2、STICK 七個新增來源及正常繪字追蹤有映射。
- 最新配對：`workplace/save-ui-v3/receipt.json`，22 點完整原版狀態相同，逐點文字區外差異 0。涵蓋名稱提示、是／否、退出、F7 及 F9／3／Enter／Right；兩側獨立建立相同 256-byte 新檔並冷讀回。速度第二行的數字保留原版像素；Right 後提示清除，該抽樣點尚未證明位移。
- 舊速度提示失敗保留 `save-ui-modal-v1/`：最長行寬誤涵蓋第二行數字，導致中文退回。已改為各行自己的前綴範圍；沒有修改原版速度或輸入處理。
- 最新建置的短 GUI：`workplace/save-ui-modal-v2/receipt.json`、`visual-review.json`、`movement-comparison.json`。完整 F7 兩行、N 取消、F9／3 中文與數字保留；Enter 後 Right 持兩秒／放開，提示消失，角色由原生 x 約 136 移至 184，城堡錨點不變。持鍵空、退出碼 0、原始資料不變。此短流程沒有存讀檔或開 HD。

## 本輪長老 HD 與武器店證據

- 長老 HD 原版兩畫格與四像素差異：`workplace/sage-hd-source-mapping.json`。手繪圖 `workplace/hd-sage-prototype.png`，1536×1024，SHA-256 `a1fa3b745422f59153239b903591db104c5c3fb320c18ee7f5181a247c33a5aa`；完整提示詞、構圖修正及權利分類在 `hd-sage-generation.json`，只留本地。
- 長老實際 X 視窗：`workplace/gui-sage-hd-v1/receipt.json`、`region-comparison.json`。正常新局、領賞、進長老房屋後暫停，以同一狀態切原貌→HD→原貌；區外 0、動畫像素不變、完整還原，steps／cycles 不變。此 GUI 使用 `359bef32` 建置，只驗此 HD 切片。
- 武器店原版正常路徑：`workplace/shop-path-oracle-v1/commands.txt`、`evidence-index.json`，領賞後進店，售價 400、舊武器折抵 200，買後 GOLD=800，告別、Space 回城鎮及 Enter 查看裝備。
- 武器店來源：`workplace/weapon-source-mapping.json`，固定段 SHA-256 `22c3e2d3f0f9beff133f5bd5737f3e662429aba5c74b198bb707fe8d33315817`。商品來源與六個靜態對話均匹配固定 SAR2 payload `0x17817`。
- 商品名稱準備與搬移：`workplace/shop-path-buffer-trace-v1/buffer-trace.jsonl`，八個事件；正常 SI=AD1Dh／AD2Ch、BX=1871h／187Bh。之前小字形入口的有效零結果已排除，不再當成待追查項。
- 位址索引：[docs/weapon-source-index.json](docs/weapon-source-index.json)，重生工具 [tools/ida_weapon_evidence.py](tools/ida_weapon_evidence.py)。最新 IDA 9.4／16-bit 匯出 `workplace/ida-weapon-indexed-v1/weapon-evidence.json`，保留原名、位址、bytes、運算元及未證實警示。
- 武器店中文配對：`workplace/weapon-overlay-v2/receipt.json`，九點完整 1 MiB 記憶體、色號、色盤、CS:IP、steps、cycles 相同，文字區外 0。初訪、五列、兩個商品名、交易後與告別中文；動態報價仍為原文。固定鍵序為 [tools/weapon-shop.commands](tools/weapon-shop.commands)。
- 動態報價原版追蹤：`workplace/shop-path-glyph-v1/glyphs.jsonl`，225 字形；16 個抽樣的色號／色盤／64 KiB 段、cycles、IP 與未掛鉤原版相同，未宣稱全域記憶體對拍。已保存靜態文字及 400／200 的實際字形順序，後續分段覆繪須保留原版數字與輸入。
- 最新實際中文交易：`workplace/shop-path-gui-cht-v1/receipt.json`、`visual-review.json`。固定 `5ad8556d` 建置及 13／27 筆 catalog，冷新局領賞、購買練習劍、GOLD1000→800、中文告別、Space 出店、Return 查看 Training Sword 通過。最後持鍵空、退出碼 0、原版 29 檔不變。
- 已知中文缺口：商品清單上方詢問及報價模態下部分舊選單退回英文；保留 kind=2 兩筆，paragraph=10／12、first_difference=2552／242。動態報價、場景名稱、裝備標籤與名稱仍英文，不能把本輪當成武器店完整中文。
