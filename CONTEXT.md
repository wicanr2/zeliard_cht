# 目前狀態

更新日期：2026-10-18，Asia/Taipei。

## 目前狀態表

| 項目 | 現況 | 證據與限制 |
|---|---|---|
| 使用者目標 | dosgolem、繁中、HD、工具列手把綁定 | 已授權開工；規則見 [AGENTS.md](AGENTS.md) |
| GitHub | 已建立並推送 [wicanr2/zeliard_cht](https://github.com/wicanr2/zeliard_cht) | 回讀 `visibility=PRIVATE`；裝備中文功能提交 `c7b754e` 已推送並核對遠端 |
| 本地 Git | 已初始化 `main`，設定 `origin` | 環境 placeholder 已不存在；使用 GitHub noreply 信箱作本儲存庫提交身份 |
| 遊戲包 | 已安全解包與逐檔盤點 | 30 個壓縮項目，含目錄；29 個檔案，清單在本地 `workplace/source-manifest.json` |
| 中文手冊 | RAR 已列目錄，17 張 JPG | 已人工讀第 1、5–8、10、13 頁；另含封面、地圖等掃描，17 張不等於 17 頁 |
| dosgolem | 正常新局、國王領賞、長老存檔與冷讀檔已驗 | 原版、中文覆繪配對及實際 X 視窗均通過；戰鬥未驗 |
| 架構 | 使用者已選定直接執行原版再覆繪 | 參考 psychic-war，不另建玩法重製引擎；見 [架構規格](docs/spec/001-runtime-overlay.md) |
| 前端 | `workplace/bin/zeliard-preview` 已建置，GUI 雛形可啟動 | 中文工具列、手把綁定設定；音效輸出尚未接入 |
| 中文 | 工具列、首章、國王、長老、存讀提示、武器店與裝備 prototype | 報價、裝備分類開關的配對與實際視窗通過；資金不足限配對。其餘物品與文字待完成 |
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

依 [WORKLIST.md](WORKLIST.md) 接武器商人HD與其他商店中文，再推進其餘物品、片頭／對話及場景／角色。已有商品、報價片段、重用數字暫存、清框及裝備繪字證據，不重新猜測這些入口。正常存讀檔已完成，不重新開啟輸入時序研究。正式字級、術語及素材待校訂。證據入口為 [RESEARCH-LOG.md](RESEARCH-LOG.md)，驗證狀態見 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md)。

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
- 前輪 HD 與首章稽核：`workplace/current-audit-hd.json`、`current-audit-scroll.json`，保留各自來源與二進位身份。提示、長老／商品、彈窗及報價輪另保留 `current-audit-ui.json`、`current-audit-weapon-sage.json`、`modal-audit-v1/current-audit.json`、`quote-audit-v1/current-audit.json`；目前為 `workplace/inventory-audit-v1/current-audit.json`。
- 同狀態區域比較：`workplace/gui-smoke-v4/intro-region-comparison.json`。原版色號完全相同，覆繪區內 3,655 個像素不同，區外 0。
- 初始化工具身份：`workplace/gui-smoke-v4/provenance.json`，記錄當時 13 個來源檔、二進位、dosgolem 快照及映像 ID。該份舊收據沒有 Git 身份，現行程式以本儲存庫 HEAD 為準。
- 最新前端二進位 SHA-256：`52826ff520b8b5af184e02dbe3a76188d54f9a4f6d610766d12a623d16cb213a`，主入口與 `zeliard-preview-inventory-observe` 相同。基於報價功能加裝備覆繪及唯讀觀測欄位，未稱正式版本；前輪 `zeliard-preview-quote-heartbeat` 保留歷史身份。
- 首章中文最新重跑：`workplace/intro-scroll-v3/receipt.json`。30 個來源行、430 個完成列事件；10,729,940／14m／30m／65m／110m／160m 指令抽樣，完整 1 MiB 記憶體、畫面色號、色盤、CS:IP 與虛擬時間均與無掛鉤原版相同。覆繪區外差異 0；結尾捲出與換章移除已驗。
- 首章實際 X 視窗：`workplace/gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`。自然片頭擷取時已觀察 7 個來源行、62 個完成列。專名暫留原文，正式術語與字級尚未定稿。
- 早期國王正常操作：`workplace/probe-door-up-70m/`、`probe-king-0-84m/` 及 `gui-king-v2/receipt.json`。探針使用既有六次 Space，再於 60m 按 Right 2m、64m 按 Up 250k。GUI 以畫面核對進入國王房間，首段文字與原版收據相符；空白鍵後 GOLD 更新至 1000。不把賞金確認稱為對話全部翻頁完成。
- 國王對話目前證據：`workplace/dialogue-v3/receipt.json`，130m 正常新局、領賞與返回城鎮。九個抽樣點的完整 1 MiB 記憶體、色號、色盤、CS:IP 及 cycles 均與無覆繪原版相同；241 字形、兩段話、50 個捲動列，覆繪區外 0。
- 國王實際 X 視窗：`workplace/gui-dialogue-hd-v5/receipt.json`。兩段中文全部顯示、按原版待續標記翻頁、1000 賞金及返回城鎮通過；最後持鍵為空、無覆繪失配。國王與回城鎮的原貌→HD→原貌分別驗完整還原。
- 國王 HD：`workplace/hd-king-prototype.png`、`hd-king-speaking-prototype.png`，完整 imagegen 提示詞組與來源在 `hd-king-generation.json`。只留本機研究區；兩個已觀察姿態，不稱為所有動畫完成。
- 對話位址索引：[docs/dialogue-source-index.json](docs/dialogue-source-index.json)，固定 84m 國王段傾印 SHA-256 `06b74a1ea26f25d7e81de50ecb9f9ece1e1851d9429b469c59cf6a80bbb1d156`。最新 IDA 匯出 `workplace/ida-dialogue-indexed-v5/`，新增 FF 多行路徑，不與片頭段混用。
- 前輪十二項具名 Go 測試、建置與 go vet 通過，無具名測試 skip；`bitmap` 套件沒有獨立測試。紀錄為 `workplace/test-weapon-rest-v1.jsonl` 十一項及 `test-weapon-v2.jsonl` 一項。該輪存讀檔為 `save-weapon-v1/receipt.json`；本輪受影響五項及新存讀檔收據見下節。
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

## 前輪長老 HD 與武器店證據

- 長老 HD 原版兩畫格與四像素差異：`workplace/sage-hd-source-mapping.json`。手繪圖 `workplace/hd-sage-prototype.png`，1536×1024，SHA-256 `a1fa3b745422f59153239b903591db104c5c3fb320c18ee7f5181a247c33a5aa`；完整提示詞、構圖修正及權利分類在 `hd-sage-generation.json`，只留本地。
- 長老實際 X 視窗：`workplace/gui-sage-hd-v1/receipt.json`、`region-comparison.json`。正常新局、領賞、進長老房屋後暫停，以同一狀態切原貌→HD→原貌；區外 0、動畫像素不變、完整還原，steps／cycles 不變。此 GUI 使用 `359bef32` 建置，只驗此 HD 切片。
- 武器店原版正常路徑：`workplace/shop-path-oracle-v1/commands.txt`、`evidence-index.json`，領賞後進店，售價 400、舊武器折抵 200，買後 GOLD=800，告別、Space 回城鎮及 Enter 查看裝備。
- 武器店來源：`workplace/weapon-source-mapping.json`，固定段 SHA-256 `22c3e2d3f0f9beff133f5bd5737f3e662429aba5c74b198bb707fe8d33315817`。商品來源與六個靜態對話均匹配固定 SAR2 payload `0x17817`。
- 商品名稱準備與搬移：`workplace/shop-path-buffer-trace-v1/buffer-trace.jsonl`，八個事件；正常 SI=AD1Dh／AD2Ch、BX=1871h／187Bh。之前小字形入口的有效零結果已排除，不再當成待追查項。
- 位址索引：[docs/weapon-source-index.json](docs/weapon-source-index.json)，重生工具 [tools/ida_weapon_evidence.py](tools/ida_weapon_evidence.py)。最新 IDA 9.4／16-bit 匯出 `workplace/ida-weapon-indexed-v1/weapon-evidence.json`，保留原名、位址、bytes、運算元及未證實警示。
- 武器店中文配對：`workplace/weapon-overlay-v2/receipt.json`，九點完整 1 MiB 記憶體、色號、色盤、CS:IP、steps、cycles 相同，文字區外 0。初訪、五列、兩個商品名、交易後與告別中文；動態報價仍為原文。固定鍵序為 [tools/weapon-shop.commands](tools/weapon-shop.commands)。
- 動態報價原版追蹤：`workplace/shop-path-glyph-v1/glyphs.jsonl`，225 字形；16 個抽樣的色號／色盤／64 KiB 段、cycles、IP 與未掛鉤原版相同，未宣稱全域記憶體對拍。已保存靜態文字及 400／200 的實際字形順序，後續分段覆繪須保留原版數字與輸入。
- 最新實際中文交易：`workplace/shop-path-gui-cht-v1/receipt.json`、`visual-review.json`。固定 `5ad8556d` 建置及 13／27 筆 catalog，冷新局領賞、購買練習劍、GOLD1000→800、中文告別、Space 出店、Return 查看 Training Sword 通過。最後持鍵空、退出碼 0、原版 29 檔不變。
- 該輪 GUI 留有 kind=2 兩筆，paragraph=10／12、first_difference=2552／242。商品框上方詢問及報價旁舊選單退回的限制已由下節補正；動態報價、場景名稱與裝備中文仍未完成。

## 前輪彈窗遮擋證據

- 本節取代上節的商品框上方／報價框旁退回限制，其餘缺口保留。`workplace/shop-path-modal-trace-v1/evidence-index.json` 為原版追蹤入口，購買 9 點、取消 7 點完整狀態相同；26 次矩形繪製均完整返回，改變像素區外 0。商品框 `(84,110,148,36)`、是／否框 `(188,43,48,25)` 已證實。
- [位址索引](docs/weapon-source-index.json) 已回填清框入口與參數。`workplace/ida-modal-indexed-v1/weapon-evidence.json` 是最新 IDA 9.4／16-bit 九函式匯出，輸入仍為固定武器店段 `22c3e2d3…`，不與國王段混用。
- 覆繪依清框事件版本裁切，框外仍逐像素比對。完整原框範圍還原後才解除遮擋；未知變化安全退回，不改價格或輸入。
- `workplace/modal-overlay-v2/receipt.json` 九點完整原版狀態相同、文字區外 0；舊對話及各舊標籤的遮擋區改變 0。商品框上方中文與報價旁舊選單中文已驗。
- 五項受影響測試通過：國王、長老、遮擋、商店在 `modal-overlay-v2/test.jsonl`；存讀檔在 `modal-save-v2/test.jsonl`。前者保留一次已過時的標籤數量斷言失敗，未稱整檔 PASS；新存讀檔測試逐一核對必要來源、遮擋區不寫入及未知變化撤除，22 點完整狀態與相同 256-byte 新存檔／冷讀回通過。未重跑未改動的 HD／手把測試。
- 最新實際 X 視窗：`workplace/shop-path-gui-modal-v1/receipt.json`、`visual-review.json`、`modal-box-comparison.json`，固定 `c355ade5` 建置。商品框上方中文完整，是／否框旁舊中文保留，確認框與前輪差異 0；1000→800、告別、裝備查看、持鍵空與退出 0 均通過。
- 限制：原版確認框仍遮掉部分舊中文字，未重新排視窗；動態報價、裝備與場景名稱仍英文。GUI 留一筆離店舊對話退回 paragraph=12／kind=2／first_difference=422，不稱零失配。其餘 HD、音效、戰鬥與實體手把仍待完成或驗證。
- 現行來源、收據、文件連結及擁有權稽核入口：`workplace/modal-audit-v1/current-audit.json`。原始遊戲、手冊、字型、圖像與完整收據維持本地忽略目錄。

## 本輪報價與資金不足證據

- 本節取代前輪第一筆報價仍英文的限制。對話TSV目前14個靜態段及4個報價模板，商品譯名共用原有27筆標籤來源，沒有另複製商品翻譯。
- 來源入口 `workplace/shop-path-quote-source-v1/evidence-index.json`；BC33h數字以單一FF結束，不是NUL或FF00。售價及折抵依原版已完成字形分別保存，避免同一暫存區被後一筆覆用。[位址索引](docs/weapon-source-index.json)已回填，最新IDA9.4／16-bit十一函式匯出為 `ida-quote-indexed-v1/weapon-evidence.json`。
- `workplace/quote-overlay-v2/receipt.json`購買9點與 `insufficient/receipt.json`不足11點，完整1MiB、色號、色盤、CS:IP、steps、cycles相同，區外與彈窗越寫0。400／200原版數字圖形完整，買後800；不足翻頁、告別及返回城鎮仍1000。
- `quote-overlay-v2/test.jsonl`三個頂層測試及五個代入欄位／字型子測試通過。最早v1失敗為測試引用舊文字擁有的遮擋版本，已改核對目前報價版本，不放寬越寫檢查。HD、手把與存讀檔程式未改，沒有重跑未受影響測試。
- 實際視窗v1在較慢更新率下耗盡舊56秒入口預算，沒有進店；v2按放已送達，但Draw可能跳過更新幀數60整倍數，觀測檔沒有心跳而被腳本中止。兩份失敗保留，不當成CPU停滯或報價缺陷。診斷觀測已改每次Draw原子寫入完整快照，不改原版cycles或輸入。
- 實際視窗 `workplace/shop-path-gui-quote-v3/receipt.json`、`visual-review.json` 使用固定 `55f458ac` 建置。冷新局、國王兩段中文與領賞、完整中文報價400／200、購後800、告別、回城鎮及裝備查看通過。最後持鍵空，由既有capture-tail擷取流程正常終止前端，退出碼0；本輪沒有另按原版Ctrl+Q。原版裝備名稱仍為Training Sword。
- 主入口已同步該固定建置。來源及收尾稽核入口為 `workplace/quote-audit-v1/current-audit.json`；不足分支只驗配對及離線圖，未冒充本次GUI路徑。
- GUI未開HD。保留一筆離店舊對話退回，paragraph=12／kind=2／first_difference=422，與前輪相同；報價完成時沒有退回，不稱全程零失配。
- 收尾稽核核對53個自有檔案、29個原版、882個執行器快照、100個文件連結及全部正式Go來源符合功能提交。458筆按放均有送達確認，擁有權無異常，本專案容器已回收。
- 限制：報價按完成片段顯示，未完成時暫留白；數字隨中文排版移動。資金不足沿既有比例揭露，沒有宣稱逐詞同步。只驗第一筆購買與第二件不足，其他報價、裝備、其餘中文與HD、音效、戰鬥及實體手把仍待完成。

## 本輪裝備頁證據

- 本節補正練習劍裝備頁仍英文的限制。來源入口 `workplace/inventory-source-v1/mapping.json`，七個來源均核對SAR2、固定段與正常呼叫。固定段SHA-256 `2ecedadb544f1d089612ca19b1f42379fa83866c447121b131c61435fcdf7e3e`；[裝備索引](docs/inventory-source-index.json)及五函式IDA9.4匯出 `inventory-ida-indexed-v1/evidence.json` 保留原始定位與分級。
- 選單TSV共34筆，新增四分類標題、「無」及練習劍兩行來源。大字逐字AA2Bh、完成AA30h；保留AH對應色號及右下陰影，小字沿既有入口。不改物品識別、效果或輸入。
- `workplace/inventory-overlay-v3/inventory/receipt.json` 11點完整1MiB、色號、色盤、CS:IP、steps與cycles相同，區外及彈窗越寫0。初開、放開、重開各九個新文字區，17359個像素改變；關閉整張覆繪差0。每個被遮擋舊標籤獨立繪製差0，未知像素撤除通過。
- v1及v2測試分別把隱藏舊標籤計入目前標籤數、把內部記錄誤作關閉殘留。已改核對七鍵九次出現、每個舊標籤零可見像素及關閉完整畫格差0，沒有放寬完整原版狀態或區外檢查。
- 初次實際視窗 `shop-path-gui-inventory-v1` 使用 `3e0aa6f` 候選，完成交易但止於完整告別，尚未到裝備；按放均送達，90秒畫面等待結束後正常回收。未執行的補Space沒有計入輸入。失敗及事件保留，不當作裝備GUI通過或原版CPU停滯。
- 原版末尾等待的最小IDA證據為 `workplace/inventory-ida-v1/wait-evidence.json`。告別完整中文可能早於末尾控制碼；診斷新增唯讀文字游標及待續旗標，驗證腳本須核對原版末尾狀態再送Space。v3已核對兩個穩定末尾快照並回城鎮，裝備GUI另用下列冷讀檔入口驗證。
- 新診斷候選 `workplace/bin/zeliard-preview-inventory-observe` SHA-256 `52826ff520b8b5af184e02dbe3a76188d54f9a4f6d610766d12a623d16cb213a`，只有觀測JSON新增兩個唯讀欄位，覆繪及遊戲輸入未變。最新武器店IDA十三函式匯出在 `inventory-wait-indexed-v1/evidence.json`，合併[分級索引](docs/weapon-source-index.json)。
- v2在120秒入口上限耗盡時仍前進，末端40,597,431指令／42,587,957cycles，城鎮HUD尚未畫完。v3改180秒入口後完成交易及原版末尾游標檢查，回城鎮的Return只持約十萬指令，未開裝備。v4將尾段按放改原版步數，但較快吞吐使舊牆鐘城門微移走過入口，未見國王。四份失敗均保留，詳見WORKLOG；不當作CPU停滯或裝備通過。
- 實際視窗 `workplace/shop-path-gui-inventory-v5/receipt.json` 使用自己正常建立的 `save-path-verified/scratch/SAVETEST.usr` 冷F7讀回，雜湊 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`。初始武器同為Training／Sword；開啟、關閉及重開通過，初開與重開九個中文ROI相符，關閉城鎮靜態區相符。金幣1000、存檔不變、持鍵空、capture-tail正常終止碼0。此GUI沒有新交易、未開HD，不稱全畫面同狀態對拍。輸入沿已驗原版步數及畫面同步，不注入位置或記憶體。
- 共用標籤回歸 `workplace/inventory-regression-v1/test.jsonl` 長老與存讀檔兩項通過。長老2點及存讀檔22點完整狀態相同、區外0；兩側獨立建立相同256-byte新檔並冷讀回。最新go vet及重建通過，重建二進位與候選逐位元相同，未重跑未改動的HD／手把測試。
- 限制：配對只驗購得練習劍後的顯示及開關，GUI只驗自己新建存檔冷載入後的相同初始武器。其他物品、使用、魔法與裝備切換未驗；完整中文、其餘HD、音效、戰鬥與實體手把仍待完成或驗證。
- 後續銀行及道具店的字面候選為 `workplace/inventory-source-v1/next-shop-candidates.json`，只證實SAR2檔案偏移與字面，尚未驗runtime映射或行為。可沿本輪已驗正常存檔的冷讀回路徑開始新商店試玩，不重做已完成的存讀檔研究。
- 武器商人HD的原版候選為 `workplace/inventory-source-v1/merchant-pose-candidates.json`，保存六個正常路徑畫格的內框色號、RGB及來源雜湊分組。這只供後續美術接入定位，未驗收新HD，不把雜湊分組當作動畫語意。
- 武器商人手繪樣圖為 `workplace/merchant-hd-v1/merchant-prototype.png`，1536×1024，SHA-256 `2660fc7892f9a5229d4cee3723e992146992a460a1678197793cc890b0dfb59b`。完整內建imagegen提示詞、構圖與風格參考及本地權利分類在同目錄 `generation.json`。已目視禿頭、紅鬍、藍背心、綠臂帶及櫃台構圖，尚未接入或驗收HD；原版衍生圖不入Git。
