# 神劍封魔 Zeliard：繁體中文化與 HD

## 每次開工

對使用者的回覆與新文件預設使用繁體中文，句子簡短、直接。`~/.claude` 維持唯讀；外部方法由 Codex 知識路由選擇，不回寫上游。

1. 先讀 [CONTEXT.md](CONTEXT.md)，以目前程式與最近收據核對狀態。
2. 再讀 [WORKLIST.md](WORKLIST.md)，只推進尚未完成且前置條件已具備的工作。
3. 比對 `/home/anr2/.codex/knowledge-base/knowledge-router.md`，說明命中列與實際載入的入口。不可預讀整個知識庫。
4. 遵守使用者提供的共用主機衛生規則。所有分析、解包、建置、測試、遊戲與 GUI 工作都在 Docker 內執行。

## 已授權目標

- 參考 `/home/anr2/cht/softworld_san`，使用 dosgolem 推進《神劍封魔》的繁體中文化與 HD。
- 工具列提供搖桿／手把選項，能選擇並綁定系統已辨識的藍牙與 USB 手把。
- 建立私人 GitHub 儲存庫。README 的故事背景以本地中文手冊為來源。
- 先完成原版可執行、可觀測的垂直切片，再接中文與 HD。不得用猜測填補原版規則。

2026-10-18 使用者已選定直接以 dosgolem 執行原版，另指定參考 `/home/anr2/cht/psychic-war` 的中文與 HD 覆繪方式。後續選定比較圖 B「手繪奇幻」，只決定畫風。不另建玩法重製引擎，不再詢問已定架構或畫風。HD 正式倍率、預設主題、構圖及發行平台仍待確認。

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

| 文件 | 唯一職責 |
|---|---|
| [README.md](README.md) | 故事背景、穩定入口、使用方式與現況摘要 |
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
| [tools/inventory.commands](tools/inventory.commands) | 冷新局交易後開啟、關閉及重開裝備頁的正常鍵序 |
| [tools/ida_dialogue_evidence.py](tools/ida_dialogue_evidence.py) | IDA 9.4 的正常國王初訪段證據，合併對話位址索引及未證實警示 |
| [docs/dialogue-source-index.json](docs/dialogue-source-index.json) | 固定國王段傾印的原始定位、語意等級與證據，不與片頭段混用 |
| [docs/weapon-source-index.json](docs/weapon-source-index.json) | 固定武器店段的商品名稱搬移與 MCGA 清框鏈，保留原始定位及證據等級 |
| [tools/ida_weapon_evidence.py](tools/ida_weapon_evidence.py) | IDA 9.4 武器店來源證據匯出，自動合併分級索引及未證實警示 |
| [docs/source-index.json](docs/source-index.json) | 固定輸入的位址語意、推論等級與證據索引；IDA 匯出自動合併，未證實項附警示 |
| [docs/spec/001-runtime-overlay.md](docs/spec/001-runtime-overlay.md) | 使用者選定架構與前端雛形邊界 |
| [text/host.json](text/host.json) | 前端雛形的可見中文 |
| [text/hud-prototype.tsv](text/hud-prototype.tsv) | 狀態欄與場景名稱53來源的單行試譯；未走過來源須經當次入口驗證 |
| [docs/hud-source-index.json](docs/hud-source-index.json) | 狀態列來源、原版完成字形、安全區與證據等級 |
| [internal/prototype/overlay/hud.go](internal/prototype/overlay/hud.go) | 當次入口唯一驗證來源的狀態列覆繪，不改原版數值或輸入 |
| [internal/prototype/overlay/hud_scene_test.go](internal/prototype/overlay/hud_scene_test.go) | 正常室內名稱配對及候選來源的唯一匹配／失配退回測試 |
| [text/intro-prototype.json](text/intro-prototype.json) | 片頭單句的覆繪試作譯文，不代表完整翻譯 |
| [text/intro-scroll-prototype.tsv](text/intro-scroll-prototype.tsv) | 首章 30 行試譯，鍵為本地 SAR 的檔案偏移；專名尚待手冊校訂 |
| [text/dialogue-prototype.tsv](text/dialogue-prototype.tsv) | 國王、長老、武器店、銀行與道具店的對話試譯；原文與控制碼仍留在本地遊戲資料 |
| [internal/prototype/overlay/dialogue.go](internal/prototype/overlay/dialogue.go) | 國王與長老對話逐字、捲動與待續標記的唯讀覆繪試作 |
| [internal/prototype/overlay/labels.go](internal/prototype/overlay/labels.go) | 核對 SAR1、SAR2、STICK 來源的提示、商品暫存列與銀行標題覆繪；保留輸入值及價格區 |
| [internal/prototype/overlay/occlusion.go](internal/prototype/overlay/occlusion.go) | 觀察固定 MCGA 清框事件，依繪製先後裁切舊中文；整框原圖還原後解除遮擋 |
| [internal/prototype/overlay/occlusion_test.go](internal/prototype/overlay/occlusion_test.go) | 多層遮擋、完整還原、未知變化與歷史淘汰的退回驗證 |
| [internal/prototype/overlay/quotation.go](internal/prototype/overlay/quotation.go) | 已核對來源順序的武器報價覆繪；商品譯名共用來源，價格與折抵保留原版數字圖形 |
| [internal/prototype/overlay/bank.go](internal/prototype/overlay/bank.go) | 銀行問句與後接結果的唯讀覆繪，逐筆保留原版數字圖形 |
| [internal/prototype/overlay/bank_test.go](internal/prototype/overlay/bank_test.go) | 正常存提款的完整原版狀態、問句保留、數字圖形與未知畫面退回驗證 |
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
| [internal/prototype/hd/item_shop.go](internal/prototype/hd/item_shop.go) | 道具店四組已觀察畫格與局部動畫覆繪 |
| [internal/prototype/hd/item_shop_test.go](internal/prototype/hd/item_shop_test.go) | 道具店HD來源、動畫範圍、原子退回及區外不寫入驗證 |
| [tools/item-shop-branches.commands](tools/item-shop-branches.commands) | 三件商品介紹、空賣、首瓶出售與售後物品頁的正常鍵序 |
| [tools/item-shop-use.commands](tools/item-shop-use.commands) | 首瓶購買、城鎮物品頁、正常入洞窟、使用藥水與重開／關閉的鍵序 |
| [internal/prototype/hd/bank.go](internal/prototype/hd/bank.go) | 銀行三個已觀察畫格的 HD 辨識、合成與安全退回 |
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

新增研究文件須同輪掛進本索引或 CONTEXT 的證據入口。不得將暫存、舊待辦或綠色單元測試當作原版一致證據。

## GitHub、授權與交付

- 使用者已授權建立私人儲存庫。建立後回讀 `visibility`，確認為 PRIVATE；不得自行轉公開。
- 儲存庫為 `https://github.com/wicanr2/zeliard_cht`，遠端名稱 `origin`，主分支 `main`。
- 使用者另已明確授權使用主機 `/home/anr2/.local/bin/gh`，不重複索取此授權。當執行環境仍拒絕主機存取時，記錄政策阻塞，不搬運憑證或宣稱憑證失效。
- 不提交原版 EXE、資料、音樂、美術、字型、手冊掃描、原始 ZIP／RAR 或含原版資料的收據。私人儲存庫也遵守此邊界。
- 自有程式與文件採 RRSAL-1.0。建立 LICENSE 時讀 `/home/anr2/.codex/knowledge-base/local/rrsal-1.0.md`，使用技能的 `assets/project-template/LICENSE` 全文，只替換占位符。
- 正式封包只放 `dist-all/<版本>/`。子目錄依 `/home/anr2/.codex/knowledge-base/local/retro-remake-dist-all-output.md`，建立前重讀。
- 版號固定為 `v.<主版>.<次版>.<修訂版>-YYYYMMDD`。定版前讀 `/home/anr2/.codex/knowledge-base/local/retro-remake-release-versioning.md`；不移動或覆寫已發布 tag。
- 發行須驗啟動、檔案、平台架構、存檔位置、必要素材與 SHA-256。未驗平台與未做的手把實測如實列出。
