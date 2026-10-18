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

## 2026-10-18：國王初訪中文、正常結束與 HD 肖像

### 對話來源與繪字契約

本節完成前節尚未驗證的國王初訪結束。工具沿用 dosgolem 固定快照、Go 1.24.13 與 IDA 9.4。原版 EXE、SAR 與 ZIP 的身份不變，見 CONTEXT。所有位址是 IDA binary base 0 的段內偏移，對應 dosgolem `01B1:<偏移>`，不與線性位址混用。

| 來源 | 位置／SHA-256 | 等級與用途 |
|---|---|---|
| `ZELRES2.SAR` | `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1` | 已證實：固定版本的國王來源 |
| 正常 84m 段傾印 | `workplace/probe-king-memory-84m/memory-01B1_0000.bin`，65,536 bytes；`06b74a1ea26f25d7e81de50ecb9f9ece1e1851d9429b469c59cf6a80bbb1d156` | 已證實：從正常新局進門，沒有狀態注入 |
| 第一段字面 | SAR `0x17250`、`0x1726A`；段內 `A436h`、`A450h` | 已證實：兩段字面屬同一個旅費段落，中間有原版控制碼 |
| 第二段字面 | SAR `0x172BB`；段內 `A4A1h` | 已證實：要求先整裝、再進迷宮的後續段落 |

這三個錨點都符合「SAR 檔案偏移等於段內偏移加 `0xCE1A`」。讀檔紀錄在第 64,066,878 指令從 SAR `93722` 讀入 1,954 bytes；此模組從段內 `A000h` 開始。此映射只適用本次國王模組，不套用其他 SAR 項目。

- 已證實：`A000h`、`A002h` 是模組 header word，分別指向 `A004h`、`A302h`。首輪導航錯把 `A000h` 當作指令入口；讀 header 後從同一唯讀輸入重建，`ida-king-v2` 才作函式邊界證據。
- 已證實：`A056h` 的 `mov [FF4Ch],si` 在正常初訪觀察 `SI=A42Fh`，接著經 `CS:6004h` 進入 `sub_706C`。`7097h` 的 `lodsb` 讀來源；`711Bh` 經 `CS:2022h` 呼叫字形繪製；`7120h` 的 `58` 是返回後的 `pop ax`。
- 已證實：首字的 `AX=0142h, BX=0038h, CX=0063h, SI=A437h`，對應來源 B、原生座標 56,99。`sub_27E9` 在 `2829h` 執行 `32 ED`，`xor ch,ch`，然後以 CX 乘每列 320 bytes。繪字的 Y 只來自 CL。
- GUI 揭露的勘誤：正常翻頁曾在 `711Bh` 帶入 `AX=016Ch, BX=0036h, CX=028Bh, SI` 對應字面 `A516h`。原型把整個 CX 當座標，錯誤拒絕這一字形並重建段落。`gui-dialogue-v3/town-final.png.json` 保存原始暫存器收據；修正為 CL 後，`gui-dialogue-hd-v5` 完整觀察 241 個字形、兩段話，沒有拒絕或重建。原版暫存器不被覆繪修改。
- 原版 `0Ch` 清框、`2Fh`／`0Dh` 換行、`FFh` 返回動作碼仍交由原版。TSV 只提供兩個來源鍵的試譯；`Garland`、`Jashiin` 暫留原文，未宣稱官方譯名。
- 正常冷開機到 130m 指令，以相同鍵盤序列對照有、無覆繪兩側。`workplace/dialogue-v3/receipt.json` 的九個抽樣點，完整 1 MiB 記憶體、色號、色盤、CS:IP 與虛擬 cycles 均相同；241 字形、50 個捲動列、1000 賞金及回城鎮，覆繪區外差異 0。這只證明國王初訪，不外推其他對話。
- 正式索引為 [docs/dialogue-source-index.json](docs/dialogue-source-index.json)。IDA 匯出 `workplace/ida-dialogue-indexed-v3/dialogue-evidence.json` 自動合併索引、輸入雜湊與未證實警示，保留原始函式名、位址、bytes 與運算元。

### 國王 HD 圖面

- 原版肖像外框是 `(112,23,96,64)`，黃色邊線維持原樣；HD 僅替換內部 `(113,24,94,62)`。兩個正常對話原版姿態來自 `probe-king-memory-84m`、`probe-king-pose-96m`，完整色號與 RGB 均核對，未知姿態或淡入整組保留原貌。
- 原版兩姿態差異有 68 像素，範圍在肖像內部相對 `(31,40)` 至 `(59,52)`。只證明這兩個口部圖樣，不把它們當成全部肖像動畫。
- 內建 imagegen 產生 1536×1024 閉口圖與說話變體。SHA-256 分別為 `d6be3a9fa2c93da491485e4167bc941f75367b8142ae1b3b6131dfffe6e78a70`、`06ca7fbe1dd91ebdf0c602d295019b8d6eb9cead56d3ae91d4b2bc308b52ff00`。完整提示詞組、來源圖、權利分類及檔案路徑在本地 `workplace/hd-king-generation.json`。
- 美術直接合成到 288×192 顯示圖面。說話變體只混合 `(104,86,80,39)` 的口部區，外圍像素固定使用閉口圖，避免整張肖像隨動畫變色。原版色號決定姿態，沒有另開動畫時鐘或改變原版位置／規則。
- `gui-dialogue-hd-v5` 從正常新局進國王房間，中文顯示後經工具列暫停，再切原貌→HD→原貌。肖像內部 52,452 像素改變，雙側邊框另有 172,017；區外 0，切回完整相同。中文、黃色外框與下方資訊不變。
- 同流程繼續顯示第二段中文，依待續標記送空白鍵，完整顯示後離開國王房間。實際視窗目視核對，金額 1000，返回城鎮時沒有殘留中文字或手繪肖像。回城鎮的同狀態 HD 邊框切換亦通過。
- `gui-dialogue-hd-v4` 的國王切換已通過，但最後城鎮比較沿用已開啟的 HD，錯把它當作原貌基線。腳本改成每次比較先明確切原貌，v5 乾淨重跑通過。這是驗證腳本狀態問題，不是產品圖面還原失敗。

八項具名 Go 測試通過，無測試 skip；`bitmap` 套件沒有獨立測試。完整中文、其餘人物／場景 HD、存讀檔、戰鬥及實體手把仍未完成或未驗。

## 2026-10-18：長老中文選單與正常存讀檔

本節接續國王切片。沿用固定原版 SAR、dosgolem 快照與 Go 1.24.13；共用執行器未修改。手冊人工核對 `SCAN0357_007.jpg` 印刷第 10 頁、`SCAN0357_009.jpg` 第 13 頁：長老房屋的 Record Experience 存檔，F7 讀檔。掃描只留本地 `workplace/save-path-manual/`。

### 來源與顯示契約

| 項目 | 原始定位與證據 | 等級與範圍 |
|---|---|---|
| SAR 目錄 | 固定 ZELRES2.SAR 開頭 232 bytes 為 58 個 little-endian uint32 偏移；每項前四 bytes 為 payload 長度，逐項恰接下一項及檔尾 | 已證實，限此固定檔案 |
| 長老模組 | 目錄第 17 項 `0x1BCC2`，payload `0x1BCC6`、6973 bytes；正常第 253641450 指令 read 成功，對應段內 A000h | 已證實，正常玩家路徑 |
| 正常長老段 | `save-path-replay/sage-next.segment`，65536 bytes，SHA-256 `22076a51f2a0f3d10115840090b8331107efe79d31776ceea90775b7c7236a00` | dosgolem 01B1 段內偏移，283900000 指令、290628496 cycles；不是 IDA 的新輸入 |
| 長老對話來源 | SAR `0x1CA64..0x1CA83`、`0x1CA86..0x1CAAF`、`0x1CC43..0x1CC64`、`0x1CC66..0x1CCA0`、`0x1CE7E..0x1CEF1` | 來源 bytes 已證實；實跑涵蓋初訪、再次詢問、記錄旅程及保存成功，第一個詢問分支未走到 |
| 長老四列 | SAR `0x1CA2B`、`0x1CA36`、`0x1CA40`、`0x1CA51`；對應 `01B1:AD65`、AD70、AD7A、AD8B 的 ASCIIZ | 已證實，來源及正常四列顯示 |
| ASCIIZ 繪字 | 共用 `01B1:22DB`／完成 `2308`；`22E7` 起計算 X=BH×4+CL、Y=BL；`2381` 的 add di,5 | 已證實，限本次長老四列；原始 bytes／運算元見 IDA 匯出 |

完整來源映射與 read 紀錄在 `workplace/sage-source-mapping.json`。IDA 9.4 仍從固定國王段輸入 SHA-256 `06b74a1ea26f25d7e81de50ecb9f9ece1e1851d9429b469c59cf6a80bbb1d156` 重建，匯出 `workplace/ida-dialogue-indexed-v4/dialogue-evidence.json`，新增共用 22CD／2345 函式。其位址空間為 IDA binary base 0，對應原版 01B1 段內偏移。索引保留原始名稱、位址、bytes 與等級，不把長老段冒充 IDA 國王輸入。

`Dialogue` 改為逐段核對模組前 64 bytes 與完整來源文字，再處理來源位置；控制碼與原版動作不翻譯。`Labels` 觀察 ASCIIZ 完成畫格，僅覆繪已核對來源、原黑底白字、色號與色盤仍相符的文字行。長老四列 X=168，Y=38／48／58／68；中文 24×24 顯示像素，保留選擇箭頭。十八筆選單試譯目前只驗長老四筆，其餘商店仍待正常操作。

`labels-sage-v3/receipt.json` 在 283.9m、284.65m 抽樣。完整 1 MiB 記憶體、色號、色盤、CS:IP、cycles 與無覆繪原版相同；國王 241 加長老 114 字形、四列標籤，兩點差異各 29080 顯示像素，對話／標籤區外 0。目視初稿發現逗號可能落行首，依實際字模寬度保留句尾標點位置；重跑通過。正式字級與術語仍為 prototype。

### 新存檔與冷讀檔

- 獨立試玩正常新局領賞後走到 Muralla Town 長老房屋，完成初訪，再選 Record Experience、輸入 SAVETEST、Enter。只用 KeyDown／KeyUp／Run，不載來源附帶存檔或注入狀態。新檔 256 bytes，SHA-256 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`。
- 原版第 318353491 指令建立成功，318353503 close 成功，Wrote 記錄 SAVETEST.usr／256。正常 Ctrl+Q／Y 後 DOS 退出碼 0。Close 後重新 Load，F7／Y／輸名／Enter，138950398 指令讀取新檔 256 bytes 成功，回到 Muralla Town 長老門口及 1000 金幣，持鍵為空。
- `workplace/save-path-verified/audit.json` 記錄來源、命令、二進位、映像與 dosgolem 身份，29 個原始檔雜湊不變。早期探索沒有完成名稱輸入，未產生新檔；完成畫面觀察與輸名後成功，沒有據此開啟鍵盤 driver 逆向。
- [正常鍵序](tools/save-load.commands) 與 [配對測試](internal/prototype/overlay/save_load_test.go) 保存可重播入口。`workplace/save-overlay-v1/receipt.json` 有十九點完整原版狀態比較；兩側各自建立新檔，bytes 及 SHA-256 相同，各自退出、冷開機並讀回。保存成功中文完整顯示，名稱及存檔識別未被翻譯。這是目前列出的正常路徑，未外推損壞存檔或所有裝置狀態。
- `workplace/save-path-gui-v1/receipt.json` 以實際 X 鍵盤走相同玩家流程，新檔 XVIEW1.usr 256 bytes，SHA-256 `ac9c392c1dc1a0767e50edbd28aee396c1cfeedefaa77604cceafb193d7806f4`。正常退出碼 0；冷重開的新局先見金額 0，再 F7 讀回 Muralla Town／1000 金幣。新檔讀後不變、持鍵空、原始 29 檔不變。
- 目視 `record-selected.png`、`record-name-entered.png`、`record-saved.png`、`restored-window.png`，長老初訪、四列及保存訊息中文可讀。名稱輸入框、檔案列表、Yes/No 與 F7 提示仍為英文；本次存讀流程未開 HD，不延伸成 HD 驗收。
- 名稱輸入框切換讓舊對話畫格失效。GUI 保留 kind=2、glyphs=355、paragraph=6、visible=34、first_difference=148，見 `observe-before-restart.json`。顯示層安全移除舊中文，再由保存訊息的來源事件接回。配對測試也出現同類退回，完整原版狀態始終相同；此診斷屬既有 UI 切換邊界。

## 2026-10-18：存讀檔、退出與速度提示中文

本節只擴充顯示適配。原版資料、dosgolem 快照及遊戲規則不變。來源映射在 `workplace/ui-source-mapping.json`，入口仍為 [對話位址索引](docs/dialogue-source-index.json)。

| 來源檔案偏移 | dosgolem 01B1 段內偏移 | 用途與等級 |
|---|---|---|
| SAR1 `0xE3C8`、`0xE3CC` | `7513h`、`7517h` | 已證實：是／否，ASCIIZ |
| SAR1 `0xE663` | `77AEh` | 已證實：讀檔名稱提示，FF 結束 |
| SAR2 `0x1C1E2` | `A51Ch` | 已證實：長老存檔名稱提示，FF 結束 |
| STICK `0x60A` | `070Ah` | 已證實：退出兩行提示 |
| STICK `0x745` | `0845h` | 已證實：速度兩行提示 |
| STICK `0x883` | `0983h` | 已證實：讀取兩行提示 |

- SAR1 SHA-256 `56a73b59e952dab47fb0171a3fb1cdb7d641baceafa97e2826057f877a590eb0`；40 項目錄中 payload `0xCEB5`、7293 bytes 對應 `01B1:6000`。SAR2 SHA-256 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`，長老 payload `0x1BCC6` 對應 A000h。STICK SHA-256 `3255d2a0f1f3d5c4a603e61fa8bca57dcf2130a44d8c37ed7e1df63c6ea7f8a9`，檔案起點對應段內 0100h。這些是逐檔／逐模組映射。
- 正常存檔名稱提示段傾印為 `save-path-verified/name-before-keyup.segment`，65536 bytes，SHA-256 `7fdeaf5ba642d54686cfa6c49c5ac41f11477e03dc05746dcbfb3d477754c3a5`。原版繪字追蹤 `ui-font-probe/events.jsonl` SHA-256 `66041678775c2e039f6c7dbf37b4a88aaef25045ca1aa07604500adad2702a0f`。來源 bytes 及各模組前 64 bytes 逐項核對。
- 已證實：FF 多行繪字在段內 `291Ah` 保存 BX／CL，`2934h` 的 `lodsb` 讀來源，`2939h` 在 FF 後返回；`294Ah` 呼叫 `27E9h`，`2950h` 每字增加 BX 八，CR 在 `2955h` 增加 Y 八並還原行首 X。IDA 9.4 的 binary base 0 對應此段內偏移；原始位址、bytes 與運算元保留，不覆寫函式名。
- 最新 IDA 匯出 `ida-dialogue-indexed-v5/dialogue-evidence.json` 仍使用原 84m 國王段 SHA-256 `06b74a1ea26f25d7e81de50ecb9f9ece1e1851d9429b469c59cf6a80bbb1d156`，設定 16-bit 後分析。沒有把長老段或動態追蹤稱為同一份 IDA 輸入。
- 原版名稱提示首字在第 312365355 指令為 AX=0149h、BX=003Ch、CX=006Ch、SI=A51Dh，對應 X=60、Y=108。是／否在 X=204、Y=47／57；兩類使用不同座標契約。

### 速度輸入範圍勘誤

舊版以最長原文行寬核對整個兩行矩形。正常 F9 後輸入 3，第二行末尾的原版數字落入該矩形，導致整張中文提示失配，退回英文。`save-ui-modal-v1/` 保留實際視窗證據；這是顯示層範圍錯誤。

修正後逐行只核對及清除原有文字前綴。速度兩行的原生範圍分別為 `(116,82)..(212,90)` 與 `(116,90)..(204,98)`，右側數字完全保留原版像素。譯文亦限制在對應行寬。來源缺失、畫面或色盤改變仍安全退回原貌。

`save-ui-v3/receipt.json` 由正常 [玩家鍵序](tools/save-load.commands) 重生，22 點完整 1 MiB 記憶體、色號、色盤、CS:IP、steps、cycles 均與無掛鉤原版相同。每點覆繪區外差異 0。速度提示輸入 3 前後，中文各改變 4002 顯示像素，數字區未改；Enter 加 Right 後提示消除。該點畫格與讀回後一致，只證明清除，尚不單獨證明移動。

兩側各自建立 256-byte SAVETEST.usr，SHA-256 均為 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`，退出後冷重開並各自讀回。完整 GUI `save-ui-gui-v2/` 另建立 UIVIEW2、退出碼 0、冷讀回 Muralla Town／1000；使用修正前 `a678bbbb` 建置，不能冒充最新 `fb93dc06` 的全程 GUI 重跑。

最新建置的 `save-ui-modal-v2/` 實際 X 視窗補足繼續操作：F7 兩行中文完整、N 取消；F9／3 後中文與數字均保留，Enter 後 Right 持兩秒／放開，提示消失，角色由原生 x 約 136 移至 184。`movement-comparison.json` 顯示原位置紅像素 45→0、新位置 0→55，城堡靜態錨點不變。這只輔證位移，不稱為全域同狀態對拍。Enter 單獨沒有清除舊提示，不推定其他返回鍵契約。

仍有中文覆蓋限制：彈窗改變舊對話矩形時整段退回，彈窗下可能顯示英文。完整中文仍需原版清框／還原事件證據。檔名與 Re-Start 保留原值；未驗商店十四筆、其他片頭與遊戲文字不算完成。

## 2026-10-18：長老手繪肖像與武器店交易

### 長老 HD

| 項目 | 固定來源與結果 |
|---|---|
| 原版畫格 | `save-path-verified/sage-menu.indexed`、`save-confirm-text.indexed`，正常存讀路徑取得 |
| 肖像外框／內框 | 原生 `(56,23,96,64)`／`(57,24,94,62)` |
| 兩個肖像色號 SHA-256 | `bbc79b87bc9d9b3bb59a049d5acd8206a7b9dfb287d557b71437672544d3f16c`／`0762364b30958db2e1948f850d449ff042a8f736ea5b62d8674be9296c90a66a` |
| 動畫差異 | 相對 `(59,28)`、`(59,29)`、`(57,30)`、`(58,30)` 四點；語意未知 |
| 手繪圖 | `workplace/hd-sage-prototype.png`，1536×1024；SHA-256 `a1fa3b745422f59153239b903591db104c5c3fb320c18ee7f5181a247c33a5aa` |

來源及差異表在 `workplace/sage-hd-source-mapping.json`。內建 imagegen 依原版肖像與國王試作畫風生成，第一張頭部比例過大，經一次構圖修正選定第二張。完整提示詞組、來源、兩次產物路徑與權利分類在 `hd-sage-generation.json`；手繪圖仍屬本地衍生試作，不加入 Git 或發行包。

HD 直接合成到 288×192，保留黃色外框。四個原版變動像素各自的三倍顯示區保持原樣，不猜成嘴、眼或魔法，也不加動畫時鐘。色號與 RGB 均須完整匹配兩個已觀察畫格；缺圖、短畫格、色盤改變或未知畫格整組退回。

`gui-sage-hd-v1/` 從冷新局領賞並進入長老房屋，在原版暫停的同一狀態切原貌→HD→原貌。內部 50,926 像素不同、區外 0、動畫像素不變，切回完整還原；steps／cycles 不變。使用 `359bef32` 建置；後續商品覆繪加入後的 `5ad8556d` 不冒充此輪 GUI 身份，長老繪製器另由最終測試集驗證。

### 商品名的實際消費端

固定 SAR2 SHA-256 仍為 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`。payload `0x17817`、7239 bytes 對應原版 `01B1:A000`；正常武器店段 `shop-path-oracle-v1/weapon-first-dialogue.segment` SHA-256 `22c3e2d3f0f9beff133f5bd5737f3e662429aba5c74b198bb707fe8d33315817`。來源映射在 `weapon-source-mapping.json`。

| 原始定位 | 已證實語意與證據 |
|---|---|
| SAR `0x18534`／`0x18543` | ASCIIZ 兩個商品名，對應段內 AD1Dh／AD2Ch |
| A2B7h → CS:6012h → 7539h | 商品表與索引送入列準備／搬移服務 |
| 7539h → CS:301Ah → 3805h → 383Eh | 讀取名稱指標；四像素墨跡、五像素步距、八列字模烘到 3D4Fh，stride=160 |
| 384Ah | 名稱 NUL 後返回，名稱暫存像素完成 |
| 7539h → CS:301Ch → 37CCh | 從 3CAFh 搬九列到 VGA，X=BH×4、Y=BL；名稱位於 Y+1 |
| 3804h | 完整搬移後返回；正常兩列 BX=1871h／187Bh，名稱 X=96、Y=114／124 |

位址基準為 IDA binary base 0，對應此固定段內偏移。索引 [docs/weapon-source-index.json](docs/weapon-source-index.json) 與 [匯出工具](tools/ida_weapon_evidence.py) 保留原始函式名、位址、bytes、運算元及等級。`ida-weapon-indexed-v1/weapon-evidence.json` 為 IDA 9.4／16-bit 的六個必要函式匯出，自動合併索引，未證實指令附 warning。

初次沿用 ASCIIZ 掛鉤時商品名未顯示，`test-weapon-v1.jsonl` 保留失敗。正常清單階段的 22DB／22CD 入口均為零事件；2345 的 108 筆、24F2 的 11 筆只屬前段已知 UI，沒有商品名。這些有效排除結果保存在 `shop-path-*trace*`，未當成工具故障。從武器店實際間接呼叫表追到 7539h，再追出暫存像素搬移鏈後停止。

`shop-path-buffer-trace-v1/buffer-trace.jsonl` 有八筆，383E／384A／37CC／3804 各兩筆。來源 SI=AD1Dh／AD2Ch、準備返回 3822h、搬移返回 7565h，原版抽樣不變。適配層只記錄名稱來源與完成的暫存名稱像素，搬移前再核對一次，完成後擷取 VGA 名稱前綴。價格區與原版資料均不改寫。

### 正常交易與剩餘中文

- [固定鍵序](tools/weapon-shop.commands) 從冷新局領賞，正常走入武器店，選第一項、確認購買、告別、Space 回城鎮並查看裝備。原版售價 400、折抵 200，GOLD1000→800；未載來源附帶存檔或注入狀態。
- `weapon-overlay-v2/receipt.json` 九點完整 1 MiB 記憶體、色號、色盤、CS:IP、steps、cycles 相同，文字區外差異 0。初訪、五列選單、兩個商品名、交易後與告別已接中文；六個靜態對話試譯中，未走過的詢問分支仍未驗。
- `shop-path-gui-cht-v1/` 以固定 `5ad8556d` 建置及 13／27 筆 catalog 完成實際 X 交易與裝備查看。原價 400／1500、報價 400／200 仍可見；買後 800、中文告別及 Training Sword 裝備頁均目視，最後持鍵空、退出 0、原版 29 檔不變。
- 商品清單上方詢問、報價模態下部分舊選單仍會退回英文。GUI 保留 kind=2 兩筆，paragraph=10／12、first_difference=2552／242；動態報價、場景名稱與裝備頁仍英文，不能宣稱武器店完整中文。
- 動態報價追蹤 `shop-path-glyph-v1/glyphs.jsonl` 共 225 字形，保存來源指標、SI、AX／BX／CX 與數字的實際順序。16 點色號／色盤／64 KiB 段、cycles、IP 與未掛鉤原版相同，不稱為全域記憶體對拍。後續中文須逐段保留數字與模態遮蔽，不能清掉價格來取得完整中文畫面。

## 2026-10-18：商店彈窗與舊中文的繪製範圍

本節補正前輪整段退回的限制。輸入仍為 `shop-path-oracle-v1/weapon-first-dialogue.segment`，SHA-256 `22c3e2d3f0f9beff133f5bd5737f3e662429aba5c74b198bb707fe8d33315817`。IDA 9.4／16-bit 以 binary base 0 分析，以下位址對應 dosgolem `01B1` 段內偏移，不是 EXE 檔案偏移。

| 原始定位 | 等級與已驗語意 |
|---|---|
| 2046h | 已證實：X=BH×4、Y=BL、寬=CH×4、高=CL。AL=0 清除黑底，非零清底並畫框 |
| 20B6h／20E7h | 已證實：邊框橫列；返回也供內部呼叫，不能單以 20E7h 當整個視窗完成 |
| 20E8h..2105h | 已證實：VGA 黑底填色；畫框內部也呼叫此入口 |
| A28Ch → CS:2000 | 已證實：商品框 `(84,110,148,36)`，返回 A291h |
| A3F1h → CS:2000 | 已證實：是／否框 `(188,43,48,25)`，返回 A3F6h |

完整 `2046h..2105h` 的 192 bytes SHA-256 為 `3e91f3bc6ad5a812f3b4c89c49aa3c8e146f8cef714fe0cade0009e4303f72c4`。首次探索輸出為 `modal-evidence-v1/evidence.json`；分級索引已回填 [武器店索引](docs/weapon-source-index.json)，持久工具重生的九個函式在 `ida-modal-indexed-v1/weapon-evidence.json`。未分級指令仍附 warning。

原版正常追蹤入口為 `workplace/shop-path-modal-trace-v1/evidence-index.json`。冷新局購買 9 點、獨立冷新局選 No 取消 7 點，完整 1 MiB 記憶體、色號、色盤、64 KiB 段、steps／cycles／IP、持鍵及輸入讀取均與同探針未啟用掛鉤的執行相同。26 次矩形操作都以入口 SS:SP、原始返回位址及返回時 SP+2 配對，改變像素全部在傳入矩形內，區外 0。取消後回原版五列選單及 1000 金幣。

覆繪依文字完成與畫框事件先後保存版本。較晚的框遮住舊中文字及其清底，框外每個像素仍須等於原版參考。解除遮擋須整個框範圍回到入口前色號，單一標籤相同不足以判定已還原。未知程式、越界參數與歷史淘汰仍安全退回；最多保留 64 事件，不猜測保存 buffer 或改寫原版。詳見 [顯示試作規格](docs/spec/001-runtime-overlay.md)。

`shop-path-gui-modal-v1/` 使用 `c355ade582ce540ccf98deeb77924e09b8981939dbb319ae422b3abb96eecd98` 建置。冷新局領賞、進店、商品框上方中文、報價框旁可見舊選單中文、購買 1000→800、中文告別及裝備查看均已逐張目視。是／否框的 144×75 顯示像素與前輪相同，差異 0。原版 29 檔未變、持鍵空、退出 0。

限制仍保留：是／否框邊界會遮掉部分舊中文字，沒有重新安排原版視窗。動態報價與裝備仍英文；離店換場保留一筆 paragraph=12、kind=2、first_difference=422 的舊對話退回，不稱為零失配或武器店完整中文。

## 2026-10-18：武器報價中文與資金不足

本節補正前輪動態報價仍英文的限制。原版來源入口為 `workplace/shop-path-quote-source-v1/evidence-index.json`，完整片段、控制 bytes、字形座標及數字像素在同目錄 `source-evidence.json`。使用既有225字形追蹤，不重跑已證實商品入口。

| 證據 | 來源身份與範圍 |
|---|---|
| 固定武器店初訪段 | `weapon-first-dialogue.segment`，SHA-256 `22c3e2d3f0f9beff133f5bd5737f3e662429aba5c74b198bb707fe8d33315817` |
| 報價完成段 | `shop-path-glyph-v1/training-sword-confirm.segment`，SHA-256 `1ed4dd99c85c8dc22037664fddc1a7026e345e1136e26455b3688f720d138ae7` |
| 字形追蹤 | `shop-path-glyph-v1/glyphs.jsonl`，225筆；SHA-256 `db4b5b82724ef14f98ccc90ce8ce00f763d6f3d635858f8b0fa9fe128bae4f9d` |
| 位址空間 | IDA binary base 0，對應各固定快照的 dosgolem `01B1` 段內偏移；不是 EXE 檔案偏移 |
| 最新 IDA 匯出 | `ida-quote-indexed-v1/weapon-evidence.json`，IDA9.4／16-bit，十一個必要函式，合併 [分級索引](docs/weapon-source-index.json) |

已證實的報價順序是 B0DDh 前綴、商品 ASCIIZ、B0E7h 問號、B106h 售價前綴、BC33h 售價、B115h 單位、B046h 折抵前綴、BC33h 折抵、B056h 後綴及 B0EDh 確認句。原版 A35Ah 格式化價格到 BC33h，A364h 設定文字指標；A3C9h 再使用同一區輸出折抵。文字來源、結束碼、字形次序與正常呼叫鏈相符，等級為已證實。沒有重新實作格式化、資金判定或價格計算。

BC33h 的數字以單一 FF 結尾。最後快照的200後為 `FF 02`，不能假定 `FF 00` 或 NUL。原版400的cell範圍為 `(137,109,21,8)`、200為 `(138,119,21,8)`，墨跡分別在 inclusive `(138,109)..(157,115)`、`(139,119)..(158,125)`，色號只含0／9。覆繪逐次在7120h完成事件保存已畫出的整組八列圖形，右端保留一欄安全黑底。售價與折抵分開保存，避免暫存區重用後只剩200。

中文模板引用同一份商品譯名，以原版數字圖形在顯示解析度合成。數字整組搬移及換行，不重畫字形、不根據資金推算。四個片段各等原版對應內容完成才顯示；未完成時暫留白。原始數字位置會隨中文排版改變，不能宣稱報價區像素完全不變；可宣稱經抽樣驗證的數字圖形相同及完整原版狀態不變。

第二件1500武器在1000金幣的新局直接走不足分支，不先出1500／200報價或是／否框。原始段 AF54h..AFAFh 對應 SAR `0x1876B..0x187C6`，88個可見字形，包含銀行提示及省略號。獨立正常探針 `shop-path-insufficient-v1/` 驗兩頁、選單操作、告別及返回 Muralla Town，金幣1000、持鍵空、無寫入。只有共用前段三點與舊基線比對，未把此原版探索宣稱為完整1MiB對拍。

正式覆繪配對由 [購買鍵序](tools/weapon-shop.commands) 與 [不足鍵序](tools/weapon-insufficient.commands) 各自冷新局重生。`quote-overlay-v2/receipt.json` 的購買9點，以及同目錄 `insufficient/receipt.json` 的不足11點，完整1MiB記憶體、色號、色盤、CS:IP、steps、cycles均與無掛鉤原版相同，文字區外及彈窗越寫皆0。確認報價400／200字形完整、買後800；不足翻頁後仍1000並回城鎮。所有輸入仍為正常原版掃描碼。

`quote-overlay-v2/test.jsonl` 的三個頂層測試與五個代入欄位／字型子測試通過。合成987／65圖形只證明數字圖片不被硬寫成400／200及未知畫面退回，不當作這兩種價格的原版行為證據。最早 `quote-overlay-v1/` 在舊測試引用已清除商品框的版本時失敗，改為核對目前報價繪製者的版本後重跑；沒有放寬不可越寫彈窗的要求。

此切片只驗第一筆購買與第二件資金不足。未驗盾牌、其他商品、無舊武器、出售及所有價格範圍；裝備與場景名稱仍待中文，原版確認框仍會遮住部分舊中文字。正式字級及出字節奏仍為DRAFT。

實際X視窗 `workplace/shop-path-gui-quote-v3/` 固定使用 `55f458acb28e01354ed95e946f26dab5749a5edec8d63f2060b23d70debca5ce` 建置。正常冷新局領賞、完整中文報價、購買1000→800、告別、回城鎮及Return查看Training Sword均通過。最後持鍵空，既有capture-tail流程完成擷取後正常終止前端，退出碼0；不是原版Ctrl+Q驗收，也沒有把不足分支離線圖冒充這次GUI路徑。

`number-bitmap-comparison.json` 以報價GUI圖核對原版21×8數字格。400從原生 `(137,109)` 搬到 `(141,101)`，200從 `(138,119)` 搬到 `(56,113)`，含黑底的兩格RGB差異均0。座標是扣除工具列後的原生遊戲座標；全視窗仍有中文與自然動畫差異，不稱完整同狀態GUI對拍。

較早GUI v1未在舊入口預算內進城，v2因觀測檔缺少心跳而中止。Ebitengine的Draw可能跨過多個Update，不能用幀數是否為60整倍數判斷進度。觀測改為每次Draw寫完整快照，再以同目錄換名原子更新；未更動原版輸入或cycles。兩次失敗原樣保留，沒有據此推論CPU停滯或報價缺陷。

本次GUI未開HD，最後保留一筆離店舊對話退回，paragraph=12／kind=2／first_difference=422。完整報價當下沒有退回，不能據此宣稱全程零失配。
