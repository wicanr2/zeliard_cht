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

## 2026-10-18：裝備頁彩色標題與兩行名稱

來源入口為 `workplace/inventory-source-v1/mapping.json`，SHA-256 `7958c33a466a5993aff8307f3aa5a3042f9132efb9ccff8395d3807a9877071c`。固定IDA輸入為 `shop-path-oracle-v1/equipment-held.segment`，65536 bytes，SHA-256 `2ecedadb544f1d089612ca19b1f42379fa83866c447121b131c61435fcdf7e3e`。SAR2 payload檔案偏移 `0x401E`、3613 bytes載入原版 `01B1:A000`；前綴與七個完整ASCIIZ來源均相符。

| SAR2檔案偏移 | dosgolem段內偏移 | 原文及已觀察路徑 |
|---|---|---|
| 0x4A1D | 01B1:A9FF | SELECT-MAGIC:，X=52、Y=18，AH=2 |
| 0x4A2E | 01B1:AA10 | WEAR:，X=52、Y=67，AH=3 |
| 0x4A37 | 01B1:AA19 | USE:，X=52、Y=113，AH=3 |
| 0x4A3F | 01B1:AA21 | INVENTORY，X=184、Y=67，AH=3 |
| 0x4AB0 | 01B1:AA92 | NOTHING，三次，AH=1 |
| 0x4D03 | 01B1:ACE5 | Training，五像素小字第一行 |
| 0x4D0C | 01B1:ACEE | 五個空格與Sword，五像素小字第二行 |

`inventory-source-v1/label-trace.jsonl`六筆既有入口事件確認物品兩行及後續Esc暫停；名稱不是單一Training Sword ASCIIZ。另一次正常冷新局 `inventory-title-trace-v1/` 的340筆事件確認AA2Bh逐字、AA30h NUL完成及AA4Ah字形呼叫。這些探索收據沒有宣稱完整1MiB配對。

IDA9.4／16-bit／binary base 0對應上述固定 `01B1` 段內偏移，不是EXE檔案偏移。AA2Bh保存AH後經CS:2022呼叫27E9h字形入口，24EAh表為0／9／18／27／36／45／54／63。AH=1白字9、AH=2紅字18、AH=3綠字27；分類標題另外以AH=5、色號45在右下各一像素畫陰影，X每字增加8。來源、正常呼叫與bytes相符，等級為已證實。五個必要函式重生於 `inventory-ida-indexed-v1/evidence.json`，由 [裝備來源索引](docs/inventory-source-index.json) 合併證據等級與未知警示，保留原名與原始定位。

中文試作只接四標題、三個「無」與練習劍兩行。原版物品識別、持有數、裝備效果與按鍵完全不改；其他物品、魔法、道具使用及裝備切換尚未驗。

[正常裝備鍵序](tools/inventory.commands)由冷新局領賞、交易、出店，再Return開啟、放開、Return關閉及重開。`workplace/inventory-overlay-v3/inventory/receipt.json`十一點完整1MiB、色號、色盤、CS:IP、steps及cycles均相同，區外與彈窗越寫0。初開、放開及重開的九個新文字區各17359像素改變，關閉整張覆繪差0；原版金幣800。每個被遮擋舊標籤獨立繪製時差0，未知可見像素會撤除該譯文。

最早v1把隱藏舊標籤計入目前數量，v2以內部記錄數判斷關閉殘留，兩次測試失敗保留。v3核對七鍵九次出現，逐一證明舊標籤沒有可見寫入，關閉改以整張原版畫格的實際像素差驗證。這些檢查不放寬原版狀態一致或區外不可改寫要求。

共用標籤儲存與繪製另由 `inventory-regression-v1/test.jsonl`重驗長老及存讀檔。兩項通過，長老2點、存讀檔22點完整原版狀態相同且區外0；各自新存檔256 bytes、SHA-256 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`相同，兩次冷啟動及新檔讀回均確認。沒有把前輪存讀檔收據改名成新結果。

初次裝備GUI `shop-path-gui-inventory-v1/failure.json`在完整中文告別停留，按放已送達，90秒畫面預算結束；尚未到裝備。`farewell-readiness-review.json`證實八列ROI已相同時，完整十一列仍差4像素，後續完整中文圖才差0。原文29個可見字形的比例揭露在27字時即可顯示完整10字中文；完整中文不是原版待續階段證據。

末尾等待的IDA9.4輸入仍為固定武器店段 `22c3e2d3…`，位址基準同本節前述規則。71DFh先清FF1D／FF1E，71E9h呼叫7042h輪詢，71ECh與71EFh合併旗標，71F3h為零時繼續，7204h返回。這些控制流已證實並回填[武器店索引](docs/weapon-source-index.json)，十三函式匯出在 `inventory-wait-indexed-v1/evidence.json`。072Ah位於共用服務，單一IP不能證明文字就緒；沒有深入其硬體或ISR。

正常冷新局 `inventory-source-v1/farewell.segment` SHA-256 `08ec86394a4a819caf06f21047406bc844a3894257c88afa5239de26a0be2472`，末尾文字游標FF4C=B1FDh。失敗GUI送鍵前IP06B2h，按下ack為A91Eh，放開ack為71ECh；舊觀測沒有文字游標，按鍵是否在末尾清旗標前到達仍是強推論。新的唯讀觀測提供文字游標及旗標，腳本同時要求完整中文、B1FDh、旗標零及不同虛擬步數的兩個穩定快照，再送正常Space。沒有據此改寫原版等待或輸入規則。

`shop-path-gui-inventory-v3/farewell-ready.json`兩個快照均為游標45565、旗標0，步數分別203,149,613及203,799,613，後續Space已回城鎮。Return按住期間的送達快照由209,493,344至209,593,344，只涵蓋約十萬指令；沒有開啟裝備。這支持以原版步數取代主機秒數控制驗證輸入，未證明裝備鍵的最低時間。v4末尾改500k指令，但前段牆鐘微移已走過國王城門，沒有抵達本次目標；不能拿Up送達事件證明進房。

替代正常入口為先前長老新建的 `save-path-verified/scratch/SAVETEST.usr`，256 bytes，SHA-256 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`。`shop-path-inventory-save-check-v1/`由既有原版探針正常冷啟動、F7／Y／輸名／Enter載入，回到Muralla Town、1000金幣，再Return開／關／重開。原版圖及段內ACE5h／ACEEh核對初始武器同為Training與Sword，沒有因圖示猜名稱。來源存檔不變、沒有寫檔。此原版核對只建立後續GUI入口，不冒充覆繪配對或實際視窗通過。

實際X視窗 `shop-path-gui-inventory-v5/receipt.json` 使用固定 `52826ff520b8b5af184e02dbe3a76188d54f9a4f6d610766d12a623d16cb213a` 建置，腳本SHA-256 `6c75ffe8319f0734c59ad21a9eed2185ac4fd44e2909703b262d7757f6c30a49`。正常冷F7讀上述新存檔，初開與重開的九個中文ROI匹配既有覆繪參考，關閉後城鎮靜態區匹配本次讀回圖。三張圖逐張目視清楚、沒有關閉殘留；金幣1000、存檔不變、最後持鍵空，capture-tail正常終止碼0。此GUI沒有新交易、未開HD，只證明文字區及靜態區，不稱完整同狀態GUI對拍。

原版配對的交易與800金幣、GUI冷讀檔的1000金幣是不同正常路徑，不混用。GUI按放由X11送入並觀察原版確認，三次裝備Return各至少500k原版指令；放開後分別前進1m／2m／1m。事件內固定source字串只適用裝備Return，F7與名稱按鍵實際來源為 [存讀鍵序](tools/save-load.commands)及本輪原版save-check。沒有從主機秒數推算按住長度，也沒有宣稱500k是最低需求。

## 2026-10-18：武器商人兩姿態HD

來源入口為 `workplace/merchant-hd-v1/source-mapping.json`。原始正常交易六畫格在 `inventory-source-v1/`，迎客至交易後肖像相同，告別畫格不同。以Python3.13讀取原版320×200色號，外框 `(56,23,96,64)`兩組雜湊為 `77dcc8e94e5125990656f7b33cc04ed7fa9a9719c20487b37424e3538a8a6baf`及 `bc69c9d588c150a9ecc491afba4c064b327d2b8306eb0dd6f8eecaaedb52d4d8`。原版色盤SHA-256 `6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c`；輸入完整畫格及RGB雜湊均保存在映射內。

兩個原版畫格有1062個色號不同，差異包圍框原生 `[65,39,143,85)`。告別圖已目視為交叉雙臂，等級為已證實的畫面觀察；尚未證明全部動畫或其他城鎮。不能將長老四個動畫像素的保留策略套用至此姿態變化。

內建imagegen沿既有迎客手繪圖產生交叉雙臂版，兩張1536×1024圖的SHA-256為 `2660fc7892f9a5229d4cee3723e992146992a460a1678197793cc890b0dfb59b`及 `90243f65aef8532303b12dbbc3e1ddd9bd06c03eb58e14548a2d15419892946f`。完整提示詞、參考角色與風格、輸入身份及本地權利分類在 `generation.json`與 `farewell-generation.json`，原版衍生素材不入Git。已目視同一禿頭紅鬍商人、藍背心及綠臂帶，未把畫風試作稱為原版像素一致。

原版完整外框色號與RGB均匹配時，HD直接合成到內框282×186顯示圖面，黃色邊線及框外資訊不變。兩姿態由原版畫格選擇，不新增動畫時鐘或輸入規則；缺任一圖、錯誤來源、色盤淡入或未知畫格時保留原貌。候選身份在 `candidate-identity.json`，測試 `hd-test.jsonl`通過來源、兩姿態、區外0、原版輸入不改動與原子退回；此時尚待正常玩家路徑GUI。

### 未購物分支與第三姿態補正

首輪GUI `merchant-hd-gui-v1/`驗迎客成功，但預設Go outside未購物離店進入另一原文與姿態，沒有進入已購物的Thank you分支。先前兩組畫格只來自冷新局第一筆交易，不能外推所有離店。本輪保留exit1與全部輸入，未外部介入或調整執行中的程序。

正常冷F7讀回先前新建SAVETEST，再往西進武器店的來源入口為 `merchant-path-v1/receipt.json`。未購物分支在同目錄 `no-purchase/source-evidence.json`，原版段SHA-256 `33124fa525ff59ed545ad63a3c64deba28a64943608bfff17fb73a347ead237d`。SAR2檔案 `[0x18A16,0x18A57)`與 dosgolem `[01B1:B1FF,01B1:B240)`共65 bytes相同，前導0C後正文從B200開始，末尾B23D為11、游標B23E。原文反斜線在畫面呈撇號，尾端含斜線及FF控制碼；只保存其bytes與實際結果，不推測控制碼內部語意。

新增試譯鍵 `zelres2.file.018a17`，沿既有原版繪字觀察，保留原始文字、比較值與控制碼。對話TSV為15個靜態段及4個模板，共19筆。[未購物正常鍵序](tools/weapon-shop-no-purchase.commands)從冷新局領賞、進店、未購物離店，三點完整1MiB、色號、色盤、CS:IP、steps與cycles相同，區外0。告別中文11846像素改變，離店整張覆繪差0、金幣1000；保留一筆paragraph14／kind2／first_difference1換場安全退回。

第三姿態外框色號SHA-256 `cc24c5722e583f61aa229b2c3a18cba1fe365c170da7a42a8655ed761859a8e6`，手繪圖SHA-256 `48f567a000f2d6c413f8090b72417763c7ff96a7a7a24b1dea059dc5fb2a0ec8`。完整內建imagegen提示詞與參考在 `merchant-hd-v1/no-purchase-generation.json`，三姿態映射在 `source-mapping-v2.json`。三組畫格的色號差異為1062／874／1341，RGB差異為1060／865／1325；少數色號對應相同RGB，不能把兩種比較值混用，也沒有因此推論新動畫變體。

### 三姿態實際視窗與獨立像素核對

`merchant-hd-gui-v2/receipt.json`固定建置SHA-256 `48e9c0bc900ce9e4b99f4e2feb1a0f752b30772d7b31ddeb62f1b9fc532fb970`，腳本SHA-256 `810ea7e7a0f8127219e00cbb54edcb75ba82ffabf5c177d67e65aa98cc0fdbbe`。正常冷F7載入自己新建存檔，進店未購物離店1000，再原地Up重進店買第一件武器，購後800並離店。粗段後實際門口已在角色錨點附近，沒有補短步；不冒稱與原版探索指令序列完全相同或整段同狀態對拍。

迎客、未購物告別與購物後告別三組暫停原貌→HD→原貌，內框依序52409／52396／52419像素改變，遊戲畫布內框外0差、原貌完整還原，steps／cycles不變。獨立ImageMagick6.9.11-60分析 `merchant-hd-v1/pixel-verification-v2.json`另確認原貌肖像與原版RGB來源逐像素相同，HD與指定圖面的282×186取樣相同。比較空間為實際960×680視窗中的遊戲畫布 `(0,48,960,600)`，肖像內框 `(171,120,282,186)`；工具列按鈕狀態不混入HD差異。

兩個原版末尾gate分別驗B23E與B1FD、旗標0及不同步數的兩個快照，再送Space。最後HD保持開啟，回城鎮商人不匹配、pose=-1、金幣800、持鍵空，capture-tail退出0，未另驗原版Ctrl+Q。84筆按放全部確認，最長0.19秒；原版29檔、來源及兩份新存檔複本、字型均未變。

本次GUI保留paragraph14／kind2／first_difference290及paragraph12／kind2／first_difference422兩筆舊對話安全退回，與配對中的1分開記錄，不稱零失配。第二訪問報價也觀察交叉雙臂，原版姿態選擇的原因未追查；HD依完整畫格匹配，沒有依流程名稱硬綁。此收據只涵蓋三個已觀察姿態與正常兩種離店，不等於完整商人動畫、所有商店、完整HD或完整中文。

## 2026-10-18：銀行文字、數字與正常交易

本節的「已證實」限固定原版、自己正常產生的SAVETEST與列明鍵序。歷史原版入口保存在 `workplace/bank-path-v1/receipt-v1.json`，SHA-256 `381dd00d70d32f09c86bc2f7e471842556fa0b2e32c87aa006234c2e7aca461c`；歷史來源映射為同目錄 `source-evidence-v1.json`，SHA-256 `5a7ebc5b976d530149fca37cb7fc2b1652946e99f0e33b9d19f182064974e341`。工具為固定882檔dosgolem快照、Go1.24.13及Docker。正常原版探索程式與唯讀觀察器的來源、二進位、鍵序與版本身份各保存在收據，沒有將兩者混用。

原始資料是 `ZELRES2.SAR`，SHA-256 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`。銀行payload檔案偏移 `0x19462`、長3384，載入dosgolem段 `01B1:A000`。正常段傾印SHA-256 `dbaefe4b9c9dee269a544e67f9df9804bdc8206a9aa9411f2a7049e587ef4f61`。檔案 `[0x19D50,0x1A180)`與執行期 `[01B1:A8EE,01B1:AD1E)`全部bytes相同。以下位址全部明列空間，不與IDA線性位址混用。

| SAR2檔案偏移 | dosgolem執行期段位址 | 已證實的顯示用途 |
|---|---|---|
| `0x19D50`／`0x19D54` | `01B1:A8EE`／`01B1:A8F2` | 銀行標題header與八字元正文 |
| `0x19D6E`至`0x19DA5`的五個NUL來源 | `01B1:A90C`、`A917`、`A926`、`A934`、`A943` | 離開、兌換、存錢、領錢、查詢選單 |
| `0x19DB3`、`0x19DC2`、`0x19DCF`、`0x19DDE` | `01B1:A951`、`A960`、`A96D`、`A97C` | 攜帶金幣、存入金額、銀行存款、領出金額 |
| `0x19DEF`、`0x19E15` | `01B1:A98D`、`A9B3` | 迎客及沒有alma寶石 |
| `0x19F2D`、`0x19FE3` | `01B1:AACB`、`AB81` | 存入及領出問句 |
| `0x19F57`、`0x19F69` | `01B1:AAF5`、`AB07` | 餘額前綴與數字後接字串 |
| `0x1A006`、`0x1A01B` | `01B1:ABA4`、`ABB9` | 領出結果前綴與數字後接字串 |
| `0x1A05A`、`0x1A073`、`0x1A07E` | `01B1:ABF8`、`AC11`、`AC1C` | 空帳戶查詢、金額查詢前綴與後接字串 |
| `0x1A100` | `01B1:AC9E` | 本路徑離店告別 |

`glyph/glyphs.jsonl`為唯讀追蹤，`glyph/ordered-runs.json`保留順序及座標。字形消費與完成仍為 `01B1:711B`／`7120`，來源游標為 `01B1:FF4C−1`。銀行重用數字buffer `01B1:AD30`，實際見100及50，各以單一FF結束。100的X為162／168／175，50依語境為182／189、163／170或119／126，不能假定固定字距。後接的`&`確實進711B、AX=0126，不能從來源順序刪掉。存入結果保留存入問句再接餘額；領出結果保留領出問句，再接給款及餘額兩筆，必須分別保存完成時的數字圖形，不能只讀最後的暫存值。

四個金額標籤由 `01B1:22DB`／`2308`繪製，X=180、Y=33或63，前景色號9、背景0。銀行標題另由 `01B1:22CD`／`2344`確認，入口SI=A8EE、返回A8FA，header為 `18 AF 02 08`，位置 `(96,175)`、八個五像素字元、前景9、背景45。完整 `[01B1:22CD,01B1:2345)`為120 bytes，SHA-256 `f7415a3888eb9e7bafd8d68b76f3e281a7fe4a6c5b15275e5536d40a7594fef4`。標題窄追蹤入口為 `title/title-evidence.json`，沒有將此銀行來源擴張為所有場景的通用契約。

中文手冊來源為 `珍084-神劍封魔.rar`內 `珍084-神劍封魔/SCAN0357_007.jpg`，印刷第9–10頁，掃描SHA-256 `a69d7e6a6cc918a5984c3e80027844a1d8d647e18050140a5669d3db72dfceb7`，已人工核對原圖。選單沿用「離開、存錢、領錢、查詢儲存金額」；兌換項縮為「以alma寶石換現金」以容納原版安全寬度。手冊只說上下增減一單位、左右增減十單位，實際方向由正常原版確認：上+1、下−1、左+10、右−10；下鍵在0、右鍵從1均停在0。已驗Space確認，未驗Enter。原始金額輸入、帳戶及存檔識別均不翻譯。

正常原版路徑：冷F7載入自建存檔→銀行→查詢0→存100→領50→查詢50→無alma寶石→正常離店。身上金幣1000→900→950。唯讀重播35點的色號、色盤、64KiB段、FileOps與完整收據相同；原件29檔及SAVETEST未變，最後持鍵空。此原版觀察器收據不稱完整1MiB對拍。銀行ledger的新存檔持久性、餘額1、提空、資金不足、有寶石兌換、其他城鎮及銀行肖像HD未驗，不由本切片推論完成。

### 銀行標題勘誤：底色、水平陰影與X餘數

前節依 `2CBE=45`寫成背景45、位置96的推論撤回。`title-v2/`保留120-byte函式、來源前綴、header、模式、DS／SI及完成當拍色號。它證明所有入口條件相符，但完成畫格含色號5，因此原先僅接受9／45的完成檢查拒絕標題。這與第一個函式雜湊長度誤植是兩個原因，不能把修正長度當成整個問題已解決。

新的獨立分析入口為 `workplace/bank-overlay-v1/bank-title-style-analysis.json`。原始完成畫格SHA-256 `a4b275b6467e5e7180ea9e289b6ac2569c137f2d54e4cd237c99093557b4cf79`，原生 `(98,175,40,8)`內底色5有174像素、字色9有85、陰影45有61。全部61個45的左鄰都是9，證實它是右移一像素的水平陰影。銀行就緒色盤中三者RGB依序為 `00007d`、`fbfbfb`、`0000fb`。完成當拍與銀行迎客的該區色號逐byte相同，不能把45當成整片底色。

header `18 AF 02 08`第三byte亦加到目標DI，因此X為 `0x18×4+0x02=98`，Y=175；前節只取第一byte乘4，漏了餘數。最新prototype只在此已知銀行來源接受5／9／45，保留深藍底、白字及亮藍水平陰影，不改其他標籤的底色或裝備頁右下陰影。原始錯誤收據與v2／v3測試失敗身份保留，不包裝成成功證據。

`title-v3/palette-geometry-evidence.json`另保存入口、完成與迎客的完整色盤，三拍SHA-256均為 `6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c`，排除這三拍的換色盤假說。現行附勘誤入口為 `receipt.json`、SHA-256 `fc1c3deaf5e1476b8d14eadef85053f326d724ee95bbbc0bfd5d6cb2fbfbea4a`；`source-evidence.json`、SHA-256 `62a5d5fd66853bc043491461c8063b76afe41a3a625a4d0ce07d874b4a1655a5`。兩份歷史原文已逐字保存，沒有改寫原始觀察事件。

### 銀行覆繪配對與實際X視窗

固定候選SHA-256 `fc7972373703d0124182e2f24331bf3cdba9913ff2364fdb239f903db816837a`，來源身份在 `workplace/bank-overlay-v1/candidate-identity-v4.json`，主入口已同步。對話TSV28筆、標籤39筆；銀行新增九個對話／模板及五個金額欄／標題鍵，五列既有選單依手冊校訂。v4配對35點的完整1MiB、原版色號、色盤、CS:IP、steps及cycles相同，逐點文字區外0差，原版數字圖形逐像素保留。存入問句接餘額100，領出問句接給款50及餘額50；離店950、持鍵空、整張覆繪差0。完整收據為 `v4/bank/receipt.json`，SHA-256 `887be426803a4cc43a4205caba0493797f17dcfc0c888621b9f33fb45685f69f`。

正常X視窗收據為 `workplace/bank-gui-v1/receipt.json`，SHA-256 `27ef801bc04ed83f84f2355784c46ba27116d596e66c6a927f3d1949da7992ca`；逐圖目視及數字區域比較為同目錄 `visual-review.json`，SHA-256 `2705aa98f2ae5fb1f86d733e8c3701e2ba3532c07dcccb9ed0ae2176eb02e4bc`。固定腳本SHA-256 `a10aaa13b4eeea9ac784cf3bdc430c6e09d03a0b11d3f18865e173d988a23075`，工具身份在 `bank-gui-v1-tools/fixed-identity.json`。同一handle、原600秒預算完成冷F7→銀行→查0→存100→領50→查50→無alma→告別→城鎮950，沒有重啟或修改live參數。每次送Space前核對實際來源、原版游標／旗標及BankState；金額框FF4C轉到共用AA48、旗標255，不能誤當仍停在問句來源，也沒有改guest旗標。

64次正常按鍵、128筆按放全部確認，最長0.34秒，每次持鍵達指定原版步數。8個金額欄與3個GOLD區域同原版RGB；這是明列區域比較，不稱整段GUI同狀態對拍。16張實際PNG已逐圖目視，沒有截字，問句與後接結果完整，最後沒有銀行文字殘留。離店checkpoint563.981秒；最終312095974steps、364905164cycles、銀行覆繪inactive／completed10、無對話失配、持鍵空，capture-tail正常退出0，未另驗原版Ctrl+Q。

原29檔、自建存檔來源及兩複本、字型、候選、TSV與參考圖面均不變，容器已回收，擁有權異常0。GUI未開HD；銀行肖像、ledger新存檔恢復、有alma兌換與其餘文字分支仍未驗。程式原始識別值、場景與其餘HUD尚未翻譯，不把此切片稱為完整中文化。

## 2026-10-18：道具店購買來源與銀行肖像

路由沿中文顯示隔離與正常玩家試玩入口。原版使用固定explore-v4、3000 cycles及自己正常建立的SAVETEST，雜湊沿用前節。原始29檔、存檔來源及dosgolem快照唯讀；證據只保存於本地研究目錄。

- 道具店收據 `workplace/item-path-v1/receipt.json` SHA-256 `4c4bf480c8a441697019b26eaadb294f62ef0e5868644492408c36cabbf8e1c8`，來源映射 `source-evidence.json` SHA-256 `4a4e1212f5ef08e9420da4b0fa1ea8297b7bda18939123f77c1b7f6dc8f677c4`。32個正常檢查點的色號、色盤、64KiB段、FileOps及KeyReads相同；這項原版探針不稱完整1MiB配對或GUI。
- SAR2檔案payload `0x1A588`、4646 bytes對應dosgolem實模式 `01B1:A000`。初訪段SHA-256 `3748f0286312470a8195befffc2eccb18d8dc98cac29e0637a9532fd5450996d`；已驗文字區 `[01B1:A81C,AB3C)`逐byte對應SAR2檔案 `[0x1ADA4,0x1B0C4)`。位址基準不可混用。
- 711Bh／7120h各201個實際字形，383Eh／384Ah／37CCh／3804h各三次商品名稱。清單來源B09Ah、B0D4h、B0F4h，原始反斜線與控制碼保留。第二件實為Magia Stone，前次讀圖Magic Stone已勘誤。B21Eh數字暫存為50加單一FF，BX137／144、CL109；後接A901h的控制字元也實際進入字形入口，不用字面搜尋推定消費順序。
- 第一件選取後即扣款1000→950，後續是是否繼續購物。歷史檔名 `first-item-quote.png`保留，但不稱購買前確認。正常離店後Return開物品頁，可見藥水圖示；沒有使用藥水。手冊第9–10頁的MAGIC四選項已核對；手冊稱五項但實列四項的差異保留。商品中文名稱為試譯，未宣稱官方譯名。
- 道具店標題22CDh入口SI=A81Ch，header `0E AF 00 19`、正文A820h起25字元，原生範圍 `(56,175,125,8)`。原貌像素563個底色5、256個字色9、181個陰影45；每個陰影左鄰均為字色9。120-byte共用繪字段雜湊與銀行相同，但來源及header獨立核對，不推廣為任意標題格式。

銀行三個已取得外框畫格的分組、原色盤與RGB對應在 `workplace/bank-hd-v1/source-mapping-v1.json`。三個畫格兩兩差84／497／413個原生像素，來源範圍為 `(56,23,96,64)`；三張1536×1024手繪圖直接取樣至內框282×186，不改原版動畫。完整內建imagegen提示詞組與原圖路徑在同目錄 `generation.json`。素材只作本地研究，不加入Git或發行包。

首輪HD視窗固定c74e9484候選，迎客pose2內框52450像素改變、區外0、原貌完整還原，三圖steps130226791／cycles134586558相同。完整harness最終exit1，原因為未交易告別與先前交易後告別不同，腳本誤等後者。部分收據 `gui-v1/partial-receipt.json` SHA-256 `a5025e3db2df9545bd8289f4201789e6f975a2d5be53dd00f6b9916e3263d394`；不換名為成功，沒有重啟或改live腳本。

未交易告別另由原版正常路徑核對。收據 `workplace/bank-hd-v1/no-purchase-source-v1/receipt.json` SHA-256 `2782656c7c4e07503c335c84f24167e98887bf178e780101b6707daffb2d8915`，來源映射SHA-256 `b9f57ba7f14b4777a5692de010a43a7c3821f6bb961d69e70edd356ba2c45059`。12點色號、色盤、64KiB段、FileOps及KeyReads相同。來源 `[01B1:AC5B,AC9D)`對應SAR2檔案 `[0x1A0BD,0x1A0FF)`；61字形連續，尾部控制碼完整，待續AC9Bh／旗標0已驗。正常Space後離店1000，持鍵空。原版此輪見pose0，首輪GUI則見pose1，不推測其切換規則。

- 整合候選SHA-256 `dc2696ed47cf8d3f7d3a2a199f7408fc89fdf00943d298f216cb406ba24c7356`，來源身份 `workplace/item-overlay-v1/candidate-identity-v2.json`。Go1.24.13／Ebitengine2.9.9，固定dosgolem882來源，36筆對話、43筆標籤。八個受影響頂層測試與十一個catalog拒絕子項全部通過，無skip；六條配對各35／12／32／9／11／3個抽樣點完整1MiB、色號、色盤、CS:IP、steps／cycles相同，覆繪區外0。已驗命名離店／關閉檢查點覆繪差0，重新開啟裝備頁的中文不當作殘留。
- 銀行HD新版收據 `workplace/bank-hd-v1/gui-v2/receipt.json` SHA-256 `5648ec301e09cff81614dd8281a732086584d3dcdc53c97207a71ed0c814ae72`，目視紀錄SHA-256 `c5ed1ea9fe7b4876df6065c16937a3b026711bf359e58816d5bcbb29475d1fc9`。正常冷F7、迎客pose2與未交易告別pose1，兩組各52450像素改變、區外0、全遊戲畫布還原0，來源RGB及直接素材取樣相同，paused steps／cycles不變。60筆按放確認、最長0.22秒，9張實際圖目視，正常離店1000、HD開但不匹配、持鍵空、capture-tail退出0。保留離店paragraph19／kind2／first_difference218安全退回，不稱三姿態GUI全驗或完整動畫。
- 道具店GUI v1在導航主動停止，未進店或購買，失敗收據保留。來源補證 `workplace/item-overlay-v1/navigation-v2/receipt.json` SHA-256 `41478c5d7355a138147142d2cd00929a297b28f7b511619529e7e5527b9223f1`證實實圖門中心144、左向角色位置下，正常Up可進MAGIC，與既有初訪完整indexed相同、色盤相同。只修harness接受已證實136及144兩點，不推廣碰撞容差；新版GUI實際遇到136，不宣稱GUI另測144。
- 道具店GUI v2收據 `workplace/item-overlay-v1/gui-v2/receipt.json` SHA-256 `98a4b392db04a5f1d0441ddae624414dd66cbfc61c9fc72fcae709616513b9d4`，目視紀錄SHA-256 `80c6076c5af9bae4837cf12454eafd417fdb825c41367398c32c5793e7b69cb2`。88筆按放確認、最長0.223秒，購買後78字形／ready7／completed1／price50；原版50的45×24圖形在三倍遊戲畫布 `(552,302)`RGB差0，六個金額／清單價格／藥水圖示區域也相同。11張實圖已逐張目視，續購否、告別、離店950、物品頁及關閉完整，持鍵空、capture-tail退出0。保留離店paragraph18／kind2／first_difference422舊文字退回，不稱零失配。
- 兩次正常GUI固定輸入、候選、41來源與8鍵序、原29檔、自建存檔與複本及字型均不變，所有相關容器已回收，擁有權異常0。主入口已同步dc2696ed，原fc797237別名保留。NO USE、其餘HUD／場景、其他物品與藥水使用仍未翻譯或驗證；沒有新存檔持久性、實體手把、完整遊戲或正式平台包聲明。

## 2026-10-18：道具店介紹、出售與四組HD畫格

原版入口為 `workplace/item-branches-v1/receipt.json`，SHA-256 `ce132fe7f07ff3fe29e7e8324dbbfd68f2de04b82727613022be42b18bb309a1`；文字映射 `source-evidence.json`為 `0779bbc09c60b7b47527943ee06e1b4a3a4fa9eb301461feeb9c8caa09386077`。使用既有explore-v4、自建SAVETEST、3000 cycles與固定882來源，Go1.24.13。窄探針程式SHA-256 `55b433c75f0e2351698d86eddc400060f5b0c70db1d0dc8ac4c6608e8bb4c13b`、二進位 `dfcea1460d7b617e5e4674ffa2e417cdd039208cb6a4a9e85e75ce0cfe240250`，只觀察已指定八個入口。

- 63個檢查點的色號、色盤、64KiB段、FileOps及含KeyReads收據逐byte相同。最終482.8msteps／489124421cycles、金幣975。原版鍵序SHA-256 `7294e3ce8e21c1d5ebc0985b0e60c67287ea7ac2294604a9640950c1c54045a1`；正式測試鍵序刪除探針專用quit命令，遊戲輸入與預算不變。
- 已證實文字對應，位址基準分別為dosgolem實模式與SAR2檔案偏移：Ken’ko `[01B1:AB4A,ABC5)`對應 `[0x1B0D2,0x1B14D)`；Magia `[01B1:ADFD,AEA6)`對應 `[0x1B385,0x1B42E)`；刀劍油 `[01B1:AF3D,AFD3)`對應 `[0x1B4C5,0x1B55B)`。這只證實說明文字與正常選取，不證明效果規則。
- 第一頁待續游標依序AB86h、AE34h、AF7Ch，旗標0。刀劍油第一頁在句中等待，第二頁才接增強劍攻擊與末句；不能依斜線單獨推定頁界。介紹清單是兩列viewport，正常Down可捲到第三件，不是只有兩件商品。
- 出售數字來源 `01B1:B21E`為 `32 35 FF`，原版711Bh兩字BX=138／145、CL=119，原始CX高位保留於trace。確認前金幣950，接受Yes後975，謝詞接四列主選單詢問；沒有觀察到續賣Yes／No。正常離店後USE:NOTHING、藥水圖示消失，未使用道具。
- HD來源入口 `workplace/item-hd-v1/source-groups.json`、`artwork-manifest.json`及 `generation.json`。十個已知室內抽樣歸四組；人物原生局部差異界為 `[18,8,48,63)`，蒸氣為 `[48,1,95,48)`，相對96×64外框。選定1536×1024圖直接合成至282×186內框，共用底圖，差異只限上述區域。前兩張右手越界稿保留並拒用；四組不是完整動畫。
- 素材是原版構圖衍生，只留本機；所有原29檔與自建存檔不變。原版探針與離線HD測試不稱正常X視窗或完整中文化完成。拒絕售價、多物品續賣、物品效果與其他城鎮仍未驗。
- 中文配對收據 `workplace/item-hd-v1/pair-v3/item-branches/receipt.json` SHA-256 `b96f132d3b65c45bf1b3b223a1a29100d4d7e2e5ad334699b9fd19909bf7c3fc`，63點完整1MiB、色號、色盤、CS:IP、steps／cycles相同，區外0。介紹完成三次538字形，購買一次78字形，出售一次90字形；離店與關閉物品頁覆繪差0，持鍵空、金幣975。保留離店paragraph18／kind2／first_difference1舊文字安全退回，不稱零失配。
- 固定候選 `881ef8dd27ec6f28ee880e27f557c135f1d40531ea43be589103a64d29951eef`，身份與來源入口 `workplace/item-hd-v1/candidate-identity-v1.json`及 `source-audit.json`。Go1.24.13／Ebitengine2.9.9，43來源與9鍵序，51筆對話及43筆標籤。新分支、舊首購32點／銀行35點／武器9點及HD共八個頂層、十九個catalog拒絕子項全部通過，無skip；go vet／建置通過。GUI另驗，不把配對PNG當成實際視窗。

### 實際視窗與捲動清單勘誤

- 881ef8dd的正常GUI收據 `workplace/item-hd-v1/gui-v1/receipt.json` SHA-256為 `3e3f028988e44dd5d98be0377125ef1c8745ed5ae8006cca5a062be87a7f659a`；目視紀錄為 `16c0ec52f878624aeb691272db7878aeb919d74bb776b4060a30ea22dd852727`。六頁介紹、原版待續三角、空出售、買50、售前950、以25出售後975、離店與物品移除通過。152筆按放回執、29張目視、權威退出0。HD實遇0／1／3，內框52406／52405／52384像素改變、區外0、完整還原0，暫停steps／cycles不變；pose2未在GUI遇到。
- 同份GUI目視發現捲動後Magia Stone／Sabre Oil兩列仍英文。舊配對參考也保留英文，ROI相同不能證明中文完成。舊收據保留此限制，沒有換名為修正後驗證。
- 窄顯示證據在 `workplace/item-hd-v1/scroll-fix-v1/pixel-source-evidence.json`。SAR2 SHA-256仍為 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`。dosgolem實模式來源Magia `01B1:B0D4`與Sabre `01B1:B0F4`分別對應SAR2檔案偏移 `0x1B65C`與 `0x1B67C`。此時trace只見383Eh／384Ah，沒有再走37CCh／3804h或22DBh／2308h；刷新consumer仍標「未知」。
- 已證實顯示條件為兩列viewport `(92,120,132,26)`，名稱起點 `(112,124)`／`(112,134)`，每列含黑色尾端的完整65×8色號圖形。viewport固定部分SHA-256為 `ac575320f6b3ef26a3208535c532847f7ccbd8bfb9f42cc349d4633804d79e2e`。原版buffer `01B1:3D4F`以160步距保存八列；只讀取已觀察來源，不推測搬移常式。
- `ObserveFrame`只在模組前綴、完整原文、色盤、固定框、唯一已知箭頭及兩列bitmap全部相符時記錄當前列，順序只接受Ken’ko／Magia或Magia／Sabre。`Draw`重新驗證並只繪製既有記錄，避免隔離繪製舊標籤時建立無關的新列。失配撤除，原版buffer、記憶體、輸入及捲動邏輯不變。
- 修正候選70a3f8e3的完整SHA-256及43來源／9鍵序身份見 `scroll-fix-v1/candidate-identity-v2.json`。配對收據 `pair-v2/item-branches/receipt.json` SHA-256為 `495ecfcc5dfb8bb890541201411dacf79f0537c91aabaa87cbb5c66d85729b23`，63點完整原版狀態相同、區外0、售後975、離店及關閉覆繪差0。首購32／銀行35／武器9點回歸也通過；七個頂層與十九個catalog子項PASS，無skip或fail。另驗兩列中文與位置、重複畫格及五類失配退回。HD與TSV沿881相同來源，沒有冒充新候選重跑HD視窗。
- 短GUI收據 `workplace/item-hd-v1/gui-v2/receipt.json` SHA-256為 `2f16bb16dcf0bbb8a5b04154d270770d8c92a2421a847252ca61255186d66655`；目視紀錄為 `99609619565d43cd3a11d04f6cc5d68ce1334cb3fe9e1d93490ebc5d6e1c878f`。同54233執行至權威exit0，初始兩列中文、Down捲至第三件、油品兩頁／三角／續問否與正常離店1000通過。102筆按放、10張目視、285固定輸入、原件／字型／自建存檔不變，持鍵空、容器清理、擁有權異常0。HD關閉、購買與出售完成次數0；不稱重驗交易或HD。最後保留kind2／paragraph18／first_difference242舊文字安全退回。
- 主入口已同步70a3f8e3，整合稽核 `scroll-fix-v1/current-audit-v2.json`同時核對兩份不同候選與範圍GUI、完整配對及素材未變。v1英文捲動限制仍可回查，v2修正另有收據。原版、圖面與字型留在本機，物品使用、效果、其餘場景、音效、戰鬥與實體手把仍未完成或未驗。

## 2026-10-18：城鎮與洞窟的道具使用顯示

來源入口 `workplace/item-use-v1/source-evidence.json` SHA-256為 `4c0284727b50cf87ab336ac28d82eb4310da71c21d4464b79029dbf7bc12f571`。原版EXE與SAR2分別沿用 `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`與 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`。Go1.24.13、3000 cycles，固定dosgolem882快照；原版explore-v4為f4a17bb5，窄trace程式為58d8ccbe、二進位7de96845，完整身份保存在證據檔。

48點中文配對在同目錄 `pair-v4/item-use/receipt.json`，完整1MiB、色號、色盤、CS:IP、steps／cycles相同，文字區外0。五個新白字來源及小圖示均驗實際可見像素，關閉與返回城鎮整張畫面差0。根端另目視使用訊息、重開空欄及關閉三圖，記在 `pair-visual-review-v4.json`；它們是配對合成圖，不是實際視窗。最新五項回歸為 `regression-v3.jsonl`，銀行35／首購32／介紹出售63／武器9／裝備11點均通過，無skip。實際視窗驗證另記收據，不由上述結果推定通過。

實際視窗入口 `workplace/item-use-v1/gui-v3/receipt.json` SHA-256為 `9ea846be928c497ccb6da771eb610ced710629669fb180b472f8c4a0663ade77`，目視紀錄 `bb06a74c3f412dc3da6eaf3ee6b5249278de97f1a7e483cacbaa534cba6c52cf`。固定b8c1be29正常入洞、選藥水、Space消耗、重開空欄與返回950，capture-tail退出0；最終294918786steps／301946427cycles、持鍵空、HD關。14圖目視、26原版ROI零差、116筆按放確認、389固定輸入與原件不變。滿生命80不證明受傷恢復量，離店保留paragraph18／kind2／first_difference242安全退回。GUI1／2失敗未改；根因及原版放開補證見同目錄上層 `navigation-v3/receipt.json`與 `gui-tools-v3/preparation.json`，沒有把持鍵搬移中間畫格當成輸入就緒。主入口已同步，原貌HUD仍原文，後續候選字面位置只供研究。

- 已證實玩家可見差異：城鎮Enter物品頁的USE為綠字，持有藥水仍不能移動該區選框；正常經右側石門進Cavern of Ialicia後，USE紅字，可用Right選首瓶、Space消耗。使用後出現訊息，關閉／重開時藥水圖示移除，原版返回Muralla Town仍950。此次生命80且上限80，不當作受傷恢復量證據。
- 原始控制定位保留於既有IDA9.4匯出 `workplace/inventory-ida-v1/evidence.json`，SHA-256 `9c742b9daa9ad5890ce81a958cecab6f2ec82821e147125fbf6323af28f39266`。輸入 `equipment-held.segment`為 `2ecedadb544f1d089612ca19b1f42379fa83866c447121b131c61435fcdf7e3e`，位址為IDA base0段內偏移，對應dosgolem實模式01B1，並非EXE偏移。A004h與A00Bh分別寫ADF8為0與FF；A09Eh、A698h及後續選用分支讀它。此次兩種正常物品頁觀測值相符，沒有改名、改寫欄位或重新解硬體輸入。
- 裝備模組SAR2檔案payload `0x401E`載入A000h。新ASCIIZ來源AA9Ah／AAACh／AB72h／AC44h／AC4Bh分別對應檔案 `0x4AB8`／`0x4ACA`／`0x4B90`／`0x4C62`／`0x4C69`。完整原文、AH=1、每字位置與AA30h完成事件已核對，分級為已證實；五筆新增索引附來源雜湊與證據，不覆蓋原始定位。
- NO USE小圖示不假定走文字入口。四張正常圖的完整16×16圖示SHA-256同為 `1eee9c14d388ec401a73b41898b6b9cfc7497315a292fd1b6828ae293c7278a3`，原生範圍 `(58,131,16,16)`。覆繪僅用內框 `(59,132,14,14)`，保留圖示框及原版紅／橙選框。原版色盤、兩組已驗無魔法頂部面板、模組前綴及完整文字來源也須相符，未知畫面退回；後續其他魔法配置沒有冒稱已覆蓋。
- 窄trace-v1嚴格64KiB比較差2bytes於01B1:0A6F／0A70。兩值分別為27563與27766，精確對應各自存檔複本UTC修改時間13:29:23與13:35:44的DOS編碼。差異紀錄SHA-256為 `b73d519d8569fbcdf9ad7c8bf0202498339d5e7851bc4e98a030321c5c7aaad5`。trace-v2執行前保留原版複本mtime，55點色號、色盤、64KiB段、FileOps及含KeyReads收據全部相同，沒有遮罩。此探針不稱完整1MiB或實際視窗驗收。
- 正式正常鍵序為 [item-shop-use.commands](tools/item-shop-use.commands)，SHA-256 `0c7ef14432a52850db25ce9e8859f3cf5e94868b0c9dbbc3e9b3d0bb6c4f18c4`。原版48點在 `workplace/item-use-v1/normal-v3/`，279200000steps／286064030cycles、950、持鍵空、沒有寫檔。導航來源 `navigation-v1/receipt.json`另記石門實圖與對齊，保留第4個非必要粗段，不宣稱最短鍵序或碰撞容差。中文完整配對及實際視窗另驗。

## 2026-10-18：狀態列四欄與三場景名稱

入口 `workplace/hud-v1/source-evidence.json` SHA-256為 `15aaf7d0749f0aeca4bd32b04bb3a53338fb4b9865a0c30eba2b9efa3c689acc`，受版控索引為 [hud-source-index.json](docs/hud-source-index.json)。原版EXE沿既有fe1c00b3，固定882來源摘要2a5ad11a；Go1.24.13、3000 cycles、自己正常建立的SAVETEST。IDA9.4輸入冷啟動01B1段SHA-256為 `51a087f2bf332b916e96ace05a72e679872646300dba39176f32584d8b7ecf81`，IDA base0等於該段內偏移，檔案偏移另列，沒有套用EXE位移。

- 已證實：01B1:200Eh近呼叫表指向22BFh，2010h指向22CDh；前者設定字色27／陰影18，後者9／45，均走2312h四byte header及2344h完成。2345h字形常式將陰影寫入DI+1，字距5原生像素。因此 `x=header[0]*4+header[2]`、`y=header[1]`、`count=header[3]`，完整字形需`count*5+1`寬及8高，包含最後一欄陰影。原始函式名保留為sub_22BF等，沒有以名稱推定語意。
- 九個來源含SAR1 payloadCEB5h載入6000h的LIFE／ALMAS／GOLD／PLACE，SAR2 payloadECh載入6000h的GOLD／PLACE，以及SAR2 payload38A1Ah／392DCh與SAR3 payloadAD02h載入C000h的城堡／城鎮／洞窟名。各筆完整來源、模組前綴、字形與色盤雜湊、安全區及完成圖在索引；23次匹配完成畫格逐筆核對。首次誤沿用城鎮module定位城堡，已由原始header斷言拒絕並更正，歷史analysis-v1-failure保留。
- 原版探針v2程式SHA-256為 `bc901fa7ca213966318b9980cf6bf248e7511303b8ecf392c4ee00f170d1fc5b`，二進位為 `6312fd4116516632f70ffc3d71648102545a981233aaf6a4aa7b679fb324359f`。同48點色號、色盤、64KiB段、FileOps及含KeyReads收據與無探針原版逐byte相同。copy2在執行前固定來源mtime，沒有遮罩；這份窄證據不稱完整1MiB中文配對或正常視窗。
- 勘誤：前輪從畫面讀為Cavern of Ialicia，原始SAR3檔案C31Ah及實際01B1:D618字串逐byte證實Cavern of Malicia。C88Fh另有同名字串，但不是此次載入模組，不接入。歷史圖片與收據不覆寫；後續以Malicia為準。中文試譯保留Felishika／Muralla／Malicia原拼法，不宣稱手冊官方譯名。
- 生命與寶石原版字形在6000h模組換入洞窟版後仍存在。新層保存已驗來源快照，每幀只比對完整原版欄位與色盤，不把卸載模組誤當清框；未知覆寫退回，同一完整字形還原可恢復。最多五欄，不移動原版數字、條形或圖示。新層完整配對、字模與安全區測試、正常GUI另留收據。

- 固定候選3aeca361的46來源／10鍵序身份在 `workplace/hud-v1/candidate-identity-v1.json`。`pair-test-v2.jsonl`三個頂層與七個catalog子項PASS，無fail或skip；新層與舊使用各48點完整1MiB、色號、色盤、CS:IP、steps／cycles相同。新層收據SHA-256為 `81f3665baa7ac52088e3e2d6b9cbdc6659baccae3c7a62ee9ea771430d2b60a1`，舊使用為 `df64f64f665d2b626c902f1040cd59b1e113eb82a92754df13793ef3143d16fd`。每個中文字模確實可見，安全區外0、數字與生命條原樣；五類失配及模組卸載後欄位保留通過。go vet與建置通過。
- 正常視窗 `gui-v1/receipt.json` SHA-256為 `18e0309bac3485a33298126c509f15b0ec4d28fbf03c80b60580a418eabc1348`，稽核為 `b901e73bea37bea960f71695f0bdf95bfbc6828dc01b0927cdde14a5c47e8892`，目視為 `e948f56d9430c553e043c5cffaa14e84e322c5e4c61d697bdde23f5c6dfb57c7`。同35152完成冷啟動城堡、冷F7城鎮、買50後950、物品頁開關、正常入洞及返回城鎮。47個原版區域及40個中文區域零差，8個HUD帶安全區外0；16圖目視、55次按放／110筆確認，最長確認0.182秒。最終286298902steps／293261430cycles，持鍵空、HD關、capture-tail退出0。保留paragraph18／kind2／first_difference422安全退回；本輪GUI保留藥水，未消耗或驗恢復量。394固定輸入、原件、字型、自建存檔及mtime不變，根端另看三圖，主入口已同步。
- 共用字形的補證在 `font-primitive-proof.json`。01B1:F504指向F856，768-byte字型SHA-256為 `f16d564fee5c4472914b4dc64cf467bcd05da89125288a83b0b065d164707d07`；依IDA2345h指令重建24個完成header，原生像素全部零差，48個快照字型相同。原版29檔尚未找到同一連續字型，producer仍未知，不由此宣稱已定位檔案字型或散布字模。
- 後續候選在 `scene-candidates.json`，SAR表邊界與項目數已檢查。175個模組掃得52個header形狀候選、42種字面，含5個本輪已驗來源。其餘消費端、載入基址、顏色與生命週期仍未知，不能直接納入中文層。此補證不改本輪固定候選或擴大CONFORMED範圍。

## 2026-10-18：場景名稱的當次消費端驗證

來源入口為 `workplace/scene-v1/consumer-evidence-v2.json`，SHA-256為 `7873562215ef08e130f64175f9ce9fd011dac3500ff1ecf32396529124a29d22`。字型與22BF至2386機器碼沿前節固定雜湊；IDA位址仍為base0等於dosgolem01B1段內偏移，SAR檔案偏移另外記錄，不混用。新觀察器來源6713d826，沿前節已驗入口只補2344完成時64KiB段，不寫原版。正常武器9點及長老1點的色號、色盤、完整64KiB、FileOps與含KeyReads收據相同；這份窄觀察不稱完整1MiB配對。

- 已證實：國王、武器店、小屋及Marid長老都走22CD／2344，載入基址A000，完整header、字串與64-byte模組前綴唯一匹配。21個場景／PLACE完成畫格依既有字形契約重建，原生像素零差；字型指標及768-byte字模、程式碼均相符。
- 檔案定位：SAR2的國王body17238／module16E1A、武器body184C9／module17817、小屋body17809／module175C0、Marid body1C997／module1BCC6。完整來源與bitmap雜湊附於 [HUD索引](docs/hud-source-index.json)，原始位址及字串保留。專名仍保留原拼法，不稱手冊正式譯名。
- 44候選的來源／字形推導在 `source-contract.json`，SHA-256為 `87db6cc153ca3465ef39d9ecdfebfe6d8c4c8c26062ccfa9b8daf300ff8b49e0`。完整header與prefix配對無歧義，固定字型推導的bitmap及53筆譯文均符合安全區。新增4筆有正常觀察，其餘40仍為強推論，不能由catalog存在推論正常可達或全遊戲完成。
- 新契約只在實際22CD入口，以當次SI減去來源相對header導出base，唯一匹配模組前綴與完整原文；字型指標／雜湊及程式碼也須相符。2344完成再驗SI、完整bitmap與色盤。這讓候選必須在實際使用時通過消費端驗證，沒有依字面或固定場景索引強套中文。銀行／道具店標題維持既有層，ENEMY未納入。
- 完整配對入口 `pair-v1/scenes/`國王／武器10點、國王／小屋／長老3點，以及 `pair-v1/hud/`48點；完整1MiB、色號、色盤、CS:IP、steps／cycles與原版相同。另 `regression-v1/item-use/`舊使用48點無覆繪殘留。兩份測試紀錄合計六個頂層及九個子項PASS、無skip。88個合成base匹配及792拒絕只證明來源閘門，不冒充未走過場景的正常遊玩；正常GUI另驗。
- 正常GUI v1第一路國王及武器店通過，第二路小屋通過但長老門未對準，權威exit1與partial保留。既有成功GUI使用房屋外觀對齊；當前失敗的Draw觀察不是精確輸入edge，沒有據此宣稱同狀態原版重現。新的最小補證 `navigation-v2/verification.json`以正常冷F7自己存檔後Up直接進Marid，來源ACCD、模組A000與原始SAR2相符，完整字形e1d3bb2e與固定色盤相同。原版105300000steps／109182814cycles、1000、持鍵空、沒有寫檔；字形證據不替代後續正常X視窗。
- GUI2收據 `gui-v2/receipt.json` SHA-256為 `8b3b2828e928d4b7d65aebbf7dbb24ade1f78e5f025b2b733c59f0b4bb39d800`，稽核為 `3c3ecd9b65bbf66974588759eefb619cad5d80fe6caf04cf9fe1d756517fdcf6`，目視為 `26b7273059c2ba8b05ca5306392c81c347ae3221ca874ce5ff5c04f88cac3a06`。正常F7自己存檔後Up進Marid，重訪中文及名稱清楚，8原版ROI／1名稱ROI零差。4圖目視、52筆按放確認、139輸入與copy2mtime不變；1000、持鍵空、HD關、capture-tail退出0、容器空。原GUI1的exit1及partial保持原樣，前三種名稱仍沿GUI1身份，不宣稱同一完整新局全部到達。
- GUI2原版indexed SHA-256為 `779029018e65b6456fcc6cbbaa3c0b527e63d8ebb9acfa7ead370654bb7cfafd`，與原版補證同圖相符。以固定補證palette解碼後，實際視窗肖像與既有中文區外零差；本次未獨立擷取palette，不稱整份當前色盤已驗。首次分析錯用208×48對話範圍，依固定源碼210×52更正，原紀錄另存；修正是測量契約，不改產品或增加任意遮罩。根端四種實際名稱圖另記 `root-gui-review-v2.json`。

## 2026-10-18：居民專用對話消費端與動態外框

本地入口 `workplace/prose-v1/consumer-evidence.json`、`ida-town-evidence-v2.json`、`world-script-inventory.json`。原版EXE、SAR2、工具鏈與位址基準見 [居民規格](docs/spec/004-town-dialogue.md)。IDA9.4的binary base0等於dosgolem01B1段內偏移；SAR2檔案偏移另外記錄。輸入 `explore/citizen-near-2.segment` SHA-256為 `dae6b666047046143c22326c2cbaadad61e3665d3add21a4887d78835d5ef405`，原名、bytes、xref與運算元保留。

- 已證實：居民使用63CA及7C58游標，與室內706C／FF4C不同。63DA以C00D表查SI，64B2／64B7繪字，6551／658F標記，65A1結束等鍵，63A3背景還原。前節的706C結論僅限室內，不能直接套居民。
- 已證實：623F檢Space旗標FF1D及C2朝向，FD居民標記須在面向前1至3格。首次Left1m太遠、再Left2m重疊，Space由BIOS讀到仍不交談。保留explore失敗；正常重播由固定自建存檔Left2.5m放開後達前三格，Space交談成功，不改NPC、角色座標或輸入旗標。
- 已證實：正常來源為SAR2檔案039FFF、模組payload0392DC、段內CD23，腳本248 bytes SHA-256 `5cd9baf5dcd8e696df71b9ea95d4ba7049daf0831ab2c3e3b9d77563557a5ed7`。九點原版與唯讀觀察的64KiB、色號、色盤、FileOps及完整收據完全相同；完整1MiB另由後續覆繪配對證實。
- 已證實：63CA至664C代碼SHA-256 `2011d284c03ced34946616a7dde903d3e9972d9af6e00301f78d4ea0c562581c`；7B82至7C41左側空白與advance表192 bytes SHA-256 `15eda0237134c09a79b7a068929c292ef9161f96c9f46d3b79b5555b5bfe9cf9`。producer未知，不由此推定原檔字型位置。
- 已證實：6609依字寬168與空格後單字預看算行數，斜線換行、單FF結束。外框最多8行，7C4E／7C5A由原版控制。正常原文11行，外框(56,24)-(232,110)、安全區(60,28)-(228,106)、紅色標記(140,98)-(148,106)。中文以24顯示像素畫到3倍目標畫布，未修改原生畫面。
- 強推論：十模組指標表118腳本，113純文字完整FF段、5混合控制碼。113譯文的鍵、字型及獨立原文行數容量全通過；只有039fff正常走過，另112須在實際63DA入口驗完整來源、前綴與表，不能稱全部可達。5混合源03BA83／03F1E2／03F291／03F294／03F346不覆繪，不猜購物／劇情回呼。
- 追加勘誤：pair-v2／v3在第77字形s退回，BX59、CL58而安全區x60。IDA64A6／64AE依左側空白1扣除，故字形原點不等於可見像素邊界。改為驗證原版計算公式，沒有拓寬中文區；兩份失敗原樣保留。
- 完整配對 `pair-v4/receipt.json`驗冷F7自建存檔、交談、續頁及離開9點，1MiB、色號、色盤、CS:IP、steps／cycles、FileOps、KeyReads及持鍵相同。第一頁44／全文63字，中文改變26080／27297像素，安全區外及紅標記0差，離開後整張覆繪差0；248次原版字形完成事件只作觀察數。八Draw拒絕、八catalog拒絕通過，未將合成條件算成正常路徑。
- 正常X視窗另驗，配對圖目視不能取代GUI收據。固定候選06b63ee7、50來源、10既有鍵序與一條town鍵序，原件、字型、文字及含原版畫面的收據均留本地。完整中文、HD、音效、戰鬥與實體手把仍未完成或未驗。


### 2026-10-19：中途呈現補證與正常視窗收斂

- GUI1尚未交談即因走路相位辨識失敗，exit1保留。原版正常冷F7後Left2.75m、Space、續頁及結束成功，`navigation-v1/verification.json`權威exit0。actor48樣本對GUI的差異由4變0，來源為合法原版另一走路姿態，未加誤差容忍。相位比較在 `actor-phase-comparison.json`。
- GUI2首頁中文與marker通過，續頁在glyph220退回英文。`mutation-baseline.jsonl`在118429284steps、DS01B1、原版6513一像素寫入後重現；色號已變、舊pending=false而清掉來源。這是產品顯示生命週期缺口，不是導航或原版遊戲失敗。候選06b與50來源及READY規格保存在 `candidate-v1-sources/`，兩份失敗不換名。
- 已證實：64FA開始十次650E／6513一像素捲動，6516才完成一列。另654C、658A、639E是標記與背景寫入前。新層在整個寫入期間pending，只暫呈原版，不移除有效來源；完成後才capture。原版資料、輸入及時間不寫入。
- `pair-v5/receipt.json`9點完整原版狀態相同，30子步及三種原版寫入前呈現通過；中文圖與已目視v4三PNG逐byte相同。候選08b3309、50來源／10既有鍵序、51／48／53／113 catalog，go vet／build-v3通過。
- GUI3同35191權威exit0，正常F7自建存檔、兩頁中文44／63字及Space結束，1000、持鍵空、rejected0、HD關閉。163輸入／58按放確認及24圖目視，20區域、頁間區外與同GUI背景還原零差；根端另讀三張實際X圖。最終raw採固定來源palette對照，未獨立擷取色盤，不宣稱整段GUI同狀態。
- GUI2下框3點在兩條原版near→page均保留進框前背景；換同navigation-v1的原版參考後352像素零差，見 `frame-corner-comparison.json`。GUI3三角點與本次pre-talk背景相同，未新增任意遮罩。
- 5個混合入口中03F291、03F294、03F346共用FF結束區間；85／87會重新分派63C5及BL=4／5。六種handler原始bytes與位址在 `mixed-control-notes.json`，只證實底層跳轉與寫入，不替byte_4／34／9A及6AED猜玩法名稱。此補證未加譯文或runtime。

## 2026-10-19：城堡與Muralla共用遠山天空

- 證據入口為 `workplace/town-hd-v1/layer-evidence.json`。輸入EXE SHA-256 `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`，SAR2 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`；dosgolem摘要 `2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`、Go1.24.13／Ebitengine2.9.9、cycles3000。
- 已證實：正常城堡、冷F7讀回、兩個步行位置及對話結束的原生矩形(48,24)-(272,78)色號全部相同，224×54 SHA-256 `22fb614b02b3d9ac436c85b7d56c62ce9004d1254f6c52160de2cf62edeb7cf9`。較大矩形從castle y78／town y80混入移動前景，不能宣稱整張背景靜態。
- 已證實：SAR2檔案payload偏移038A1A、0392DC對應執行期01B1:C000兩個模組前綴，64-byte SHA分別 `968819bb6c4906514b4fce3e2ce99764b73cfa003f48d7d57d4cf8b12da8ffb6`、`025bf2982660c942cbefed303ccdd0523ff824756a0fe0bb2587b99e0931e7ec`。檔案偏移、段內位址與VGA座標分開記錄。遠景producer、跳躍遮擋及其他城鎮適用性未知，不追硬體driver。
- 內建image_gen依原版構圖產生手繪圖，完整實際提示詞與來源／權利分類在同研究區。實際PNG1973×797，SHA `275f14de1b18ea6dea297999cdf76e2574bcdc9a4a975782d5c01901480058de`；畫帶(0,207)-(1973,589)直接取樣到672×162。這是顯示差異與本地研究素材，不是原版像素一致或可公開資產。
- `pair-v1/receipt.json`正常9點完整原版1MiB、顯示、CPU時間、檔案與輸入相同。GUI2實際走冷F7、步行、中文兩頁、結束，權威exit0；兩組暫停HD切換的時間與持鍵不變、兩層聯集外0、原貌還原0。最後raw比對採固定來源palette，不稱當次獨立色盤。GUI1腳本點錯按鈕的exit1保留，不列為產品失敗。
- 此層只完成上方54列。手繪與原版下層的水平接縫可見；樹木、建物、角色及完整城鎮HD仍待完成。規格與證據索引見 [遠景規格](docs/spec/005-town-skyline-hd.md)及 [CONTEXT](CONTEXT.md)。

## 2026-10-19：完整遠山、已畫快取與前景位元

- 來源入口為 `workplace/town-hd-v1/foreground-v1/`的 `ida-foreground-evidence-v3.json`、`normal-v1/receipt.json`、`layer-proof-v2.json`及 `display-contract.json`。原版EXE／SAR1／SAR2及dosgolem來源沿前節固定雜湊；Go1.24.13、cycles3000。IDA9.4輸入正常citizen-near.segment SHA-256 `3291a8aeb5cfb46026e0bcfd4fd26ddbf52a7b6cba80044d8d518eef94a5751f`，binary base0等於dosgolem01B1段內偏移，不是檔案偏移。
- 已證實：3028..3050從VGA A000實模式段擷取原生(48,78)-(272,102)到01B1:A000，欄交錯排列。兩次正常3050快照的5376pixels與當時VGA全部相同；四個正常場景的完整1MiB與前輪原版配對相同，沒有寫原版。快取row-major SHA-256 `d248d2ed7371838652df05bd46908fab397446896e96860ba680492e7a9d1e6a`，完整遠山224×78為 `17a6bbaa433a3dc7123dcc20fab641eab8eab2d743c7a6f70dafee08a656c63b`。
- 已證實：3123的3bytes解出4個6位元色號，31D8依資源段8000的旗標表及D000逐列位元，選圖塊色號或背景快取；本輪資源段11B1由01B1:FF2C讀得，旗標表8006。原始指令／bytes與來源身份在上述IDA與display-contract，語意不改原名。圖塊48bytes、8×8，遮罩8bytes，位元1露出背景。
- 勘誤：最初直接讀FF2A地圖重建城堡差65pixels，其他三個場景0差。城堡地圖與E000已畫快取有五個項目不同；原版32C5..32F9存在動畫映射及地圖寫入。顯示必須讀E000，不能把下一次繪圖來源當作當前畫格。`layer-proof-v1.json`與原腳本保留，v2四場景5376pixels全部零差，沒有容忍或自訂透明mask。
- 未知：SAR壓縮資料內圖塊／遮罩producer。對三個SAR搜尋完整runtime atlas／mask／旗標均無連續原bytes匹配，不能把runtime位址換算成SAR偏移。這不阻塞固定runtime來源、全區重建與原版位元覆繪；其他城鎮及跳躍／戰鬥仍未驗。
- 手繪圖由內建image_gen依完整原版背景參考生成，實際2126×740，SHA-256 `3a6791169fb9fa28f3fac63310155249654ce3e80da21e6fec5b9d80bdf0d160`，直接合成672×234。固定ed01c199的9點完整原版配對相同；四點HD各134708／133826／133367／133367像素改變，前景及區外0。18失配類型與直接取樣通過。
- GUI62531正常冷F7、走路、居民兩頁及結束退出0，200輸入不變。獨立Python讀當次E000及固定原版資源重建原貌5376pixels零差；兩組paused切換的前景／聯集外／還原0，時間不變。最後raw採固定來源色盤，不稱當次獨立色盤。根端三張actualX圖已讀，前輪水平矩形切口已消除；樹木、建物、角色本身仍為原貌，見 [完整背景規格](docs/spec/006-town-background-mask.md)。

## 2026-10-19：Muralla前景與原版角色遮擋

- 來源、工具與位址空間沿前節固定EXE／SAR與IDA9.4。`ida-foreground-evidence-v3.json`原始bytes／xref、`decode_scene.py`／`scene-proof-v1.json`與 `scene-proof-walk-v2.json`逐像素重建原生(48,78)-(272,142)。四正常點及八步行項全部14336pixels零差，restored重複，共11個不同狀態；此為已證實顯示格式，不外推玩法。
- 已證實：01B1:E000為28×8已畫圖塊，FF／FE覆寫項回查FF2A當次地圖，FD從C00F八byte居民紀錄+3取底圖。NPC六圖索引在11B1:4000，圖像4100、21B1:7000遮罩；主角6975..6A3A以C2、E7及6A3B／6A59選圖，圖像11B1:6000、遮罩21B1:8000。正常樣本keep位元為1時來源色號皆0，故可保留原版不透明位元。未知nested FD與非零OR混合整層拒絕，不猜補。
- 已證實：固定參考的地圖為C017、215欄，`world-foreground-v1.json`與 `canonical-map-proof-v1.json`保存來源。末步NPC圖像改1176bytes、C000已切城堡，`scene-contract-v3.json`保留五組完整身份並拒絕；v1／v2失敗與新SAR讀取保留，不以放寬雜湊處理跨場。右行及正常Up的 `scene-cave-v1/receipt.json`、`scene-proof-cave-v1.json`補第十城鎮來源，14336pixels亦零差；入洞後來源拒絕。
- 手繪v1三張多出背景與地基，未採用。v2三張透明2172×724圖分段覆繪；來源、實際提示詞與本地權利分類在兩份 `town-art-generation`。`town-art-layout-v2.json`保存六入口約3至9nativepixels偏移與UV校正，不改原版入口熱區。
- `scene-prototype-v1/receipt.json`離線Go1.24.13重建10城鎮來源，原版角色／框外0差，三個其他來源整層拒絕。根端檢視五圖：樹木、房舍與洞口手繪可辨，三招牌中文在安全框；原生地面、輪廓邊緣與角色仍保留。`visual-review.json`只授權接可撤回前端；正常視窗尚待，規格007達READY，不稱整鎮HD完成。
- 後續人物資源候選：`actor-inventory-v1.json`沿固定restored的1MiB SHA-256 `9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9`、Go前景來源相同的atlas／mask身份，以Python3.13標準函式庫解碼六圖記錄。主角兩組各5項，共10項9個不同RGBA，C2兩組phase4相同；當次居民group0..4各8項，共40候選。50候選均無keep位元1且非零色號。根端已讀 `player-source-sheet-v1.png`，尚未賦予方向／姿態語意或接人物HD；未正常觀察項維持資源候選等級。
- 追加勘誤：前景GUI2等待居民首頁時中文撤除，`gui-scenery-v2/`權威63247 exit1，未續頁。窄探針 `scene-prototype-v1/mutation-probe-v1/receipt.json`不載HD，只讓原版沿同一正常鍵序每50k呈現；88900000steps／92762567cycles、IP11A1:145B、DS11A1重現。全部15136個外框pixels相同，代碼SHA2011d284、字寬15eda023、prefix／table／body及色盤6dc0b8c0完全相符。已證實舊Draw誤套callback的DS01B1條件；不需要追原版driver。原候選0b4b與首次basic收據保留，規格004回READY，修正只拆CPU上下文與固定段身份，尚待新候選驗收。

## 2026-10-19：首場戰鬥正常路徑與中文狀態列

- 入口為 `workplace/combat-v1/`。輸入沿前節固定原版 EXE、SAR 與 dosgolem 快照，工具為 Go1.24.13、cycles3000。正常冷啟動後 F7 讀取自己建立的 SAVETEST.usr，SHA-256 `78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb`，mtime_ns 725847808000000000。手冊 `SCAN0357_005.jpg` 印刷第6頁的方向、Space 出劍、上跳、下蹲已目視；掃描 SHA-256 `8ddd761c499b6161bbdfc2d587cedc3fa701875073cd37df1ae3c2ebb1936d7c`。
- 已證實：entry-v1、v2、v3 分別由正常鍵序執行，沒有記憶體寫入或座標注入。共用14／16點的完整原版狀態逐項相同，收據為 prefix-comparison-v1.json 與 prefix-comparison-v2-v3.json。v3見出劍、敵人消失後遺留物、左移拾取後寶石數0變1、離地跳躍及蹲姿。這些是可見結果，不推導傷害、掉落機率或敵人規則。
- 已證實：`tools/combat.commands` 保留同一鍵序，25點中文配對與原版v3的完整1MiB／色號／steps／cycles相同。兩台原版執行器的Regs、色盤、FileOps、KeyReads及持鍵也相同。19洞窟點的中文來自SAR1生命00DB4C／寶石00DB54、SAR2金幣000D34／地點000D3C、SAR3場景00C31A，各偏移均為原始SAR檔案偏移。既有HUD消費端未改，原版數字、生命條及狀態列外像素不變。pair-v3/receipt.json、pair-review-v1.json及根端三圖目視為入口。
- 追加勘誤：pair-v1把內部保留標籤數當作可見殘字，pair-v2把金幣／地點固定為SAR1來源，兩份失敗保留。回查既有消費端後只修測試預期。pair-v3證實舊F7標籤可見像素為0，25點測試通過，正式程式未改。
- 已證實：當次dosgolem實模式01B1:C000前64bytes的SHA-256為 `3f5edbc8b51aa017211fb0ca808f807a1ade51d95e14b49dadf593e90ada24eb`。洞窟的主角程式、圖像與遮罩區不同於城鎮，不能把城鎮E7相位定義直接外推到洞窟。洞窟HD來源、繪製順序與其他動作仍未知。
- 遊戲亂數種子欄位未知。本項固定完整初始機器狀態、輸入與預算，不重擲，不宣稱隨機規則對拍。正常視窗另驗，原版配對圖不取代GUI。

### 洞窟HD前置來源

- 工具沿既有IDA9.4 py312-v1映像，輸入為正常cave-entered的01B1段，SHA-256 `12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93`。IDA binary base0等於dosgolem實模式01B1段內偏移，以下位址均採此基準，並非SAR檔案偏移。工具與逐檔SHA入口為 `workplace/combat-v1/cave-research-index-v1.json`。
- 已證實：39B9、3B80..3BF2及3C05..3D21將本體與前後手臂的九格來源依序合成。當次資源段由01B1:FF2C取得，為11B1；來源表在11B1:6000..6333，圖像6333、遮罩D000。每8×8圖塊32bytes，16-bit little-endian字的高四位元先解，經01B1:4F8C指標表選16色映射。3626..3689保留遮罩位元1，其他位置覆寫。不得沿用城鎮的六位元角色圖格式。
- 已證實：01B1:511D起的九格暫存區以row-major順序，由40F0..412E送至4FED指示的VGA位置。正常19點中17點的24×24暫存區與原版畫面全部相同；本體／手臂18點共3483不透明像素相同、6種圖面。cave-entered暫存差144但角色不透明像素相同；sword-standing的角色差22、暫存差3保留，伸出武器與完成點尚待核對，不把圖面差猜成新規則。
- 已證實：4192／41AC..41D0及4500從FF31指示的循環地圖取28×18格，行距36bytes、左側略過4bytes，E900回捲E000。遊戲矩形為原生(48,14)-(272,158)。3272／3290..32E3的低位元地形來源為0清零，1..127取11B1:8030+(id-1)×48的8×8六位元圖塊。19正常點共9317格、596288原生像素完全相同；2格像素差及257格動態／未完成項保留。
- 本批只解實遇的顯示分支。敵人、伸出武器、完整下層與完成畫格生命週期未閉合，HD尚未接入。沒有改原版記憶體或共用dosgolem。

### 完成點與正常視窗的補證

- 3123為本批IDA圖像函式302C的return指令。唯讀探針沿固定v3正常鍵序，在此保存當次完整1MiB與VGA，不拼造live記憶體；25個原live檢查點與原先v3完全相同。13候選完成來源的角色與24×24暫存區均零差，6363格靜態地形亦零差。入口為frame-review-v1.json及frame-v1/completed-receipt.json。初始4192批次、伸出武器與完整呈現生命週期仍待核對。
- GUI首戰同29699／31475退出0；73張actualX逐張讀，57洞窟圖的285中文HUD區域零差。初始寶石0，最早於cave-right-2-held-287892.png已為1，四次Left只再次確認1；不能將固定指令原版的pickup檔名當GUI拾取時點。86按放確認、末持鍵空，456固定輸入／29原件／存檔與mtime不變。最終raw與顯示比較使用固定原版色盤，範圍與例外見pixel-review.json，不稱全段GUI同狀態。GUI1的低姿與頭盔閃現原因未知，跳／蹲姿態未證實。
- 移動GUI初版30606／17515退出1，未送洞內Up或Down。reference誤把整張960×680視窗縮成320×200，混入工具列與頁尾，導致站姿gate假失敗。正確先裁960×600遊戲畫布後，根端與tester各自核原生(146,94)16×24區域零差；未增加容忍或改遊戲。初版460輸入與失敗原樣保留。
- 正常原版83623退出0，固定冷F7入洞後Left1m／release250k離門，再Up500k／落地2m、Down500k／release1m，12點、前7點與原v3完整狀態相同。五張根端目視見離地、落地、蹲及恢復，寶石0、金幣1000，沒有戰鬥。證據入口為movement-original-review-v1.json。移動GUI修正版沿此窄路徑，不重跑首戰或重擲掉落。

- 移動GUI修正版43055／65371退出0。正常Left離門後，當次actualX清楚顯示站立、離地、落地、完整低蹲、放開恢復；根端五圖入口root-movement-review-v2.json。生命條仍綠、寶石0、金幣1000，無交戰或再次擲取掉落。500固定輸入與存檔mtime不變；與首戰gui-v1分別保存身份，不將兩個不同狀態合稱同狀態整段重播。

- 移動GUI修正版共37圖，其中24張新SHA逐張讀、13張引用既有完整PNG雜湊；19洞窟圖的95HUD區域零差。金幣1000／寶石0及生命條初末零差，before-jump到恢復站姿的當次原生畫布零差。最終raw比較採固定原版色盤，非全段同狀態。原events的三筆movement動作source欄繼承combat.commands字樣，後處理明示實際依movement-original-v1.commands，不改舊script或events。

## 2026-10-19：洞窟角色下層、武器與HD離線研究

- 入口為 `workplace/cave-hd-v1/research-index-v1.json`。原EXE、SAR與dosgolem快照沿首戰固定身份。IDA輸入01B1段SHA-256 `12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93`，匯出沿combat-v1/ida-cave-v3.json，IDA9.4 binary base0等於實模式01B1段內偏移。Go1.24.13、Python3.13.15、cycles3000。
- 已證實：3AA9的原版角色下層連至同次4129／3123完成來源。必須在完成callback保存對應下層，不能在後來shot以最新下層拼接舊完成畫格。各完整1MiB與來源時間獨立保存。三觀測版本25live點與原首戰完整RAM、Regs、VGA、色盤、時間、檔案及輸入相同；左移離門／左出劍另補8共用前段相同。
- 勘誤R011-composition：較早cave_actor_v2只核產生的opaque像素，未證明覆蓋完整。全框重建揭露左右攻擊前手分支選錯與跳躍背手漏畫。依IDA3ABE..3B7D、3C05..3CDA修正後，首戰54份及左向13份的下層＋角色與原版暫存及VGA均零差。舊資料保留，正式城鎮010與中文消費端不受影響。
- 已證實：3F42將32×32武器下層存於VGA實模式A000:FA40，1024bytes。4013..407E模板為欄優先四欄各四列；圖像base取資源段B000的word，每8×8用16bytes、兩位元色號，0保留下層、1／2取01B1:4FF5、3取4FF6。首戰17份與左向5份全框零差，含重複來源。其他裝備、魔法與未見旗標不外推。
- 4192在兩條正常鍵序未觀測，不再假定是必經初始化；其他情境用途未知，不追無玩家依賴helper。實際已觀察302C、39A3及40F0等入口。
- 內建imagegen產生兩張1254×1254逐格圖與一張1254×1254連續岩面圖，完整提示詞與SHA在art-index-v1.json。逐格重複感強，連續圖改以循環地圖地址定位。圖塊黑色會直接覆寫VGA，不是通用透明色；ID3..8連續岩面欄保留原版空地與首末可見像素，內部黑點換材質。這是明示美術規則，不稱原版碰撞遮罩，素材alpha不改寫。
- 離線工具先移除原貌武器／角色，核對整張320×200可原樣重建，再畫高清地形並還原原貌角色／武器。四張完成來源各64000色號重建零差，角色／武器opaque及繪製區外零差。只複製五個已驗中文標籤並核對當次原版位圖相同；數字及生命條取當次完成來源。
- `tools/cave_hd_preview.py`以研究輸入唯讀、空白輸出另掛可寫的Docker命令重跑，四PNG與prototype逐byte相同。CLI不等於正式接入或GUI驗收；當輪011仍DRAFT，後續接入另見下一節。

## 2026-10-19：洞窟完成來源、動態格及回城勘誤

- 輸入固定為EXE `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`、dosgolem882檔摘要`2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`。IDA9.4匯出`combat-v1/ida-cave-v3.json` SHA `3e11f43d6826ccdcd3b31f8c8ab74e18b745d4e36c421d699e71f4b9f89cb519`；以下定位均為IDA binary base0／實模式01B1段內偏移，資源在11B1，不能當SAR偏移。Go1.24.13、Python3.13.15、cycles3000，原始資料與共用dosgolem唯讀。
- 已證實：703C／7136開始、7114／714C完成、4129保存主角來源。舊3123漏掉獨立武器更新；R011-lifecycle回填011至012及resolution-index-lifecycle-v1.json。`lifecycle-v2/`1082取樣，957當次完成、125批次保留；原版回城換場199像素改動按契約撤除。`full-frame-v2/`229完成事件整個224×144色號零差。主角資源與完成來源分開保存，未消費資源區可變，不能要求整段相同。
- 已證實：動態地圖低七位從ED20取底圖，C010指標指向16-byte記錄，+4選A030／A070型別表，+6低四位選5-byte動畫。資源4000四位元圖格、A000遮罩及4F8C色表依3599重建，395次／77圖格零差。只描述圖像來源，不命名怪物或推論規則。地形1..14、動態、主角、武器完整重建後才接受呈現與高清合成。
- 已證實：正常GUI原生日誌可冷重播。`input-replay-v4/`3319筆前後暫存器／時間相同，最終1MiB、色號及暫存器等於GUI。固定延伸捕捉FF39=255；獨立Python整區零差，334完成事件零差。v5首戰另見E8=255、回城另見FF3A=255，`actor-profiles-proof-v2.json`三來源各32256像素零差，角色171／194／240像素。圖像分支與原始位址見012，沒有給未知欄位附玩法名稱。
- 限制：v5兩條失敗GUI無最終live RAM檔。`combat-profile-replay-v6/`及`return-profile-replay-v6/`只稱逐筆日誌暫存器／週期相同及242＋74完成事件重建零差。回城68BA來源的主段、資源、extra與色號SHA另等於GUI已呈現來源，不能外推完整live RAM對拍。三個profile只接受0／255且最多一個非零；0093／FF41與其他未驗組合仍拒絕。
- 已證實R007-return-resource：Town主角模板最高ID45，實用資源6000..68A0及8000..D800不變，68A0..8000保留區回城可變。007／008與return-resolution-index-v1.json附原定位與審查，不取消程式／NPC／遮罩守衛。
- 已證實R009-ground-phase：`return-replay-v5/last-town.memory` SHA `db512d91d79d856b8a94779c573623c8a4d1aebc48fbaa7d96f1dfb9bac12e9f`，完成色號SHA `c0e8fd4c11b3c295f53096dace6c4f1ac3956181101c5a7a164ba83d0c2b6023`。camera179時上下兩帶相位皆0，原公式差970／926像素。009保留原行走證據並附勘誤，改驗每帶唯一完整圖樣與RGB；零個或多個相位均拒絕。初始化producer仍未知，不阻塞顯示實作。
- READY審查分開保存：ready-layer-review-v1、ready-terrain-review-v1、ready-ff39-review-v1、ready-actor-profiles-review-v2、ready-ground-return-review-v1。素材仍沿三份內建imagegen完整提示詞及art-index-v1，不重新產圖，不提交素材／原版／字型／收據。

## 2026-10-19：原版增量覆寫與寶石驗證勘誤

- 輸入、工具及位址空間沿上節。已證實：v6首戰116像素失配來自原版3316..3319在主角之後寫VGA。`combat-write-replay-v6/original-writes.json`保存148筆實際寫入，分列實體VGA位址、oracle IDA位址與01B1偏移。3272靜態格以DX乘2定位，整8×8覆寫；3599動態格以DX定位，底圖0整格覆寫，其他底圖依資源A000遮罩的0位元覆寫。只補靜態格仍差50／51像素，恰為動態格的後寫像素。
- `native-paints-proof-v1.json`獨立Python重建三個實際來源各32256像素零差。首拒絕來源memory SHA `ee0f321df3579a1d1725b60a806a38a1f2efde9d858bd14c28a5c4e077da16a8`、body SHA `32828873904c436a8271db7e890fc95a21a51d66a0063f2dfcadf5cc3207bbdb`。依真實寫入扣除舊主角像素，不由最終顏色猜遮擋。READY審查為`ready-native-paints-review-v1.json`，正式實作再驗64筆上限、來源／時間順序、深複製及原子退回。
- `current-journal-pair-v7/receipt.json`兩個原版冷啟動重播3349筆日誌，每筆完整RAM、暫存器、VGA／色盤及虛擬時間相同；最終另等於原失敗GUI實際RAM及色號。267完成畫格零差、零拒絕，674次HD繪製包含30次增量來源。未知分支仍退回，不推論敵人或掉落規則。
- R-gem-reference：v7工具將`advance-strike-2.indexed`的寶石0誤當作1。該ROI與cave-entered的0相同，色號SHA `dc8686f5ea84b3bf3a77759d8e6f4084c7286d762ba215045b3123114550f2b2`；正確`pickup-attempt-left-1.indexed`的1為`ffab415a8d3c239e8a6dccf0d40e50bc04f42a058d703ec8719c224b6d5afed0`。撤銷舊GUI的first_gem_bitmap判定，HD／輸入證據各自保留。
- 修正工具v8沿同候選0227c040，不改原版或種子。移除持鍵中抓圖的額外耗時，以正常按鍵完成拾取。`gem-reference-review-v8.json`獨立核對原版0反例、正確1參考與實際X視窗：第一拾取1，最終2與當次原版色號相同。300完成事件零拒絕；`pixel-current-v10/`另核該次原版全區重建及290304高清像素零差。原錯誤收據不覆寫，不以碰巧通過的結果替換原版日誌。

## 2026-10-19：其餘片頭來源與魔王宣言中文

- 固定原版EXE、SAR1及dosgolem身份沿[規格013](docs/spec/013-intro-narration.md)。新正常段SHA `2e9158ad005957b2d298e6ab970e182a6cd1a71427fbcee4525d09f2071319fe`，65,536 bytes；IDA9.4以binary base0匯出，對應實模式01B1段內偏移。SAR偏移另列。Go1.24.13、Python3.13.15，所有工作在既有非root Docker內執行。
- 已證實：冷啟動無輸入，首章之後接魔王三行、再進標題；正常200m送Space100k／放開後，可自然走製作名單與故事。宣言用62D1／62FD消費時序碼及FF換行，不走32C9捲字端。631E開始字形，6331以AH2在X+2／Y+1繪陰影，6341以AH7在X／Y繪前景；6346完成後原版才加X8。FF77=255映射色號34與119，不套用其他介面AH語意。
- `threat-proof-v1.json`由獨立Python依原版F500指向F506的8列字型重建，88完成來源各9280原生像素零差，首字前整區為0。字型768bytes SHA `8ff0729de496f9d84d4cb3e4931a349a028981232507939f23af95a821d955d6`；來源／程式碼／字型與固定三行位置同列索引。1..4時序碼保留原樣，不推論音效或玩法。
- IDA首輪JSON因預設ASCII編碼失敗，修正UTF-8後另存v2。v2強制6300切點落在指令中間，改由原始call目的62FD與631E重建v3；錯誤邊界不當作函式語意。v3原始匯出SHA `03ee16a05544bf770adfac8eafa3a082d9d836ab5b105b8c2c245c186ed55311`，保留原名、位址、bytes與未知警示。
- `ready-threat-review-v1.json`核來源、字型與三行中文容量後，013才轉READY並接唯讀覆繪。正常雙原版88字形完成點及200m末端完整狀態相同，88陰影途中呈現通過。正常GUI末端另冷啟動、依同50k週期與音訊取出重跑，全部RAM、暫存器及VGA等於實際視窗；`gui-pixels-v1.json`用獨立字模重建，宣言與標題各576000遊戲像素零差，25932文字像素為明示中文差異。
- `story-source-v1.json`核5083個正常6A75字形，原字元、SI及SAR偏移逐筆相同，位置唯一、無漏字或重複，共374原文字面片段。最初P後接原版清框，不當作可見敘事。末端Indexed黑圖尚不證明正常UI結束，完整故事／名單中文與控制碼生命週期仍待接入，不能把374片段稱為374段劇情。

## 2026-10-19：故事正文來源、字形及中文生命週期

- 固定EXE、SAR1及882份dosgolem來源沿[規格014](docs/spec/014-intro-story.md)。原版冷啟動200m按Space、持100k後放開，總預算2.2b指令。Go1.24.13、Python3.13.15、cycles3000。IDA9.4沿前節v3原始匯出，binary base0等於實模式01B1段內位移；SAR檔案偏移另列。
- 已證實：故事池SAR `[1A6A,3104)`，對應01B1:`[79C6,9060)`。5786原始bytes依序消費，5083字形唯一。6A8B觀察讀byte，6ADE開始陰影，6AE3陰影完成，6AEF前景完成；6BFD／6C0A觀察FE清框。52清框各18240色號全0；47頁有正文，另五頁無正文。初始P以前景0／陰影0畫黑底，確實不可見。
- 已證實：FA、FB、F9分別取陰影／前景0／7、1／7、2／6，FF77=255轉色號乘17。F7、F3、F2、F1設原版第0、1、2、3列。依清框、字色及明示列分58片段，原版音效碼仍留原始池。374字面片段是先前盤點的細切，不稱374段劇情。
- 獨立Python用原生8×8字型重建5083字形的前圖、陰影及完整前景，15249次文字全區雜湊相同。字元主位置X2..315，部分字模依原版線性VGA位址跨列，實際寫入仍在文字區。58片段的來源、字色、明示列及52清框均核對。入口為`story-flow-proof-v1.json`，原版完整初始段SHA `0899d08fba7a6c4ab9d21b11bd7c4c11076e57af55451d75fc29b32064daa683`；各子區雜湊列於014。
- 58片段TSV與47頁獨立容量／版面核對後轉READY，審查保存為`ready-story-review-v1.json`及`ready-story-spec-v1.md`，再接IntroStory。譯文不進入原版比較、查找、存檔、時間或亂數。固定全頁排版後依已完成原文字元比例揭露中文，原版字形途中只接受前圖／陰影／完整像素，原版清框完成才換頁。未知來源或畫面整層撤除，不把未驗控制碼命名成規則。
- 頭段首章SAR `[1094,13DB)`為30行；故事後SAR `[13DC,14D2)`為六行，製作名單SAR `[14D3,1A69)`為51行，各後接FF。六行由sub_6D04設定原生Y20、高160；首章／名單由32C9／300E／332C使用Y32、高120。六行不能直接沿用首章硬編碼視窗。這兩組仍待中文接入，原版結尾保持13h，不把較早黑圖稱為已驗回新局。
- 正式雙原版`story-pair-v1/`共163點完整RAM、Regs、VGA／色盤、時間、FileOps及KeyReads相同，5083字形／52清框／58片段零拒絕。58陰影途中及46清框途中呈現另驗；47頁與先於實作的獨立Python排版整個遊戲畫布相同，文字區7715520像素零差。
- 固定候選574094bc／80來源的正常GUI逐頁完成47頁及故事後撤除；同一原生日誌由冷啟動重跑42895筆，每筆Regs／時間相同，末端全部1MiB／Regs／VGA等於GUI。首頁後Space跳過另驗9917筆；原版淡出改圖造成一筆`unobserved display write`，中文安全撤除，保留拒絕而不放寬來源守衛。兩路共50張actualX遊戲畫布，以各路冷重跑原版色號／色盤及獨立中文字模核28800000像素零差。人工只抽10張，清單在`story-root-visual-review-v1.json`。014僅故事正文限定CONFORMED。

## 2026-10-19：三組捲字完整來源

- 輸入EXE、SAR1、882檔dosgolem快照及IDA9.4原始匯出沿[規格015](docs/spec/015-intro-scroll-blocks.md)。以下程式位址為IDA binary base0／實模式01B1段內偏移，SAR檔案偏移另列。Go1.24.13、Python3.13.15、cycles3000。正常冷啟動200m Space／持100k／放開，固定觀察3.2b指令，不注入狀態。
- 已證實：首章30行、名單51行、收尾6行，各有一次FF來源，共90來源事件。4511起3200bytes含8列字形及2空白列。3323原始bytes `02C01AD202C01AF6C3`依兩次ADD的carry，以SBB DL／DH展開原生字模為00／FF；每字8像素。90個完整字形緩衝皆由獨立8×8原生字模重建相同，不從中文譯文推論原文形狀。
- 已證實：332C先將21B1整個65536-byte背景緩衝向上移320bytes，65216bytes被複製，最後320bytes保留；所選字模列寫入視窗下界之外。完整VGA視窗取前圖AND99h與捲後背景AND66h合成。1300次完整背景緩衝及64000色號畫面各自雜湊相同，區外無差。首章430列、名單640列、收尾230列，其中FF各10列後再送120／120／160空白列。
- 已證實：全部90份正常來源的01B1:`[6FF0,79C6)`與SAR`[1094,1A6A)`相同，SHA `994c9f06daed0ce9715a5984f1b3ed0dfc1e265c69f797fdbd8ebc1db0dd7009`。32C9..33B7程式SHA `504241bf9c5662c6e2ee2059e11f3915468d94035e7de83dc2860fbf15537623`，字型身份沿014。原版收尾在2171445528指令、2248381885cycles的6D55返回，保持13h；這不證明其後完整遊玩流程。
- `scroll-blocks-proof-v1.json`與`preview-scroll-blocks-v1/receipt.json`分開保存原生證據及獨立中文排版。三組87行試譯皆有字模，單行最大寬384／432／276顯示像素，90張離線圖通過。根端只目視READY審查列的六張離線圖。審查及READY全文先保存，再接正式程式；正常原版配對及實際GUI另驗。
- 用語勘誤R015-buffer-name：本節較早稱CS+2000h為「背景緩衝」，現統一稱「原版離屏緩衝」。3359移動的是21B1的65536bytes；33A0的99h背景位元來自既有VGA，33A4的66h文字位元來自離屏緩衝。原始位址、bytes與雜湊不變，規格015已修正用語，READY歷史全文保留。
- 正式三組原版配對95點完整狀態相同，1300列零拒絕，90張中文畫面與獨立排版相同，90次真實AND後途中與四種原子退回通過。正常GUI同載五組片頭中文，三組全列與故事5083字形均零拒絕。44826筆原生日誌另冷重跑，每筆Regs／時間及末端完整RAM／VGA相同，1390原生來源／完成列事件獨立保存；九張actualX遊戲畫布5184000像素與獨立合成零差，八張有文字圖皆有正差異對照。015只對固定MCGA三組及故事共存限定CONFORMED，不外推片頭美術或其後完整遊玩。

## 2026-10-19：共用物品文字及Magia正常使用

- 固定輸入、SHA及工具位址空間見[規格016](docs/spec/016-shared-inventory-text.md)。Go1.24.13、Python3.13.15／GUI3.11.2、IDA9.4既有27函式匯出，dosgolem882檔摘要`2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`。以下AAB8等為IDA binary base0／原版實模式01B1段內偏移，不是SAR檔案偏移。
- 已盤點：SAR2檔案401E起3613bytes，AAB8、AAF3、AB62、AC32、ACD9、AD67六表與九個固定來源，去重後72非空字串。字型分類為44白色大字、四彩色大字、24白色小字；原版空第二行不另造文字。既有IDA A35B bytes `8bb732ac`補上AC32讀取端，原相對運算元保留在匯出；未正常走過的語意保留強推論。
- 已證實，限本輪正常路徑：Magia買入金幣1,000→0，持有槽00A6..00AA由全零→`0500000000`→全零。SAR檔案004CB4名稱與004BF0使用句走AA2B／AA30、AH1；白字9、無陰影。兩路原版共61次文字完成事件，F500的8×8與F504的五像素步距字型逐筆重建，全幅320×200零差，來源、前圖、後圖與字型段各存本地。不據此推論Magia效果或全部物品玩法。
- 已證實，限所見狀態：Sabre售價1,200，起始1,000的購買分支拒絕，未取得或使用。正常首場交戰取得第一寶石，走回同一洞門、回城再進銀行，報價1寶石6金幣；未確認換錢，不推論完整兌換公式、掉落規則或亂數。證據保留shared-sabre-v1及shared-funding-v1..v3。
- 手冊SCAN0357_007印刷9–10頁與008印刷11–12頁核對商店、裝備、Enter、方向鍵及Space操作；五張既有掃描逐檔核對原RAR大小與CRC，未重新解包。手冊未提供逐項物品中文名，專名保留原文，通名為試譯。Kashi／靈藥及Acero／聖水按整組雙行翻譯；72個表內來源皆通過字型及原文寬度，新增60筆，原48筆不改。
- 016先DRAFT，再以來源／字型／原生完成圖／容量審查轉READY，才接指標表分類與試譯。新Magia49點全部1MiB、Regs、VGA／色盤、時間、輸入及檔案操作相同，存檔SHA／mtime不變；5616次50k呈現及兩種原子退回通過。首輪測試錯把完全遮擋的舊標籤數當可見殘留，保留失敗，改依實際像素還原驗收，未改正式繪圖邏輯。
- 固定候選4766810c／85來源的五張完整960×600與獨立Unifont排版共2880000像素零差。正常GUI v1城鎮面板通過，但舊工具對四次Right8m後的位置假設失敗，保留原身份；v2以八段有界正常步行至已驗角色範圍，仍嚴格核洞門及角色像素後送Up。五個穩定面板共1484820像素零差，根端逐張人工讀圖，退出0、持鍵空、存檔不變。只開標籤中文，不宣稱整段GUI全RAM同狀態或本輪HD重驗。

## 2026-10-19：洞窟主角25姿態與原版配色保留

- 輸入為[017](docs/spec/017-cave-player-hd.md)所列固定ZELIARD.EXE、SAR2及dosgolem882檔快照。Go1.24.13、Python3.13.15／GUI3.11.2；本段4129、4FED等為原版01B1段內偏移，資源11B1，不是SAR檔案偏移。沒有新增反組譯或推測性改名，沿既有012來源與九格／手臂解碼。
- 已證實，限來源盤點：67份角色暫存區證明，加六份全幅來源為73份、14個完整色號／透明姿態；不是73個姿態。另三個原版後畫來源的0bd51c6d只剩底部八列，九次真實呼叫已完全蓋住人物，最終Actor0。先補畫第15姿態是錯誤判斷，原失敗、圖面及改動0結果保留，不算新增可見HD。
- 已證實，限正常鍵序：23姿態候選的首戰25點與回城10點完整原版狀態相同；183個呈現來源全有對應。九個新增姿態與原14的整個576bytes沒有精確平移／鏡射等價，須新美術。初補圖底列朝向不符，透過內建imagegen修正後才採用。
- 已證實，限固定GUI：23姿態GUI交戰遇2b975f19而安全回退。保存原失敗與3099筆輸入日誌，從冷啟動重播，每筆完整RAM／Regs／VGA／時間及最終GUI原版相同；206完成來源、519次高清繪製零拒絕。補出8e1b514c與766bb52b兩個新行走出劍姿態；索引在`player-gui-missing-v1/index.json`，三份原始遊戲區各32256像素獨立重建零差。
- 已證實，限顯示資料：2b975f19與134def85的完整576bytes透明及幾何一致，各原色號有唯一映射。C2=2、E7=128、FF36=255；不據此稱受傷、無敵或推論持續時間。該畫面保留原人物色號，普通HD不能蓋掉原版配色提示，亦不算HD匹配。
- 25姿態採三張固定1254×1254 RGBA圖集；原23直接取樣不變，藝術來源由固定裁切直接縮至原版包絡三倍。原版完整來源／色盤／32256像素先驗，下層、真實後畫遮罩與最後武器順序保持。三圖任一缺失、損壞或身份不同即整組拒絕，未知姿態保留原貌。
- 25姿態單元93具名pass、無fail／skip，22份獨立全幅6386688像素零差，含18可見HD、三完全遮擋與一配色原貌。這證明已列來源與合成，不能取代正常視窗或全遊戲完成。原先圖面、alpha裁切、測試工具與GUI失敗均保留本地身份，入口見CONTEXT。
- 已證實，限c0e66326／86來源：正常35點完整原版相同、183完成來源皆有圖；同一3099筆日誌重播206來源／519次合成通過，兩缺圖已補、原配色保留。三路GUI及Y78半空補拍通過；兩份洞窟與回城地面612864高清像素零差，actualX與程式擷取相同。半空v2早拍站姿無效，v3使用正常Up放開後工具列暫停，原貌還原與HD重複零差。根端實圖目視及各版本工具身份見CONTEXT，不外推整段逐幀或全遊戲HD。

## 2026-10-19：洞窟十個武器來源與五個主角增量

- 輸入SHA與位址基準見[018](docs/spec/018-cave-weapon-hd.md)。Go1.24.13、Python3.13.15與GUI3.11.2，沿既有IDA9.4證據，不新增推測性符號。3F42、3FD0..4085、4FF5..4FF9及FF44均為原版實模式01B1段內偏移，圖像資源段11B1，不是SAR檔案偏移。
- 已證實，限正常28點鍵序：229個選定完成來源中48份有武器、181份空白，十個唯一32×32色號／透明鍵。獨立Python逐個重建完整32256原生遊戲像素及1024武器bytes零差；模板欄優先、2-bit透明與原黑色0保留。來源、主角下層與完成時間分開記錄，沒有拼造快照、注入狀態或重擲。
- 已證實，限新主角來源：左向三鍵及連續動作兩鍵的整個576bytes不等價於既有姿態的平移／鏡射，分兩張PNG補圖。原25登錄與取樣不變。三格初圖的多餘手臂已拒絕，局部修正後才採用；新素材只供顯示，不作玩法證據。
- 十武器圖面保留原生非透明外包絡，直接取樣到96×96位置；三十主角直接取樣到72×72位置。原版全區與來源先驗，再依當次下層、地形、動態、主角、最後武器合成。主角原貌、已驗配色與HD分開處理；未知或空白來源不強留舊武器。素材SHA、固定裁切、RGBA、完整提示詞與拒絕產物沿CONTEXT索引，只留本地。
- 已證實，限d8666d0a與凍結89來源：正常28點首戰／左右動作加10點回城的完整RAM、Regs、VGA、色盤、時間、檔案與輸入操作相同，存檔SHA／mtime不變。238個有限完成來源中48武器已知、190空白，主角與武器均無缺圖。125具名測試無fail／skip；29份武器組合加22份既有主角全幅共14805504高清遊戲像素與獨立合成相同。這些有限取樣不證明其他裝備、全動畫、動態物件或其他洞窟HD完成。
- 已證實，限同候選的實際GUI v4：兩路正常退出，左右出劍、收劍、切換及Up回城撤除通過。2837筆實際原生日誌冷啟動重播，每筆完整原版狀態與最終GUI RAM／VGA相同，107完成事件、268次合成、759筆FileOps、零來源拒絕。按觀察器步數130716645及134389083擷取的兩個可見武器，完整來源／主角下層／原生圖SHA及兩個真實時間相同；不能以其他站姿來源替代。
- 實際左右兩幅遊戲區與回城地面共612864高清像素零差，兩份原貌區與原生來源相同。左向最終及回城actualX與程式擷取相同；右向由指定來源獨立驗證。六張actualX根端目視。原貌X圖採索引PNG，第一版像素工具只支援RGB／RGBA而停止；補驗證解碼器的PLTE／tRNS及CRC後，同批圖v5通過，原版／候選／89來源未改。完整美術提示詞與最終索引見CONTEXT。

## 2026-10-19：洞窟動態物件來源及四圖登錄

- 固定EXE、SAR2、882檔dosgolem摘要與位址基準見[019](docs/spec/019-cave-object-hd.md)。Go1.24.13、Python3.13.15／GUI3.11.2，沿既有IDA9.4的36D1、36D8與3599證據；本段C010、A030／A070、4F8C均為01B1段內偏移，圖像4000與遮罩A000為資源11B1段內偏移，不是SAR檔案偏移。
- 已證實，限正常首戰／左右28點鍵序：229有限完成來源取得352次物件觀察，去重28個完整16×16色號／透明鍵。原始type1／2／10／18／20各8／12／3／3／2圖像來源，不據此命名怪物或效果規則。獨立Python逐個核完整32256原生遊戲像素與256物件bytes零差；色表、模板、四底圖、原始record與兩個真實時間各留索引。
- 已證實，限所見來源：四底圖皆0，因此16×16全畫，黑色0是原版清黑。部分物件由畫面外進入，不能以出界概括拒絕來源。019逐筆原順序繪製，已知來源先清原矩形再畫48×48高清；未知依自身Raster及255透明重畫，0仍可見。主角後畫遮罩與最後武器不變，矩形與HD都裁在原遊戲區。
- 28個來源沒有完整256bytes的直接鏡射等價，未借圖；分紅色、綠色與特效三張PNG。既有後畫來源combat-paint-replay-v7/source-01另有type2／phase2、record01B1:D79E、模板01B1:A128，完整鍵fc4bd4ad。來源完整SHA與原始bytes見019；另外補第四張單格圖，原28登錄與直接RGBA不變。所有圖由內建imagegen生成，原PNG與alpha不修改，完整四份提示詞、原始工具產物及SHA留本地。
- 29來源固定裁切的主體未碰格界，直接取樣至原版非零包絡三倍；原清黑矩形與新美術alpha分開。35份獨立完整圖涵蓋28來源、四原貌主角／武器分支及三真實後畫，78物件皆有圖、51部分出界、未知0，先審READY才接Go。173具名測試無fail／skip，35幅10160640高清遊戲像素與獨立Python相同；裁切、重疊順序、黑色、未知與整組缺圖／壞圖拒絕通過。
- 已證實，限2ac1b51c／91來源：正常首戰／左右28點與回城10點完整RAM、Regs、VGA、色盤、時間、FileOps及KeyReads相同，存檔SHA／mtime不變。238有限完成來源含352已知物件、230次部分出界、48空白畫格，未知0。這不證明其他配色、非零底圖、全部敵人或全遊戲HD完成，實際GUI另驗。
- 已證實，限同候選GUI v2：正常首戰／寶石1與回城兩路退出0，原貌／HD切換與撤除正常。3240筆實際原生日誌冷重播，每筆原版完整狀態及末端GUI RAM／VGA相同，243完成／613合成／759 FileOps、零來源拒絕；309物件皆已知，193次部分出界、94空白畫格。兩指定來源136994278／153002060的完整來源SHA與真實兩個時間相符；兩洞窟及回城地面612864高清像素與獨立合成相同，五張actualX目視。v1第二段右移取樣假設失敗保留，v2只移至已記完整物件的前一段，不改原版鍵序或步數預算。
- 已證實，限主角顯示資料：新c29c7a73與pose1的011099c1完整576透明幾何相同，雙向色號映射唯一；既有2b975f19同樣對pose0。兩者FF36=255，c29另有E7=128、FF40=255、FF3F=3，不據此命名受傷／無敵或推論時序。保留原色，不算HD匹配。兩個真實來源再以正式繪製器驗580608單元圖面像素零差；未保存cycles的缺圖傾印只用明示單元時序邊界，不稱新原版時鐘或actualX。真正原版RAM／時序由原3240日誌獨立證明，程式91來源未改。來源與完整提示詞索引見CONTEXT。

## 2026-10-19：Malicia至Muralla入口020

- 已證實，限CONTEXT的portal-source-v1/index.json：38去重真實完成來源，16完整入口、3左側部分出界、19不在視野；26657個未遮擋入口像素與完整原生遊戲區零差。入口是5×5靜態矩陣，21個非零id及各48bytes串接1029bytes，SHA 3fd046b8c1281a013bfe1c8dfa627a663fec2eac5cebe7e3df025a63c1db6633；原生40×40 SHA 9374206133321e2cb058c7c94512e230cde8183afc6896282dd94989180fb716。
- 來源為01B1:FF31環形地圖及01B1:ED20底圖索引；圖塊為11B1:8030+(id-1)×48。這些為dosgolem實模式段內偏移，不是SAR檔案位址。Go1.24.13、Python3.13.15／GUI3.11.2，沿012的IDA9.4既有證據，未另深挖硬體。EXE、三SAR、882來源快照及工具版本由020與CONTEXT固定。
- 已證實，限圖牌與正常返回路徑：原圖可讀MURALLA，正常Up回到Muralla鎮。SAR2檔案0399B8的12bytes是Muralla Town；中文沿既有HUD鍵zelres2.file.0399b8。內建imagegen圖面字牌留白，原PNG與alpha不改，固定裁切直接120×120；Unifont字模直接畫高清文字，先量寬、陰影及邊距，不以裁切或省略字處理溢出。
- 已證實，限0d70040a／93來源：232具名回歸無fail／skip，57份獨立全幅16547328高清像素零差，正常38點全部原版與存檔相同。首戰v2及回城v1退出0；3241實際日誌冷重播268完成／677合成／759 FileOps，每筆原版及末端GUI相同。四指定來源SHA與真實時間吻合，四洞窟加回城1193472實際像素與獨立合成相同，五張真正X圖已讀。
- GUI v1的尾段恰等寶石1假設被原始圖推翻，同批記錄早段1、尾段2；保存失敗、獨立位圖核對及離線修正，不將v1改名為成功。像素v1工具變數衝突亦保留，v2以同一來源重驗。c29c7a73沿019已證實配色保留，不新增受傷／無敵推論。其他入口、洞窟、全部中文與HD未證實完成，額外HUD3個跳過不計驗收。完整圖面與提示詞索引、來源SHA及稽核入口見CONTEXT。

## 2026-10-19：正常昏迷送返長老021

- 已證實，限`exploration-source-v1`及同鍵序冷重播`exploration-return-source-v1`：17點完整1MiB RAM、64000色號與768色盤相同，174300000指令及原版時間相同，自建存檔SHA／mtime不變。原版自然走到昏迷、送返Marid、四段對話並回鎮，未注入生命或金幣，也未恢復快照。
- 已證實，固定SAR2 SHA `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`：四字段的檔案範圍為01D72E..01D76C、01D76C..01D79C、01D79C..01D7B2及01D7B2..01D7D5，右端不含。模組檔案起點01BCC6對dosgolem實模式01B1:A000；字段相應BA68..BAA6、BAA6..BAD6、BAD6..BAEC、BAEC..BB0F。兩種位址空間分開記錄，完整script SHA列[021規格](docs/spec/021-exhaustion-return-overlay.md)。
- 已證實，沿既有IDA9.4掛鉤定位及Go1.24.13固定882來源：337事件含711B／7120各157、71A7／71AC各10、7205／71C9／71D9各1。157原字元、指標、DS、AH、座標與711B的`2E FF 16 22 20`均相符。只在首字段前清框，第二字段後待續，第四字段中捲動，前文不在字段間清除。末尾11 FF FF與金幣／配色旗標的完整語意不由此推定。
- 已證實，置中(112,23)-(208,87)的兩幅完整6144色號與既有左側(56,23)-(152,87)完全相同。兩SHA為`bbc79b87bc9d9b3bb59a049d5acd8206a7b9dfb287d557b71437672544d3f16c`及`0762364b30958db2e1948f850d449ff042a8f736ea5b62d8674be9296c90a66a`。沿既有長老PNG SHA `a1fa3b745422f59153239b903591db104c5c3fb320c18ee7f5181a247c33a5aa`，保留外框及四動畫像素，不新增圖面或推測動畫語意。
- 已驗證，四段42字的獨立三行版面及九整圖經READY審查後接入。正式配對17點全部原版狀態、檔案與存檔相同；九幅5184000高清像素零差。回鎮原版清框造成唯一已確認kind2撤除，沒有放寬產品守衛；其餘拒絕零。真正GUI及其原生日誌、像素另記，不以此取代。原版資料、圖面與收據只留本地，入口見CONTEXT。


- 021限定完成證據，3ae68018／96來源：正常17點全部原版狀態與九整圖5184000像素相同；真正GUI v6六時刻已目視含完整Muralla實景，3849筆每筆原版及末端GUI相同，850 FileOps，六時鐘及全SHA相同。六時刻中文框／肖像的兩HD及原貌2764368像素與獨立Python相同。只有最後清框kind2撤除，正常first_difference0，實際430，不稱零拒絕。原件及存檔SHA／mtime不變，收據沿CONTEXT的exhaustion現行入口。
- 已證實，限兩正常GUI的完整RAM比較：城鎮十組守衛中，只有dosgolem實模式11B1:8000..D800來源不同，差異五bytes在AE60、AE62至AE65；完整輸入SHA及差異在exhaustion-town-source-v1.json。五bytes用途未知，不能稱未用區，也未放寬守衛；送返後城鎮目前原貌。v4黑畫面被根端目視退回，v5距兩肖像各3及1像素的動畫中途也保留失敗；不以退出0或工具失敗改名宣稱正常完成。


## 2026-10-19：送返城鎮的圖塊尾端022

- 已證實，限固定Muralla profile及return-town-footprint-v1.json：完整215×8地圖與居民底圖最高ID239，地形資料為dosgolem實模式11B1:8100+ID×48。AE60等五bytes屬242號塊，未被這個profile讀取。原版17點全RAM／色號與既有來源相同；靜止、右行、左行44355次原生讀取均不在AE00..D000。原始指令IP、段內／線性位址、時間與數值留三份reads.jsonl，監看期間沒有觀察器自行讀取該區。
- 已證實，五份完整原版圖層共71680色號獨立重建相同。共同前綴與地形使用區8000..AE00 SHA為6cddbaa678370307422a6a60798c0d75a53c46deacc1a9f299a716c40d234b45；atlas8100..AE00 SHA為889318f3979aaa81803facbf6d14ed4f7f45c78b4a7c2fe2881d64ea1e5da8c2；完整遮罩D000..D800保持0bdb78b2702490f3309f72ea3046fda2417bd088a9a6eec1c3f179fcad4cef71。範圍右端不含，均為實模式段內偏移，非SAR檔案位址。工具Go1.24.13、Python3.13.15、固定882檔dosgolem，原檔SHA見022。
- AE00..D000只排除於已驗Muralla圖形profile；寫入原因未知，其他城鎮與ID240以上未擴張。人物／主角圖像、遮罩、程式及原生逐像素重建仍嚴格驗證。五份四圖層的獨立合成2880000像素零差屬單元圖面，不當作真正GUI。來源與READY審查沿CONTEXT的return-town入口。


- 022限定完成，d72e260a／97來源：34具名、五整幅2880000高清像素、GUI v2四HD層恢復與原貌切換通過。3775筆每筆完整原版及末端GUI相同，850 FileOps，七張實際X圖已目視；城鎮810432及同輪長老／中文2303640實際像素零差，呈現來源時間與原CPU保持分開。來源與稽核見CONTEXT的return-town現行入口。只有已確認原版清框kind2撤除，不稱零拒絕；未知ID240以上仍拒絕，沒有變更原版玩法或存檔。

## 2026-10-19：正常攀繩與下層主角來源023

原始EXE與三SAR的固定SHA沿CONTEXT，dosgolem882檔來源摘要仍為2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada。工具Go1.24.13、Python3.13.15；實際X使用既有psychicwar-go-ebiten與Python3.11.2。位址空間為dosgolem實模式01B1段內位移與11B1資源段，不是SAR檔案偏移。來源、完整RAM／色號／色盤SHA、原始欄位與工具身份見expedition-build-v2/input-identity.json及expedition-inventory-v2/index.json。

已證實：冷F7自建存檔、原版按鍵與146110000指令的18份來源，完整遊戲區各32256像素獨立重建零差。遮擋前24×24來源陣列以255表示透明；12個鍵中4個已有HD、7個需新圖、1個為新蹲姿的配色變化。七鍵、原生包絡及正常來源由規格023逐項列出，不借用相似姿態或鏡射猜補。每個新圖鍵至少一份正常來源有可見HD改動；eb7307bb在14-west-open全遮擋，15-leave-vine-left則有131個可見原生像素，不能以全遮擋證明補圖完成。

已證實：39ca5302與acde2d3e的完整576bytes透明及座標相同，每種非透明色號各有唯一雙向映射，原始FF36=255；配色來源保留原貌。FF38、FF39、FF3D、E8等欄位只按原偏移記值，不由外觀推論受傷、無敵、死亡或時間規則。下層探索未進下一洞窟，地圖摺頁不當作已走過路徑。

七格AI圖集的SHA為7691d0862c5a98e49112ef2022ec5ac47529afc16a07221dc44e3a661925a0c3，1254×1254 RGBA，完整提示與參考SHA留本地。固定窗口alpha主體外擴4像素，只裁切登錄不修改原PNG，原30直接RGBA不變。獨立18份主角合成與正式18份完整遊戲區5225472高清像素相同，實際X攀繩的290304高清像素另由完整地形／入口／物件／角色／武器合成比對零差。

實際GUI的3494筆原生日誌已從原版冷啟動重生，每筆Regs／時間及末端1MiB RAM／VGA／Regs／時間相同；原版與存檔不變。暫停CPU166889374、完成來源166861689及Body166770124各自保留，不混稱同時刻。觀察到的攀繩姿態33有匹配、原貌切換保留中文HUD、零洞窟來源拒絕；不外推七姿態每幀GUI、其他洞窟或全遊戲完成。

## 2026-10-19：來源存檔與盾牌文字的後畫數字024

原版EXE SHA為fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341，SAR2為bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1。五個來源USR的SHA沿source-manifest，正常冷啟動F7全部讀256bytes且停在Muralla，並非後段城鎮入口；來源取得過程未知。觀察沒有機器快照還原、記憶體注入或原存檔改寫。原型與輸出沿CONTEXT的source-saves與shield入口。

既有IDA9.4、equipment-held.segment SHA 2ecedadb544f1d089612ca19b1f42379fa83866c447121b131c61435fcdf7e3e。IDA base0對應dosgolem 01B1段內位移，SAR2檔案偏移另列。A7A5原bytes8BB767AD從AD67表取雙行來源；A7B8的2EFF163820呼叫既有五像素字；A7BD的E88400呼叫原始sub_A844，下一定位A7C0為F6069800FF。保留原函式及相對運算元，不重新命名或推論數值規則。

已證實：兩份來源SAVEM3／SAVEM各正常F7、開啟、關閉及重開，四組回呼的文字完成至A7BD原圖相同；A7C0後83個原生像素變化、範圍204,105至230,112，其中73在舊整列守衛內。字母區238,105至268,113共240原生像素不變，因此舊原貌退回由前置空白被數字寫入造成。六個第二行固定SAR2來源為004d96、004dad、004dc0、004dd2、004de4、004df9，前置5或6空白；完整raw及指標不變。

原版數字皆為(080)，數字玩法意義及公式未知。初稿「耐久」不作語意證據，READY v2追加勘誤且保留v1；沒有為此深入sub_A844。正式修正只以AD67第二行分類導出字母安全區，UTF-8 TSV仍108筆，原24個小字與大字的物理字型不改。六個第二行是其中的小字子類，不代表六種都正常取得。

10具名、八點新完整原版配對、Magia49點及5616次呈現、2304000獨立中文像素通過。兩GUI四面板1161216像素及重開時同原版4896個數字像素相同；4291筆原生日誌與末端完整RAM／VGA／Regs／時間相同。024原始定位另加入inventory-source-index，016保留明文勘誤與自動backlink護欄。其餘五種盾牌、混合來源與數字規則未升級證據等級。

## 2026-10-19：刀劍油正常使用與刀劍第二行025

來源為原包SAVEM1.USR，SHA-256 03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c。取得過程未知；原件唯讀，副本保存mtime，從冷啟動正常F7讀取，未注入或還原機器。固定EXE、SAR2及882份dosgolem來源身份沿CONTEXT與[025](docs/spec/025-sword-inventory-label.md)。Go1.24.13、Python3.13.15與既有IDA9.4匯出；IDA base0對應01B1段內偏移，來源鍵為SAR2檔案偏移。

已證實：原版選取刀劍油為SAR2 004cd9，使用前句004aca，名稱句004c20。原版正常使用後，01B1:00A6..00AA由0501010107變為0501010100；重新開關頁面仍保持消耗結果。現有TSV已能顯示訊息，無新增或內嵌譯文。數字含義、效果公式、正常買入及來源存檔取得過程保持未知。

刀劍第二行原始指標來自IDA A76E的8BB7D9AC、mov si,[bx-5327h]；A781的2EFF163820呼叫文字介面，A786的E8E500呼叫sub_A86E，返回定位A789為F6069300FF。九份正常回呼快照確認完字到呼叫前全圖相同。使用前兩組呼叫未改圖，使用後重開增加37個原生數字像素，範圍204,87至218,94；238,86至263,94的200個字母區像素均不變。因此舊含空白的整列守衛失配，分類六個ACD9第二行並沿024方式只覆繪字母區。原始定位、raw、前綴及來源鍵全部保留。證據在workplace/prose-v1/sword-source-v2/，實作前READY正文與SHA為sword-ready-review-v1.json。

來源v1只到使用前，原版呼叫沒有數字變化。工具對空diff取min失敗是驗證前提錯誤，保留preflight.json；v2加入已觀察的正常刀劍油使用及重開，不改金額、物品或亂數。15點刀劍油及8點盾牌完整原版配對、11具名測試、vet／build通過。五整圖2880000獨立中文像素零差；實際視窗與最終狀態另驗，尚未計入本筆通過範圍。

同一路徑的洞窟名稱仍英文。完成快照中的原版字形、字型與繪字機器碼吻合，C000前綴相較SAR3 AD02只差相對13／14兩個bytes，由1A00變為FFFF。差分已證實，是否在HUD入口導致拒絕仍為強推論。下一步只觀察22CD／2344及必要消費範圍，不放寬守衛、不追無關寫入原因。入口為workplace/prose-v1/sword-hud-preflight-v1.json。

025視窗補證：sword-gui-v2四張實際面板共1161216像素與已驗圖相同，根端目視使用訊息、刀劍油消耗及重開後原版(1)與中文劍並存。sword-gui-pair-v2冷重播3130筆原生日誌、最終完整原版相同；同一暫停時刻2448數字區像素零差。來源存檔及副本SHA／mtime保持，025限定CONFORMED。v1的Space鍵名失敗與隔離Xvfb錯誤證據保留，不算產品缺陷或通過收據。

## 2026-10-19：洞窟名稱來源前綴與只讀診斷026

輸入與工具身份見[026](docs/spec/026-cave-hud-source-prefix.md)：ZELIARD.EXE SHA-256 fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341，ZELRES3.SAR為9c19b9c317d3006e11f8a5cad8ea1f226f8043e11c6467fa27f10eef33175ef0，兩份USR的SHA、Go1.24.13、固定882來源及原版鍵序一併保存。實模式01B1段內偏移與SAR3檔案偏移分開，既有IDA9.4定位沿hud-source-index。

已證實：22CD至2344的兩次正常繪字，SI從D614至D629，讀取全部21byte來源；DS區讀取11545／11654次，C000前綴64bytes均0次。兩側完字SHA為2c3ad1e64e2227cb08a8a95cd7a17a9d67d7edc3ec1c58828769288074d86d8c。原靜態前綴3f5edbc8b51aa017211fb0ca808f807a1ade51d95e14b49dadf593e90ada24eb，SAVEM1完整前綴6c853b5dab1da5682a959dd804b0f23e5acdc3aa67099596335e1293694f03b2，只差C013／C014。新識別只接受此完整身份，未遮罩未知值。原欄位意義及寫入原因未知，不阻塞唯讀名稱覆繪。

正常導航補證：SAVEM1回鎮時原始0083為22；短左鍵100k未動，300k後為21，再Up入洞。此為本次位置及輸入結果，不推定完整碰撞規則。現行observe.json的native_primary_0083只讀原始值，實際GUI在已確認Muralla後正常左移22至21；原生日誌證明沒有注入。首輪未對齊路徑與原測試保留，當時入洞、裝備頁、回鎮均已顯示五欄中文。

現行101來源、ae87b6b7，23個完整原版配對點與十整圖5760000獨立像素通過。實際GUI五個暫停點及3654筆日誌冷重播，逐筆暫存器／時間及最終完整原版一致；342720完整HUD與279720原數字／區外像素零差，原始檔及存檔SHA／mtime不變。入口均在CONTEXT，原始資料及畫面只留本地。HD解碼仍拒絕FFFF前綴與0093非零，須另以原版圖層證據推進，026未證明該分支可HD。


## 2026-10-19：027持盾角色圖層與高清

原版輸入EXE `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`，SAR3 `9c19b9c317d3006e11f8a5cad8ea1f226f8043e11c6467fa27f10eef33175ef0`，原包SAVEM1.USR `03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c`，取得過程未知。Go1.24.13／Python3.13.15；IDA9.4的py312映像。IDA輸入cave-entered.primary `12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93`，binary base0對應01B1段內偏移，資源11B1與SAR檔案偏移分開。

已證實，限53份正常原版候選：0093=2經原始sub_39B9內loc_3D22分類1，改變前後手模板；後手、本體、前手及後畫／武器可獨立重建224×144區1709568像素零差。保留原始名稱、bytes、位址與ES:SI呼叫，不作推測性改名。原生讀取只見C012於01B1:315F／7054，觀察快照間隙執行零指令；兩個完整64byte前綴限定接納，不遮罩未知值。各檔SHA與完整362指令、22參照見cave-shield-ida-v1／v2，分級索引為docs/cave-shield-source-index.json。

17個持盾來源完整576色號鍵分別登錄編號37..53。原圖1536×1024，SHA `ea679274d3f144bef9841e660a26b0ec4a860957cb19989d98102c87f7f9633a`，直接高清取樣及完整提示詞見cave-shield-art-v3。獨立高清53整圖30528000像素零差。正常F7完整原版122次配對與實際視窗三次同狀態切換閉合限定顯示鏈；存檔與mtime不變，未注入或還原。原始數值含義、取得與特殊組合仍未知。

正常離洞的0093=2／FF3A=255先保持原貌。實際日誌3678筆，退回原因與真實候選分開記錄於cave-shield-gui-pair-v2，不把缺少新候選的守衛事件包裝為另一份FF3A畫格。v1重播重複引用同一候選的勘誤保留。此項限定CONFORMED，不外推其他盾牌、特殊姿態或全遊戲HD。


## 2026-10-19至20：028持盾離洞與批次前主體

- 原版EXE SHAfe1c00b3、SAR3 SHA9c19b9c3、SAVEM1.USR SHA03df9abc的完整值沿028規格與來源清單。來源存檔取得未知，唯讀原件、正常F7與保留mtime的隔離副本。工具Go1.24.13／Python3.13.15／IDA9.4，IDA原輸入primary SHA12d91918的完整值見028；binary base0對應01B1偏移，資源11B1與SAR檔案偏移分開。
- 已證實，限四份正常原版：sub_39B9的3B80固定610E本體；3C27分類1取7E偏移，前手6195／6237九格ID及完整576色號結果相同。原生ABI及完整129024像素零差，保留原始bytes／operand／名稱與位址。相同角色鍵a119ce2b...由單一姿態54覆繪，完整鍵及素材SHA見028與registration.json。
- 已證實：實際GUI v1日誌2580筆在118082163步具有來源守衛有效的新主體118073860步，晚於上一完成117902674步；原畫面改變且drawing=false，selectFrame整組clear導致後續兩次缺少主體。真實快照與完整雜湊見shield-transition-lifecycle-v1，不把舊RejectedSource當新候選。
- READY-v2後只撤除舊呈現並保留時間／來源有效的新主體，原版完成仍需整區零差。實際同日誌回歸新主體保留、舊圖撤除、拒絕2→0，原版最終RAM／VGA相同。十個生命週期邊界、3686完整原版配對與正常雙朝向視窗閉合限定鏈。
- 實際視窗3068筆原生日誌、五暫停原版狀態與四呈現來源相同，四次同原版HD切換逐像素零差。012／027回填索引為docs/re-resolution-backlinks.tsv，以固定模組SHA與原始位址定位。其他裝備與特殊組合仍未知，不外推全遊戲。


## 2026-10-20：029原版音訊輸出

- 沿復古規格閘門及硬體規格優先路由，使用固定882份dosgolem來源。EXE、RESOURCE.CFG與29原件保持原清單；源音訊為MSCSTD／SNDSTD，不替換音源。音訊時間沿oracle已知契約，屬硬體規格近似，不深入PIT／DAC／ISR。
- `audio-output-source-v1/`重播698筆正常冷啟動前綴，34825129 cycles得到511929個44100 Hz有號16位元單聲道取樣，整數時鐘公式相同；18610非零、峰值8192，PCM SHA621896c6。沒有原版寫檔、狀態還原或注入。
- 依本地Ebitengine2.9.9官方介面轉成浮點雙聲道，單向佇列不持有原版。實作前READY保存；八項競態測試及第九項完整原版698配對通過。受控轉換4095432位元組，原版及輸入／檔案操作不變。
- 沿既有eob-audio-capture:20261009-r3，SHA d831720f，建立容器內私有PulseAudio，未連主機。transport-v2直接編入正式audio.go與原版PCM，七個512音框片段相符；正常遊戲另跑GUI-v1，四生命週期及三種原版凍結通過。虛擬輸出與人耳／實體輸出分開。
- 實際710筆原生日誌重播逐筆暫存器與時間相同，最終完整RAM／VGA／色盤相同；520749生成取樣及WAV逐位元組相同。正常裝置監聽十二個512音框片段相符，左右聲道0差，全部18610非零音框可見。
- 品質限制已證實：實際重度觀察視窗欠載1904751音框，約43.19秒補零；滿載／暫停丟棄0。未測無觀察負載，不能把根因歸給觀察器，也不稱穩定播放完成。量測與修正節奏是必須待辦。人耳1.37秒試音已提出，尚待回覆，物理喇叭未驗。


## 2026-10-20：030原版RGB合成

- 同程式近似600／603正常run，無觀察仍慢於即時。capture-frames為擷取下限，ERRATUM明示非同狀態。研究用來源覆寫只加runtime/pprof，原600run、末端RAM及源WAV與無觀察相同。7.37秒CPU取樣中Draw2.56、SetRGBA1.51秒，屬已量到的成本；原版RunCycles2.97秒，共用主機時間有變動。
- 未載洞窟HD就不建立CaveFrames，故本次開場不能歸因於其1MiB快照。go tool pprof列CaveFrames無匹配，以及trimpath找不到原始碼的限制保留，不當產品缺陷；不為硬體時序重新逆向。
- 先存rgb-ready-v1.json再改既有SetRGBA九次寫入為一列RGB及兩列複製。五份028實際原版來源，整幅含alpha與區外3264000像素零差；兩具名及兩子項含競態偵測通過。三次微型量測中位數約9.2倍，零配置。
- cea16753／112來源固定後，正常音效GUI四生命週期及退出0，721筆日誌、最終完整原版與全部528344生成取樣／WAV重播相同；監聽左右0差、22329非零、六段512音框相符。42812欠載、22846滿載與1470暫停清空仍是品質待辦；不把不同鍵序／負載的差額當純加速效益。


## 2026-10-20：031音訊預取與正常節奏

- 沿規格閘門與平台規格入口，實際讀本地Ebitengine2.9.9／Oto3.4.0官方原碼。Play可同步預填，Read請求17640位元組即2205音框；兩個實際生命週期第一次只有735音框，舊碼補1470。源PCM與正式audio.go身份固定，競態偵測啟用；原版882份來源不改。
- READY先存stream-ready-v1.json，改空佇列等待、真實資料短讀、先備妥2205音框才Play，Close先喚醒來源。11具名含10項競態、698完整原版配對與vet／build通過，37d86d66／113來源固定。
- 獨立虛擬輸出整段39120訊號音框、18610非零與左右聲道相同；仍9924滿載丟棄，不外推全時間軸。實際正常GUI-v1四生命週期及退出0，708日誌與519279源取樣／WAV相同；前端補零／丟棄0，但0.887秒訊號被拉長到4.248秒。
- 另以原生日誌按放確認、不開每畫格1MiB觀察，707筆及518450源取樣／WAV相同，RAM／VGA／暫存器／時間相同，實際色盤未另擷取。非零內容與順序相同，但0.819秒拉長至3.442秒，排除只有重度觀察才失敗的解釋。
- 正常Space鍵序CPU剖析加入runtime/pprof、音效裝置off：24.67秒程序與8.77秒CPU取樣，RunCycles52.68%、Draw14.37%。無CPU取樣時間的原因尚未知，不直接歸因於主機或GPU。關閉垂直同步研究版確實加快流程，709日誌及源WAV仍相同，但遺失120非零音框，整段失敗；不進正式路徑。
- 所有complete-signal失敗及移除靜音後內容對比各自保存，不換名成功。不因來源計數0冒稱穩定播放，不改取樣率或擴大容量掩蓋。031尚未CONFORMED，候選不升主入口；接續量送樣／顯示等待／更新節拍。人耳及實體裝置未驗。


## 2026-10-20：Muralla六筆新交談與排程近似

輸入沿004的EXE fe1c00b3、SAR2 bf7c792a、882份執行器摘要2a5ad11a；Go1.24.13與Python3.11，字型20e8b505。自建SAVETEST SHA-256為78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb，mtime_ns725847808000000000。原版段位址基準為01B1，來源鍵數字為SAR2檔案偏移；本輪未反組譯或改名。

已證實：正常冷F7、左右與Space可觸及`zelres2.file.03a0f8`、`039f30`、`039a8e`、`039b90`、`0399f5`、`039e6c`，分別一／二／二／二／一／一頁，接續與關閉由原版處理。原版與中文19個完整狀態相同，實際視窗4510筆日誌、16個暫停點及9216000像素通過。舊039fff兩頁保持原驗收，合計七筆來源十一頁；不能由來源數推論獨立居民數或其餘106筆可達性。混合控制碼五筆語意仍未知。

工具勘誤：residents-pair-v1以JSON位元組比較typed Rectangle與map，Min／Max順序不同造成假失敗；v2保留所有欄位改結構比較通過，獨立像素亦確認。ERRATUM原文保留，confirmation.json追加確認，不重寫歷史。Python最初語法錯誤在讀來源前停止；不是產品缺陷。

032僅DRAFT試作：依固定Ebitengine2.9.9平台時鐘契約及psychic-war的累計cycles回饋方法，正常180筆日誌與525015生成取樣相同。實際訊號47317對48652音框，21585非零值／順序相同，完整精確驗收仍FAIL，不能以移除靜音後相同稱為時間軸通過。此差異標為平台近似，不深挖DAC／PIT逐週期一致。100毫秒追趕上限與卡頓操作取捨未獲使用者確認，不進正式路徑。各工具、輸入身份與原始收據入口見CONTEXT。

## 2026-10-20：銀行餘額的新檔冷讀

輸入沿EXE fe1c00b3、SAR2 bf7c792a、882份dosgolem來源摘要2a5ad11a及字型20e8b505；Go1.24.13、Python3.11。自建SAVETEST SHA-256為78359e3684cf8ebc4cb52762bc7afc0b551567cd01726b8014dfb043109042bb、mtime_ns725847808000000000。原始29檔與參考來源唯讀，只有各側scratch及本輪輸出可寫。

已證實：正常銀行存100、領50及離店後，可回長老另存LEDGER，正常退出，再冷F7讀回。原版確實寫出並讀入256 bytes，冷讀後銀行來源仍是SAR2檔案偏移01a073，原版01B1:AD30數字暫存及已完成字形為50，身上01B1:0086為950。兩側各自產生的新檔SHA均為c3b21bcfe48c5b68fe2051bc9e818c054d7ea9e604ea2df6742b0186fb9444ba；來源SAVETEST及mtime未變。此項只驗玩家可見保存結果，不推測USR帳戶欄位格式。

56個完整原版／中文狀態及來源畫面相同，銀行金額沿原版數字圖形，不由中文層計算或重新印數字。11252筆原生日誌、613行可重生輸入、56個來源圖及測試收據入口見CONTEXT。離線來源圖不當成實際X視窗，這段完整存讀流程及獨立完整排版仍待後續驗證。

混合來源的入口限制另存mixed-reachability.json：五份既有USR冷F7後的C000前綴均匹配SAR2 payload0392dc，五筆混合來源分屬03ae89與03e813。這只證明既有起點，不能宣稱後段不可達；原包存檔取得過程未知。本輪不猜控制碼、劇情或交易分支。

稽核勘誤：此固定dosgolem的internal/dos/files.go之write呼叫h.f.Write並記Wrote，不加入FileOps。FileOps中的LEDGER只有create／close，不代表沒有寫入；新檔原先不存在、原版獨立生成的256 bytes及第二次F7讀取共同證明保存。audit-v1-failure／confirmation保留工具錯誤及來源雜湊，不把日誌缺項當成產品缺陷。


## 2026-10-20：銀行存讀的實際視窗收據

原始EXE fe1c00b3、SAR2 bf7c792a、882份執行器摘要2a5ad11a與字型20e8b505保持。實際主入口cea16753／112來源；目前115份來源只有既有測試差異，銀行覆繪與譯表未改。自建SAVETEST的78359e36及mtime_ns725847808000000000保持。Go1.24.13、Python3.11／Pillow9.4；位址沿dosgolem01B1段，來源鍵為SAR2檔案偏移，兩者不混用。

已證實：GUI v4正常存款100、提款50、查詢50及長老另存，原版退出0；新256-byte LEDGER SHA為c3e68b98908a5f4423222af68a9e55c12c0e4a156bb4792e0b499f5d5cc463ad。GUI v5冷F7讀同一檔、查詢50、正常離店950及退出0。v4中間被外層時限停止的冷進程未送出查詢，保留但不計完整通過。

27653筆真實原生日誌、37個實際暫停點的完整原版雜湊／時間／持鍵與銀行／對話狀態相同。原版重播自行生成同一新檔及mtime，再冷開機讀自己的產物。沒有共享機器狀態或注入。查詢來源為SAR2 01a073、動態數字仍由01B1:AD30完成字形提供；身上01B1:0086為950。

獨立Python由原版完成的50圖形、UTF-8 TSV及原始字模重建兩頁查詢的196560像素，差異0；31824原版金幣欄像素不變。37整圖的21312000像素另是同套Go覆繪的實際合成核對，不能稱為全部獨立排版。

工具勘誤：原生按放ack不保證每個短按都被原版選單消費，實際箭頭與金額像素是確認前的依據。回呼完成文字亦不代表末尾等待已就緒；銀行告別須核ACD2及等待旗標0。重播v1漏了既有Bank.Draw失配撤除，原版雜湊已相同；v2加入實際暫停Draw後保留全部狀態條件並驗繪製冪等。這些工具修正沒有改正式遊戲或輸入規則。

顯示缺口已證實：v3實際無存款提款仍是英文，v4長老離開段亦見英文。完整來源範圍及控制尾碼仍待核對，不能僅由畫面推定新的SAR偏移或劇情規則。前述失敗、誤名截圖與中斷進程入口見CONTEXT，舊資料不重寫成成功。

## 2026-10-20：兩段英文缺口的來源已確認並補譯

追加解析：上述未知已由正常GUI日誌冷重播確認。輸入EXE fe1c00b3、SAR2 bf7c792a、dosgolem凍結882份摘要2a5ad11a保持，工具Go1.24.13，原版字形觀察為dosgolem記憶體01B1:711B／7120。原始檔案偏移與記憶體位址分列於規格001，不視為同一空間。

已證實：無存款提款的SAR2範圍019f95..019fe3、payload019462，對應01B1:AB33，75字元後尾碼FF010C。長老普通告別的SAR2範圍01cab2..01cace、payload01bcc6，對應01B1:ADEC，25字元後尾碼11FFFF。來源原始bytes與payload前綴在當次繪字均相同；證據入口為CONTEXT的bank-text-gaps-v1下兩份source.json，原始圖面與記憶體只留本機。

新增正常路徑17點及既有存提款35點回歸通過。實際視窗的7052筆執行／按放、9個完整原版停點及中文狀態相同；正常退出0且USR來源雜湊／mtime不變。兩頁獨立字模核對196560像素、原版金幣欄31824像素差異0，身上1000。這證實已列顯示切片，不推論其他拒絕分支、USR格式或新HD範圍。

## 2026-10-20 教堂來源與受傷返城

已證實的五個純文字來源為SAR2檔案偏移01a453、01a491、01a4cd、01a508及01a568，payload為01a19e，載入dosgolem位址01B1:A000。原始範圍、控制尾碼及兩條順序見[規格001](docs/spec/001-runtime-overlay.md#教堂與旅店中文)與本機`services-v1/source-v1/glyph-events.json`，SAR2 SHA-256為`bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`，EXE為`fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`，凍結dosgolem882檔摘要為`2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`。工具Go1.24.13，位址基準是dosgolem實模式段位址與SAR2檔案偏移，兩者分列，不是IDA線性位址。

正常原版來源的生命值75→77→80只記錄結果，不推定中段增加原因。獨立實際GUI從同一自建USR冷開機，以真實鍵序受傷後50回鎮，教堂第一回復句完成80；這條9241筆日誌與16個暫停點另重播相同。全部原生日誌、原始RAM及畫面留本機，證據入口為CONTEXT的services-v1。教堂兩個牧師姿態各觀察三個燭光畫格，共六組96×64色號／RGB來源；它們只證明已觀察動畫，兩張手繪候選未進執行路徑。

## 2026-10-20 教堂五手勢／三燭光來源

- 輸入ZELIARD.EXE SHA為fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341，ZELRES2.SAR為bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1。凍結dosgolem882檔摘要2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada；Go1.24.13、Python3.11／Pillow9.4。各收據SHA與映像身份由本機hd-provenance-v2.json串接，入口見CONTEXT。
- 位址空間分列：SAR2檔案payload偏移01a19e；dosgolem實模式01B1:A000；原生320×200色號；肖像外框(112,23)-(208,87)；高清內框282×186，視窗原點(339,120)含48像素工具列。這些不是IDA線性位址或遊戲世界座標。
- 已證實：三個真實首頁時點各自冷重播、正常Space後，手勢區(176,67)-(187,86)分五組，每組三種燭光，15種來源反覆出現。至少七次IP在01B1:A000..A1FF的抽樣只作來源重複證據；IdleWitnesses不是已證實的函式名稱、等待語意或完成回呼。五個其他來源只列候選。
- 勘誤：初見四種未知房間不是四個手勢，其中兩種是同手勢不同燭光，實際新增三個。原快照與初次分類留在hd-gui-pair-v1/unknown，五組與見證在hd-gesture-source-v1/inventory.json，不重寫舊收據。
- 已證實：新版431筆觀察與同時點原版匹配相同。唯一未知來源SHA為7c696f045cf275ace1387ed0cfa654f4e5e698d809e01ce068d2b96794bc446e，IP01B1:3770，與最近穩定圖相差8個燭焰像素；繪製中更新是強推論，不覆蓋位址或加映射。原子退回及原生輸入不變另驗。
- 手繪圖是原版構圖衍生的本機研究素材；15來源與圖面固定SHA見規格001、註冊索引及程式。原版、字型、畫面與圖面不入Git或發行包，不藉此宣稱其他場景或全遊戲完成。

### 教堂場景名稱的正常載入證據

- 固定EXE fe1c00b3、SAR2 bf7c792a與dosgolem 882份2a5ad11a沿CONTEXT完整雜湊。Go1.24.13，原版位址空間為dosgolem實模式01B1段內偏移；檔案偏移屬ZELRES2.SAR，不作EXE或IDA線性位址解讀。
- 兩次正常冷讀存檔路徑均由22CD收到A2A6完整14 bytes，對應SAR2 header01a444／body01a448、module01a19e。入口原文SHA a7abf555、前綴77e9c7f2、字型f16d564f、繪字碼a4564867，完整值及兩次steps／cycles見church-scene-v1/pair/receipt.json。2344完成SI A2B4，字形b4d572d4及安全區(94,175,51,8)相同，證據等級已證實。
- 實際視窗及獨立字形證據見同研究根gui-v1、replay-v1與pixels-v1.json。場景名稱顯示與原版規則分離，回鎮改用已驗Muralla來源；不從固定SAR檔案候選外推旅店可達。

## 2026-10-20：全遊戲 HD 資源盤點與解壓常式

- 輸入為固定三個SAR，完整SHA-256見CONTEXT及 [覆蓋台帳](docs/hd-coverage.json)。共194筆，40／58／96筆目錄及payload長度均連續至檔尾。來源位址為各SAR從0起算的檔案偏移，未改原版。
- IDA9.4沿既有 `ida-pro-9.4-idapython:py312-v1`，輸入為 `probe-intro-memory/memory-01B1_0000.bin`，SHA-256 `c9f188ea847e8d364ccf8dbc9f6a9b9c74e192c0c83bd7a012b552e6d89b13d8`。一次性DB為binary base0、16位元，對應此快照dosgolem實模式01B1段內偏移。v2保留原始候選指令，v3合併 [分級位址索引](docs/source-index.json)，未證實語意附警示。
- 已證實：DAD以來源首byte低三位選DBC的八項分派。從DAD至F16擷取的原始常式SHA-256為 `42f4f6392954f85fa2e8b5b1b1b3c0a1ca2c3f6ff2aab6e1126ed4c14d7a6d61`。測試MZ保留段內原始碼位址，只以框架設定來源、目的地及長度後呼叫DAD；不是原版正常遊戲路徑。
- Go1.24.13、既有dosgolem隔離快照。`decoder-oracle-v4/receipt.json`記84筆原始stream／框架／輸出SHA、步數及週期，與Python離線解碼逐byte相同，全部DOS退出0。合計58,085,998道指令，單筆最高3,694,917道。真實資料涵蓋類型0、1、5、6、7，其餘三分支僅有原始指令定位。
- 勘誤：`decoded-v1`把類型7的替換值與長度順序讀反，第一筆原常式比較揭露。EF5至F16的LODSB／XCHG原始指令顯示標記後依序是替換值與長度，長度加3。修正後 `decoded-v2` 全84筆與原常式相同。v1保留失敗身份。
- 首次IDA輸出因locale為ASCII而留下空JSON，v2使用明確UTF-8後核工具9.4、輸入SHA及3,143筆候選定位。原常式首個2M指令預算不足以跑完字典逐項搜尋，依12KiB輸入與256項字典的指令上界改為32M，沒有改遊戲bytes。兩項是工具／預算問題，不作產品缺陷。
- 強推論：名稱descriptor候選對應174筆記錄，依副檔名暫分80圖像、41地圖、17音樂及36程式，另20未知。記錄名稱包含31個洞窟地圖及10個城鎮地圖候選。名稱與副檔名只供追查，不證明全部可達或圖像已解完。
- 未知：GRP內部形狀與圖形驅動前處理。國王A021透過2044指向2C2A，候選指令複製CX乘48bytes再轉換，但本輪未證明完整圖像語意。直接把解壓結果當六位元圖格的 `sixbit-candidates.jpg` 未成立，不當作原圖來源；RLE解壓成功與圖像解碼分開記錄。

### 五類居民的手繪候選

- 來源沿007已驗MCGA圖格及遮罩契約，固定1MiB SHA-256 `9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9`。資源11B1:4000為五組各八相位、每相位六個圖格ID，11B1:4100為圖格，21B1:7000為遮罩。原生16×24，40來源有34種RGBA；不宣稱40種姿態都已正常遇到。
- `tools/hd_town_npc_sources.py` 重生 `town-npc-source-v2/`，40張RGBA均與初次擷取相同。五張原版4×2圖集供內建imagegen沿已確認B手繪奇幻風格重畫，原圖、全部提示詞及來源在 `town-npc-art/g0..g4-generation-v1.json`，產物1448×1086、真實alpha。只留本地，不加入Git或既有封包。
- `town-npc-registration-v1/manifest.json`記每姿態裁切與原版alpha邊界，產生40張48×72註冊預覽。原RGBA相同的姿態共用相同候選，避免生成差異增加原版沒有的動畫。五組圖與註冊接觸表已目視；目前未驗正式合成、連續動畫、遮擋或實際視窗，不計HD接入完成。
- 此階段收據為 `workplace/full-hd-v1/stage-audit.json`，入口是 [規格034](docs/spec/034-full-game-hd.md)。後續先接居民圖面及核GRP內部格式，不重開已驗的RLE常式或前版音樂。

### Muralla居民的共同合成驗收

- 已證實：035沿007／010已驗原版繪製順序、模板及遮罩。固定參考1MiB SHA仍為9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9，五張圖與40裁切的註冊SHA為4633a11f282060e43c5a6111875d4c913faa79624772e67637c178c8b5cef6f0。原生16×24、HD48×72；面積取樣直接使用1448×1086原圖，不拿LANCZOS預覽作原版證據。
- Go1.24.13候選SHA-256為3bdc2098989064bf4d315fbc7e5042764004324877859a0086d2e9cc915407c6。town-npc-pair-v2六點正常冷F7／居民交談／步行的完整RAM、VGA、色盤、暫存器、IP、steps、cycles、檔案及鍵盤狀態相同。town-npc-gui-v1四組原貌切回每組576000像素相同，正常DOS退出0，固定自建USR未改。
- 視窗抽樣居民鍵為g1-p4與g0-p1；原版配對另見g1-p5。40項資產身份檢查不代表40種姿態都已正常遇到。合成色塊遮擋測試只證明內部圖層順序，原版行為依既有007／010。
- 已知覆蓋缺口：交談期間既有完整畫格守衛撤除HD，顯示原貌背景。其他城鎮與主角相位不外推；這些仍阻擋全遊戲HD。稽核與逐檔程式身份見workplace/full-hd-v1/town-npc-final-audit-v1.json。

### 三平面前處理與城堡來源

- 已證實：IDA9.4既有resident-v3的01B1:2C2A..2CBC原始指令先複製CX×48，再逐6bytes讀三個big-endian word，依C／B／A最高位順序交織。每24bits輸出前16bits little-endian與後8bits，重複兩次形成6bytes。隔離MZ連同可寫scratch保留2C2A..2CC8，其SHA-256為eab920e45d2253c63f119d20c3e2823a6b1fb5d45f2a2b6ebe7acd4559afb7ab。Go1.24.13／既有dosgolem執行80筆補零研究buffer，與Python逐byte相同、退出0；收據為planar-oracle-v2。這不認定header與可見形狀。
- 已證實：MMAN／TMAN三平面色7清為0，色0的兩個子像素均透明時設定native遮罩位元。MMAN原256bytes header與正常11B1:4000相同，152圖格7296bytes與11B1:4100相同，1216遮罩bytes與21B1:7000相同。TMAN46圖格2208bytes與11B1:6000相同，368遮罩bytes與21B1:8000相同。原始SAR2解壓檔SHA及輸出SHA逐項列grp-sources-v3。
- 已證實：CPAT157及MPAT242圖格的header三個指標加8000後與11B1:8000相同。flag0保留色、flag1..3清色4／2／1，兩個子像素同為該色時設定遮罩；flag4在兩個正常profile的遮罩均為八個FF。兩組19152圖像bytes及672個非零flag遮罩bytes與11B1:8100／D000相同。無效或未採用的flag0遮罩不稱已驗。
- 勘誤：冷城堡1MiB SHA為5c5b610c97bd2cd5266ae25b3da3764d0b46f02b30e8392a3b09c286a0db45a2，載入的居民模板仍為MMAN；CMAN檔名不能證明此入口使用它。CMAN164與DPAT191圖格只列結構候選。MMAN尾123bytes、CMAN尾59bytes用途未知，未據此開無關RE。
- 城堡全景來源沿007的C017地圖與C00F底圖契約，114欄×8格，912×64。第一次靜態比對(5,0)揭露flag4背景遮罩；第二次(106,0)揭露地圖ID4、已畫快取ID3差異。按E000實際圖格、A000背景、原版居民與主角重建後，完整14336畫帶像素與正常原版相同，見castle-source-v3。沒有跳過失配點，來源守衛未放寬。
- 已證實的快取差異為世界欄37／43的第1列、欄47／53／59的第0列。兩圖格是不同旗幟，動畫原因為強推論；計時與更新端仍未知。生成圖候選不能凍結這些差異。
- 四段城堡候選使用內建imagegen、兩張參考及完整提示詞，生成來源與SHA在castle-art-v1/generation.json。候選PNG SHA為cd6af931976ab0d55b5ea9ab8e4fca76ed70d7ebf8575664907b6006f01a29a0，1214×1295、真實alpha。根端目視四段、主門及亭台，裁切高度與對齊仍待036審查，不稱正式HD。

### 城堡旗幟更新端與主角右邊界勘誤

- 已證實：正常冷新局及左右往返共13點完整原版狀態相同；CPAT正常來源的旗幟表在11B1:80A6，四對為2→3、3→4、4→5、5→2。原始sub_31D8在01B1:32EA讀ES:[DI+1]，32EE保留原運算元MOV [SI-1],AL寫入DS地圖，32F9後觀察754次更新；保存128筆涵蓋四對且寫後值均相符。E000已畫快取與未來地圖值分開，故HD只按完成來源的已畫ID選圖。這項新證據取代前輪「動畫原因為強推論」，不重寫舊研究段落。
- 輸入為既有固定EXE fe1c00b3、SAR2 bf7c792a及冷城堡1MiB 5c5b610c。Go1.24.13沿現行隔離dosgolem執行；工具及完整SHA見`castle-normal-v2/receipt.json`。正常鍵序SHA-256為bcaee48cf75e3bd737ba64597d5c8b74a2aa5d6ce88ad88562fdee0b105eff6f；八個城堡最近完成來源各14336像素重建相同，保留自己的steps／cycles，不冒稱當次CPU畫面。
- 已證實：連續正常往返在88800000步、01B1:07BF揭露舊主角模型的右側半身錯誤。真實1MiB快照`castle-pair-v2/uncovered.memory` SHA-256為c6a54aa329a9483f7a6cbf03af49756e840242a94e80a247233c7f827f6799e8；該快照主角83欄位為27，舊模型錯誤裁切後仍畫半身，原版完全不畫。
- IDA Pro9.4／IDAPython匯出輸入為該快照01B1段，65536bytes SHA-256 aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d。IDA binary base0對應該段內位址，不能與SAR檔案偏移混用。保留sub_6950原名：6950讀byte_83、6953 CMP AL,1Bh、6955 JB只接受0..26；其餘6957 RET。通過時才標E000兩欄各三格為FF。sub_68AC依序呼叫6B1C、6975、6950、3004，與原版失敗畫面一致，不需追無關helper。
- [城堡來源索引](docs/castle-source-index.json)保留32EA／32EE／68AC／6950原始定位、已證實等級與證據。IDA匯出`castle-edge-ida-v2/evidence.json`及正常探針自動合併索引，索引SHA-256 dbd4cc1e8d9341620c40dfae79554d61ff4209cb540fddaab5acab7e57670e5e。未改原始函式、欄位名或運算元。6950..6975精確SHA守衛及真實邊界14336像素核對補入既有模型，居民仍可顯示HD。
- Castle守衛只採CPAT157圖格與232個有效遮罩bytes；MMAN152、TMAN46及其遮罩沿已驗來源。背景、原生重建、圖面及alpha分層核對；共用地面224×16來源SHA為80322194e1e97e020451cdd184bfbfd6dd4e4de8bfbc074453b8f1484c53a83b。其他GRP、phase4與交談背景仍未知或未接入，不因城堡正常往返通過而升格。

### 11組洞窟地形與31份地圖

- 已證實：SAR3記錄74..84解壓後的MPP1..MPPB，各保留48-byte header，圖格數依序25、34、43、21、52、36、61、43、38、20、5。全部378格三平面變換沿已核原常式；MPP1的header與正常11B1:8000相同，全部1200像素bytes與11B1:8030相同。正常1MiB SHA-256 e44c7555e68e61d3176cd40306ae7fc909a53b6216481de1290f5a013b20e76a可回查prose-v1/progression-v1原收據，並非新注入來源。其餘十組只升結構／靜態來源，不稱正常到達。
- 已證實：固定正常段cave-entered.primary的01B1:6CED以壓縮byte高兩位選四類run。類型0為count=byte+1、下一byte為圖格；類型1為count=((byte>>4)&3)+2、圖格=(byte&15)+1；類型2為count=byte&63、圖格0；類型3為單格byte&63。sub_6D57每欄解至64格，以36-byte列距寫ES:DI。SAR3記錄20..50寬42..320，C01B起、C019為結束指標，31份合306048格與隔離原始常式逐byte相同，最終SI皆正確。
- 輸入段65536bytes SHA-256 12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93；擷取6CED..6D6D原始bytes SHA-256 cc712d2efe92fb89ca3d93f5da41e15d98716dd21fbac994cbd12718aed2537b。Go1.24.13沿既有隔離dosgolem來源workplace/music-release-v1/release-build/dosgolem執行，來源身份沿該release-build/identity.json；原常式31次退出0，2881666指令。這是格式契約，不是正常玩家場景路徑或遊戲通關。
- 已證實：原主流程61EE從word_C000取地圖尾，61F2先LODSB後呼叫sub_7E93。7E9B將後續四byte複製至9EF6..9EF9，故9EF7讀地圖尾+2。7ECD讀此byte、乘11加9C43，7EDF指定8000後經原載入服務取得圖集。固定段的11項descriptor為02、4B..55及MPP1.GRP..MPPB.GRP，索引byte減1對應SAR3記錄74..84。31份真實地圖selector都在範圍內，MP73選MPPB。
- IDA Pro9.4／IDAPython輸出在cave-map-ida-v2/evidence.json，binary base0對應此正常01B1段內偏移。SAR3 record payload偏移及解壓buffer偏移另列，不能混作IDA位址。新分級索引保留原始sub_6C98、6CFA間接跳躍、sub_6D57、7E9B與7ECD；匯出自動合併，未分級指令仍為未知，不用自訂名稱替代證據。
- 正常來源補證據：四份包內USR冷讀共12點，加新版SAVEM1三點，兩獨立dosgolem的完整RAM、VGA、色盤、暫存器、IP、steps、cycles、檔案與鍵盤狀態相同；所有正常入洞來源仍為MPP1。新版在01B1:6CEC保存初始化完成1MiB SHA-256 cc6023f47fbd9e7a4af12655a09067e9638711a6e696abd8676fe5f9b7da08c6，水平起點45，36欄×64列的E000來源2304格與原始解壓結果相同。完成畫格另有18格門口標示寫入，不將運行中差異當壓縮錯誤，也未忽略差異宣稱完整畫面相同。
- 冰窟MPP4的21格原圖與MP40靜態研究圖作構圖，原有岩面材質只作手繪風格參考。內建imagegen兩版2048×768候選及完整提示詞只留本機。v1 SHA為226b6af3407afd9ed3a5d8cd7cf6645fe8c8ceae17e5ae0dc6e32a53614f0913；v2為16fb57a71cf5ec0fe377afec7fc972e1d8b51ec4cdb83f6f52b05454e9ce21fc。兩版透明像素分別764397及829549，但未用的三格仍有殘留alpha；不升READY、不稱已註冊或正式HD。

### 來源身份增量與冰窟相鄰圖格勘誤

- 31個C000..C013的19-byte原始前綴互異，與各地圖尾+2選擇的11組MPP形成固定來源鍵。新增Go資料表只保存名稱、圖格數與SHA；逐組實際48-byte header及其後圖格數乘48的完整資源雜湊須同時成立。C013..C016不由本覆繪解碼消費，排除於新身份鍵；舊兩個已驗64-byte前綴另沿原契約，不混稱同一個守衛。
- 同家族MP40／MP41可有不同地圖，不能只憑MPP4相同保留舊完整快照。合成結構fixture刻意沿用相同假色號仍須按來源鍵撤除；主角／完成來源不屬同一已知地圖也拒絕。此為資料與生命週期測試，不是正常玩家到達冰窟的證據。原版規則及全部32256原生像素重建不變。
- 新候選Go1.24.13 SHA-256 bc925caafdfb44d5dfd70bf221ddaf3b9dd2bf5bd4eceb7861335b534311ad46，沿既有隔離dosgolem與原EXE。正常MPP1首戰25點、回城10點完整狀態及自建存檔相同；實際視窗兩組原貌切回1152000遊戲區像素相同，正常回城及DOS退出0。其他家族正常來源仍未知；來源鍵涵蓋31圖不等於已HD或可達驗收。
- 勘誤：MPP4的15..19是圖形文字，原圖相鄰合成可見HELADA，不能按冰塊描述重畫。原圖13與14是三角形兩半，各25個非零像素；14的原始色號含0、4、32、36、54、56、62、63，不是空格。初次註冊的空格斷言失敗後回查固定SAR3記錄77解壓圖格，沒有把失敗改名或刪掉來源像素。
- 新參考first-14-reference.png SHA-256 617f2a250c21a387fd06cc14c8904153d03bff6a1d296bc199646c597e9dd6ec。內建imagegen第三版1254×1254 SHA-256 9a09979a2c70a893c73a1dc876fe248862202573f4ce0b0c898a596bbd71b118，實際輸出保留13／14的連續三角形；原prompt錯稱14空格保留在generation.json並附原生勘誤，不能反過來以prompt當原版證據。圖形文字與20..21未放入本次14格註冊。
- 不透明材質同為1254×1254，SHA-256 a8fef8473861ffe9a41d7f7cfe44dd6fda8ec94db154023034714c8fd88d4bc2。候選直接取432×768並鏡射成864×1536週期；3..12按連續欄首末可見像素形成美術包絡，其他非零形狀按原生邊界合成。13／14共用生成圖裁切(0,940)-(625,1254)取48×24後分半。這是手繪呈現研究，不是原版透明、碰撞或正式runtime資料格式。
- 註冊與MP40／MP41／MP4D預覽為cave-ice-registration-v3，原始AI圖面未改寫；完整提示詞及每張輸入SHA分別保存在cave-ice-art-v3、cave-ice-material-v1的generation.json。仍未有冰窟正常GUI，保留DRAFT。完整身份與分層限制見cave-catalog-final-audit-v1.json及037。

### MPP4圖形文字、光球與21格圖集

- 已證實的靜態來源仍為SAR3記錄77，解壓SHA-256 41edc9380a37931fec4c0ee81ec79aff61c64e71b7c8e78101a3a88c44036995。原始15..19五格相鄰組成HELADA六字母，不能按單格猜字母。三份已核原始地圖MP40／MP41／MP4D均未使用20或21，故不能把這兩格命名為已知敵人、道具或動畫階段。
- 內建imagegen以15..21原圖與既有手繪冰窟風格為參考，產生1691×930透明圖，SHA-256 ac8fab5311db8071ddf4a90013cfe078e15aa8966b48b365721f70bbee21986c。完整提示詞、兩份輸入SHA與原圖保存在workplace/full-hd-v1/cave-ice-details-v1/generation.json；原始輸出未改寫。兩個光球保留白色左上高光及藍紫陰影，玩法用途仍未知。
- 註冊後字帶120×24再分五格，清除舊字形黑底；光球24×24按原生非零外形限制。與第三版1..14合成504×24圖集，SHA-256 4305348c37fa0ff0c5bf0f9e15546f7e5b4bcffdb3416ec16f4beea75d816fee。432×768直接HD材質SHA-256 9f2dc6eed83d9582edeef8ac4f64124c382c93550b5f6fe295fbdec064d9e5ac，資料與裁切在cave-ice-complete-v1/manifest.json。
- 原版身份的地址空間仍為01B1主段、11B1資源段，MPP4為11B1:8000的48-byte header及8030起21乘48bytes。Go1.24.13的結構測試從已驗MPP1取得程式與角色，組合真實MPP4資源、地圖鍵與靜態格，不執行VM；三份離線HD預覽共870912像素與正式呈現一致，畫布區外及來源記憶體不變。此結果不當正常冰窟原版對拍或可達證據。
- 新候選SHA-256 4948bc709f551827d74b451925e41dc4d9b3e2e898d564a78e8e5f1efd23f98e，增加可選冰窟圖集。正常首戰25點與回城10點原版狀態、存檔及mtime相同；三項新冰窟測試、舊地形獨立像素、來源／轉場／未知分支與原子退回均通過，無skip／fail。正式冰窟進出、捲動與事件仍待正常玩家路徑驗證。

### MPP2的34格來源與手繪接入

- 已證實的原始來源為SAR3記錄75，解壓SHA-256 66d849269082061d658ab07dab2f33a44fe5e231ebce64599bb8e48d54ba34e6。header的SHA-256為d9d84ec687bb991cbe0157800ce53c137c810b6edfa6be84f83eaf2966093bc8，變換後34乘48bytes為a84dd4687f8ec67d320ef0fdd2d83c3050544e31cc804867ea216f1616afef99。地址仍為11B1:8000／8030，不與SAR檔案偏移混用；原版圖集選擇端沿已核01B1:7E9B／7ECD。
- 原圖21..25相鄰構成PELIGRO，五格共同表示完整字帶。MPP2的22..24與MPP1的17..19 RGBA雜湊相同，只證明這三格相同，不推定整組資源相同。MP20、MP21、MP2D靜態格皆未使用26..34；九個尾端圖格只按形狀與配色註冊，不命名敵人、道具或動畫階段。
- 內建imagegen三張原圖及完整提示詞／輸入SHA在workplace/full-hd-v1/cave-peligro-art-v1/generation.json。材質1254×1254 SHA-256 8963f811a25b14b8a81f31a446bccee3ac2b3bc96b5f2a9a3e7ec787c75d5568；11格圖面1448×1086 SHA-256 bf2061bfbe0fcb91aa754cb3ad139c52fa08edbc8cfc9a485ebadfbb1d61e752；字帶2172×724 SHA-256 4c96a91ed3139bce0690655ab2ab06192865ef6ddc16f1166abb63ad4c62f95b。原圖未改寫，材質原檔為RGB，不把缺A通道稱為透明度失敗。
- 34格24×24圖面直接從生成解析度取樣；1／2及26..34限原生非零外形，21..25從完整120×24字帶分格，3..20選連續黃紅岩面及連續欄首末可見包絡。此為美術映射，不是碰撞或原版透明語意。816×24圖帶SHA-256 d5b5b4addc544e008d1b37ea9e2ee24926bf7ee72a1ed346d42410cf134c9e1f，432×768材質35d1aa52b716b925bd75f34f85d682a69e9f8aa5729e4e868f2aafc54f91d6d9；裁切與預覽在cave-peligro-complete-v1/manifest.json。
- Go1.24.13候選e9f89d5296eb04d7b2fb667882efaf614e7cf41a9f39de9d047ec680a7e19126，以三份真實靜態格組合進結構快照，870912像素與獨立Python預覽相同；全部34格、尾端九格與原子拒絕通過。同一呈現器的冰窟三份870912像素及Malicia舊圖通過，避免依載入順序選錯家族。結構快照不執行VM，不當新增家族正常可達證據。
- 正常MPP1首戰25點、回城10點完整原版狀態、USR與mtime相同；實際視窗兩組切回1152000遊戲區像素相同，正常回城及DOS退出0。來源仍MP10／MPP1，未升MPP2或冰窟的正常進出、捲動、事件或切回驗收。稽核入口cave-peligro-final-audit-v1核三份預覽各504格與原始地圖相同，舊收據保留各自原身份。

### MPP3的43格及五格BOSQUE標示

- 已證實的來源結構沿SAR3記錄76，解壓SHA-256 66e9e07a7fc7677853baafb82c15a653338979fe73ce57f291f939719ba54b91，48-byte header為4911af867d864e1616a9a0b3fdbd8dfcc438810f7fc67c8278f1f3d146647cde，變換後43乘48bytes為b19a3e201325e198d44ef5399e721a148e7f17a804f69ebccd402d1b2e179cdf。地址仍為11B1:8000／8030；原版選擇端沿01B1:7E9B／7ECD，不混用SAR偏移。
- 放大原圖後，九個綠色材質ID為3、4、6、7、8、9、10、32、33；圖格5及紅褐色格架不能當成同一連續材質。這是有原圖依據的美術分組，不是碰撞、透明或玩法分類。
- 初次只看35..38四格，沒有足夠文字證據。MP30及MP31的靜態網格均在預覽第8列、第9欄起明列34、35、36、37、38；原圖完整五格拼合後目視轉錄為BOSQUE。原始片段、完整拼圖及勘誤在cave-mpp3-source-v1，未把四格片段改稱完整文字證據，也未推定官方中文譯名。
- 內建imagegen先產生一張綠色岩面與三組逐格圖樣，後續另生完整BOSQUE字帶。五張原圖及完整提示詞／輸入SHA在cave-mpp3-art-v1/generation-v2.json，初版generation.json及逐格標示保留原身份。字帶原圖SHA-256為4b62eee54e9bbb1a4e8350532184b7a9ff451c00633d4ad3b5c54fcc09a0add0；沒有覆寫任何原始生成圖。
- 正式候選註冊為43個24×24圖格：九個連續材質、29個逐格圖樣及五格完整字帶。1032×24圖帶SHA-256 40f9be56644f647d6f5b7247aa3339de6e890f8a5a8d5875fd490cfc319aa678，432×768材質4ceb8428d6d7231e764bdcccea20520d09bf079424fd645fa58ffc612cc555a6；裁切、三份真實靜態網格與預覽在cave-mpp3-complete-v2/manifest.json。
- Go1.24.13候選148bdd99e3a5b6a308cb3cb3247c39c305cd313c96fc741e4fa040cd4ad9c1eb，三份MPP3結構快照870912像素、全43格、字帶與原子拒絕通過。相同呈現器的MPP2及冰窟各三份870912像素、Malicia四份獨立像素通過。結構快照只組合來源資料，不執行VM，不當正常MPP3到達或原版規則對拍。
- 正常MPP1首戰25點及回城10點完整原版狀態、USR與mtime相同；11項圖面及三項正常／來源測試無skip／fail，vet及建置通過。其他家族正常路徑仍待驗，不由來源表或相同seed數字推定玩法一致。

## 2026-10-20 MPP5來源與HD圖集

來源為SAR3記錄78，解壓SHA-256 `964ab425d1f2d215fd4286ef53c0780090fa2663c8fc3b221f95b45f68005588`，dosgolem實模式11B1:8000的48-byte header為`bb1d943bb0659bfabff564b21cec7256b4d56ed9d03c78ba4b3dcbdf2252a903`，8030起52×48bytes三平面像素為`51ae7fec2bd8b14f5affe35a0337eaa0ea632efa5ee15554d29279e27a68faf0`。來源身份沿037既有原版解壓及01B1:7E9B／7ECD選擇證據，不新增位址語意或覆蓋原始名稱。正常MPP5載入仍未知。

- 工具：素材流程`tools/hd_cave_family_art.py`，固定容器eob-audio-capture:20261009-r3，Python 3.11.2、NumPy 1.24.2、Pillow 9.4.0；Go1.24.13。腳本、輸入及產物SHA在`cave-mpp5-complete-v1/manifest.json`、`cave-mpp5-art-v1/generation.json`及`cave-mpp5-final-audit-v1.json`。
- 已證實的靜態資料：MP50／MP51／MP5D已解網格使用同一MPP5家族；48..52未出現於三份網格。完整52格原圖及鄰接樣本見`cave-mpp5-source-v1/`。外觀分組只供美術，20個連續岩面ID不作原版碰撞或用途證據；尾端圖樣事件未知。
- 原圖不覆寫。四張內建imagegen原圖直接註冊為52個24×24圖格、1248×24圖帶與432×768材質；圖帶SHA為`58e76f6b1ef046960a2fabcf4f3268bb1df514d23d9e500ae6a80e2a49b7e02b`，材質為`d640e43e0b4dec2aa3235a2b038a64cbb7b18cac9fe139c09ed19b1229633c8b`。三圖及額外水面預覽已目視，沒有將原圖像素放大當新手繪。
- MPP5三圖870912像素、全52格及原子拒絕通過，同時載入MPP2／MPP3／冰窟後各三圖870912像素與Malicia舊圖通過。結構快照只組合資料，不執行VM，不當正常原版MPP5可達證據。
- 候選`7d623290989ca2790a50ac893248a97e8a1a62d677850fce517c8179af420cad`的正常MPP1首戰／回城35點完整原版狀態、USR與mtime相同，實際視窗v2兩組切回1152000遊戲區像素相同並正常回城、退出0。正常MPP5進出、捲動、事件及全遊戲HD仍待完成。

## 2026-10-20 MPP6雙材質與非連號字帶

SAR3記錄79的解壓SHA-256 `44dc5d86e69709bbe571b1cf86008a745a91e80d897437b68aeecc4cbe8b4325`，預期dosgolem實模式11B1:8000的48-byte header為`9e25d4d349978ee06e5b05efdb291ce8e70a363f5a269c34bf395a28f87f5fc3`，8030起36×48bytes為`a90b2faacf3e4895cdfef2690805d80fec4ddd9141a791dc64d42d93e3abe812`。來源及選擇端沿037的原版解壓、01B1:7E9B／7ECD證據；不新增推測性命名，正常MPP6載入仍未知。

- 工具為`tools/hd_cave_family_art.py`、固定eob-audio-capture:20261009-r3、Python 3.11.2、NumPy 1.24.2、Pillow 9.4.0與Go1.24.13。腳本、來源與產物SHA在`cave-mpp6-complete-v2/manifest.json`、`cave-mpp6-art-v1/generation.json`及`cave-mpp6-final-audit-v1.json`。
- 已證實的靜態資料：MP60／MP61／MP62／MP6D同選MPP6；30、31、34、35、36未出現於四份網格，事件用途未知。金黃磚牆ID3、4、5、7、8與紫岩9..13是外觀分組，不作碰撞或玩法語意。
- MP60／MP61原始相鄰五格順序26、27、28、23、6。完整40×8字帶目視為DORADO，不能以連號範圍覆繪。先前CEMENTO候選在生成前訂正，候選與勘誤保留於`cave-mpp6-source-v1/references/`；黑色色號只含部分字形，不能單獨判字。[攻略地名列表](https://gamefaqs.gamespot.com/pc/575849-zeliard/faqs/13336)只輔助比對Dorado拼字，不作像素、正常可達或官方中文譯名證據。
- 五張選定原圖由內建imagegen產生，紫岩及details_b第二版修畫風；七次請求及所有輸入／輸出SHA均保留。圖帶864×24 SHA為`059853dce8d844f218194172e1d6bc1a586371750191f729993e849a4d2e6703`；432×768金黃材質為`d96f553170cf4870f68ec20811dfeb424ab4bed935e881dbea298b567ec2f59b`，紫岩為`5916121f2097de409fc8e744f7af64c8ac8f3764b08c881c44b87bd4b7839c6a`。按生成圖比例取樣並垂直重複，不把原生8×8放大冒稱HD；原圖及初版拉伸研究不覆寫。
- 四圖1161216像素、全36格、雙材質與非連號字帶、缺第二材質及其他原子拒絕通過。五個新增家族共存，舊四家族各三圖及Malicia回歸通過。結構快照不執行VM，不當正常MPP6可達證據。
- 候選`52c0716372c7b527efb0b062ce7db2fabb0d515aa1b0409a120fe894b290226d`的MPP1正常首戰／回城35點完整原版狀態、USR與mtime相同。實際視窗v2兩組切回1152000遊戲區像素相同並正常回城、退出0；v1門框邊緣Up未轉場保留原身份。正常MPP6進出、捲動、事件及全遊戲HD仍待完成。

## 2026-10-20 MPP7的可見圖格與原版空白槽

SAR3記錄80的解壓SHA-256 `89bf9b7d268985362e099b6b0db29c04ca4b48444857f70a7e0bacf82df26b14`，預期dosgolem實模式11B1:8000的48-byte header為`80d6b9f0ff2e7230cc10a610acf8e286ffc2271dae2b8ba386895619406afbc1`，8030起61×48bytes為`8f653d969c43b66220777fdd56ce331c5297df9d006163b90be3b1fbe51b32f8`。來源及選擇端沿037既有原版解壓、01B1:7E9B／7ECD證據，不新增原始函式命名；正常MPP7載入仍未知。

- 工具為`tools/hd_cave_family_art.py`、固定eob-audio-capture:20261009-r3、Python 3.11.2、NumPy 1.24.2、Pillow 9.4.0與Go1.24.13。腳本、輸入與產物SHA在`cave-mpp7-complete-v1/manifest.json`、`cave-mpp7-art-v1/generation.json`及`cave-mpp7-final-audit-v1.json`。
- 已證實的資料：MP70／MP71／MP72／MP7D同選MPP7。40、41、42的64個RGB像素及48個轉換資源bytes各全零，資源header與61格完整雜湊亦核回來源；保留相同ID的透明24×24槽，不當缺圖。原始證據在`cave-mpp7-source-v1/blank-slots.json`。
- 赤紅岩面15個ID、白熱核心五個ID是外觀分組，不作傷害、碰撞或事件規則。其餘33個細部及五格字帶各自註冊。MP70兩處相鄰36、37、38、39、46，完整字帶目視為LLAMA；不推定官方中文譯名。
- 六張內建imagegen原圖直接註冊為1464×24圖帶與兩張432×768材質，不放大原版8×8冒稱HD。圖帶SHA為`74e79828e3f1921388118ac82de7b1550670de109df4ab0954d7a9dfce5adea7`；赤紅材質為`b593a591f787a075fb1a9b37bfa458048b744b0cca18992c462ae4a69b8eaf47`，白熱材質為`297dd13c6bcc41807856f979aaaf455fd708f5f99a3a7ee160b758e91b07baa0`。51..61的註冊RGBA雜湊十一種皆不同，沒有合併成同一材質；六份實際生成參考重生SHA相同。
- 四份靜態地圖未使用的ID為[34, 45, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61]，其事件或動畫用途仍未知。四圖1161216像素、全61槽、58個可見圖格、空白槽及原子拒絕通過；MPP6四圖、其他四個新增家族各三圖與Malicia回歸通過。結構快照不執行VM，不當正常MPP7可達證據。
- 候選`1df207def6a9fa90add5f808887114b2fe683fbd7a20d8e4a9da89d911d2c750`的MPP1首戰／回城35點完整原版狀態、USR與mtime相同。實際視窗兩組切回1152000遊戲區像素相同，正常回城與DOS退出0。MPP7正常進出、捲動、事件及全遊戲HD仍待完成。

## 2026-10-20 MPP8／MPP9／MPPA／MPPB圖面來源

已證實：SAR3記錄81..84的原始解壓、48-byte header及逐槽48-byte轉換資源沿cave-terrain-sources-v1/receipt.json與既有原版常式證據。fixture從檔案偏移0放header，對應預期11B1:8000；偏移48開始逐槽資料，對應預期11B1:8030。本輪未以正常後段RAM證實可達性，不把檔案偏移當正常載入位址證據。

| 家族／記錄 | 原始解壓SHA-256 | 完整轉換fixture SHA-256 |
|---|---|---|
| MPP8／81 | `47ddb3770034d6dfa890e41daaa8c1337f1838c9b752371fc6e8da103a367966` | `8a5a3e0c71172f033449795e2650906b6e62540ef6e99759aa2a90ea07efa063` |
| MPP9／82 | `b60665deae37a452a76bd79ff1f9bc1695b2ddc33752058fdc8122a3e9ee253e` | `9035c8c51d1c0a350ddb5a18ec38d34b70ac77a64d191e76d8f9df26961a8891` |
| MPPA／83 | `6610bddf655550d5157ae228172851b92b5204b30a7ae2b6e5d1a30122995521` | `bbad71f3dd200ddec78763dba359fc878210cabae849492b1f18da711276789c` |
| MPPB／84 | `cd053fb5fc7025b95207cb0d238c5bf17bf4b08611c2e44112cdd099b72684a6` | `43f014e537acd84f22a8ed36cbc46291256fc9074b0e945155aa05a91091392f` |

已證實的靜態地圖關係為MPP8六圖MP80／MP81／MP82／MP83／MP84／MP8D、MPP9的MP90、MPPA的MPA0及MPPB的MP73。各source-review.json保留逐ID使用數及鄰接樣本。MPP8的29..33完整拼為PUREZA，34／35／36／13／14拼為ESCO；原始SAR2檔案偏移0x410c7的Esco village僅核拼字，原檔SHA與bytes保存在label-review.json，不當圖形映射或正常到達證據。

已證實的空白來源為MPP8的17／18／19與MPPA的20，RGB及每槽48個轉換bytes全零。四份reference-check.json保留檢查；正式Go解碼另以64個全零色號核對。其餘102個可見圖格皆有生成及註冊圖面。連續材質、灰綠／深藍分組與字帶覆繪是明示外觀處理，不推定碰撞、事件或動畫語意。

素材工具使用Docker內Python3.11.2、Pillow9.4.0、NumPy1.24.2，Go1.24.13只接入呈現層；共用dosgolem來源未改。九圖2612736像素比對的是正式Draw與獨立離線預覽，結構快照不執行VM。正常MPP1的35點與GUI v2另作回歸；四個新家族正常到達、捲動及事件仍為未知。證據入口為[規格037](docs/spec/037-all-cave-terrain-hd.md)及workplace/full-hd-v1/cave-final-families-audit-v1.json。

## 2026-10-20 MPP1補充圖格與原版共用入口

已證實的SAR3記錄74解壓SHA-256為`0b991410e2069c785b2db5781f2f662862d2985a7c536438cbe892b4ebdce32c`；48-byte header SHA為`2e4ad5efcf25dfafa873c222d950ad8d524048329b1b6c452dba159a2eee953a`，25乘48bytes轉換圖格為`bb873820c81d32ff533a5fb1e74ec0f0fca64f917d89d6f7dc57796dfe516de7`。來源位址為既有正常11B1:8000及8030契約，fixture從檔案偏移0／48保存；兩種位址基準分開。原版MP10靜態x127／y31有16..20，x60／y5有21..25，完整拼合目視為SATONO及MURALLA；第15格出現32次。MP1D沒有15..25，不因該圖未用就刪除美術。

勘誤：舊terrain-handpaint-v2.png第15格為透明，最初憑整張圖誤認右側片段已畫好，已由complete-v1接觸表及空alpha推翻。complete-v2改用新生成第15格，舊14格及材質不改。原始細部alpha最大254是有效透明圖，不歸零或改成255；註冊按原生非零邊界裁切。初版灰色字面診斷用128錯誤，實際色盤為125；後續暗色字形與完整圖牌、SAR2原文拼字共同核讀，失敗診斷不當空白證據。

正常首訪兩條鍵序在cave-mpp1-pair-v1/的cave-entered.presented.memory與body.memory顯示共用入口ID73、74、75、76、77、78、88..97。已證實的完整5×5矩陣及21個去重ID的資源包SHA仍為020的`3fd046b8c1281a013bfe1c8dfa627a663fec2eac5cebe7e3df025a63c1db6633`。來源、檔案SHA、原始矩陣、環形圖格位置及共用ID在cave-mpp1-source-v1/normal-common-evidence.json。這只回查既有Muralla入口，不由名稱猜一般MPAT索引或用途。完整矩陣與兩側資源守衛通過才能保留原版共用下層；孤立73、任意26及來源變動仍拒絕。

新圖由內建imagegen製作，generation-v2.json保留三次完整請求與輸入／產物SHA；註冊圖帶SHA為`eeb7ce0286a08aafdca396379109a37dd6e03e163680fd58b2714ddfedcbe7cd`。Docker內Python3.11.2、Pillow9.4.0、NumPy1.24.2以分軸整數權重重建舊14格／材質，Go1.24.13以既有caveArea呈現，三份構圖870912像素相同。結構測試不執行VM；正常MPP1的35點及GUI另驗。Satono正常到達及後段家族仍未知。入口見[規格037](docs/spec/037-all-cave-terrain-hd.md)與workplace/full-hd-v1/cave-mpp1-final-audit-v1.json。

## 2026-10-21 城鎮背向圖與正常進門居民

已證實：正常城堡進門的兩個TMAN背向表為01B1:6A71及6A53，皆取20、22、24、21、23、25，原16×24 RGBA的SHA-256同為70ca7bed8492e9f7c4a49a951ba602782c57cc94ba62d3b07ee4b1ddd5fbf2d8。這是段內偏移，不當SAR檔案定位。完整原始記憶體、TMAN、遮罩及色盤SHA見[038](docs/spec/038-town-player-back.md)。正常快照兩份1MiB記憶體SHA分別為50ba7bef235d3c9ac684feffd86b2b318a9e9f43b9b45f1bdeb8184d9e19be0d及368b1ce2e672bfe5c48c88537401c3fbe5b3917c8a8216c60659cb5379ad9d00。

已證實：這兩張來源的四筆居民記錄仍在，但C017起地圖各該欄第5列已還原為[SI+3]的底圖。IDA9.4由既有唯讀castle.i64建立一次性副本，原始輸入primary.segment為65536bytes，SHA-256 aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d。IDA base0對應此快照dosgolem01B1段內偏移，回1MiB檔案定位另加1B10。sub_6C4E在6C5A讀底圖、6C67算C01C加欄號乘8、6C6B以原MOV [BX],AL還原。原名與bytes及分級索引沿[town-back-source-index.json](docs/town-back-source-index.json)，匯出town-back-ida-v4/evidence.json自動合併語意。

勘誤：TownScenery接受phase4編號不足以證明正常畫格可重建。首次兩個來源分別在96,40及64,40失配，原因是解碼器仍畫居民。以原版底圖加背向主角獨立重建28672色號像素完全相同後，才允許phase4的已驗居民還原形狀。未驗的混合居民配置保持拒絕，不從資料表存在推論可見。

圖面沿內建imagegen保存於town-neutral-art-v1/generation.json，完整提示詞與兩參考SHA保留；48×72 PNG為d84525e69589decf0045cc32b35032eaea862c967e9876df5294cf208d3096f4，重生相同。Python3.11.2／Pillow9.4.0另核兩朝向共6912個HD像素，最大每色偏差2，來自整數alpha取捨，僅屬明示高清外觀容差，不用來容許原版畫面或原版狀態差異。

已證實的正常路徑為兩個獨立dosgolem冷新局，以原版方向鍵從左右朝向走到主門再按Up；兩條都在完成回呼取得phase4並合成HD，完整原版狀態相同。實際Xvfb視窗尚未捕捉背向畫格，未知其可見時窗與前端取樣的關係，不因此宣稱已驗GUI或修改原版等待。其他城鎮與其餘phase4配置仍未知。稽核入口town-back-audit-v1.json明確保留此限制。

## 2026-10-21 小屋、旅館與室內圖塊字典

已證實：長老正常96×64畫格的96個8×8圖格可全部反查KENJYA字典。重建索引表與SAR2檔案偏移116348及正常01B1:A9B6的96bytes完全相同。連續排GRP像素及直接排圖塊的兩種研究圖皆不能當完整原圖；前兩列近似相同是早期索引多為連號，之後重用空白等圖塊才顯出錯位。此反例保存在interior-source-v1/。

已證實的OMOYA程式輸入為SAR2記錄11，原始payload SHA-256 738f6ff33e36685ae0908605d4db50e4f5df7486e098401b0bc5e033972aee37；圖像記錄19解壓SHA ec6b6a1c4511d3ac627e6f13dfc6e8bc3476fa72051ee9794aec63653eb7fe25。INNAPRO為記錄16，程式SHA 7b3dedeb10111d671b0d9245ea457d6235cb620d629ef9f49616348cf8073da5；INN圖像記錄24解壓SHA f01edaecf5f67a7ff9c8dea2209a3d73a66d5b97b738c40af31541e3e36cc8b8。兩者的RLE及三平面算法沿既有原版常式收據，不重開解碼研究。

IDA9.4對每個BIN採base A000對應檔案偏移0，未與SAR偏移混用。sub_A104從A129取272個ID，BX=0C1E、16列17欄，得到136×128、位置96,30的小屋石像。sub_A1AA從A1CF取96個ID、BX=0717，得到96×64、位置56,23的旅館。旅館A279兩份2×2及A281四份5×4表由原函式分別讀取，位置皆64,39。七個合成表項有四種不同像素，基底／ambient-0／rest-0／rest-1相同；別名只作導覽，不推定未驗事件觸發或時間。原名、bytes、xref與分級索引沿[室內索引](docs/interior-source-index.json)及interior-ida-v2/。

已證實的正常小屋入口：冷新局城堡01B1:C009指向C3CA，原始兩筆為95、1及52、0。先保留原數值，再正常右行至世界位置760並Up，實際載入OMOYA及顯示石像，才將95、1與此入口關聯。正常hut-stable完整RAM SHA為54ee22787d0acfe9f07331543cd1a0ce06bfbd2f2842adef1a526e9c3768c98e，轉換6480bytes及17408像素與原圖相同。正常長老完整RAM SHA為9846e900784145498af99f4589a19503fb96dd9821a54ce7525aa5850aed14d1，另核7968bytes及6144像素。這兩條來源不代替旅館正常可達性。

手繪圖由內建imagegen生成，六次完整請求及五張選定圖保存在interior-art-v1/generation-v2.json。Python3.11.2、Pillow9.4.0及既有整數面積取樣註冊到3倍，不先縮成原版。小屋姿態修正與旅館共用背景限制為明示外觀處理，原版表及程式不改。GUI小屋原版／HD各156672像素與獨立PNG完全相同，正常原版配對及GUI另列；旅館仍為靜態構圖與來源守衛。證據入口為[039](docs/spec/039-interior-resources.md)及interior-final-audit-v2.json，不宣稱全室內完成。
