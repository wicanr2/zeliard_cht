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
