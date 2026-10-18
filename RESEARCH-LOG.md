# 原版與手冊證據

## 2026-10-18：輸入盤點與中文手冊初讀

- 原始輸入雜湊見 [CONTEXT.md](CONTEXT.md) 的「原始輸入」。本節未引用程式位址，只引用壓縮檔成員與手冊印刷頁碼。
- ZIP 以 Docker 內 Python 3.13 `zipfile` 檢查 CRC，30 個項目均通過；含 `ZELIARD.EXE` 與 `.jsdos/dosbox.conf`。
- RAR 以既有 `dkk-inventory:bookworm-r1` 的 `lsar` 列目錄，含 17 張 JPG、`Thumbs.db` 與補完計劃文字。17 張掃描不等同手冊頁數。
- 手冊暫解至 `/tmp/zeliard-manual-preview-20261018/珍084-神劍封魔/`，原 RAR 唯讀掛載。此目錄只供本輪目視核對，不是永久證據來源。

| 證據 | 已讀內容 | 等級／適用範圍 |
|---|---|---|
| `SCAN0357_003.jpg`，印刷第 1 頁「前言」 | 公主被邪惡魔龍石化；接受邪靈挑戰，帶著人民期待與神靈祝福進入地下迷宮 | 已證實：手冊文字，不外推遊戲內劇情完整性 |
| `SCAN0357_003.jpg`，第 1 頁「硬體配備」 | 鍵盤／搖桿，搖桿非必備 | 已證實：手冊要求，不證明現代手把可用 |
| `SCAN0357_006.jpg`，第 7 頁「其它控制鍵」 | Esc 暫停；F1 音樂、F2 音效；Ctrl＋J 搖桿、Ctrl＋K 鍵盤、Ctrl＋Q 退出 | 已證實：手冊按鍵，仍需固定原版實跑核對 |
| `SCAN0357_006.jpg`，第 7–8 頁「開始遊戲」 | 城鎮交談、空白鍵推進訊息、方向鍵選擇、進建築與武器店選單 | 已證實：手冊操作說明，不稱為本專案驗證通過 |

README 只以以上已讀內容改寫背景與城鎮探索描述，不抄未核對的網路劇情或譯名。

## dosgolem 起始邊界

已讀 `/home/anr2/cht/dosgolem/README.md`、`CLAUDE.md`，並核對 `cmd/probe/main.go` 現有 `-exe`、`-root`、`-scratch`、`-steps`、`-dump-screen-png` 等旗標。目前尚無本遊戲的啟動收據。是否能執行及所缺能力均為未知。

## 2026-10-18：啟動、架構與覆繪實測

本節追加開工後的新證據，取代上一節「尚無啟動收據」的起始狀態。現況只以 CONTEXT 為準。

### 工具身份

- RAR 目錄工具：`lsar v1.10.1`，既有 `dkk-inventory:bookworm-r1`。
- Go：`go1.24.13 linux/amd64`，Ebitengine `v2.9.9`。
- 建置映像：`dosgolem-go-ebiten:1.24-r1`，ID `sha256:8ed6bdee2bce1f32f804e888f942b25bbf62f2165a43344bb4587f3371de0581`。
- GUI 映像：`psychicwar-go-ebiten:latest`，ID `sha256:c3d8835a9d39ec7229078a6f49c0036af2aaf1cee0f15e1a8b2cbb7f6d232ff3`。沿用其中的 Xvfb、xdotool 與 ImageMagick，沒有另建映像。
- dosgolem HEAD 與工作樹快照身份見 [CONTEXT](CONTEXT.md)。探針 SHA-256 為 `d763b2c7fbf846e65e5d90c35067583929251c80abd3ab1f0b5db58f440731e0`。

### 原始檔與手冊

- `ZELIARD.EXE`：3,050 bytes，SHA-256 `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`。
- `ZELRES1.SAR`：256,952 bytes，SHA-256 `56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0`。
- 原版啟動主控台自報 `ZELIARD Version 1.208`，Game Arts 1987–1990、Sierra On-Line 1990。這是固定輸入的程式輸出，不外推其他版本。

| 掃描成員 | SHA-256 |
|---|---|
| `SCAN0357_003.jpg`，第 1–2 頁 | `24c97021888208d1f745d19c2ff4d4e3babc1290cb33e1d6f32dc5dfb132041d` |
| `SCAN0357_005.jpg`，第 5–6 頁 | `8ddd761c499b6161bbdfc2d587cedc3fa701875073cd37df1ae3c2ebb1936d7c` |
| `SCAN0357_006.jpg`，第 7–8 頁 | `5f164cef7c1b4d1f590f6a2cb155a440047b8b350905d8da48eba6bf54d47bcf` |

第 6 頁明列數字鍵盤方向與斜跳，空白鍵為武器／交談、Alt 為法術、Enter 查看裝備；搖桿按鈕 1、2 分別對應武器與法術。這是現代手把映射的手冊證據，尚未驗證正常遊戲中的全部動作。

### 啟動基線與停止線

- 固定 200 萬指令、無輸入，進入視訊模式 13h，擷取標題。讀取了設定、BIN／DRV 模組與 SAR。兩次執行的色號、色盤、PNG、末端 40 道指令軌跡完全相同。
- 40m 指令探針送入七次空白鍵按下／放開，原版送出 14 次 IRQ1，擷取到英文片頭。未驗新局與城鎮，不能稱為完整可玩。
- 初次 `-dump-screen-png` 失敗是探針旗標只支援平面模式；改用 mode 13h 的 `-dump-vram` 與 `-dump-palette` 成功。原始失敗保留在 `workplace/probe-startup/`，不列產品缺陷。

### 單句 prototype 的基線勘誤

早期 40m 收據中的英文墨跡位於 y=142–149；即時前端無輸入片頭位於 y=143–150。兩者文字墨跡各為 337 像素、色號均為 102，垂直平移一像素後完全相同。它們的時鐘與輸入狀態不同，早期畫格不能直接當作即時前端同狀態基線。

以與前端相同的 3000 cycles／毫秒，在無輸入狀態執行 10,729,940 指令，dosgolem 自行重生的 320×200 色號畫格與診斷 GUI 的原版色號完全相同，SHA-256：`de520c43e8b0098258cf5938f9afb42b60e7f04a8742d3eee5e897fcbcbcad5d`。來源為 `workplace/probe-intro-clock3000/`。舊收據保留，不重命名或覆寫。

### 覆繪與 GUI 驗證

依使用者選定的架構，參考 psychic-war 的原版執行、HD 圖面、中文疊字分層。片頭試作使用 dosgolem `xlate.Watcher` 與 `xlate.Layer`，只匹配已驗證矩形，不寫入原版畫格或執行記憶體。

- 純輸入三項測試與真實原版畫格覆繪一項測試均通過，沒有 skip。
- `workplace/gui-smoke-v4/receipt.json`：原貌、滑鼠開啟手把頁、片頭中文三條流程正常結束，並擷取實際 X 視窗。
- `intro-region-comparison.json`：GUI 內原版色號與上述 dosgolem 基線完全一致。實際視窗遊戲區對最近鄰 3 倍原版，差異僅落在 `(88,136,160,16)` 覆繪區內，3,655 像素；區外 0。
- 中文譯文只有一句「兩千年前，」。來源 SAR 內偏移、其餘片頭文字、完整玩法與 HD 美術均未解完。此 prototype 不等於正式中文化完成。

## 2026-10-18：冷開機、城鎮操作與片頭文字來源

本節追加後續證據，取代前文對新局及文字偏移的未解狀態。原始輸入與工具快照身份不變，現行狀態見 CONTEXT。

### 正常城鎮切片

- 已證實：從 `ZELIARD.EXE` 冷開機，不載入包內存檔、不注入記憶體，能進入 `Felishika's Castle` 城鎮外景。原版收據為 `workplace/probe-newgame-60m/`。
- 原版探針固定 3000 cycles／毫秒，於第 12m、20m、28m、36m、44m、52m 指令各按住 Space 50,000 指令。右移版本另於第 60m 指令按住 Right 3m 指令，兩側都跑到 66m。
- 右移冷開機重播兩次，`screen.bin`、`palette.bin`、`screen.png` 與末端 40 道指令軌跡全部相同。畫格 SHA-256：`3f7fcdebe8a640f33be563046e3d7f267f8bfc4bcb08b983845d68be09685c48`。比較為 `workplace/town-replay-comparison.json`。
- 實際前端以 X11 視窗按鍵進入城鎮，再按右方向鍵約一秒並放開。目視核對角色與城堡的相對位置改變，放開後停止平移。收據為 `workplace/gui-town-v2/receipt.json`，沒有狀態載入或記憶體寫入捷徑。
- 狀態辨識使用原生畫面 `(48,162,136,22)` 的 LIFE／PLACE／地名區，與已核對城鎮基線逐像素完全相符後才開始右移。它不是全畫面一致或完整玩法的判定。
- 放開後持鍵集合為空。城堡靜態區 `(100,103,92,13)` 在右移前後不同，右移後與放開兩秒後相同。比較為 `movement-region-comparison.json`，不將動畫角色或特效要求為整張畫面相同。
- 第一輪 `gui-town-v1` 使用固定牆鐘送鍵，按鍵落在轉場，沒有驗到城鎮。這是驗證腳本的時機問題；後續改為觀察城鎮狀態列，未改原版流程來通過測試。
- `probe-castle-entry-70m` 的較短 Right 與 Up 嘗試仍擷取城堡外景，未證明進建築。對話、存讀檔與戰鬥保持未驗。

### 檔案字面與記憶體映射

`workplace/literal-text-search.json` 只做原始 bytes 搜尋，不宣稱整個 SAR 的格式或文字編碼已解明。

| 來源與位址空間 | 字面／範圍 | 等級 |
|---|---|---|
| `ZELRES1.SAR`，從 0 起算的檔案偏移 `0x109F` | `Two thousand years,` | 已證實：ASCII bytes 與片頭原文相符 |
| `ZELRES2.SAR`，檔案偏移 `0x184DF` | `Go outside` | 已證實：原始 bytes；商店 consumer 未追完 |
| `ZELRES2.SAR`，檔案偏移 `0x184F8` | `Buy weapon` | 已證實：原始 bytes；不外推所有選單控制碼 |

`ZELRES2.SAR` SHA-256：`bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`。其他來源雜湊見前文與來源清單。

以無輸入、10,729,940 指令及 3000 cycles 重生 dosgolem `01B1:0000` 起的 65,536 bytes。只讀傾印 SHA-256：`c9f188ea847e8d364ccf8dbc9f6a9b9c74e192c0c83bd7a012b552e6d89b13d8`，收據為 `workplace/probe-intro-memory/`。

| 檔案偏移，ZELRES1.SAR | dosgolem 段：位移 | dosgolem 實模式線性位址 | IDA base 0 的段內位移 |
|---|---|---|---|
| `0x109F` | `01B1:6FFB` | `0x8B0B` | `0x6FFB` |
| `0x10B4` | `01B1:7010` | `0x8B20` | `0x7010` |
| `0x11B9` | `01B1:7115` | `0x8C25` | `0x7115` |

三個文字錨點都符合「段內位移減 `0x6000` 等於檔案偏移減 `0xA4`」，見 `workplace/intro-memory-anchors.json`。此結果只涵蓋本次載入區段，不套用其他 SAR 記錄。

### 片頭單行 consumer

IDA 9.4、metapc、16 位元，從上述唯讀段傾印新建一次性 `.i64`。原始名稱、位址、bytes 與運算元保存於 `workplace/ida-intro-v4/intro-evidence.json`，不改名。主題只限目前片頭顯示，不追查其他 helper。

| IDA 原始定位與 bytes | 同一段 dosgolem 的執行證據 | 等級與語意 |
|---|---|---|
| `sub_6358` 的 `0x6363`：`BE F0 6F`，`mov si, 6FF0h`；`0x6366`：`2E FF 16 0C 30`，`call word ptr cs:locret_300C` | 第 7,808,914 指令於 `01B1:6366`，`DS=01B1, SI=6FF0`；下一指令進入 `01B1:32C9` | 已證實：首行起點含前導空格，由間接呼叫送入 `sub_32C9` |
| `0x300C` 的原始 word 為 `C9 32` | 段傾印與實際跳轉都指向 `01B1:32C9` | 已證實：本次快照的呼叫目的地，不外推其他模組版本 |
| `sub_32C9` 的 `0x32D8`：`AC`，`lodsb`；`0x32D9` 比較 `FFh`；`0x32DE` 減 `20h` 並在小於 `20h` 時返回 | 逐次記錄 `DS=01B1`，`SI` 從 `6FF0` 遞增；第二次呼叫於第 10,676,524 指令，`SI=7010` | 已證實：讀取原文 bytes 與行間推進；尚未窮舉所有控制碼 |
| `sub_32C9` 從 `ES:4511` 清除 `640h` 個 word；非空格字元經索引、8 列與 `ES:[DI]` 等寫入處理 | 首行執行時 `ES=01B1`；指令、原始 buffer 位址保留在 IDA 匯出 | 強推論：建立單行字形緩衝；送往畫面的定位、遮罩、捲動及特效仍待追查 |

動態暫存器收據為 `workplace/probe-intro-consumer/`。只觀察，不修改資料或暫存器。正式翻譯掛鉤與 catalog 仍未達 READY，現有畫格匹配 prototype 保留。

## 2026-10-18：手繪 HD 邊框與片頭顯示區

### HD 邊框

使用者看過兩種畫風的比較圖，明確選擇 B「手繪奇幻」。這次只決定畫風，未定稿構圖、角色比例、正式倍率或預設主題。

- 比較圖：本地 `workplace/hd-style-preview.png`，SHA-256 `58e93ba96f1c9891cf0d8c71d47194d6590740021ee0d534336ffa96a0010fef`。
- 試作圖：內建 imagegen 產生透明手繪雙欄石雕，869×1810，SHA-256 `bb9cc2e211b1d2223e3baf0e185bab25bc484bb26477078e9d769e67d5f9582d`。本地 `workplace/hd-border-prototype.png` 與 `hd-border-generation.json` 保存原圖、完整提示詞、視覺參考及本機試作分類。沒有把它當成已定稿發行素材。
- 實作：完整核對兩側 `(0,0,48,200)`、`(272,0,48,200)` 的原版色號與 RGB。圖面直接合成至 144×600 顯示區；任一側、色盤、畫格長度或素材身份不符時整組退回原貌。正式素材格式仍未定案。
- 測試：真實城鎮基線與生成圖測試通過，涵蓋邊界、唯讀輸入、單點失配、RGB／淡入變化、不完整畫格及缺圖。這是顯示層測試，不證明原版玩法規則。
- 實際 X 視窗：正常冷開機進城鎮、右移及放開後，用工具列暫停。相同畫格下切原貌→HD→原貌，差異只有兩側共 172,017 像素，中間區外差異 0；切回原貌後完整畫格相同。
- 收據：`workplace/gui-hd-frame-v1/receipt.json`、`hd-region-comparison.json`；截圖 `town-original-paused.png`、`town-hd-paused.png`、`town-returned-original-paused.png` 已目視核對。原版來源保持不變，沒有直接進入狀態或記憶體注入。
- 同版前端的原貌、手把頁及片頭單句三條流程在 `workplace/gui-smoke-v6/` 通過。完整場景、角色、動畫與實體手把仍未驗，不能用邊框推論其完成。

### 片頭顯示區定位

仍用固定 10,729,940 指令、3000 cycles、無輸入的段傾印，雜湊與位址空間沿用前節。IDA 9.4 重新建庫，加入原始間接目的地 `32C9h`、`332Ch`、`33B7h` 的最小函式邊界。只追本次片頭文字需要的 `sub_332C`；另一目的地不深挖。

| 原始定位 | bytes／運算元及動態觀察 | 等級 |
|---|---|---|
| IDA `sub_332C`，對應 dosgolem `01B1:332C` | `mul dl`、`add ax,ax`、`add ax,4511h`；來源行由 AX 選取。第 7,813,922 指令的 `AX=0000, BX=0020, CX=5078`，後續 AX 依次為 1、2、3，首行共有 0–9 十次 | 已證實：本次片頭的來源行序與調用；不外推所有畫面 |
| IDA `0x3369`，dosgolem `01B1:3369` | `8B F8`，`mov di,ax`；前 11 次 `AX=2800h` | 已證實：本次顯示起點 |
| IDA `0x339C`，dosgolem `01B1:339C` | 初始 `ES=A000, DI=2800, CX=0078`；下一列 `DI=2940`，之後 `2A80`。末段 `81 C7 40 01`，`add di,140h` | 已證實：MCGA 視訊列間距 320 bytes；本次目的區為原生 `(0,32,320,120)` |
| IDA `0x33A0`、`0x33A6` | `26 21 15`，`and es:[di],dx`；`26 09 05`，`or es:[di],ax`；本次 `DX=9999h, BP=6666h` | 已證實：使用遮罩混合寫入；遮罩的完整美術語意與其他特效仍未知 |

動態收據：`workplace/probe-intro-display/`。顯示區內還有背景與特效，不能直接用整個 320×120 矩形蓋中文。正式中文仍需文字墨跡範圍、行距／換頁與失效契約；證據足以定位掛鉤後停止追查無關 helper。

位址語意改由 [docs/source-index.json](docs/source-index.json) 保存。IDA 匯出自動附原始定位、語意、等級、證據與未證實警示，不從散文自動匯入，也不改原始函式名稱。

## 2026-10-18：首章捲動與國王對話

### 首章捲動中文試作

沿用前節固定 SHA-256 的 `ZELRES1.SAR`、`memory-01B1_0000.bin`、dosgolem 快照與 IDA 9.4。位址空間仍為 IDA base 0 的段內位移，對應 dosgolem `01B1:<位移>`，不與實模式線性位址混用。

- 已證實：首章從 SAR 檔案偏移 `0x1094` 起，30 個 CR 分隔行，含空白行；`0x13DB` 為 FF 結束碼。執行來源 DS:SI 對應 `01B1:6FF0` 至 `01B1:7337`。30 行逐次來源 bytes 均與本地 SAR 相符，不把此格式外推其他章節。
- 已證實：IDA `33A0` 的 `26 21 15` 以 DX=`9999h` 保留背景；`33A4` 的 `23 C5` 以 BP=`6666h` 選來源文字；`33A6` 的 `26 09 05` 合入目的視訊字。這個已驗首章的位元分離足以覆繪，不必重解其餘圖像特效。
- 已證實：`sub_332C` 的 `33B6`，原始 bytes `C3`，是逐列 viewport 複製完成的觀察點。掛鉤只讀事件，原版自行執行返回。對照前節 IDA 指令與本輪 430 個完成列事件。
- 實作為可撤回 prototype：TSV 只保存來源鍵與試譯，原文從本地資料讀取；中文直接合成到三倍畫布，每字 24×24、每行 30 顯示像素。等待原版一列複製完成才更新中文位置；複製途中暫留上一個完整 viewport。正式字級、專名及版面未定稿。
- `workplace/intro-scroll-v2/receipt.json`：無輸入、3000 cycles，六個抽樣點為 10,729,940、14m、30m、65m、110m、160m 指令。每點的完整 1 MiB 記憶體、CS:IP、虛擬 cycles、色號及色盤均與無掛鉤原版相同。Draw 後原版色號不變，viewport 區外差異 0。
- 結尾需保持中文捲出：首輪 `intro-scroll-v1` 在讀 FF 時過早清除，110m 畫面重現英文。修正為停止新增譯文、繼續沿原版空白列捲動；v2 的 110m 尚有 6 行中文，160m 換章後完全移除。舊收據保留作勘誤，不改名成通過版本。
- 實際 X 視窗 `workplace/gui-intro-scroll-v2/` 顯示多行中文；擷取時 7 個來源行、62 個完成列，背景首飾保留。收據記錄二進位、字型及腳本 SHA-256。生成 PNG 與原版畫面只留本地。
- 索引增補後，從相同唯讀段傾印重建一次性 IDA 資料庫，匯出 `workplace/ida-intro-scroll-indexed/intro-evidence.json`。原始函式、位址、bytes 與運算元不改名；JSON 自動合併現行索引及輸入身份。

### 國王首段對話

- 原版探針從冷開機走既有六次 Space 新局序列；60m 按住 Right 2m，64m 按住 Up 250k，跑到 70m。`probe-door-up-70m` 出現國王肖像及對話框，尚未顯示文字。
- 同序列自然跑到 84m，`probe-king-0-84m` 顯示完整首段：國王給旅費 1000 金幣。沒有跳入函式、載入測試狀態或寫入記憶體。畫面不代表對話已全部結束。
- `gui-king-v2` 冷開機後由實際 X 按鍵進城鎮，小步 Right／Up 操作，核對下方國王名稱，再等待文字區 `(54,98,208,40)` 與原版收據逐像素相符。按 Space 後目視核對金額為 1000、對話框出現待續標記。輸入放開，原始檔案不變。
- `gui-king-v1` 用固定 0.65 秒移動，角色仍在門左邊，未進城堡。這是視窗更新與驗證腳本的時機問題；v2 改為觀察實際畫面及小步操作，不改原版進門規則。此失敗不列為產品缺陷。
- 對話翻頁／結束、存讀檔與戰鬥仍未驗。國王對話尚未中文化，首章掛鉤不能套用其繪字端。
