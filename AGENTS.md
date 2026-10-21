# 神劍封魔 Zeliard：繁體中文化與 HD

GUI來源重生：[hd_gui_frame_probe.go](tools/hd_gui_frame_probe.go)，核保存日誌的每筆暫存器／步數與目標RAM及完成來源雜湊，不以來源重播代替GUI驗收。
無招牌共用門框：[規格045](docs/spec/045-cave-untitled-doors.md)及[來源盤點](tools/hd_cave_door_sources.py)，不把新變體套成Muralla入口。
門框圖面註冊：[十變體註冊工具](tools/hd_cave_doors_art.py)，內建生成基底及標記直接取樣120×96，保留原始輪廓。
門框呈現：[來源守衛與十圖載入](internal/prototype/hd/cave_doors.go)，依045的原始記錄與完整矩陣唯讀覆繪。
門框驗證：[正常來源及十變體測試](internal/prototype/hd/cave_doors_test.go)，分列正常站位、結構圖面與原子退回證據。
門框稽核：[正式收尾工具](tools/hd_cave_door_runtime_audit.py)，核來源／十圖／正常配對及實際視窗，再銜接覆蓋清單。
已存裝備缺圖：[規格046](docs/spec/046-cave-saved-gear-hd.md)及[來源研究](internal/prototype/hd/cave_saved_gear_study_test.go)，直接重解正常平台配對已存快照，不重蒐相同姿態。
青藍刀光註冊：[十姿態圖面工具](tools/hd_cyan_weapon_art.py)，以完整原生輪廓核對既有姿態，另註冊已生成的青藍圖面。
青藍刀光接入：[固定鍵與原子圖組](internal/prototype/hd/cave_weapon_cyan.go)，沿046保留舊十圖及未知來源退回。
青藍刀光驗證：[已存來源與原子回歸](internal/prototype/hd/cave_weapon_cyan_test.go)，核十鍵、直接高清像素及來源未寫入。
青藍刀光稽核：[來源／圖面／正常路徑／GUI](tools/hd_cyan_weapon_runtime_audit.py)，分列獨立混色與視窗抽樣範圍。
主角保存姿態註冊：[十二來源圖面工具](tools/hd_saved_player_art.py)，核既有持盾輪廓，直接註冊新圖與六種可見色彩變體。
主角保存姿態接入：[十二鍵與原子圖組](internal/prototype/hd/cave_player_extra.go)，依046保留原55姿態及原版當次配色時機。
主角保存姿態驗證：[來源／混色／遮擋／原子載入](internal/prototype/hd/cave_player_extra_test.go)，核12鍵、六個變色及原55圖不變。
主角保存姿態稽核：[來源／圖面／正常配對／GUI](tools/hd_saved_player_runtime_audit.py)，區分完成來源缺圖歸零與其餘來源拒絕。
門框原始載入端：[IDA9.4窄範圍匯出](tools/ida_cave_door.py)，保留原始函式、位址與運算元，未分級項標為未知。

魔法石正式來源與高清圖層：[cave_magia.go](internal/prototype/hd/cave_magia.go)，依[044](docs/spec/044-cave-magia-presentation.md)的完整來源及可見／清除契約接入。
魔法石正式測試：[cave_magia_test.go](internal/prototype/hd/cave_magia_test.go)，核27正常來源、獨立高清像素、清除／重疊、來源及圖面原子拒絕。
魔法石正式稽核：[hd_cave_magia_runtime_audit.py](tools/hd_cave_magia_runtime_audit.py)，核原版狀態、GUI獨立像素及原貌切回；未匹配地形站位仍列出。

魔法石呈現研究：[規格044](docs/spec/044-cave-magia-presentation.md)、[正常拒絕來源探針](tools/hd_cave_rejected_probe.go)與[差分研究](internal/prototype/hd/cave_rejected_study_test.go)。研究輸出不作HD完成證據。
魔法石原始定位：[分級索引](docs/cave-magia-source-index.json)與[IDA9.4窄範圍匯出](tools/ida_cave_magia.py)，沿用既有一次性資料庫，不改原始函式與位址。
魔法石高清圖面註冊：[hd_cave_magia_art.py](tools/hd_cave_magia_art.py)，依102原生非零輪廓直接取樣24×24，素材與提示詞留本機研究目錄。
魔法石研究階段稽核：[hd_cave_magia_audit.py](tools/hd_cave_magia_audit.py)，核27來源重建、原始定位與圖面註冊；只證READY，不當正式HD完成。

平台來源、圖面及拒絕回歸測試：[cave_common_test.go](internal/prototype/hd/cave_common_test.go)，契約見[規格043](docs/spec/043-cave-common-scenery.md)。
正常平台HD配對命令：[verify_cave_platform.sh](tools/verify_cave_platform.sh)，在既有Go工具容器內執行，所需掛載與輸出見WORKLOG。
平台實際視窗：[hd_cave_platform_smoke.py](tools/hd_cave_platform_smoke.py)，沿既有Xvfb入口，只送正常鍵盤與工具列操作。
平台收尾稽核：[hd_cave_platform_audit.py](tools/hd_cave_platform_audit.py)，核來源、測試、高清像素及原貌切回，歷史收據保留身份。

## 每次開工

對使用者的回覆與新文件預設使用繁體中文，句子簡短、直接。`~/.claude` 維持唯讀；外部方法由 Codex 知識路由選擇，不回寫上游。

1. 先讀 [CONTEXT.md](CONTEXT.md)，以目前程式與最近收據核對狀態。
2. 再讀 [WORKLIST.md](WORKLIST.md)，只推進尚未完成且前置條件已具備的工作。
3. 比對 `/home/anr2/.codex/knowledge-base/knowledge-router.md`，說明命中列與實際載入的入口。不可預讀整個知識庫。
4. 遵守使用者提供的共用主機衛生規則。所有分析、解包、建置、測試、遊戲與 GUI 工作都在 Docker 內執行。

## 已授權目標

2026-10-20 最新完成標準由使用者明確限定：第一個洞窟能玩、抽樣對話中文、HD正常、手把串接，工具列內可設定手把與鍵盤，選單中文。以此清單及代表性整合抽測收尾，不要求完整通關、逐筆文字／場景驗證或完成所有後段美術。未抽驗實體裝置與範圍照實列出，不延伸成新的完成閘門。

2026-10-20 使用者調整驗收標準：採代表性正常流程抽測，不要求完整通關。已有共同來源與顯示契約的批次內容，抽代表樣本，不為同一機制逐筆重跑。適當測試通過後轉向製作與交付，保留已知問題及未抽驗範圍，不把抽測等同所有內容均已製作完成。

- 參考 `/home/anr2/cht/softworld_san`，使用 dosgolem 推進《神劍封魔》的繁體中文化與 HD。
- 工具列提供搖桿／手把選項，能選擇並綁定系統已辨識的藍牙與 USB 手把。
- 建立私人 GitHub 儲存庫。README 的故事背景以本地中文手冊為來源。
- 先完成原版可執行、可觀測的垂直切片，再接中文與 HD。不得用猜測填補原版規則。

2026-10-18 使用者已選定直接以 dosgolem 執行原版，另指定參考 `/home/anr2/cht/psychic-war` 的中文與 HD 覆繪方式。後續選定比較圖 B「手繪奇幻」，只決定畫風。不另建玩法重製引擎，不再詢問已定架構或畫風。HD 正式倍率、預設主題、構圖及發行平台仍待確認。

2026-10-20 使用者新目標為「完成全遊戲HD」。沿既有3倍畫布、手繪奇幻、dosgolem覆繪及已交付三平台方向擴充。第一洞窟只是前版完成範圍；現行完整覆蓋與抽樣閘門見 [規格034](docs/spec/034-full-game-hd.md)。不重問既定畫風或架構，不把安全退回原貌當成該素材已HD。

## 參考來源與適用範圍

| 來源 | 用途 | 不可直接沿用 |
|---|---|---|
| `/home/anr2/cht/softworld_san/AGENTS.md` | 證據分級、規格閘門、唯讀原始資料 | 該遊戲的編譯器、位址、文字編碼、玩法與兩版本規則 |
| `/home/anr2/cht/softworld_san/docs/spec/021-hd-art.md` 的第 5、6 節 | 原貌／HD 分離、工具列、疊層與正常路徑驗收 | 寫實手繪畫風、4 倍倍率、640×408 畫布、AI 選項與 Esc 配置 |
| `/home/anr2/cht/softworld_san/internal/ui/windowbar.go` | 工具列繪製、狀態與命中區的分離 | 固定寬度、欄數與遊戲專屬標籤 |
| `/home/anr2/cht/softworld_san/cmd/san1/windowbar.go` | 前端攔截輸入、畫面合成與語言切換 | 已實作手把支援的假定；該檔目前只提供鍵盤／滑鼠選擇 |
| `/home/anr2/cht/dosgolem/README.md`、`CLAUDE.md` | 決定性 DOS 執行器、分層與工具鏈 | 已支援 Zeliard 全遊戲的假定 |
| `/home/anr2/cht/psychic-war/AGENTS.md` 第 3、5、6 節及 `cmd/psychicwar/main.go` | 原版執行 → HD 圖面 → 中文疊字的合成方式、輸入與適配層歸屬 | 遊戲專屬位址、8×8 幾何、字型散布許可與已定稿美術 |

## 原始資料與手冊

- 原始輸入清單、SHA-256 與現況放在 [CONTEXT.md](CONTEXT.md)。原始 ZIP、RAR 不覆寫、不重新打包。
- 先檢查壓縮檔路徑，拒絕絕對路徑與上層跳脫。保留每個遊戲檔案的原名、大小與雜湊。
- 遊戲包內附帶的存檔視為來源檔。正式存檔、設定修改與測試寫入使用另外的可寫層。
- 中文手冊是繁中術語的首要參考。每項轉錄附壓縮檔內的掃描檔名、印刷頁碼與核對方式；OCR 結果必須對照原圖。
- 手冊規則與執行結果衝突時，保留差異並以固定版本的原版實跑為行為依據。
- README 故事只改寫已讀到的手冊內容。沒有手冊證據的角色譯名、年代、道具與劇情細節不補寫。

## dosgolem 接入與證據

1. 先讀 `/home/anr2/cht/dosgolem/docs/spec/000-index.md`，再依需求讀 `001-scope-and-mvp.md`、`005-oracle-api.md`、`006-layering.md`。
2. 首次接入使用該專案 `cmd/probe`。先檢查當前命令列旗標，再用固定指令預算跑本地 `ZELIARD.EXE`，記錄未實作指令、DOS／BIOS 服務、顯示模式、檔案存取與停止原因。
3. 正式收據須由 dosgolem 自行重生。DOSBox／DOSBox-X 只補足能力缺口並提供輔助基準，不得換名成 dosgolem 收據。
4. 本專案使用的 dosgolem 來源唯讀掛載。若需修改共用執行器，先建立獨立工作副本，記錄來源 commit 與差異；不直接改動 `/home/anr2/cht/dosgolem` 或其他專案工作副本。
5. 程式專屬位址、流程、翻譯掛鉤與美術對應留在本專案適配層。機器層不得出現 Zeliard 專屬條件。
6. 若使用 Go 套件接入，參考 `/home/anr2/cht/dosgolem/README.md` 的 Go workspace 方法。在容器內接本機來源；不填入不存在的模組版本。
7. 收據記錄來源檔及雜湊、dosgolem commit、工具版本、初始設定、輸入序列、指令／虛擬時間預算、輸出雜湊與限制。
8. 亂數測試於執行前固定條件，記錄兩側種子或受控亂數方法；不得重擲到通過。

## 規格與逆向邊界

流程為：原版證據 → DRAFT 規格 → 證據審查 → READY → 實作 → 同狀態驗證 → CONFORMED。

- 未達 READY 的 prototype 不進正式路徑。正式架構、資料格式、存檔相容性或玩家體驗的改變須先確認。
- 證據等級統一為「已證實、強推論、假說、未知」。原始 bytes、位址空間、工具版本與出處不可省略。
- 只追查目前玩家功能需要的最小充分證據，不逐行翻譯整支 EXE。
- 需要反組譯時讀 `/home/anr2/.codex/skills/use-ida-pro-9-4/SKILL.md`，沿用既有 IDA 9.4 工具鏈。保留函式名、位址、結構偏移與運算元，用附加註解記語意。
- 硬體時序先讀 `/home/anr2/.codex/knowledge-base/local/retro-platform-spec-first.md`。規格近似不稱為原版逐週期一致。

## 中文化

- 先查實際文字來源、編碼、控制碼與繪字流程，不因有中文手冊就假定遊戲是 Big5。
- 譯文與顯示掛鉤須能回查原文字串、來源位置與使用情境。資源鍵、判定字串與存檔識別不可被翻譯取代。
- 名稱優先採手冊有證據的譯名；手冊未收錄者另建待校訂項，不自行宣稱官方譯名。
- 字型記錄來源與授權。中文顯示須指定安全文字區、字距、行距、換頁與溢出政策。
- 驗收至少涵蓋開場、對話、商店、狀態／裝備、存讀檔與長字串。翻譯不得改變遊戲流程或數值。

## HD 與工具列

- 原貌保留獨立、可擷取的基準畫面。HD 屬明示的顯示差異，不取代原版驗證。
- HD 只在已確認的資源與繪製位置接入，保存來源鍵、原圖／產物雜湊、尺寸與權利分類。缺圖、損壞或不符來源時須安全回退。
- 高清素材直接以目標解析度合成；不得先縮成原版大小再放大並稱為高清。
- 不改碰撞、地形、角色位置、動畫節點、命中範圍或規則。切換主題不得重開局、重擲亂數或改寫存檔。
- 工具列獨立於遊戲畫面，不遮擋或裁切遊戲資訊。先量本遊戲畫布，不照抄參考專案尺寸。
- 工具列的開關、暫停政策與快捷鍵先核對手冊及實跑。手冊列 Esc 為暫停，不得直接複製參考專案的 Esc 攔截方式。

## 藍牙與 USB 手把

| 功能 | 必須滿足的行為 |
|---|---|
| 入口 | 工具列提供「搖桿／手把」，可用鍵盤、滑鼠與已綁定手把開啟 |
| 裝置選擇 | 列出系統已辨識裝置，顯示名稱、連線狀態與目前選擇；不承諾平台無法提供的連線類型或電量 |
| 綁定 | 可把方向鍵、類比軸與按鈕綁定至已驗證的遊戲動作；可重新綁定、取消及還原設定 |
| 操作區隔 | 綁定中的按鍵不得同時送入遊戲；工具列操作不得誤觸攻擊、跳躍、購買或退出 |
| 類比輸入 | 處理死區、軸方向、持續按住與放開；數值由規格與測試決定 |
| 生命週期 | 支援啟動後接入、拔除／斷線及重連；失焦、切換裝置與斷線時釋放已按住的動作 |
| 設定 | 裝置與綁定設定獨立於遊戲存檔，重啟可恢復；不得只用當次列舉索引辨識裝置 |
| 備援 | 未接手把或綁定失效時仍可用鍵盤操作；不可鎖死設定頁 |

藍牙配對及 USB 裝置辨識先由作業系統完成，遊戲工具列負責裝置選擇與映射。若要在遊戲內管理藍牙配對，另立平台權限與介面規格，不把它默認為已完成。

先建立動作與原版輸入的對應，再決定送入掃描碼或原版搖桿介面。`STICK.BIN` 的存在只證明包內有該檔，不證明現代手把已接通。

自動測試驗證映射、按下／放開、死區、輸入隔離與設定恢復。實機另驗 USB 與藍牙連線、斷線及重連，記錄型號、平台與結果。合成事件不算實體手把驗收。

## Docker 與工作區

- 一次性容器用 `--rm`，設定記憶體、CPU、程序數及逾時。預設 `--network none`；只有下載或取依賴時開網路。
- 掛載前檢查來源存在、類型與 UID/GID。原始輸入及參考專案唯讀，只讓本專案明確輸出可寫；可寫容器使用目前使用者 UID/GID。
- 研究、中間產物與測試收據集中在 `workplace/`。建立新結構前讀 `/home/anr2/.codex/knowledge-base/local/project-document-responsibilities.md`，並在 WORKLOG 記錄用途。
- 優先沿用既有映像；工具缺件時修正可重現工具鏈，不回主機執行工作負載。
- 收尾檢查本專案容器與 root-owned 檔案，停止並刪除不再需要的容器。不得全域清理其他專案資源。

## 文件索引與權責

- [武器店動畫規格041](docs/spec/041-merchant-frames-hd.md)：武器店完整圖塊、動畫及正常路徑的增量契約。
- [武器店動畫來源索引](docs/merchant-frame-source-index.json)：原始位址、圖塊表與證據分級。
- [室內模組IDA匯出](tools/ida_room_frames.py)：依固定來源索引匯出原始函式、bytes、交叉參照及分級語意。
- [室內正常畫格擷取](tools/hd_room_probe.go)：保存工具快照與存檔mtime的正常鍵序來源，不注入遊戲狀態。
- [武器店完整來源](tools/hd_merchant_sources.py)：八姿態、兩組局部動畫、正常匹配與未引用圖格的台帳。
- [武器店完整圖組註冊](tools/hd_merchant_register.py)：直接3倍圖面、既有基底及局部動畫的取樣契約。
- [武器店完整姿態合成](internal/prototype/hd/merchant_frames.go)：規格041的唯讀完整來源及直接3倍圖面，不改原版選姿態。
- [武器店完整姿態驗證](internal/prototype/hd/merchant_frames_test.go)：十張正常來源、原子退回、兩基底及購買／離店的原版狀態配對。
- [武器店購買鍵序](tools/merchant-hd-purchase.commands)、[未購物離店鍵序](tools/merchant-hd-departure.commands)：正常冷F7自建存檔，不注入機器狀態。
- [武器店實際視窗抽測](tools/hd_merchant_smoke.py)：正常局部動畫、未購物離店姿態、原貌切回與正常退出。
- [武器店收尾稽核](tools/hd_merchant_audit.py)：核已知254個引用圖格、十張圖、正常交易／告別及實際視窗；未引用兩格不計完成。
- [後續城鎮規格042](docs/spec/042-all-town-resources.md)、[分級來源索引](docs/town-resources-source-index.json)及[IDA選擇端匯出](tools/ida_town_resources.py)：核八份後續城鎮的地圖、圖像與入口，不把靜態地圖當正常到達。
- [後續城鎮正常路徑探針](tools/hd_town_path_probe.go)：只送正常按鍵，保存完整原版快照及新城鎮完成點；`--save`只選固定雜湊USR，副本沿SAVETEST檔名、保留bytes及mtime。`shot`可能落在繪製中，`cave`另取已核01B1:7114完成來源，不寫入位置、道具或存檔狀態。
- [洞窟共用圖格規格043](docs/spec/043-cave-common-scenery.md)及[來源擷取](tools/hd_cave_common_sources.py)：核DCHR39格與正常資源段11B1:8C00..934F，分開已見平台64..66與其餘用途。使用既有Pillow映像產生原圖及來源收據，DRAFT不接正式前端。
- [共用平台候選註冊](tools/hd_cave_common_art.py)：以既有NumPy／Pillow映像直接取樣內建生成原圖至72×24，再依原版平台包絡裁切及分成64～66三格。保存提示詞、SHA、裁切及透明缺口，未經正常配對與GUI不計HD完成。
- [共用洞窟分級索引](docs/cave-common-source-index.json)與[IDA匯出](tools/ida_cave_common.py)：固定正常平台段，保存DCHR descriptor、載入及平台表／圖格查找的原始位址、函式、bytes、xref與推論等級，不修改原版控制。
- [平台正常鍵序](tools/cave-platform.commands)：冷F7讀固定包內SAVEM1副本、交戰後跳上平台及升降再升，不注入狀態；正常配對工具僅接受既有自建USR或此固定SHA來源，後者前史未知。[魔法石路徑](tools/cave-platform-magia.commands)另保存已觀察的呈現來源缺口，不能以一般路徑代替魔法效果驗收。
- [平台來源與圖面審查](tools/hd_cave_platform_contract.py)：核IDA消費端、31份C004三byte表、72個靜態表項、正常升降與1287個不透明HD像素。只作043限定READY依據，不代替正式配對與GUI。
- [共用平台HD](internal/prototype/hd/cave_common.go)：043的原子載入、固定DCHR／程式、三byte平台表及快取三格守衛；從工具列既有HD切換顯示，不改原版平台控制。
- [十份城鎮來源重建](tools/hd_town_sources.py)：按原版descriptor選CPAT／MPAT／DPAT及背景程式，輸出原始前景與明列未核圖格；靜態圖不代替正常畫面。
- [後續城鎮美術準備](tools/hd_town_art.py)：按八份原圖準備固定分格、參考雜湊與內建生成提示詞；未核Pureza圖格與正常接入另列，不直接進正式前端。
- [後續城鎮階段稽核](tools/hd_town_stage_audit.py)：核十地圖17784格、八份生成候選、提示詞與原版路徑收據；新城鎮未到達時保持待驗，不升HD完成。

- [長老手勢規格040](docs/spec/040-sage-frames-hd.md)：長老室內未覆蓋圖塊、手勢及正常路徑的增量契約。
- [長老動畫IDA匯出](tools/ida_sage_frames.py)：固定KENJPRO模組的原始函式、bytes及分級語意匯出。
- [長老正常來源擷取](tools/hd_sage_probe.go)：正常鍵序的手勢與來源擷取，不注入原版狀態。
- [長老完整圖塊來源](tools/hd_sage_sources.py)：八手勢與四動畫圖格的覆蓋、結構組合及正常來源綁定。
- [長老手勢位址索引](docs/sage-frame-source-index.json)：原始位址、圖塊表及證據分級；IDA匯出自動合併。
- [長老手繪圖組註冊](tools/hd_sage_register.py)：八手勢、兩高清動畫圖面及不變區驗證，完整圖面留本機。
- [長老完整手勢合成](internal/prototype/hd/sage_frames.go)：規格040的唯讀完整來源匹配與3倍合成。
- [長老手勢驗證](internal/prototype/hd/sage_frames_test.go)：38結構來源、正常來源、原子拒絕及八手勢原版狀態配對。
- [長老正常鍵序](tools/sage-hd.commands)：冷F7自建存檔、查看力量與離開，不注入狀態。
- [長老實際視窗抽測](tools/hd_sage_smoke.py)：正常查看力量、代表手勢、工具列原貌切回與正常退出。
- [長老收尾稽核](tools/hd_sage_audit.py)：來源、圖組、原版狀態、實際畫面與目前程式身份的完整證據入口。

本次音色與三平台交付研究入口為 `workplace/music-release-v1/`，見 CONTEXT。`internal/prototype/fm/` 載入作品中立的音源動態函式庫；`third_party/nuked-opl3/` 保存固定來源、LGPL 條款與重建所需橋接程式，不適用本專案的 RRSAL。

`cmd/zeliard-preview/bundle.go` 負責原生封包的資料定位與獨立使用者存檔，未指定 `-orig` 才套用。封包提供原版音源核心可替換機制，不限制替換 LGPL 函式庫所需的除錯與逆向。

音訊與封包契約見 [規格033](docs/spec/033-original-music-release.md)。固定來源與素材入口為 [tools/prepare_release.py](tools/prepare_release.py)，建置入口為 [tools/build-release.sh](tools/build-release.sh)，通用執行器差異為 [tools/dosgolem-fm.patch](tools/dosgolem-fm.patch)。

[tools/package_release.py](tools/package_release.py) 組裝原生封包與跨平台獨立存檔入口。Linux使用AppImage，Windows與macOS為ZIP；完整素材與影片維持本機使用。

[tools/release_smoke.py](tools/release_smoke.py) 沿既有第一洞窟正常鍵序抽測，並以X11與PulseAudio錄下實際視窗／音訊；完整錄製只留本機。

[tools/platform_smoke.py](tools/platform_smoke.py) 檢查實際AppImage與Windows ZIP的原生資產定位、獨立存檔及工具列啟動。Wine為輔助基準，macOS交叉建置不當作真機驗收。

[tools/make_promo.py](tools/make_promo.py) 以實際錄影、中文介面截圖及原版AdLib事件音樂合成60秒推廣影片，保存分鏡、字幕邊界與素材身份。

[tools/verify_promo.py](tools/verify_promo.py) 檢查影片串流、音量、靜音、黑幀、凍結與抽樣圖；[tools/audit_release.py](tools/audit_release.py) 核對三平台架構、封包、原版不變與SHA-256總表。

[tools/verify_audio.py](tools/verify_audio.py) 比較實際原版生成取樣與PulseAudio監聽的短段及節奏；不以片段相符聲稱全音訊逐取樣一致。

[tools/capture_music.go](tools/capture_music.go) 以固定原版資料及音源核心錄取未跳過的原版曲段，供本機推廣片使用。

2026-10-20 本輪新增原版配樂音色、三平台本機完整資料包與推廣片；使用者已確認實際時間推進及100毫秒追趕上限。實作在 [internal/prototype/sound/clock.go](internal/prototype/sound/clock.go)，契約在 [規格032](docs/spec/032-presentation-clock.md)。原生封包供使用者直接執行；Codex驗收仍須在Docker內進行。

[.gitattributes](.gitattributes) 只為通用音源差異檔保留unified diff所需的context空白與Tab；Go原始碼仍以gofmt檢查。

本機第一洞窟試玩版由 [VERSION](VERSION) 指定版號，以 [tools/build-local.sh](tools/build-local.sh) 呼叫 Docker 內的 [tools/prepare_local.py](tools/prepare_local.py) 建置至dist-all。啟動用 [tools/play.sh](tools/play.sh)，容器包裝見 [tools/run-local.sh](tools/run-local.sh) 與 [tools/start_game.py](tools/start_game.py)。原版、字型與本機圖面只在full-local，未加入Git或Release。

代表性視窗抽測為 [tools/integration_smoke.py](tools/integration_smoke.py)，只在有界Docker／Xvfb內對實際本機包執行，收據存於同版dist-all的smoke。檢查設定重綁、解除、還原、冷重開，以及中文／HD和第一洞窟基本操作；不擴張到完整通關。

限定範圍的收尾核對使用 [tools/audit_local.py](tools/audit_local.py)，核實際GUI收據、Linux架構、素材manifest與來源身份，輸出同版smoke/audit.json。實體手把未接入時保留此限制，不宣稱已做USB／藍牙實機測試。

工具列鍵盤設定入口為 [cmd/zeliard-preview/keyboard.go](cmd/zeliard-preview/keyboard.go)，穩定鍵名、交換與映射見 [internal/prototype/controls/keyboard.go](internal/prototype/controls/keyboard.go)，相容與保存抽測見 [keyboard_test.go](internal/prototype/controls/keyboard_test.go)。契約沿規格001的「工具列鍵盤設定」，研究與建置收據沿CONTEXT。

[盾牌名稱與數字規格](docs/spec/024-shield-inventory-label.md)保存前置空白、後畫數字與中文安全區的證據閘門；來源存檔只經原版F7讀取隔離副本，取得過程保持未知。

[盾牌文字測試](internal/prototype/overlay/shield_test.go)驗六來源分類、縮小後的容量、兩個正常F7來源存檔的開關／重開、完整原版狀態與數字像素保留。

新增補圖入口：[正常洞窟探索主角補圖規格](docs/spec/023-cave-expedition-player-hd.md)與[獨立圖面及退回測試](internal/prototype/hd/cave_expedition_test.go)。本地原版來源與獨立重建沿CONTEXT的接續探索入口。

[洞窟正常探索鍵序](tools/cave-expedition.commands)保留首次冷F7與原版按鍵，不包含原版資料、還原狀態或注入。

| 文件 | 唯一職責 |
|---|---|
| [tools/bank-persistence.commands](tools/bank-persistence.commands)、[internal/prototype/overlay/bank_persistence_test.go](internal/prototype/overlay/bank_persistence_test.go) | 銀行交易、長老另存、正常退出與冷讀後查詢的正常鍵序與原版／中文配對；只含輸入及自有測試 |
| [tools/residents-muralla.commands](tools/residents-muralla.commands)、[internal/prototype/overlay/town_residents_test.go](internal/prototype/overlay/town_residents_test.go) | 七筆Muralla居民來源的正常冷讀、移動、交談與關閉，以及19點原版只讀配對；不含原版資料 |
| [docs/spec/032-presentation-clock.md](docs/spec/032-presentation-clock.md) | 顯示更新與原版執行時鐘的可撤回比較，未定正式追趕與節奏 |
| [docs/spec/031-audio-stream-scheduling.md](docs/spec/031-audio-stream-scheduling.md)、[預填取消測試](cmd/zeliard-preview/audio_test.go) | 原版串流的後端預取、送樣、等待與暫停生命週期；不改原版時鐘 |
| [docs/spec/030-native-rgb-presentation.md](docs/spec/030-native-rgb-presentation.md)、[internal/prototype/bitmap/rgb.go](internal/prototype/bitmap/rgb.go)、[整幅比較測試](internal/prototype/bitmap/rgb_test.go) | 原版三倍RGB底圖的逐像素等價加速，不改HD與中文順序 |
| [.gitattributes](.gitattributes) | 保留製作名單TSV空欄的分隔Tab；欄數、空白來源與字型由捲字譯表驗證器檢查 |
| [README.md](README.md) | 故事背景、穩定入口、使用方式與現況摘要 |
| [docs/spec/034-full-game-hd.md](docs/spec/034-full-game-hd.md) | 全遊戲HD覆蓋範圍、完成閘門及資源盤點入口 |
| [tools/hd_inventory.py](tools/hd_inventory.py) | Docker內固定三個SAR的194筆邊界、長度與雜湊盤點，分類不冒稱美術完成 |
| [tools/ida_hd_resources.py](tools/ida_hd_resources.py) | IDA9.4一次性資料庫的常駐程式候選指令匯出，保留原始定位及未知等級 |
| [docs/hd-coverage.json](docs/hd-coverage.json) | 全遊戲194筆容器記錄的唯一覆蓋台帳，分類候選與HD／正常路徑狀態分列 |
| [tools/hd_catalog.py](tools/hd_catalog.py) | 核對原版、名稱descriptor及84筆原版解壓收據，重生覆蓋台帳；參數為`--project`及`--output` |
| [tools/hd_decode.py](tools/hd_decode.py) | Docker內解開零前綴RLE候選，輸入由`--original`指定，既有空輸出目錄由`--output`指定 |
| [tools/hd_decoder_oracle.go](tools/hd_decoder_oracle.go) | dosgolem執行擷取原版常式，比對84筆解壓bytes；參數為`-project`、`-decoded`及`-output`，不作正常玩家證據 |
| [tools/hd_town_npc_sources.py](tools/hd_town_npc_sources.py) | 依007已驗模板及遮罩擷取五類40姿態的居民製圖來源，參數為`--project`與既有空`--output` |
| [tools/hd_register_npcs.py](tools/hd_register_npcs.py) | Docker內以`--sources`、`--art`及空`--output`校準40姿態圖集；只產生候選註冊與預覽，不代表正式HD接入 |
| [docs/images/provenance.json](docs/images/provenance.json) | README 四張文件預覽圖的來源雜湊、裁切與輸出身份；原始素材留在本機 |
| [tools/readme_images.py](tools/readme_images.py) | 在具 Pillow 的 Docker 內製作 README 圖片，以 `--project`、`--output`、`--font` 指定唯讀來源、可寫圖片目錄及 Noto Sans CJK 字型 |
| [LICENSE](LICENSE) | 自有程式與文件的 RRSAL-1.0 完整條款及原版素材排除範圍 |
| [CONTEXT.md](CONTEXT.md) | 有日期的目前狀態、決策、來源及下一步 |
| [WORKLIST.md](WORKLIST.md) | 可執行工作與完成條件 |
| [WORKLOG.md](WORKLOG.md) | 追加式工作歷程與環境阻塞 |
| [RESEARCH-LOG.md](RESEARCH-LOG.md) | 手冊、原始檔與執行證據 |
| [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md) | 實際驗證範圍與未驗證項 |
| [tools/prepare.py](tools/prepare.py) | Docker 內安全解包、逐檔盤點與隔離來源快照 |
| [tools/probe_receipt.py](tools/probe_receipt.py) | Docker 內固定預算探針與完整收據 |
| [tools/ida_intro_evidence.py](tools/ida_intro_evidence.py) | IDA 9.4 的片頭段輸入雜湊、原始交叉參照與函式證據匯出 |
| [docs/inventory-source-index.json](docs/inventory-source-index.json) | 裝備頁彩色標題與兩行物品名稱的原始位址、證據等級及來源 |
| [tools/ida_inventory_evidence.py](tools/ida_inventory_evidence.py) | IDA 9.4 裝備頁五函式證據重生，自動合併分級索引 |
| [internal/prototype/overlay/inventory.go](internal/prototype/overlay/inventory.go) | 裝備標題、物品使用白字與已驗NO USE圖示覆繪；操作由原版處理 |
| [docs/spec/025-sword-inventory-label.md](docs/spec/025-sword-inventory-label.md)、[正常鍵序](tools/sabre-source-use.commands)、[配對測試](internal/prototype/overlay/sword_test.go) | 原包SAVEM1冷F7、刀劍油使用與刀劍第二行數字區隔離；驗收狀態依規格，不以試譯代替驗收 |
| [tools/inventory.commands](tools/inventory.commands) | 冷新局交易後開啟、關閉及重開裝備頁的正常鍵序 |
| [tools/ida_dialogue_evidence.py](tools/ida_dialogue_evidence.py) | IDA 9.4 的正常國王初訪段證據，合併對話位址索引及未證實警示 |
| [docs/dialogue-source-index.json](docs/dialogue-source-index.json) | 固定國王段傾印的原始定位、語意等級與證據，不與片頭段混用 |
| [docs/weapon-source-index.json](docs/weapon-source-index.json) | 固定武器店段的商品名稱搬移與 MCGA 清框鏈，保留原始定位及證據等級 |
| [tools/ida_weapon_evidence.py](tools/ida_weapon_evidence.py) | IDA 9.4 武器店來源證據匯出，自動合併分級索引及未證實警示 |
| [docs/source-index.json](docs/source-index.json) | 固定輸入的位址語意、推論等級與證據索引；IDA 匯出自動合併，未證實項附警示 |
| [docs/spec/001-runtime-overlay.md](docs/spec/001-runtime-overlay.md) | 使用者選定架構與前端雛形邊界 |
| [docs/spec/008-town-frame-presentation.md](docs/spec/008-town-frame-presentation.md) | 城鎮繪製批次、完整畫格與原始／呈現收據的試作契約 |
| [docs/spec/009-town-ground-hd.md](docs/spec/009-town-ground-hd.md) | Muralla雙層地面來源、取樣與限定驗收規格 |
| [docs/spec/010-town-player-hd.md](docs/spec/010-town-player-hd.md) | 主角資源鍵、下層還原、手繪合成與待驗閘門 |
| [internal/prototype/hd/town_ground.go](internal/prototype/hd/town_ground.go)、[地面測試](internal/prototype/hd/town_ground_test.go) | 原版雙層地面守衛、高清直接取樣與原子退回 |
| [internal/prototype/hd/town_frames.go](internal/prototype/hd/town_frames.go) | 原版捲動／完成回呼、單一快照、來源時間與撤除 |
| [internal/prototype/hd/town_scenery_reload_test.go](internal/prototype/hd/town_scenery_reload_test.go) | 重讀來源、角色保留及已用／保留區邊界驗證 |
| [text/host.json](text/host.json) | 前端雛形的可見中文 |
| [text/hud-prototype.tsv](text/hud-prototype.tsv) | 狀態欄與場景名稱53來源的單行試譯；未走過來源須經當次入口驗證 |
| [text/town-prototype.tsv](text/town-prototype.tsv) | 城鎮／城堡居民純文字腳本試譯，原版條件與控制碼不依賴譯文 |
| [docs/spec/004-town-dialogue.md](docs/spec/004-town-dialogue.md) | 居民專用繪字流程、動態安全區與原版續頁契約 |
| [docs/spec/005-town-skyline-hd.md](docs/spec/005-town-skyline-hd.md) | 城鎮遠山天空的來源、直接HD合成與安全退回 |
| [docs/spec/006-town-background-mask.md](docs/spec/006-town-background-mask.md) | 完整遠山背景、原版圖塊快取與樹梢遮罩契約 |
| [docs/spec/007-town-scenery-hd.md](docs/spec/007-town-scenery-hd.md) | Muralla前景、原版角色遮擋與商店招牌的試作閘門 |
| [internal/prototype/hd/town_scenery.go](internal/prototype/hd/town_scenery.go) | Muralla來源驗證、前景覆繪與原版角色保留 |
| [internal/prototype/hd/town_scenery_test.go](internal/prototype/hd/town_scenery_test.go) | 原版畫格、獨立人物遮罩、失配及招牌安全區驗證 |
| [internal/prototype/hd/town_scenery_pair_test.go](internal/prototype/hd/town_scenery_pair_test.go) | dosgolem冷讀檔、城鎮步行、居民與跨場的完整原版狀態配對 |
| [text/town-signs-prototype.tsv](text/town-signs-prototype.tsv) | 三個店面招牌的來源與中文試譯 |
| [internal/prototype/hd/town_background.go](internal/prototype/hd/town_background.go) | 依原版已畫快取與透明位元合成完整遠山背景 |
| [internal/prototype/hd/town_background_test.go](internal/prototype/hd/town_background_test.go) | 正常來源、前景保留、失配退回與完整原版配對 |
| [internal/prototype/hd/town_sky.go](internal/prototype/hd/town_sky.go) | 已確認遠景的手繪圖面覆繪，不改原版狀態 |
| [internal/prototype/hd/town_sky_test.go](internal/prototype/hd/town_sky_test.go) | 遠景來源、邊界、退回及正常交談配對驗證 |
| [internal/prototype/overlay/town.go](internal/prototype/overlay/town.go) | 唯讀居民中文覆繪、來源閘門與背景還原 |
| [internal/prototype/overlay/town_test.go](internal/prototype/overlay/town_test.go) | 正常交談配對、113筆容量與來源失配拒絕驗證 |
| [docs/hud-source-index.json](docs/hud-source-index.json) | 狀態列來源、原版完成字形、安全區與證據等級 |
| [docs/spec/026-cave-hud-source-prefix.md](docs/spec/026-cave-hud-source-prefix.md)、[原包存檔鍵序](tools/cave-hud-source-save.commands)、[自建存檔鍵序](tools/cave-hud-prefix.commands)、[配對測試](internal/prototype/overlay/hud_prefix_test.go) | 名稱前綴、原版讀取範圍與兩條正常回鎮／重入路徑；原始0083僅供診斷 |
| [docs/spec/027-cave-shield-profile-hd.md](docs/spec/027-cave-shield-profile-hd.md) | 0093=2、FFFF前綴與原版洞窟角色圖層；來源、素材及HD接入閘門 |
| [docs/cave-shield-source-index.json](docs/cave-shield-source-index.json) | 固定IDA輸入的盾牌圖層分類、原始定位、推論等級及證據；新匯出自動合併 |
| [internal/prototype/overlay/hud.go](internal/prototype/overlay/hud.go) | 當次入口唯一驗證來源的狀態列覆繪，不改原版數值或輸入 |
| [internal/prototype/overlay/hud_scene_test.go](internal/prototype/overlay/hud_scene_test.go) | 正常室內名稱配對及候選來源的唯一匹配／失配退回測試 |
| [internal/prototype/overlay/church_scene_test.go](internal/prototype/overlay/church_scene_test.go) | 教堂場景名稱的正常進出、完成字形、唯讀配對與舊名稱撤除；不修改既有執行器 |
| [text/intro-prototype.json](text/intro-prototype.json) | 片頭單句的覆繪試作譯文，不代表完整翻譯 |
| [docs/spec/013-intro-narration.md](docs/spec/013-intro-narration.md) | 其餘片頭顯示來源與魔王三行宣言的試作閘門 |
| [docs/spec/014-intro-story.md](docs/spec/014-intro-story.md) | 完整故事的58片段、52清框與唯讀中文生命週期 |
| [docs/spec/015-intro-scroll-blocks.md](docs/spec/015-intro-scroll-blocks.md) | 首章、製作名單與六行收尾的來源、視窗與捲動生命週期 |
| [docs/spec/016-shared-inventory-text.md](docs/spec/016-shared-inventory-text.md) | 物品、裝備及使用訊息的共用來源分類與顯示閘門 |
| [docs/spec/017-cave-player-hd.md](docs/spec/017-cave-player-hd.md) | 洞窟主角完整姿態來源、高清透明素材與原版後畫遮擋契約 |
| [internal/prototype/hd/cave_player.go](internal/prototype/hd/cave_player.go)、[測試](internal/prototype/hd/cave_player_test.go) | 規格017的主角姿態登錄、完整遮擋、直接高清合成與原貌退回 |
| [docs/spec/018-cave-weapon-hd.md](docs/spec/018-cave-weapon-hd.md) | 洞窟32×32武器來源、高清下層與最後合成閘門 |
| [internal/prototype/hd/cave_weapon.go](internal/prototype/hd/cave_weapon.go)、[測試](internal/prototype/hd/cave_weapon_test.go) | 規格018的十武器圖面、下層還原、空白／未知退回與最後高清合成 |
| [docs/spec/019-cave-object-hd.md](docs/spec/019-cave-object-hd.md) | 洞窟動態物件來源、美術、層序及正常驗收閘門 |
| [docs/spec/020-cave-portal-hd.md](docs/spec/020-cave-portal-hd.md) | Malicia至Muralla入口的圖塊、裁切、高清及中文圖牌閘門 |
| [docs/spec/021-exhaustion-return-overlay.md](docs/spec/021-exhaustion-return-overlay.md) | 正常昏迷送返長老的中文、續頁／捲動、置中肖像及離開生命週期 |
| [docs/spec/022-return-town-hd.md](docs/spec/022-return-town-hd.md) | 送返後Muralla地形來源、原版讀取範圍、嚴格上界及HD恢復閘門 |
| [internal/prototype/hd/town_return_source_test.go](internal/prototype/hd/town_return_source_test.go) | 五份正常送返來源的四圖層、完成畫格、ID上界及排除區／使用區原子退回 |
| [internal/prototype/overlay/dialogue_retained.go](internal/prototype/overlay/dialogue_retained.go)、[送返測試](internal/prototype/overlay/exhaustion_return_test.go)、[GUI日誌配對](internal/prototype/overlay/exhaustion_gui_test.go) | 四字段完整性、前文保留、失序回退、正常原版／獨立全幅及真實GUI同狀態 |
| [internal/prototype/hd/cave_portal.go](internal/prototype/hd/cave_portal.go)、[入口測試](internal/prototype/hd/cave_portal_test.go) | 完整矩陣與資源守衛、直接高清圖牌及獨立全幅驗證 |
| [internal/prototype/hd/cave_object.go](internal/prototype/hd/cave_object.go)、[測試](internal/prototype/hd/cave_object_test.go) | 規格019的物件圖組、原版清黑、繪製順序、裁切及未知退回 |
| [tools/weapon-hd.commands](tools/weapon-hd.commands) | 正常首戰及固定左右出劍延伸，供018完整原版配對 |
| [internal/prototype/overlay/inventory_sources.go](internal/prototype/overlay/inventory_sources.go)、[測試](internal/prototype/overlay/inventory_sources_test.go) | 規格016的指標表字型分類、來源拒絕及正常物品中文配對 |
| [tools/item-shop-magia.commands](tools/item-shop-magia.commands) | 自建存檔冷F7、正常買入Magia Stone、選用、移除及回城鍵序 |
| [text/intro-credits-prototype.tsv](text/intro-credits-prototype.tsv) | 製作名單51行來源鍵與試譯，保留人名及空白行 |
| [text/intro-outro-prototype.tsv](text/intro-outro-prototype.tsv) | 故事後六行捲動文字試譯 |
| [internal/prototype/overlay/scroll_blocks_test.go](internal/prototype/overlay/scroll_blocks_test.go) | 三組捲字正常原版配對、獨立排版、途中呈現與退回驗證 |
| [text/intro-threat-prototype.tsv](text/intro-threat-prototype.tsv) | 魔王三行宣言的來源鍵與中文試譯，控制碼仍由原版消費 |
| [text/intro-story-prototype.tsv](text/intro-story-prototype.tsv) | 完整故事58個文字／字色片段的來源鍵與中文試譯 |
| [internal/prototype/overlay/intro_story.go](internal/prototype/overlay/intro_story.go) | 規格014的原版字形、清框與唯讀中文生命週期 |
| [internal/prototype/overlay/intro_story_layout.go](internal/prototype/overlay/intro_story_layout.go) | 故事固定來源分段、譯表及有界排版 |
| [internal/prototype/overlay/intro_story_test.go](internal/prototype/overlay/intro_story_test.go) | 完整原版故事配對、47頁獨立排版與途中／退回驗證 |
| [internal/prototype/overlay/intro_threat.go](internal/prototype/overlay/intro_threat.go) | 規格013的唯讀逐字來源、完整畫面守衛與三行中文 |
| [internal/prototype/overlay/intro_threat_test.go](internal/prototype/overlay/intro_threat_test.go) | 正常冷啟動完整原版配對、字形途中與失配退回驗證 |
| [text/intro-scroll-prototype.tsv](text/intro-scroll-prototype.tsv) | 首章 30 行試譯，鍵為本地 SAR 的檔案偏移；專名尚待手冊校訂 |
| [text/dialogue-prototype.tsv](text/dialogue-prototype.tsv) | 國王、長老、武器店、銀行與道具店的對話試譯；原文與控制碼仍留在本地遊戲資料 |
| [internal/prototype/overlay/dialogue.go](internal/prototype/overlay/dialogue.go) | 國王與長老對話逐字、捲動與待續標記的唯讀覆繪試作 |
| [internal/prototype/overlay/labels.go](internal/prototype/overlay/labels.go) | 核對 SAR1、SAR2、STICK 來源的提示、商品暫存列與銀行標題覆繪；保留輸入值及價格區 |
| [internal/prototype/overlay/occlusion.go](internal/prototype/overlay/occlusion.go) | 觀察固定 MCGA 清框事件，依繪製先後裁切舊中文；整框原圖還原後解除遮擋 |
| [internal/prototype/overlay/occlusion_test.go](internal/prototype/overlay/occlusion_test.go) | 多層遮擋、完整還原、未知變化與歷史淘汰的退回驗證 |
| [internal/prototype/overlay/quotation.go](internal/prototype/overlay/quotation.go) | 已核對來源順序的武器報價覆繪；商品譯名共用來源，價格與折抵保留原版數字圖形 |
| [internal/prototype/overlay/bank.go](internal/prototype/overlay/bank.go) | 銀行問句與後接結果的唯讀覆繪，逐筆保留原版數字圖形 |
| [internal/prototype/overlay/bank_test.go](internal/prototype/overlay/bank_test.go) | 正常存提款的完整原版狀態、問句保留、數字圖形與未知畫面退回驗證 |
| [tools/bank-text-gaps.commands](tools/bank-text-gaps.commands) | 冷F7後無存款提款、正常離店、長老普通告別與回城鎮的正常鍵序 |
| [tools/church.commands](tools/church.commands)、[教堂配對測試](internal/prototype/overlay/church_test.go) | 自建存檔冷讀、兩種教堂入口、受傷存活返城、回復及離店的正常鍵序與來源核對 |
| [tools/bank.commands](tools/bank.commands) | 自建存檔冷F7至銀行、存100、提50、查詢、無寶石與離店的原版鍵序 |
| [internal/prototype/overlay/quotation_test.go](internal/prototype/overlay/quotation_test.go) | placeholder／字型／數字圖形合成與未知畫面退回驗證；合成數字不當原版行為證據 |
| [internal/prototype/overlay/labels_test.go](internal/prototype/overlay/labels_test.go) | 冷新局走到長老選單的完整原版狀態與覆繪區域比較 |
| [internal/prototype/overlay/save_load_test.go](internal/prototype/overlay/save_load_test.go) | 中文覆繪開／關的新存檔、退出與冷讀檔完整原版狀態比較 |
| [tools/save-load.commands](tools/save-load.commands) | 上述測試的正常玩家鍵序，不含存檔或注入狀態 |
| [tools/weapon-shop.commands](tools/weapon-shop.commands) | 正常領賞、武器交易、告別與裝備查看鍵序 |
| [tools/weapon-insufficient.commands](tools/weapon-insufficient.commands) | 正常領賞後選第二把武器，走資金不足翻頁、告別及回城鎮的鍵序 |
| [internal/prototype/overlay/shop_test.go](internal/prototype/overlay/shop_test.go) | 武器店中文覆繪與完整原版狀態、金額及區域配對 |
| [text/labels-prototype.tsv](text/labels-prototype.tsv) | 選單、存讀檔、速度、商品名、裝備及銀行金額欄試譯；未走過分支與正式術語待校訂 |
| [internal/prototype/overlay/scroll.go](internal/prototype/overlay/scroll.go) | 首章來源與逐列完成事件的唯讀覆繪；字級與版面仍為 prototype |
| [tools/gui_smoke.py](tools/gui_smoke.py) | 有界 Xvfb 前端啟動與滑鼠開啟設定頁驗證 |
| [cmd/zeliard-preview/main.go](cmd/zeliard-preview/main.go) | 直接執行原版的可撤回前端入口 |
| [internal/prototype/hd/town_player.go](internal/prototype/hd/town_player.go) | 規格010的Muralla主角八姿態、下層還原與高清合成 |
| [internal/prototype/hd/town_player_test.go](internal/prototype/hd/town_player_test.go) | 主角拒絕原子性、素材身份及獨立合成證據比對 |
| [tools/combat.commands](tools/combat.commands) | 正常冷F7入Malicia、出劍、交戰、第一寶石、跳躍與蹲下鍵序 |
| [docs/spec/011-cave-hd.md](docs/spec/011-cave-hd.md) | 洞窟地形、角色、武器與動態圖層的覆繪閘門 |
| [tools/cave_hd_preview.py](tools/cave_hd_preview.py) | Docker內從原版完成快照合成洞窟HD離線研究圖，不接正式前端 |
| [docs/spec/012-cave-frame-sources.md](docs/spec/012-cave-frame-sources.md) | 洞窟完成來源、圖層解碼、守衛與呈現生命週期 |
| [internal/prototype/hd/cave_shield_test.go](internal/prototype/hd/cave_shield_test.go) | 027持盾53份原生／HD獨立合成、未知來源拒絕與正常冷F7完整原版配對 |
| [docs/spec/028-cave-shield-transition-hd.md](docs/spec/028-cave-shield-transition-hd.md) | 正常離洞時持盾FF3A姿態、原版來源重用與高清連續顯示閘門 |
| [docs/spec/029-runtime-audio.md](docs/spec/029-runtime-audio.md) | 同一原版音源的取樣、前端播放、緩衝與停止政策閘門 |
| [internal/prototype/sound/stream.go](internal/prototype/sound/stream.go)、[測試](internal/prototype/sound/stream_test.go) | 029的有界取樣緩衝、雙聲道浮點輸出、欠載／溢位與並行安全 |
| [internal/prototype/sound/wav.go](internal/prototype/sound/wav.go)、[測試](internal/prototype/sound/wav_test.go) | 029的原版單聲道來源錄音、新檔保護及完整取樣驗證 |
| [internal/prototype/sound/original_pair_test.go](internal/prototype/sound/original_pair_test.go) | 029正常冷開機前綴的698次完整原版狀態、PCM及播放緩衝隔離比對 |
| [cmd/zeliard-preview/audio.go](cmd/zeliard-preview/audio.go) | 029可選播放器、暫停／失焦／手把頁停止與新來源恢復 |
| [docs/re-resolution-backlinks.tsv](docs/re-resolution-backlinks.tsv) | 固定模組雜湊與原始位址對應的後續證據、舊規格及必要回填標記 |
| [internal/prototype/hd/cave_shield_transition_test.go](internal/prototype/hd/cave_shield_transition_test.go) | 028四份真實完成、直接高清合成、主體重用與正常原版完整狀態配對 |
| [internal/prototype/hd/cave_transition_lifecycle_test.go](internal/prototype/hd/cave_transition_lifecycle_test.go) | 028實際視窗日誌在Draw時刻的主體生命週期、舊呈現撤除與原版狀態回歸 |
| [internal/prototype/hd/cave_sources.go](internal/prototype/hd/cave_sources.go) | 規格012的原版地形、動態物件、主角與武器色號解碼 |
| [internal/prototype/hd/cave_frames.go](internal/prototype/hd/cave_frames.go) | 規格012的完成來源選擇與撤除；高清美術另由011追蹤 |
| [internal/prototype/hd/cave_terrain.go](internal/prototype/hd/cave_terrain.go) | 規格011的Malicia地形合成；動態、主角與武器保留原貌 |
| [internal/prototype/hd/cave_frames_pair_test.go](internal/prototype/hd/cave_frames_pair_test.go) | 正常鍵序與已記錄GUI原生日誌的完整原版唯讀配對；不取代實際視窗驗收 |
| [internal/prototype/hd/cave_frames_test.go](internal/prototype/hd/cave_frames_test.go) | 洞窟真實來源、增量遮罩、時間界限、深複製與失配退回 |
| [internal/prototype/hd/cave_terrain_test.go](internal/prototype/hd/cave_terrain_test.go) | 洞窟高清直接取樣、角色保留、素材身份與原子退回 |
| [internal/prototype/overlay/combat_path_test.go](internal/prototype/overlay/combat_path_test.go) | 原版交戰與中文狀態列的完整原版狀態、區外及來源驗證 |
| [internal/prototype/hd/item_shop.go](internal/prototype/hd/item_shop.go) | 道具店四組已觀察畫格與局部動畫覆繪 |
| [internal/prototype/hd/item_shop_test.go](internal/prototype/hd/item_shop_test.go) | 道具店HD來源、動畫範圍、原子退回及區外不寫入驗證 |
| [tools/item-shop-branches.commands](tools/item-shop-branches.commands) | 三件商品介紹、空賣、首瓶出售與售後物品頁的正常鍵序 |
| [tools/item-shop-use.commands](tools/item-shop-use.commands) | 首瓶購買、城鎮物品頁、正常入洞窟、使用藥水與重開／關閉的鍵序 |
| [internal/prototype/hd/bank.go](internal/prototype/hd/bank.go) | 銀行三個已觀察畫格的 HD 辨識、合成與安全退回 |
| [internal/prototype/hd/church.go](internal/prototype/hd/church.go)、[教堂HD測試](internal/prototype/hd/church_test.go) | 教堂15個已觀察穩定房間畫格、同快照匹配、共同底圖與五種手勢／燭光局部覆繪；契約見規格001 |
| [internal/prototype/hd/bank_test.go](internal/prototype/hd/bank_test.go) | 銀行 HD 邊界、來源不變及失配退回測試 |
| [internal/prototype/overlay/item_purchase.go](internal/prototype/overlay/item_purchase.go) | 道具店購買、出售確認與三件商品說明的唯讀顯示鏈 |
| [internal/prototype/overlay/item_title.go](internal/prototype/overlay/item_title.go) | 魔法物品店標題的來源、幾何與色彩核對 |
| [internal/prototype/overlay/item_test.go](internal/prototype/overlay/item_test.go) | 道具店正常購買／介紹／出售配對、原版字形與顯示失配退回驗證 |
| [tools/item-shop.commands](tools/item-shop.commands) | 冷讀自建存檔到道具店購買、離店及物品頁的原版鍵序 |
| [tools/bank-no-purchase.commands](tools/bank-no-purchase.commands) | 冷讀自建存檔到銀行未交易告別及正常離店的原版鍵序 |
| [internal/prototype/hd/merchant.go](internal/prototype/hd/merchant.go) | 已觀察武器商人三畫格的HD覆繪與未知畫面退回 |
| [internal/prototype/hd/merchant_test.go](internal/prototype/hd/merchant_test.go) | 商人三姿態、區外不寫入、原版不改動與素材缺失驗證 |
| [tools/weapon-shop-no-purchase.commands](tools/weapon-shop-no-purchase.commands) | 冷新局到武器店未購物離店的正常配對鍵序 |
| [internal/prototype/hd/frame.go](internal/prototype/hd/frame.go) | 手繪雙側邊框試作，完整色號／RGB 匹配與整組退回 |
| [internal/prototype/hd/king.go](internal/prototype/hd/king.go) | 國王兩個已觀察姿態的手繪覆繪試作；未知姿態或色盤保留原貌 |
| [internal/prototype/hd/sage.go](internal/prototype/hd/sage.go) | 長老兩個已觀察畫格的手繪覆繪試作；保留原版外框及四個動畫像素 |
| [internal/prototype/hd/sage_test.go](internal/prototype/hd/sage_test.go) | 實際長老兩畫格、動畫像素保留、區域邊界與未知素材退回驗證 |

居民增量入口：

- [規格035](docs/spec/035-town-residents-hd.md) 定義Muralla居民與主角共同合成及限定驗收。
- [town_actors.go](internal/prototype/hd/town_actors.go) 與 [固定資產表](internal/prototype/hd/town_npc_assets.go) 處理五類40姿態；[資產編譯工具](tools/hd_compile_npcs.py) 由本地已審查註冊重生資產表。
- [合成測試](internal/prototype/hd/town_actors_test.go)、[正常原版配對](internal/prototype/hd/town_actors_pair_test.go) 與 [正常鍵序](tools/town-residents-hd.commands) 保存40來源、圖層、退回及正常玩家路徑檢查。
- [實際視窗抽測](tools/hd_npc_smoke.py) 與 [收據稽核](tools/hd_npc_audit.py) 保存原貌切回、中文交談及程式身份；完整輸出留 `workplace/full-hd-v1/`，不外推其他城鎮。

全遊戲後續來源入口：

- [洞窟來源表編譯](tools/hd_compile_cave_catalog.py)將31份地圖及11組圖集身份寫入[固定資料表](internal/prototype/hd/cave_catalog_data.go)，只含名稱、數量及SHA；[來源辨識](internal/prototype/hd/cave_catalog.go)與[測試](internal/prototype/hd/cave_catalog_test.go)依037來源READY契約處理，結構fixture不當正常路徑。
- [冰窟註冊研究](tools/hd_register_ice.py)將角色無關的MPP4原圖、手繪圖面與不透明材質合成三份地圖離線預覽；不改原始AI圖、不接正式前端，圖形文字與光球缺口另列。
- [冰窟完整圖集](tools/hd_complete_ice.py)補齊HELADA五格標示及兩格藍白光球，保存21格高清圖集、材質、來源與三份離線預覽；正常事件用途及冰窟玩家路徑仍須另驗。
- [MPP2手繪素材流程](tools/hd_peligro_art.py)按實際來源準備標示與11個獨立圖格參考，後續註冊維持原始來源與正常玩家路徑分列；原圖與生成圖只留本機。
- [洞窟家族素材流程](tools/hd_cave_family_art.py)支援MPP3、MPP5、MPP6、MPP7的43／52／36／61槽來源，按明列ID分開材質、字帶與逐格圖樣。MPP7三個原版全零槽另列，不把缺圖視為空白；保存原生外形與輸入／產物SHA，不以候選場景名猜寫文字。使用`eob-audio-capture:20261009-r3`的NumPy／Pillow執行；既有Pillow專用映像不含NumPy。
- [MPP5圖面載入](internal/prototype/hd/cave_mpp5.go)按037固定52格、20個材質ID及32個細部；[MPP5測試](internal/prototype/hd/cave_mpp5_test.go)核三份結構地圖、全格、原子拒絕及四家族共存，正常到達另驗。
- [MPP6圖面載入](internal/prototype/hd/cave_mpp6.go)按037固定36格、金黃／紫岩雙材質與非連號DORADO字帶；[MPP6測試](internal/prototype/hd/cave_mpp6_test.go)核四份結構地圖、全格、第二材質原子拒絕及五家族共存，正常到達另驗。
- [MPP7圖面載入](internal/prototype/hd/cave_mpp7.go)按037固定61槽、赤紅／白熱雙材質與LLAMA字帶，保留三個原版空白槽；[MPP7測試](internal/prototype/hd/cave_mpp7_test.go)核四圖、58個可見圖格、空白資源、原子拒絕與六家族共存，正常到達另驗。
- [MPP8／MPP9／MPPA／MPPB圖面載入](internal/prototype/hd/cave_remaining_families.go)按037固定四組圖帶、材質與MPP8雙字帶；[四家族測試](internal/prototype/hd/cave_remaining_families_test.go)核九份結構地圖、106槽、四個原版空白槽及原子拒絕，正常到達另驗。
- [四家族稽核](tools/hd_cave_remaining_audit.py)從原始來源、註冊、生成提示詞、測試與實際視窗重生收尾收據；舊MPP7收據保留原身份，正常新家族可達性另驗。
- [MPP1補充圖面](tools/hd_mpp1_art.py)註冊第15格及SATONO／MURALLA兩圖牌；1..14與岩面不重畫，來源與接入契約見037。舊圖集第15格實際透明，不能當可用美術。
- [MPP1載入器](internal/prototype/hd/cave_mpp1.go)原子補齊15..25並沿既有14格及材質；[MPP1測試](internal/prototype/hd/cave_mpp1_test.go)核三份獨立構圖、25格、相容與拒絕，不把Satono靜態圖牌當正常到達。
- [MPP1稽核](tools/hd_mpp1_audit.py)核補充圖面、十一家族共存與正常第一洞窟回歸；三份靜態構圖與正常路徑分列，收據入口在CONTEXT。
- [城鎮主角背向姿態](docs/spec/038-town-player-back.md)記錄phase4兩朝向相同的原版來源、美術、正常路徑與接入閘門；不推定其他姿態或場景已完成。
- [城鎮背向素材工具](tools/hd_town_back_art.py)從固定TMAN表重建兩朝向原圖，直接將手繪圖取樣為48×72；面積取樣沿hd_mpp1_art.py的整數權重，不修改原圖或原版。
- [城鎮背向姿態測試](internal/prototype/hd/town_player_back_test.go)核正常進門兩朝向的完整原版畫格、圖面與退回；原版來源與限制沿038，不以注入phase值代替正常路徑。
- [背向來源索引](docs/town-back-source-index.json)及[IDA匯出](tools/ida_town_back_evidence.py)保存主角表與居民底圖還原的原始位址、bytes及分級語意，從既有城堡資料庫的一次性副本匯出。
- [背向視窗抽測](tools/hd_town_back_smoke.py)沿既有城堡GUI工具，在Docker內走正常進門、暫停與原貌切回；輸出只放workplace，不以離線圖代替實際視窗。
- [背向稽核](tools/hd_town_back_audit.py)核來源、圖面、正常原版配對與實際視窗，將GUI待驗和已通過項目分開；輸出至workplace的新檔，不覆寫舊收據。
- [室內資源規格](docs/spec/039-interior-resources.md)及[室內IDA匯出](tools/ida_interior_evidence.py)保存民房／旅館的原始模組、圖塊排列與來源等級；格式對照不代替正常可達性。
- [室內原圖重建](tools/hd_interior_sources.py)依原版索引表重建OMOYA、INN及正常長老對照，列出已用／未用圖格與動畫候選，不以直接排列圖塊字典當完整原圖。
- [室內圖面註冊](tools/hd_interior_register.py)將內建生成圖直接取樣至3倍，保存五張圖及七個旅館來源的別名關係；固定旅館共用背景，正常入口仍另驗。
- [室內位址索引](docs/interior-source-index.json)分OMOYA與INNAPRO保存原始名稱、位址基準、圖表尺寸及分級語意；IDA匯出自動合併，不改原函式名。
- [小屋與旅館覆繪](internal/prototype/hd/interior.go)以原版程式、資源、色盤與整張原生畫格選圖；小屋正常路徑及GUI限定通過，旅館七個表項共用四張圖，正常到達與GUI仍待驗。
- [室內測試](internal/prototype/hd/interior_test.go)核完整來源、五張圖、來源別名與原子拒絕，依[小屋正常鍵序](tools/hut-hd.commands)從冷新局走到亭台、入小屋及離開；旅館結構fixture不冒稱正常路徑。
- [室內視窗抽測](tools/hd_interior_smoke.py)從正常冷新局進城堡小屋，核石像完整圖面、原貌切回與離開，不把結構旅館圖當正常到達。
- [室內稽核](tools/hd_interior_audit.py)核247圖格、五張圖、正常來源與實際小屋GUI像素，分列旅館結構通過與正常路徑待驗，供覆蓋清單重生引用。
- [MPP3圖面載入](internal/prototype/hd/cave_mpp3.go)按037固定43格、九個材質ID與完整BOSQUE字帶；[MPP3測試](internal/prototype/hd/cave_mpp3_test.go)核結構像素、全43格、界限及三家族共存，正常到達另驗。
- [MPP2圖面載入](internal/prototype/hd/cave_peligro.go)按037固定34格圖面與材質，與冰窟共用整組載入；[MPP2測試](internal/prototype/hd/cave_peligro_test.go)核三份結構快照、尾端九格及來源／載入拒絕，不能當正常到達證據。
- [洞窟家族稽核](tools/hd_cave_family_audit.py)接受MPP2／MPP3／MPP5／MPP6／MPP7，核原始靜態地圖格、完整註冊、合成像素、MPP1正常回歸及實際視窗；MPP7另核全零資源與十一尾端圖樣的RGBA差異。舊收據重生須用當輪來源工作樹與相符driver。新增家族先補037來源及接入契約，不以任意參數跳過守衛。
- [冰窟圖面載入](internal/prototype/hd/cave_ice.go)依037固定SHA與尺寸原子載入，地形依完整來源身份選擇；[冰窟圖面測試](internal/prototype/hd/cave_ice_test.go)區分離線像素、合成結構來源與正常MPP1回歸，不把合成快照當正常冰窟。
- [冰窟圖集收尾稽核](tools/hd_ice_audit.py)核21格資產、三份結構快照像素、MPP1正常35點及實際視窗身份；冰窟正常玩家路徑維持待驗，不升全家族HD完成。
- [洞窟來源視窗回歸](tools/hd_cave_catalog_smoke.py)以新版候選正常冷讀、入洞、HD切回及回城，保存來源鍵、實際命令與圖片。可由ZELIARD_GUI_BINARY、ZELIARD_CAVE_ICE_ART、ZELIARD_CAVE_PELIGRO_ART、ZELIARD_CAVE_MPP3_ART、ZELIARD_CAVE_MPP5_ART、ZELIARD_CAVE_MPP6_ART及ZELIARD_CAVE_MPP7_ART選候選與新增圖集；不外推未到達家族。
- [洞窟辨識收尾稽核](tools/hd_cave_catalog_audit.py)核目前程式、來源測試、正常35點、實際視窗還原與冰窟離線候選的各別身份；不將美術候選或結構fixture計為全家族正常HD完成。
- [全洞窟地形來源](tools/hd_cave_sources.py) 擷取SAR3的MPP1..MPPB原生圖格與分組總覽；[規格037](docs/spec/037-all-cave-terrain-hd.md)分開正常RAM已核與結構候選。輸出留`workplace/full-hd-v1/cave-terrain-sources-v1/`，不把解圖當HD完成。
- [洞窟存檔來源探針](tools/hd_cave_save_probe.go) 在兩個獨立dosgolem冷讀包內四份USR，只用原版按鍵擷取場景與圖像來源；存檔前史未知，不當新局全流程或HD驗收。
- [洞窟地圖解碼](tools/hd_cave_maps.py) 保存31份地圖原始柱狀資料與靜態研究圖；[原版地圖常式對照](tools/hd_cave_map_oracle.go)執行原始6CED..6D6D逐欄核對，與正常可達性分列。
- [洞窟地形分級索引](docs/cave-terrain-source-index.json) 保留固定正常段的原始定位；[IDA9.4來源匯出](tools/ida_cave_resource_evidence.py)自動合併索引，未達已證實的項目加警示。
- [洞窟來源稽核](tools/hd_cave_source_audit.py)核378圖格、31份地圖、原版常式、正常初始化及美術候選身份；參數為`--project`、`--output`，只接受新輸出檔，不把來源通過升格為HD完成。
- [三平面研究工具](tools/hd_planar.py) 與 [原版常式對照](tools/hd_planar_oracle.go) 在Docker內對80筆補零研究buffer核對轉換，不等同80筆圖像格式或HD完成。
- [GRP來源擷取](tools/hd_grp_sources.py) 保存四個正常資源共597圖格及有效遮罩的原版RAM對照；CMAN／DPAT仍為結構候選。
- [城堡來源擷取](tools/hd_castle_sources.py) 與 [規格036](docs/spec/036-castle-hd.md) 保存912×64前景、原版快取變體及限定CONFORMED契約。圖面與完整提示詞在 `workplace/full-hd-v1/castle-art-v2/`、`castle-animation-v1/`，註冊產物在`castle-registration-v1/`；只接候選版，未更新封包。
- [城堡來源索引](docs/castle-source-index.json) 保留原始32EA／32EE／68AC／6950、證據等級及來源；[正常原版探針](tools/hd_castle_probe.go) 與 [IDA9.4匯出](tools/ida_castle_evidence.py) 自動合併此索引，不用散文或推測名稱替代原始定位。
- [城堡合成](internal/prototype/hd/castle.go) 載入2736×192世界圖及四張24×24旗幟，按獨立來源選擇Castle／Muralla；[測試](internal/prototype/hd/castle_test.go) 核八個正常來源、原子退回、主角右邊界及跨場撤除。
- [城堡素材註冊](tools/hd_castle_register.py) 固定原生座標、門口與接縫，直接取樣HD；[正常鍵序](tools/castle-hd.commands) 保存冷新局、步行、離場及返回，禁止以注入狀態代替。
- [城堡視窗抽測](tools/hd_castle_smoke.py) 與 [最終稽核](tools/hd_castle_audit.py) 保存四旗幟、原貌切回、獨立像素及已驗程式身份；輸出在 `workplace/full-hd-v1/`，限036所列範圍。

新增研究文件須同輪掛進本索引或 CONTEXT 的證據入口。不得將暫存、舊待辦或綠色單元測試當作原版一致證據。

本地失敗畫格研究入口：[tools/hd_gui_rejected_probe.go](tools/hd_gui_rejected_probe.go)依保存的GUI按鍵與RunCycles日誌重生被拒來源，核對觀察RAM及候選雜湊。輸出僅供來源研究，不當成HD通過。

[規格047](docs/spec/047-cave-ff41-source.md)管理FF41非零來源的證據與實作閘門；[IDA匯出](tools/ida_cave_ff41.py)保存原始指令與xref，未分級語意維持未知。

[FF41來源研究](internal/prototype/hd/cave_ff41_study_test.go)僅以複製的解碼輸入隔離守衛，驗證原生畫面差異；不改原版機器，不當正常HD驗收。

[FF41正式檢查](internal/prototype/hd/cave_ff41_test.go)使用未修改的真實保存來源，驗證原生整幅、既有HD、唯讀及未知分支拒絕；正常來源與邊界合成分開記錄。

[FF41視窗抽樣](tools/hd_cave_ff41_smoke.py)在Docker／Xvfb正常冷F7、步行與出劍，依完成主體欄位確認FF41=1且FF40=0，再驗原貌／HD切回與正常退出。

[FF41分級索引](docs/cave-ff41-source-index.json)保存兩個繪製consumer及寫入端的原始位址、bytes與證據等級，由IDA匯出自動合併；未列指令維持未知警示。

[FF41收尾稽核](tools/hd_ff41_audit.py)核對被拒／修正後來源、IDA分級、正常配對、GUI原貌切回與來源修訂鏈，只支援047限定範圍。

## GitHub、授權與交付

- 使用者已授權建立私人儲存庫。建立後回讀 `visibility`，確認為 PRIVATE；不得自行轉公開。
- 儲存庫為 `https://github.com/wicanr2/zeliard_cht`，遠端名稱 `origin`，主分支 `main`。
- 使用者另已明確授權使用主機 `/home/anr2/.local/bin/gh`，不重複索取此授權。當執行環境仍拒絕主機存取時，記錄政策阻塞，不搬運憑證或宣稱憑證失效。
- 不提交原版 EXE、資料、音樂、美術、字型、手冊掃描、原始 ZIP／RAR 或含原版資料的收據。私人儲存庫也遵守此邊界。
- 2026-10-20 使用者另指定 README 加入原版／HD 截圖及素材總覽。此例外只涵蓋 `docs/images/` 的四張文件縮圖及來源清單；不擴張至原始遊戲、完整解析度素材、手冊掃描或封包，也不代表取得公開散布權。
- 自有程式與文件採 RRSAL-1.0。建立 LICENSE 時讀 `/home/anr2/.codex/knowledge-base/local/rrsal-1.0.md`，使用技能的 `assets/project-template/LICENSE` 全文，只替換占位符。
- 正式封包只放 `dist-all/<版本>/`。子目錄依 `/home/anr2/.codex/knowledge-base/local/retro-remake-dist-all-output.md`，建立前重讀。
- 版號固定為 `v.<主版>.<次版>.<修訂版>-YYYYMMDD`。定版前讀 `/home/anr2/.codex/knowledge-base/local/retro-remake-release-versioning.md`；不移動或覆寫已發布 tag。
- 發行須驗啟動、檔案、平台架構、存檔位置、必要素材與 SHA-256。未驗平台與未做的手把實測如實列出。
