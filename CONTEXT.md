# 目前狀態

更新日期：2026-10-19，Asia/Taipei。

主入口816634ad，102份Go、測試、譯表與依賴檔固定。[027持盾洞窟HD](docs/spec/027-cave-shield-profile-hd.md)已限定CONFORMED：17個新姿態，累計54個主角來源姿態。53份原生及高清獨立合成、122次完整原版配對與正常視窗通過；未驗特殊組合安全退回。026、025、024、023及其他既有收據保持原身份。全遊戲中文與HD仍在開發。

## 目前狀態表

本輪補上17個持盾姿態，原版172個完成畫格中取53份獨立來源，1709568原生及30528000高清整幅像素零差。26具名測試含正常冷F7完整原版配對通過；正常視窗另有三個同原版切換。原版資料、玩法及存檔不改。

| 項目 | 現況 | 證據與限制 |
|---|---|---|
| 使用者目標 | dosgolem、繁中、HD、工具列手把綁定 | 已授權開工；規則見 [AGENTS.md](AGENTS.md) |
| GitHub | 私人 [wicanr2/zeliard_cht](https://github.com/wicanr2/zeliard_cht) | 主機gh可用，保持PRIVATE；最近提交與遠端核對見WORKLOG |
| 本地 Git | 已初始化 `main`，設定 `origin` | 環境 placeholder 已不存在；使用 GitHub noreply 信箱作本儲存庫提交身份 |
| 遊戲包 | 已安全解包與逐檔盤點 | 30 個壓縮項目，含目錄；29 個檔案，清單在本地 `workplace/source-manifest.json` |
| 中文手冊 | RAR 已列目錄，17 張 JPG | 已人工讀外封面及第 1、5–13 頁；物品操作五張掃描另核RAR大小／CRC。另含地圖等掃描，17 張不等於 17 頁 |
| dosgolem | 正常新局、國王領賞、長老存檔與冷讀檔已驗 | 首場交戰／第一寶石及跳／蹲另已通過原版、中文配對與正常視窗；不代表全遊戲已驗 |
| 架構 | 使用者已選定直接執行原版再覆繪 | 參考 psychic-war，不另建玩法重製引擎；見 [架構規格](docs/spec/001-runtime-overlay.md) |
| 前端 | `workplace/bin/zeliard-preview` 已建置，GUI 雛形可啟動 | 中文工具列、手把綁定設定；音效輸出尚未接入 |
| 中文 | 工具列、首章、魔王、故事正文、製作名單／收尾、國王、長老、存讀、商店、裝備、使用、狀態列及居民交談 prototype | 首章30／名單51／收尾6行與故事58片段、47頁已驗，故事Space跳過另有收據。113居民純文字試譯、一位居民完整兩頁及續頁／還原已驗。其他112居民與40新增場景名未正常走過；專名與字級待校訂 |
| 銀行中文 | 最新35點配對及正常視窗通過 | 查詢0、存100、領50、查詢50、無alma寶石及離店950；標題、問句與數字圖形已驗，ledger存檔持久性及其他分支未驗 |
| 本輪工作 | 027持盾洞窟HD已限定CONFORMED | 主入口816634ad／102來源。17新姿態、53份原生及高清合成、26具名、122完整原版配對、15觀察點、172完成畫格。實際日誌3678筆，三次HD／原貌／HD切換通過；全部退回事件另核原因，不稱整段零退回 |
| 接續工作 | 未驗特殊裝備組合及其他中文／HD | 027保留0093=2配合FF3A的特殊姿態原貌，已保存真實來源。先判斷正常玩家功能依賴，再走窄RE及規格；不重開026前綴、已閉合數字區或無關效果公式。其他物品／法術及其餘居民仍待完成 |
| HD | 手繪奇幻邊框、室內人物、遠山、Muralla前景／地面與主角八姿態；Malicia地形、主角54姿態，其中17個持盾姿態、十武器與29物件圖像已接入 | 地形1..14、主角、武器與物件可選覆繪；已驗配色提示保留原貌。未知來源安全回退；其餘洞窟、phase4、居民與正式倍率／預設待完成 |
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

027已閉合其限定範圍，不重做53份與既有026／025／024驗收。正常離洞出現0093=2且FF3A=255的特殊組合，依規格保留原貌；真實來源及來源／主體守衛退回已保存，可作下一窄任務入口。其餘居民、場景名稱、混合控制腳本、物品／魔法與洞窟美術仍待正常路徑抽樣；正式字級、術語、HD倍率／預設、音訊及實體USB／藍牙手把驗收仍待完成。

前輪武器商人HD已接入可撤回前端，三姿態、區外不寫入、原版畫格不改動及缺圖／未知畫格／色盤／短畫格退回測試通過。正常冷讀檔到商店、未購物離店1000、重進店購買後離店800與三姿態HD切換均通過，該輪收據使用固定48e9c0bc建置。三張1536×1024圖與完整內建imagegen提示詞在 `workplace/merchant-hd-v1/`，只有本地研究用途。該輪測試與建置紀錄為同目錄 `hd-test-v2.jsonl`、`no-purchase-test-v2.jsonl`及 `build-v2.log`。

本輪商人來源、色號／RGB、座標空間及三張圖面雜湊的入口為 `workplace/merchant-hd-v1/source-mapping-v2.json`；固定建置與Go來源身份在同目錄 `candidate-identity-v2.json`。正常路徑收據為 `workplace/merchant-path-v1/receipt.json`，未購物告別65 bytes的來源映射在同目錄 `no-purchase/source-evidence.json`；對話TSV共19筆。配對3點完整原版狀態相同，離店覆繪差0；GUI入口為 `workplace/merchant-hd-gui-v2/receipt.json`、`visual-review.json`與 `region-comparison.json`。首輪失敗保留原身份，不換名為成功。

獨立逐像素核對為 `workplace/merchant-hd-v1/pixel-verification-v2.json`，確認三個原貌肖像與原版RGB來源相同、HD與指定圖面取樣相同、遊戲畫布內框外0差、原貌完整還原。三組暫停steps／cycles不變；最終HD開啟但商人不匹配、持鍵空、capture-tail退出0。84筆按放均確認、最長0.19秒。GUI保留兩筆舊對話安全退回，paragraph14／first_difference290及paragraph12／first_difference422，不稱零失配。第二訪問的報價也見交叉雙臂，不依流程名稱硬綁姿態，不推測原因。三姿態不等於完整商人動畫。

## 工具與本地證據入口

- [027持盾洞窟HD](docs/spec/027-cave-shield-profile-hd.md)原版入口為`workplace/prose-v1/cave_shield_probe_v1.go`、`cave_shield_actions_v1.commands`及`cave-shield-source-v1/`，172個完成畫格／53份候選。`cave_shield_inventory_v1.py`／`cave-shield-inventory-v1/receipt.json`獨立重建1709568原生像素零差。IDA9.4沿既有py312映像，`cave-shield-ida-v1/query.json`保存原始強推論，`cave-shield-ida-v2/query.json`自動合併[已分級索引](docs/cave-shield-source-index.json)；原始`sub_39B9`及362條指令不改名。READY先存`cave-shield-ready-v1.json`；素材及完整提示詞見`cave-shield-art-v3/registration.json`、`generation.json`，原圖在`workplace/cave-hd-v1/player-shield-atlas-handpaint-v2.png`。`cave_shield_composition_v2.py`／`cave-shield-composition-v2/`核53整圖30528000像素；v1路徑錯誤保留。`cave-shield-unit-v1/`記26具名、122完整原版配對與vet／build，主入口816634ad。102來源身份與快照為`cave-shield-candidate-identity-v1.json`、`cave-shield-candidate-sources-v1/`。正常視窗`cave_shield_gui_v2.py`／`cave-shield-gui-v2/`，v1累計退回檢查過嚴的失敗保留；`cave-shield-gui-preflight-v1.json`記分類修正。實際日誌及全部退回核對為`cave-shield-gui-pair-v2/`，3678筆；v1重播的候選重複引用勘誤保留。實際像素入口`cave_shield_gui_pixels_v1.py`／`cave-shield-gui-pixels-v1/`，三個時刻HD與原貌各870912像素、重複整畫布1728000與中文HUD205632像素零差。來源／擁有權／連結稽核、舊主入口備份及同步見`cave-shield-unit-v1/current-audit.json`、`mainalias-sync.json`與`previous-main-ae87b6b7`。原始資料與完整收據只留本機，不作Release。

- [026洞窟名稱前綴](docs/spec/026-cave-hud-source-prefix.md)原版入口為`workplace/prose-v1/hud_prefix_probe_v1.go`／`hud-prefix-source-v1/`，各擷取的輸入、工具及位址身份在同目錄`identity.json`。分析v2核兩次21位元組讀取、前綴0次讀取及完整字形；v1路徑metadata錯誤保留。READY先存`hud-prefix-ready-review-v1.json`；`hud_prefix_navigation_probe_v1.go`／`hud-prefix-navigation-v1/`正常對齊後，v2審查補原始0083診斷。`hud-prefix-unit-v1/`未對齊失敗保留，現行`hud-prefix-unit-v2/`驗13具名、23完整配對點、十整圖5760000獨立像素及vet／build。身份`hud-prefix-candidate-identity-v1.json`／`hud-prefix-candidate-sources-v1/`固定101來源，GUI為`hud_prefix_gui_v1.py`／`hud-prefix-gui-v1/`；根端目視五圖。`hud_prefix_gui_replay_v1.go`／`hud-prefix-gui-pair-v1/`核3654筆及五個真實暫停快照，`hud_prefix_actual_pixels_v1.py`／`hud-prefix-unit-v2/actual-pixels.json`核342720完整HUD與279720原數字／區外像素零差。稽核為同目錄`current-audit.json`、主入口同步／備份為`mainalias-sync.json`／`previous-main-439baeea`，Git／Docker收尾為`repository-sync.json`。此項已閉合；HD裝備分支另追。

- 持有物品觀測入口為`workplace/prose-v1/held_item_probe_v1.go`，二進位在`source-item-tools-v1/held-item-probe`，正常F7、原生鍵序與擷取在`held-item-v1/`。來源為原包SAVEM1.USR，SHA沿來源清單；原始檔唯讀，scratch副本保留mtime。刀劍油選取、使用、槽位消耗及重開已觀察，不推定取得過程或效果公式。該次4209f976觀測保持原身份，現行驗收依下列025入口。

- [025刀劍第二行](docs/spec/025-sword-inventory-label.md)原版來源為`workplace/prose-v1/sword_probe_v1.go`與`sword-source-v2/`，九個回呼與37像素差分確認字母區不變，READY先存`sword-ready-review-v1.json`。`sword-unit-v1/`通過11具名、23完整配對點、五整圖2880000獨立像素及vet／build。主入口439baeea，100來源身份為`sword-candidate-identity-v2.json`，快照`sword-candidate-sources-v1/`；v1漏列兩個Go依賴檔的紀錄保留。正常GUI為`sword_gui_v2.py`／`sword-gui-v2/`，四面板1161216像素通過並根端目視；v1的Space按鍵工具錯誤保留於`sword-gui-preflight-v1.json`。`sword_gui_replay_v1.go`／`sword-gui-pair-v2/`核3130筆實際日誌與最終完整原版，`sword_actual_pixels_v1.py`／`sword-unit-v1/actual-pixels.json`核同一暫停的2448原數字像素。來源／擁有權／連結稽核為`sword-unit-v1/current-audit.json`，主入口備份及同步在同目錄`mainalias-sync.json`與`previous-main-4209f976`。洞窟HUD另從`sword-hud-preflight-v1.json`接續，不重開已閉合刀劍／盾牌字區或猜效果公式。

- 024現行入口為`workplace/prose-v1/shield-unit-v1/candidate-identity.json`、`current-audit.json`及`actual-pixels.json`；99來源快照在`shield-candidate-sources-v1/`。原版差分為`shield_probe_v1.go`與`shield-source-v1/`，實作前READY為`source-saves-build-v1/shield-ready-review-v2.json`，原v1的數字語意稱呼已追加勘誤。`shield-unit-v1/tests.jsonl`、`receipt.json`、`independent-pixels.json`與`shield-regression-v1/`驗10具名、八點新配對、49點既有流程及2304000獨立像素。GUI為`shield-tools-v1/`、`shield-gui-v1/`，實際日誌重生為`shield_gui_replay_v1.go`及`shield-gui-pair-v1/`，共4291筆；四實際面板1161216及同時刻數字4896像素零差。舊fab主入口備份在`shield-unit-v1/previous-main-fab8f646`，Git／Docker收尾為`shield-unit-v1/repository-sync.json`。只驗已列來源存檔的`004dad`顯示，不代表其他盾牌或從新局取得。

- 後段中文的定位入口為`workplace/prose-v1/source_save_inventory_v1.go`，建置及輸出為`source-saves-build-v1/`、`source-saves-v1/`。五份來源USR的SHA沿`workplace/source-manifest.json`，唯讀掛載來源、另複製到各自scratch並保存mtime。每份冷啟動／F7最多140M指令，僅觀察所在地與檔案讀取；不使用機器快照還原、記憶體注入或推測來源存檔的取得經過。原五混合腳本的靜態證據沿`mixed-control-notes.json`，尚未據此更改正式中文顯示。

- 023現行入口為`workplace/cave-hd-v1/expedition-unit-v1/candidate-identity.json`及同目錄的`tests.jsonl`、`pixels-receipt.json`、`pair-receipt.json`；98份固定來源在`expedition-candidate-sources-v1/`。實作前READY為`expedition-build-v2/ready-review.json`，原版與獨立盤點為`expedition-source-v2/`、`expedition-inventory-v2/`；七圖固定登錄與18份獨立合成為`expedition-art-preview-v2/`及`expedition-composition-preview-v1/`。正常鍵序為`tools/cave-expedition.commands`。實際GUI工具／輸出為`expedition-tools-v1/`及`expedition-gui-v1/`，3494筆日誌與獨立像素為`expedition-gui-pair-v2/`、`expedition-gui-pixels-v2/`。45具名、18點完整原版配對與攀繩視窗通過，主入口fab8f646；前d72備份在`expedition-unit-v1/previous-main-d72e260a`。最後輸入／來源／擁有權稽核為`expedition-unit-v1/current-audit.json`，Git／Docker收尾另存`repository-sync.json`。不重開021／022已閉合工作。

- 接續探索沿`workplace/cave-hd-v1/expedition-build-v2/`與`expedition-source-v2/`，重新建置既有`exploration_probe_v1.go`，冷啟動鍵序沿`exploration-startup-v1.commands`。讀原版、固定882來源與原存檔皆唯讀；僅輸出／存檔副本可寫。原版進程有800M指令、512命令及96擷取上限，不注入、不還原或挑選亂數。目的是推進正常路徑，不重開021／022或首場首寶石已閉合驗證。

- 送返後城鎮HD依[022](docs/spec/022-return-town-hd.md)先DRAFT。`workplace/cave-hd-v1/return_town_probe_v1.go`／`return-town-build-v1/`／`return-town-read-v1/`重播既有104命令，逐點核17原件，再監看原版靜止及左右步行的11B1:8000..D800原生讀取。監看期間不由觀察器讀取該區；使用既有研究根，尚未READY不改正式程式。
- `return_town_footprint_v1.py`／`return-town-footprint-v1.json`核完整215×8地圖、居民底圖、原生前景／人物／主角重建與原版讀取日誌。探針v1只建置未執行，啟動前檢查將觀察結束與放鍵後擷取時間分欄，實際執行為probe-v2；來源與binary SHA另記足跡收據。未把中途畫格當完整驗收。
- 022 READY原文／審查為`return-town-ready-v1.md/json`，完成後才改前景與遠山的Muralla圖塊上界。`return_town_validate_v1.py`／unit-v1因GUI未另存palette檔停止；來源／log保留。v2改讀已核同一完整CPU RAM／色號的原版重播palette，產品不改。`return-town-unit-v2/`34具名、vet及候選d72e260a通過。獨立`return_town_pixels_v1.py`僅地面因舊camera公式失敗，診斷`return-town-pixel-diagnostic-v1.json`保存；v2沿009既有原圖唯一循環匹配，同五來源2880000整幅像素零差，不改正式地面或重開其研究。
- `return_town_freeze_v1.py`／`return-town-candidate-identity-v1.json`／`return-town-candidate-sources-v1/`固定97來源、247輸入；`return-town-tools-v1/gui_return_town.py`／`return-town-gui-v1/`實際冷F7及送返後恢復HD執行中。期間不改程式／測試／譯表與固定工具。後續日誌為`return_town_replay_v1.py`／`return-town-gui-pair-v1/`，實際像素為`return_town_gui_pixels_v1.py`／`return-town-gui-pixels-v1/`，未完成者不預先計通過。
- GUI v1四個HD層與完成畫格均已匹配，但工具誤要求舊Sky旗標也true而停止。`return-town-gui-preflight-v1.json`保存失敗及主前端程式核對勘誤：Sky先嘗試繪製，完整Background後繪並涵蓋天空，沒有「Background成功就跳過Sky」分支。`return_town_gui_prepare_v2.py`／`return-town-gui-v2/`只改驗四個正式圖層，產品97來源不變；日誌、像素及稽核工具明示讀v2，不把v1失敗改名成功。筆記追加工具為`return_town_notes_v1.py`，後續稽核／同步為`return_town_audit_v1.py`、`return_town_promote_v1.py`及return-town-current-audit／mainalias-sync前綴。

以下按研究版本保留證據入口。舊候選與舊失敗保持原身份；目前主入口與下一步以上方狀態表為準。

- 022現行：`return-town-ready-v1.md/json`保存實作前審查；`return-town-candidate-identity-v1.json`及97來源固定；`return-town-unit-v2/`34具名與候選d72e260a，`return-town-pixels-v2/`五整圖2880000像素。真正視窗`return-town-gui-v2/`，`return-town-gui-pair-v1/`原生日誌3775筆／850 FileOps，每筆原版及末端GUI相同；`return-town-gui-pixels-v1/`城鎮810432及中文／肖像2303640像素零差，`return-town-root-visual-review-v1.json`七張X圖目視。`return-town-current-audit-v1.json`核153自有檔、29原件、882來源與字型／存檔SHA／mtime；主入口同步／舊3ae備份為`return-town-mainalias-sync-v1.json`／`return-town-previous-main-3ae68018`。Git及容器收尾`return-town-repository-sync-v1.json`，文件更新工具`return_town_finalize_docs_v1.py`。以下DRAFT／READY／執行中敘述只保留歷史，不重開已閉合範圍。

- 021現行入口：`exhaustion-candidate-identity-v1.json`／96來源；正常`exploration-return-pair-v3/`；真正視窗`exhaustion-gui-v6/`；`exhaustion_replay_v2.py`／`exhaustion-gui-pair-v2/`；`exhaustion_gui_pixels_v2.py`／`exhaustion-gui-pixels-v2/`及`exhaustion-root-visual-review-v1.json`。3849筆每筆原版相同、850 FileOps、六時刻2764368像素零差。清框保留一筆kind2／first_difference430，不稱零拒絕。`exhaustion_audit_v1.py`／`exhaustion-current-audit-v1.json`核151自有檔、29原件、882來源、壓縮包、字型與存檔SHA／mtime。主入口同步／舊0d備份為`exhaustion-mainalias-sync-v1.json`／`exhaustion-previous-main-0d70040a`，Git及容器收尾為`exhaustion-repository-sync-v1.json`；文件更新工具為`exhaustion_finalize_docs_v1.py`。既有長老PNG及完整提示詞仍沿workplace/hd-sage-generation.json，未生成新圖。以下v1至v5只保留歷史證據，不代表現行未完成。

- 正常洞窟探索沿既有cave-hd-v1下的`exploration_probe_v1.go`、`exploration-startup-v1.commands`、`exploration-build-v1/`及`exploration-source-v1/`。單一原版冷啟動、自建存檔與持續正常按鍵，逐點保存當次RAM／VGA及完成來源，不注入或還原。先限定800M指令、96擷取點及512命令，錯誤與輸入留events／receipt。預覽只用現有覆繪，未配對及GUI前不當新功能驗收。
- 下一段探索的手冊掃描沿`exploration_manual_v1.py`／`exploration-manual-v1/`，只從已固定RAR抽取SCAN0443五張，逐檔核大小、CRC32、SHA及擁有權。`exploration_manual_review_v1.py`／同目錄review.json記根端逐張核讀：001為洞窟摺頁圖，002為音樂帶，003至005為磁片標籤。摺頁只作導覽，不推論正常入口或版本一致性；不把標籤當新增手冊頁，不修改原始包。
- 探索自然走到昏迷後送返Marid分支，沿[021](docs/spec/021-exhaustion-return-overlay.md) DRAFT。文字源自SAR2檔案01D72E等字段、模組01BCC6；先不替原版配色欄位命名。置中肖像完整6144色號與既有姿態0相同，第二姿態及事件仍待核。重生工具、鍵序及原版事件沿同目錄`exploration-return-*`，保持原探索身份。
- 021後續事件與字段已核，READY審查前由`exploration_return_preview_v1.py`／`exploration-return-preview-v1/`產生九獨立整圖，譯文草稿為`exploration-return-translation-v1.tsv`。原探索與加只讀事件的冷重播17點全RAM／VGA／色盤相同；337事件含157字元、一次待續及10個捲動完成事件。四字段保留前文，兩個置中肖像都與既有完整6144來源相同。規格的READY正文及審查另保存exploration-return-ready前綴，不覆寫原探索。
- 正式送返配對沿`exploration-return-pair-v1/`及v2保留失敗，編譯v1只因測試索引未使用；v2已核17原版點及九整圖，但最後誤要求無離場撤除紀錄。原版回鎮造成舊字框快照失配，現有守衛安全reset；測試改為只接受已確認15-return-after-4的唯一kind2／first_difference0及完整最後段，其他拒絕仍失敗。未改正式守衛，來源與失敗全文存`exploration-return-pair-preflight-v1.json`；新結果另用v3。長老兩位置／兩姿態、四動畫像素及失配測試在`exploration-return-unit-v1/sage.jsonl`已通過。
- 021候選3ae68018與96份Go／測試／譯表凍結於`exhaustion-candidate-identity-v1.json`及`exhaustion-candidate-sources-v1/`。`exhaustion_freeze_v1.py`固定245輸入；GUI v1視窗啟動超過15秒，程序未自行結束，保留原失敗及`exhaustion-gui-preflight-v1.json`。`exhaustion_gui_prepare_v2.py`只把等待改60秒，候選與來源不變，`exhaustion-gui-v2/`已啟動。只讀原生日誌配對入口為`exhaustion_replay_v1.py`／`exhaustion-gui-pair-v1/`；獨立像素為`exhaustion_gui_pixels_v1.py`／`exhaustion-gui-pixels-v1/`；收尾沿`exhaustion_audit_v1.py`／`exhaustion-current-audit-v1.json`。未完成者不預先計通過。
- GUI v2五個對話／切換時刻已保存，最後中文complete早於原文末字，Space過早而停在長老，超出40M回鎮預算。失敗身份見`exhaustion-gui-preflight-v2.json`；未改產品或重擲亂數。`exhaustion_gui_prepare_v3.py`／`exhaustion-tools-v3/`只在最後等待157原字元及2M末尾處理後才續行；`exhaustion-gui-v3/`重跑中，後續日誌及像素工具明示使用v3，不把v2當成功。
- GUI v3已正常回Muralla並撤除新覆繪，但工具仍要求TownHD used而停止；送返後來源目前安全保留城鎮原貌，不能算全程TownHD。`exhaustion-gui-preflight-v3.json`保留失敗與工具準備的整圖恆等假設勘誤，兩原版圖只在角色矩形差27像素，模組前64bytes、色盤及HUD場景字形相同。`exhaustion_gui_prepare_v4.py`／`exhaustion-gui-v4/`改核原版模組／色盤／撤除及2M繪圖完成，實際圖與日誌仍另驗。候選96來源不改，工具只讀原版，不把拒絕改成HD成功。下一步來源差異盤點沿`exhaustion_town_source_v1.py`／`exhaustion-town-source-v1.json`，未有讀寫端證據與READY前不放寬TownHD。
- GUI v4退出0，但根端目視returned-town仍是轉場黑畫面，因此不算回鎮完成，記`exhaustion-gui-preflight-v4.json`。`exhaustion_replay_v1.py`／`exhaustion-gui-pair-v1/`只保留v4原生日誌，不升格為完成路徑。`exhaustion_gui_prepare_v5.py`／`exhaustion-gui-v5/`等先前兩個已核完整城鎮原版畫面之一，再有界對齊暫停擷取；候選96來源不變。最後正常日誌改用`exhaustion_replay_v2.py`／`exhaustion-gui-pair-v2/`，獨立像素v1明示讀v5及pair-v2，不借黑畫面當完成。
- v4的3744筆日誌冷重播及五個對話時刻2303640實際像素已通過，僅作對話證據；工具為`exhaustion_gui_active_pixels_v1.py`／`exhaustion-gui-active-pixels-v1/`，原獨立像素v1保留作其固定基底。v5最後一段暫停在四動畫像素更新中間，距兩完整肖像分別3及1像素，守衛正確退回；`exhaustion-gui-preflight-v5.json`保存完整差異。`exhaustion_gui_prepare_v6.py`／`exhaustion-gui-v6/`把所有對話擷取也改有界完整來源對齊，產品96來源不變。最終重播v2明示讀v6，完整像素改用`exhaustion_gui_pixels_v2.py`／`exhaustion-gui-pixels-v2/`，不覆寫已驗基底或將v5失敗改名成功。

- 入口HD依[020](docs/spec/020-cave-portal-hd.md)限定CONFORMED；原生矩陣、資源、裁切與完整來源核對沿`portal_inventory_v1.py`／`portal-source-v1/`，沿既有已驗完成來源，不注入或重打。美術參考為同目錄`reference.png`；圖面、完整提示詞與登錄另沿portal前綴，不覆寫既有地形／物件／主角。
- 入口盤點初次在舊native-paints來源的獨立遮罩檔假設停止，該證據實際把真實CavePaints存於既有JSON。失敗全文與SHA記`portal-inventory-preflight-v1.json`，改讀明示的inline紀錄後對同批來源重跑；輸出原先仍空，原件與正式程式未改。
- 入口盤點去重38真實完成來源，16完整／3部分裁切／19不在視野，26657個可見未遮擋入口像素與完整遊戲區零差，無未匹配字牌片段。21個圖塊按id+48bytes串接共1029bytes，SHA `3fd046b8c1281a013bfe1c8dfa627a663fec2eac5cebe7e3df025a63c1db6633`；40×40原生圖SHA `9374206133321e2cb058c7c94512e230cde8183afc6896282dd94989180fb716`。根端已核讀MURALLA。手繪圖與完整提示詞另存`portal-handpaint-v1.png`／`portal-prompt-v1.txt`及`portal-art-index-v1.json`，透明字牌留白、文字另畫；未READY不接正式路徑。
- 入口美術裁切及直接120×120取樣沿`portal_register_v1.py`／`portal-art-preview-v1/`，保留原PNG與alpha。中文／原文圖牌的字型、字寬與安全區在同份登錄核對；不增新catalog格式，中文沿已驗HUD譯表，未提供該來源時用MURALLA。全幅合成另沿`portal_preview_v1.py`／`portal-composition-preview-v1/`，只作DRAFT研究，不算GUI。
- 020的READY原文與審查為`portal-ready-spec-v1.md`／`portal-ready-review-v1.json`。57份獨立完整圖面及中英圖牌核對後才實作。候選0d70040a與93份Go／測試／譯表來源凍結於`portal-candidate-identity-v1.json`／`portal-candidate-sources-v1/`；實作後回歸為`portal_validate_v1.py`／`portal-unit-v2/`，232項具名無跳過，HUD、vet及建置另通過。正常38點在`portal-pair-v1/paths/`；不以單元圖面當GUI。
- 入口GUI沿`portal_freeze_v1.py`／`portal-tools-v1/`，固定237輸入；正常首戰、切換／捲動與回城分別輸出`portal-gui-combat-v1/`、`portal-gui-return-v1/`。`portal_gui_targets_v1.py`保存真實完成時鐘，原版重播輸出`portal-gui-pair-v1/`，獨立完整合成由`portal_gui_pixels_v1.py`／`portal-gui-pixels-v1/`核對。工具輸出尚未完成者不預先算通過。
- 入口GUI v1已在`combat-right-2.png`及`portal-scroll-HD-b.png`看見第一寶石1，尾段四步卻是2。舊腳本只接受尾段恰等1，因此停止，不是遊戲缺陷。v1失敗及原圖保留，勘誤為`portal-gui-preflight-v1.json`；`portal_gui_prepare_v2.py`／`portal-tools-v2/`改驗所有正常檢查點曾出現1。程式0d70040a、93來源及原鍵序預算不改，不固定／重擲亂數。正常首戰另輸出`portal-gui-combat-v2/`，回城沿v1。後續時鐘及像素工具明示使用v2首戰，不將v1改名為成功。
- 020首戰v2與回城v1已正常退出0。`portal_replay_v1.py`／`portal-gui-pair-v1/`核3241實際日誌、268完成／677合成／759 FileOps，每筆原版及末端GUI RAM／VGA相同；336物件已知、206部分出界、104空白、未知0。僅既有c29c7a73主角配色保留原貌，不算HD或新語意。指定來源129643829／135129901／139943461／153125859的兩個真實時間與完整SHA相同。
- 獨立像素工具v1因全域sprite被後續人物／武器取代而停止，完整失敗與一張未完成輸出保留`portal-pixel-preflight-v1.json`。`portal_gui_pixels_v2.py`只隔離變數，用同一GUI與重播來源核四洞窟加回城1193472高清像素零差；不重跑遊戲或改守衛。五張真正X截圖目視為`portal-root-visual-review-v1.json`。
- 完成稽核沿`portal_audit_v1.py`／`portal-current-audit-v1.json`，核147自有檔、93固定來源、29原件、882快照、兩壓縮檔、字型及存檔SHA／mtime。額外HUD篩選9具名通過、3舊fixture跳過不計驗收，初次稽核誤當無skip的勘誤為`portal-audit-preflight-v1.json`。主入口同步與舊2ac1備份為`portal-mainalias-sync-v1.json`／`portal-previous-main-2ac1b51c`。完整提示詞、原PNG與登錄收`portal-art-final-index-v1.json`，嵌入建置資訊為`portal-build-info-v1.json`；Git及容器收尾另記`portal-repository-sync-v1.json`。沒有建立Release、tag或交付包。
- 登錄首輪的「開口alpha必須全0」檢查停止。實際864個中心樣本中830為0、34為1，非零者皆RGB0；原版先清黑後合成，輸出RGB逐點仍0，人物與物件後畫不受影響。原圖alpha不改，改驗既有清黑契約下的可見輸出，失敗全文及數值記`portal-art-review-v1.json`，不放寬來源／色盤／區界或生成另一張圖挑結果。

- 洞窟動態物件沿[019](docs/spec/019-cave-object-hd.md)先DRAFT，使用既有正常首戰／左右鍵序；探針`object_probe_v1.go`／`object-source-v1/`保存原始物件定位、16×16色號／透明來源及完整原版快照。獨立原生核對與來源圖表沿`object_inventory_v1.py`／`object-inventory-v1/`。不改主角、武器、共享dosgolem或原件，不以視覺來源推論敵人規則。
- 019正常229完成來源取得352筆物件觀察、28個完整16×16鍵，分原始type1／2／10／18／20，不替欄位命名玩法。28份全遊戲區及256bytes各由獨立Python零差；所有已見四格底圖皆0，仍保留原黑色與出界裁切。精確平移／鏡射等價與美術參考分組沿`object_art_reference_v1.py`／`object-art-reference-v1/`，圖面及完整提示詞沿`object-art-index-v1.json`與`object-*-prompt-v1.txt`、`object-*-handpaint-v1.png`。未READY不接正式路徑。
- 三組美術保留原PNG及alpha，固定登錄／直接48×48取樣沿`object_register_v1.py`／`object-art-preview-v1/registration.json`及目標圖表。不先降成16×16，不把原生黑色0轉透明；美術非零包絡與原版清黑矩形分開處理，出界只裁遊戲區。完整來源合成另驗，不以目標圖表當GUI。
- 動態物件獨立解碼與完整高清圖面沿`object_decode_v1.py`、`object_preview_v1.py`／`object-composition-preview-v1/`。固定圖組只接受已證實四底圖0來源，按原版物件順序先重畫清黑或未知原圖，再合成HD，最後主角與武器。非零底圖、未知鍵仍原貌；未完成完整圖面審查前019保持DRAFT。
- 35份獨立合成的第三個既有後畫來源另有一個綠色動作缺圖，v1如實退回。補圖來源與完整原生核對沿`object-extra-v1/`，追加PNG與完整提示詞為`object-green-extra-handpaint-v1.png`／`object-green-extra-prompt-v1.txt`；不覆寫原三圖。29登錄及同35來源全幅另用`object_register_v2.py`／`object-art-preview-v2/`及`object_preview_v2.py`／`object-composition-preview-v2/`，舊v1保持身份。
- 019的29來源／四圖與35完整圖已READY，全文及審查記`object-ready-spec-v1.md`／`object-ready-review-v1.json`。正式實作、單元與候選另沿`object-unit-v1/`、`object-preview-v1`及`object-build-v1.log`／`object-vet-v1.log`；正常配對／視窗另驗，不提前覆寫d8666d0a主入口。
- 019正常鍵序原版配對沿`object-pair-v1/paths/`／`object-pair-test-v1.jsonl`。完整候選與全部Go／測試／譯表來源凍結沿`object_freeze_v1.py`／`object-candidate-identity-v1.json`／`object-candidate-sources-v1/`；正常GUI工具與輸出沿`object_gui_prepare_v1.py`／`object-tools-v1/`、`object-gui-combat-v1/`及`object-gui-return-v1/`。GUI期間停止修改所有凍結來源與工具，待收尾檢查完成後才解除。
- GUI工具首次準備在return腳本錨點檢查停止，尚未產生輸出目錄或啟動遊戲；只修本地生成工具，準備失敗身份記`object-gui-prepare-preflight-v1.json`。不得把未啟動當產品失敗，亦不刪改來源或舊工具。
- 實際GUI來源目標、原生日誌配對及獨立像素沿`object_gui_targets_v1.py`／`object-gui-source-targets-v1.json`、`object-gui-pair-v1/`／`object-gui-pair-test-v1.jsonl`及`object_gui_pixels_v1.py`／`object-gui-pixels-v1/`。收尾以`object_audit_v1.py`／`object-current-audit-v1.json`、`object-art-final-index-v1.json`、`object-root-visual-review-v1.json`、`object-mainalias-sync-v1.json`及`object-repository-sync-v1.json`保存本輪身份，自有清單為`object-owned-files-v1.list`。不能借用前輪圖或原版日誌當本輪GUI成功。
- GUI v1在combat-right-2的「至少一完整可見物件」取樣斷言停止；實際known2／clipped2／unknown0，不能算產品來源故障。此前combat-right-1已記known2／clipped1，可作正常取樣點。v1收據、圖與原生日誌保留；v2只把取樣切換移到已驗right-1，原版按鍵與步數預算、2ac1b51c及91來源不變。工具／輸出沿`object_gui_prepare_v2.py`／`object-tools-v2/`、`object-gui-combat-v2/`、`object-gui-return-v2/`；目標與像素另用`object_gui_targets_v2.py`／`object-gui-source-targets-v2.json`、`object-gui-pair-v2/`／`object-gui-pair-test-v2.jsonl`及`object_gui_pixels_v2.py`／`object-gui-pixels-v2/`。
- v2兩GUI與612864實際像素通過；3240筆日誌重播243完成／613合成、309物件匹配、零物件未知。重播另列既有2b975f19配色保留及新c29c7a73主角鍵，不能直接當「無主角缺圖」。分類沿`object_player_variants_v1.py`／`object-player-variants-v1/`，先核完整576幾何與雙向色號映射，再決定保留原色或補動作；不重打挑結果。
- c29c7a73與既有主角pose1完整576透明幾何相同，雙向色號映射唯一，FF36=255；與2b975f19同屬已證實配色差異，保留原色，不猜受傷／無敵。兩個來源的正式繪製器視覺補驗沿`object_variants_pixels_v1.py`、`object_variant_pixels_v1.go`／`object-player-variant-pixels-v1/`，只驗圖面；未保存cycles的舊缺圖傾印用明示單元時序邊界，不稱新原版時鐘收據。完整原版日誌的3240筆時序證據仍由原重播獨立提供，程式91來源不改。
- 019現行限定成果的收尾由`object-current-audit-v1.json`及`object-art-final-index-v1.json`保存，四個PNG的最終路徑、完整四份提示詞、參考SHA與原始工具產物皆可回查。主入口同步記`object-mainalias-sync-v1.json`，舊d8666d0a保存在`object-previous-main-d8666d0a`；推送與PRIVATE兩端HEAD另記`object-repository-sync-v1.json`。不將前輪或失敗GUI改名作本輪成功。

- 洞窟武器HD依[018](docs/spec/018-cave-weapon-hd.md)先DRAFT盤點，沿既有`workplace/cave-hd-v1/`追加`weapon_probe_v1.go`、`weapon-normal-v1.commands`及`weapon-source-v1/`。來源表與原生獨立核對沿`weapon_inventory_v1.py`／`weapon-inventory-v1/`，原件唯讀，不注入或重新挑亂數。主角25姿態及017的既有收據保持身份。
- 正常28點、229個有限完成來源中48份武器可見，去重10個32×32來源；逐個獨立重建完整遊戲區與1024bytes皆零差。十格手繪圖沿`weapon-atlas-prompt-v1.txt`、`weapon-atlas-handpaint-v1.png`及`weapon-art-index-v1.json`，原始刀刃／殘影輪廓不加劍柄、光暈或新攻擊範圍。固定登錄與全幅離線合成沿`weapon_register_v1.py`／`weapon-art-preview-v1/`及`weapon_preview_v1.py`／`weapon-composition-preview-v1/`，未READY不接前端。
- 左向出劍帶出三個新可見主角鍵，不重開原25姿態結論；新增來源等價核對沿`weapon_player_coverage_v1.py`／`weapon-player-coverage-v1/`，若非整個576bytes的平移／鏡射等價則另補美術，不借用相似姿態。原武器離線24圖保留原身份，主角退回不得稱全段HD。
- 三個新左向主角均無精確平移／鏡射等價，另補圖。參考圖沿`weapon_player_reference_v1.py`／`weapon-player-coverage-v1/source-sheet.png`；新三格圖集與完整提示詞為`player-left-atlas-handpaint-v1.png`／`player-left-atlas-prompt-v1.txt`，固定登錄沿`player_left_register_v1.py`／`player-left-art-preview-v1/`。017新三姿態先另審查READY，原25素材及像素不改。
- 018十武器來源／圖面／獨立全幅已審查READY，全文及審查另存`weapon-ready-spec-v1.md`／`weapon-ready-review-v1.json`。主角新三格仍另審，不先混入正式路徑。後續武器實作、測試與可選前端入口依018契約驗收，不將原單一探針當完整原版配對。
- 左向主角初圖v1的第一格有多餘手臂，未採用；局部修正提示詞沿`player-left-atlas-edit-prompt-v2.txt`，新產物為`player-left-atlas-handpaint-v2.png`。先核兩臂與原姿態，再登錄，原v1保持身份。
- 三格修正圖的登錄為`player-left-art-preview-v1/registration.json`，原25登錄及像素沿原件不變。全幅增量沿`weapon_preview_v2.py`／`weapon-composition-preview-v2/`，包含武器HD、主角HD開／關及新三姿態配原版武器；未正常通過不升CONFORMED。
- 017左向三姿態READY全文另存`weapon-player-ready-spec-v1.md`，27份獨立圖已無新增主角缺圖，原25登錄不變。正式武器／主角／地形單元與候選沿`weapon-unit-v1/`、`weapon-preview-v1`及`weapon-build-v1.log`／`weapon-vet-v1.log`，正常配對與GUI另留獨立身份。
- 單元v1主角與武器誤用同一輸出pixels目錄，武器全幅測試尚未執行即停下；95具名項目已通過，日誌保留。產品不改，v2改用`weapon-unit-v2/player/`與`weapon-unit-v2/weapon/`獨立輸出後乾淨重跑。
- 正常兩原版首戰／左右出劍與回城配對沿`weapon-pair-v1/`，固定舊失敗GUI日誌沿`weapon-gui-replay-v1/`／`weapon-gui-replay-test-v1.jsonl`。另逐個完成來源記武器已知／空白／缺圖，缺圖原始來源保存，不以原版繼續執行稱HD通過。
- 武器單元v2共123具名通過、無fail／skip，候選4886361f及vet／build通過。38正常點全部原版相同；武器229來源中48有圖、181空白、無缺圖，回城另九來源空白。連續動作卻有兩個新主角鍵、十個可見退回畫格。增量來源／等價核對沿`weapon_player_extra_v1.py`／`weapon-player-extra-v1/`，參考圖沿`weapon_player_extra_reference_v1.py`；補圖及提示詞沿`player-motion-extra-handpaint-v1.png`／`player-motion-extra-prompt-v1.txt`，不能把v1的退回當主角全段HD。
- 候選身份與87份runtime來源由`weapon_freeze_v1.py`保存為`weapon-candidate-identity-v1.json`／`weapon-candidate-sources-v1/`，供4886361f的123項及38點舊收據回查。最終新候選另用v2，不改舊快照。後續30姿態登錄／獨立圖沿`player_extra_register_v1.py`／`player-motion-extra-preview-v1/`及`weapon_preview_v3.py`／`weapon-composition-preview-v3/`，正式單元與建置另沿`weapon-unit-v3/`、`weapon-preview-v2`及v2的build／vet日誌。
- 最終候選v2的正常GUI工具沿`weapon_gui_prepare_v2.py`／`weapon-tools-v2/`，實際左右出劍及回城分別沿`weapon-gui-active-v2/`與`weapon-gui-return-v2/`。暫停在真正可見武器來源再切換，不以放開後已無武器的站姿圖驗武器HD；完整來源與原生日誌另核。
- 正常首戰加左右延伸鍵序以自有[tools/weapon-hd.commands](tools/weapon-hd.commands)保存，與本地weapon-normal-v1.commands位元組相同。30姿態READY全文存`weapon-player-ready-spec-v2.md`，最終同鍵序重跑沿`weapon-pair-v2/`；來源／主入口、圖面完整提示詞、收尾與PRIVATE遠端身份另存weapon的最終索引，不覆寫player前輪索引。
- 由實際GUI觀察器的完成步數指定只讀擷取來源，工具與目標沿`weapon_gui_targets_v1.py`／`weapon-gui-source-targets-v1.json`；新GUI整份日誌兩原版重播、指定來源及存檔／FileOps驗證沿`weapon-gui-pair-v2/`／`weapon-gui-pair-test-v2.jsonl`。實際左右刀刃及回城地面獨立像素沿`weapon_gui_pixels_v1.py`／`weapon-gui-pixels-v1/`，不得以站姿或不同來源替代可見武器驗收。
- GUI v2兩路的動作／程式退出已完成，但收尾固定來源檢查發現開工中修改cave_frames_pair_test.go，整輪仍判失敗，原失敗與圖面保留。實際差異僅該測試檔，產品來源與d8666d0a不變。v3凍結全部Go含新測試，完整來源快照另存`weapon-candidate-sources-v3/`／`weapon-candidate-identity-v3.json`；工具與輸出沿`weapon_gui_prepare_v3.py`／`weapon-tools-v3/`、`weapon-gui-active-v3/`與`weapon-gui-return-v3/`。所有來源檔在該批GUI期間停止編輯，不能放寬SHA檢查。
- 最終固定工具的完整單元、vet／build沿`weapon-unit-v4/`、`weapon-vet-v3.log`／`weapon-build-v3.log`與`weapon-preview-v3`；同鍵序原版重跑沿`weapon-pair-v3/`，實際GUI v3目標與日誌另用v3身份。此輪來源凍結失敗屬工作流程錯誤，未修改原版或產品守衛。
- GUI v3的實際來源目標、重播與像素工具為`weapon_gui_targets_v3.py`／`weapon-gui-source-targets-v3.json`、`weapon-gui-pair-v3/`／`weapon-gui-pair-test-v3.jsonl`及`weapon_gui_pixels_v3.py`／`weapon-gui-pixels-v3/`。收尾、三張新增美術與完整提示詞、根端實圖及主入口／PRIVATE遠端身份另存`weapon_audit_v1.py`／`weapon-current-audit-v1.json`、`weapon-art-final-index-v1.json`、`weapon-root-visual-review-v1.json`、`weapon-mainalias-sync-v1.json`與`weapon-repository-sync-v1.json`。自有檔案清單為`weapon-owned-files-v1.list`，不覆寫前輪player索引。
- v3兩GUI在共用CPU負載下耗盡180秒冷入口總牆鐘，已到61891795步、焦點正常、12筆按放確認、無20秒停滯；尚未進入洞窟，不算產品故障或武器驗收。v4改逐條執行，冷入口牆鐘360秒、外層600秒，所有原版指令／輸入預算與來源SHA不變。工具、輸出與時限說明沿`weapon_gui_prepare_v4.py`／`weapon-tools-v4/`、`weapon-gui-active-v4/`、`weapon-gui-return-v4/`及`weapon-gui-timing-review-v4.json`。來源仍用已凍結89檔及同一d8666d0a；只改驗證排程與牆鐘上限。
- 最終GUI v4目標／原生日誌配對與像素各用`weapon_gui_targets_v4.py`／`weapon-gui-source-targets-v4.json`、`weapon-gui-pair-v4/`／`weapon-gui-pair-test-v4.jsonl`及`weapon_gui_pixels_v4.py`／`weapon-gui-pixels-v4/`；不讀v2失敗或v3冷入口未完成圖面作成功替代。
- 像素v4已重建兩份獨立HD，但原貌X截圖為8-bit索引PNG，既有解碼器僅支援RGB／RGBA而停止。失敗工具及兩張輸出保持身份，不改輸入圖；`weapon_png_decoder_v1.py`補PLTE／tRNS與CRC核對，`weapon_gui_pixels_v5.py`／`weapon-gui-pixels-v5/`以同批實際GUI v4重跑。只修驗證工具，原版、候選與89來源不變。
- 現行018及017增量已限定驗收：`weapon-current-audit-v1.json`、`weapon-art-final-index-v1.json`與`weapon-root-visual-review-v1.json`保存125具名、38正常點、2837實際日誌及612864視窗像素結果。主入口同步記`weapon-mainalias-sync-v1.json`，舊c0e66326原執行檔保存在`weapon-previous-main-c0e66326`；PRIVATE提交／遠端與工作樹身份由`weapon-repository-sync-v1.json`保存。舊失敗與前輪收據不改名。

- 洞窟主角HD沿既有`workplace/cave-hd-v1/`接續，來源盤點為`player_inventory_v1.py`／`player-source-index-v1.json`與`player-source-v1/`。規格為[017](docs/spec/017-cave-player-hd.md)，先沿既有67份完整重建及已驗特殊圖像分支整理完整姿態，不重開011／012的來源或生命週期研究。生成素材與完整提示詞沿`player-art-index-v1.json`及`player-*-handpaint-v1.png`，只留本地；未READY不接前端。
- 上述盤點最初的隔離Python命名空間及收據範圍錯誤保留於`player-inventory-preflight-v1.json`與兩份`player_inventory_v1.failed*.txt`。完整合成證明的54／13份不含3AA9繪製前來源，不能把completed-receipt全部73／18份當同一範圍。修正後67份角色下層至24×24暫存區獨立相同，加六份特殊／GUI全畫格來源共73份、14個完整姿態。美術圖集參考另用`player_art_reference_v1.py`及`player-source-v1/reference-atlas.png`，只用解碼資料排版，不改原始圖像。
- 單一14姿態透明圖集使用內建imagegen生成，完整提示詞為`player-atlas-prompt-v1.txt`，產物`player-atlas-handpaint-v1.png`已複製到本工作區，原始輸出仍保留。透明度、格界、原版包絡及直接72×72取樣審查沿`player_art_review_v1.py`／`player-art-review-v1.json`與`player-art-preview-v1/`；未審查不接前端。
- 圖集v1的等分格界審查失敗：格界有alpha1–2的黑色殘留，第四姿態下一行的腳底另跨等分界線數像素。原PNG不修改；`player_art_review_v2.py`／`player-art-review-v2.json`與`player-art-preview-v2/`改按各欄可見主體的分離列群登錄固定裁切，保留裁切內全部原alpha。空格不取樣，主體裁切必須互不相交且包含完整可見墨跡；這是美術登錄，不放寬原版來源或遮擋驗證。
- 主角／地形／後畫遮擋的獨立離線合成沿`player_preview_v1.py`／`player-composition-preview-v1/`，使用既有Python全畫格解碼與取樣，再按原始來源鍵取14姿態。未有素材的來源明示原貌退回，不把未知來源換成相似姿態。
- 十份離線來源中七份有14姿態素材，三份真實原版後畫遮擋來源共同缺一個新姿態。擴充盤點沿`player_inventory_v2.py`／`player-source-index-v2.json`與`player-source-v2/`，包含該三份既有來源；只為新增的第15姿態生成獨立透明PNG，索引沿`player-art-index-v2.json`，提示詞`player-pose14-prompt-v1.txt`。不以相似姿態替代，也不縮減後畫遮擋驗收。
- 單姿態首次生成未遵守前手抬高位置，未採用，來源與拒絕原因記`player-pose14-generation-rejected-v1.json`。改以既有圖集為編輯目標，在原空格14補該姿態，提示詞`player-extra-atlas-prompt-v1.txt`、產物`player-atlas-extra-handpaint-v1.png`。只取新增格，原14姿態仍取舊圖集，不以再生成的其他格取代已審美術。
- 第15姿態的固定裁切及全部15項登錄沿`player_register_extra_v1.py`／`player-art-review-v3.json`與`player-art-preview-v3/`；兩張PNG原子載入的接入契約先記017。十份含後畫遮擋的全幅離線合成重跑沿`player_preview_v2.py`／`player-composition-preview-v2/`，原v1七張成功及三張缺圖保持身份。
- 後續核對`native-paints-proof-v1.json`：上述三份角色可見像素均為0，新鍵0bd51c6d只有底部八列且已被九次原版後畫完全覆蓋。v2的三張「有素材」其實可見HD改動也是0，不能計為新增可見姿態驗收。017正式規則先判定可見性；完全被覆蓋時不畫主角，無須選素材。新增第15張及第二圖集均不採用，14姿態主圖集保持原身份。裁切／合成審查及原判斷勘誤沿`player-ready-review-v1.json`，不推論E8的玩法名稱。
- 017的READY全文保存於`player-ready-spec-v1.md`；正式實作為`internal/prototype/hd/cave_player.go`與測試。來源登錄只取14姿態主圖集。14個取樣與七張可見／三張完全遮擋的獨立圖面用於正式對照，其他未正常走過姿態仍回退原貌，不冒充全遊戲HD完成。
- 主角正式像素、可見性及退回測試集中`player-unit-v1/`，日誌為同目錄`test.jsonl`，輸入／美術唯讀。正常原版配對及新增姿態覆蓋盤點另沿後續player-pair入口，不能由本項離線測試取代。
- 主角測試第一輪尚未執行：舊cave-hd-v1/go.work只列早期/probe模組，當前掛載無該路徑。保留v1環境錯誤，使用獨立`player-build.work`列/project與唯讀/dosgolem，正式重跑輸出`player-unit-v2/`。未另建重複映像或改共享執行器。
- v2的14直接取樣、十張獨立全幅及既有來源／地形回歸通過；最後的純局部遮擋測試誤用全部九筆原版呼叫，實際覆蓋整個24×24，因此沒有可見側。保留v2及`player-failed-mask-test-v2.go.txt`；該純合成測試改取真實中央8×8單次呼叫，完整原版來源仍用九筆不改，乾淨重跑`player-unit-v3/`。
- 正常兩原版冷F7、首場交戰與回城配對及實際姿態覆蓋盤點沿`player-pair-v1/`。除完整狀態外，逐個完成來源記已知姿態／完全遮擋／缺少可見素材；缺素材另存原始來源供補圖，不以測試綠色稱完整角色動畫已完成。此輪FileOps明示啟用記錄，並核對各可寫存檔副本與mtime。
- 主角候選前端建置沿`player-preview-v1`、`player-build-v1.log`及`player-vet-v1.log`。主入口仍為已驗4766810c，未通過正常路徑與視窗前不覆寫它。
- 正常配對v1的回城10點通過；首戰在tools/combat.commands的down／up事件停下，舊測試工具僅接受key／keys。補完整持鍵邊沿支援，不改遊戲前端或原鍵序，沿`player-pair-v2/`重跑。v1部分來源與回城收據保留，不當整批通過。
- v2正常首戰25點／回城10點完整原版相同，存檔不變。首戰174個完成來源中102已知、72可見來源缺九個唯一素材鍵；回城九個完成來源全已知。接續來源盤點沿`player_missing_inventory_v1.py`／`player-missing-v1/`，獨立重建各缺圖來源，先核是否與既有圖面來源具有逐像素相同的平移／鏡射，再決定補圖，不能從相似外觀猜姿態。
- 九個新可見姿態均已獨立核完整原版遊戲區相同，與原14姿態沒有完整576bytes的平移／鏡射等價，須補新美術。補圖沿`player-motion-atlas-prompt-v1.txt`、`player-motion-atlas-handpaint-v1.png`及`player-motion-art-index-v1.json`，以九來源3×3圖集作姿態目標，只以已採用主圖集參考造型。017補圖增量仍DRAFT，審查後才擴正式圖組。
- 九格初圖底列朝向不符，保留`player-motion-atlas-rejected-v1.png`，內建imagegen局部修正的完整提示詞為`player-motion-atlas-edit-prompt-v1.txt`。採用圖的固定登錄與23姿態獨立取樣沿`player_motion_register_v1.py`／`player-motion-preview-v1/`；全幅增量合成沿`player_preview_v3.py`／`player-composition-preview-v3/`。後續驗證分別沿`player-unit-v4/`、`player-pair-v3/`、`player-preview-v2`及`player-build-v2.log`／`player-vet-v2.log`，不覆寫原失敗或舊來源身份。
- 23姿態READY全文另存`player-motion-ready-spec-v1.md`。候選凍結與正常視窗沿`player_gui_prepare_v1.py`、`player-candidate-identity-v1.json`／`player-candidate-sources-v1/`、`player-tools-v1/`及`player-gui-movement-v1/`、`player-gui-combat-v1/`、`player-gui-return-v1/`。GUI像素核對沿`player_gui_pixels_v1.py`／`player-gui-pixels-v1/`，收尾身份另沿`player-current-audit-v1.json`、`player-mainalias-sync-v1.json`及`player-repository-sync-v1.json`。
- GUI v1跳蹲及回城通過；交戰時出現新可見鍵2b975f19而安全回退，嚴格HD檢查停止。該次冷啟動輸入原生日誌及失敗收據保留，固定日誌配對與缺圖盤點另沿`player-gui-replay-v1/`及`player-gui-replay-test-v1.jsonl`。自有檔案入口為`player-owned-files-v1.list`；未完整通過前主入口保持4766810c。
- GUI增量來源分類及原生逐像素／幾何核對沿`player_gui_missing_v1.py`／`player-gui-missing-v1/`。先判定是否為已驗姿態的精確顏色變化，再決定補圖；不以相似外觀合併來源鍵。
- 固定3099筆GUI日誌、206完成來源及519次高清繪製原版相同，零來源拒絕。三個缺圖鍵中，兩個確為新肢體姿態；2b975f19與既有站姿完整576透明／幾何相同，但配色不同，保留原版配色提示，不推論受傷或無敵語意。補圖沿`player-gui-atlas-prompt-v1.txt`／`player-gui-atlas-handpaint-v1.png`，固定登錄沿`player_gui_register_v1.py`／`player-gui-art-preview-v1/`；增量獨立合成沿`player_preview_v4.py`／`player-composition-preview-v4/`。
- 25姿態及原版配色保留READY全文為`player-gui-ready-spec-v1.md`，後續單元沿`player-unit-v5/`，建置／vet為`player-preview-v3`、`player-build-v3.log`／`player-vet-v3.log`。新候選凍結／正常GUI沿`player_gui_prepare_v2.py`、`player-candidate-identity-v2.json`／`player-candidate-sources-v2/`、`player-tools-v2/`及三個`player-gui-*-v2/`。像素沿`player_gui_pixels_v2.py`／`player-gui-pixels-v2/`，不覆寫23姿態候選或GUI v1失敗。
- 25姿態的固定原版35點重跑沿`player-pair-v4/`；同一份失敗GUI日誌重播沿`player-gui-replay-v2/`／`player-gui-replay-test-v2.jsonl`。來源及藝術完整身份與收尾沿`player-art-final-index-v1.json`、`player_audit_v1.py`／`player-current-audit-v1.json`，主入口只在驗證通過後同步。
- GUI v2三路已通過；第一張jump-held截圖過早仍是站姿，雖日誌已有pose5／原生Y78，不能稱該圖驗了半空外觀。同候選補拍工具沿`player_gui_movement_prepare_v3.py`、`player_gui_movement_v3.py`／`player-movement-identity-v3.json`與`player-gui-movement-v3/`，在正常Up放開後以工具列暫停半空來源再切換，程式及86來源不變。獨立腳本首次啟動缺helper搜尋路徑，遊戲尚未啟動、輸出仍空；同命令補PYTHONPATH指向唯讀player-tools-v2後重跑。根端實際目視身份記`player-root-visual-review-v1.json`。
- 本輪最終為017限25姿態CONFORMED，c0e66326／86來源；`player-current-audit-v1.json`核137自有檔、93具名測試及全部已列收據，`player-mainalias-sync-v1.json`記主入口已同步。舊4766810c另存`player-previous-main-4766810c`。完整三圖路徑、SHA與四份完整提示詞存`player-art-final-index-v1.json`，均只本地。GUI像素工具先誤用cycles欄名而未出圖，後按main契約elapsed_cycles重跑；最終兩洞窟加回城地面612864像素零差。收據及歷史版本保持原身份；目前工作以上方狀態表及WORKLIST為準。

- 共用物品文字沿既有`workplace/item-use-v1/`接續，來源盤點為`shared-inventory-source-v1.py`／`shared-inventory-source-v1.json`。正常原版工具為`shared-probe-v1.go`，鍵序及輸出分別為`shared-magia-v1.commands`／`shared-magia-v1/`與`shared-sabre-v1.commands`／`shared-sabre-v1/`；譯表與正式接入須依[規格016](docs/spec/016-shared-inventory-text.md)審查。
- 共用物品的手冊研究沿同根`shared-manual-v1.py`／`shared-manual-v1/`，逐檔對RAR原始大小與CRC核既有掃描副本，掃描只供本地術語核對，不入Git。既有7z映像缺RAR codec，已找到所需既有副本，不重試舊unar選檔。原生字形與完成區驗證另沿`shared-glyph-proof-v1.py`／`shared-glyph-proof-v1.json`。
- Sabre的1,200售價高於起始1,000；`shared-sabre-v1/`保留正常資金不足結果。正常戰鬥取得寶石及換錢的接續研究沿`shared-probe-v2.go`、`shared-funding-v1.commands`／`shared-funding-v1/`，不改金額或亂數。補齊AC32的A35B讀取端與試譯容量審查沿`shared-ready-review-v1.py`／`shared-ready-review-v1.json`及`shared-labels-candidate-v1.tsv`。
- 016正式配對與候選沿`shared-test-v1.jsonl`／`shared-pair-v1/`，原版戰鬥回門的連續鍵序延伸為`shared-funding-v2.commands`／`shared-funding-v2/`，保留v1尚未走回門口的畫面。
- 第一寶石後正常回城及銀行換錢延伸為`shared-funding-v3.commands`／`shared-funding-v3/`。016候選建置與檢查沿`zeliard-preview-shared-v1`、`shared-build-v1.log`、`shared-vet-v1.log`；獨立中文像素沿`shared-pixels-v1.py`／`shared-pixels-v1.json`。來源凍結及GUI入口為`shared-freeze-v1.py`、`shared-candidate-identity-v1.json`／`shared-candidate-sources-v1/`及`shared-gui-v1.py`／`shared-gui-v1/`。
- 016首輪測試誤把完整遮擋但仍保留的舊標籤數視為可見殘留，保留`shared-failed-test-v1.go.txt`、v1日誌及部分畫面。關閉驗收改依實際像素完全還原，沿`shared-test-v2.jsonl`／`shared-pair-v2/`乾淨重跑，不改正式繪圖生命週期。
- 既有藥水使用及裝備頁回歸沿`shared-regression-v1.jsonl`／`shared-regression-v1/`。本輪自有檔案清單保存於`shared-owned-files-v1.list`，供來源及提交邊界核對。
- 共用物品收尾沿`shared-audit-v1.py`／`shared-current-audit-v1.json`、根端實際圖目視清單`shared-root-visual-review-v1.json`。候選、主入口同步及私人遠端身份分開保存為`shared-mainalias-sync-v1.json`、`shared-repository-sync-v1.json`；通過前不覆寫舊fca020a3入口。
- GUI v1已驗城鎮物品頁，洞口驗證工具沿舊藥水路徑假設四次Right8m已走至右側角色範圍，但實圖角色仍在左側，未送Up即停止。v1原身份及失敗保留；`shared-gui-v2.py`／`shared-gui-v2/`延長正常右行至八段有界預算，仍只在固定洞門與完整角色像素同時吻合後送Up，不改產品守衛。候選程式與85來源不變，GUI腳本身份另存`shared-candidate-identity-v2.json`，來源快照仍指v1。

- 其餘片頭中文先沿`workplace/intro-narration-v1/`研究，入口為`probe_intro_v1.go`與`original-natural-v1/receipt.json`。只走無輸入自然片頭，觀察原版繪字與模組來源；未達READY不改正式覆繪。
- 同研究區`probe_intro_v2.go`／`original-title-space-v2/`在正常200m標題送一次Space，再自然觀察後續流程。魔王宣言的來源在同一SAR1模組末段，與首章32C9捲動端分開追蹤。
- 原版完整敘事的固定預算延伸為`probe_intro_v3.go`／`original-story-v3/`。只沿同一正常Space鍵序擴大觀察；IDA來源匯出沿`ida_intro_narration_v1.py`／`ida-v1/`，保留原名、位址、bytes與未證實警示。
- IDA首輪匯出遇ASCII預設編碼，未產有效JSON；修正明示UTF-8的`ida_intro_narration_v2.py`輸出至`ida-v2/`，保留首輪錯誤，不當作反組譯能力缺口。
- 魔王宣言的逐字完成觀測為`probe_threat_v1.go`／`original-threat-v1/`，另核底圖與兩次字形返回。IDA v2的6300是錯誤切點；`ida_intro_narration_v3.py`／`ida-v3/`改由原始call目的62FD與631E重建，不沿用該錯誤函式邊界。
- 魔王字形及三行來源的獨立核對為`prove_threat_v1.py`／`threat-proof-v1.json`，規格013依此審查。後續中文與完整故事仍各自驗收，不把資料庫函式名當作語意證據。
- 魔王三行的來源／字型／容量審查為同根`ready-threat-review-v1.json`。正式實作及正常視窗尚未完成前，不把READY當作中文完成。
- 魔王中文接入測試集中同根`threat-pair-v1/`與`threat-test-v1.jsonl`，建置為`zeliard-preview-threat-v1`／`threat-build-v1.log`，來源與視窗另凍結後驗證；目前主入口仍0227。
- 首輪測試錯把185m當作標題，但原版宣言仍在。保留`threat-failed-test-v1.go.txt`及失敗紀錄，改用原版200m標題收據驗撤除，輸出`threat-pair-v2/`／`threat-test-v2.jsonl`，不改覆繪接受條件。
- 宣言候選的76份來源保存於`candidate-identity-v1.json`／`candidate-sources-v1/`。正常視窗工具為`gui_threat_v1.py`，輸出`gui-threat-v1/`；原版全劇情觀測只證明已讀到的來源，末尾等待尚未作玩家路徑結束驗收。
- 完整故事文字盤點為`inventory_story_v1.py`／`story-source-v1.json`，核對SAR原始偏移與正常6A75字形消費，不把劇情中的文字敘述當作玩法規則。
- 故事清框、字形與字色的完整流程觀測沿同根`probe_story_flow_v1.go`／`original-story-flow-v1/`。先核每個完整字形與FE清框，再定譯文段落；未審查不接正式前端。
- 故事字色／換行分段草案沿`story_segments_v1.py`／`story-segments-v1.json`。試譯在`text/intro-story-prototype.tsv`，只保存58個正文來源；最初P與控制碼不當作待翻譯敘事。
- 故事原生字形與58片段的獨立驗證沿`prove_story_flow_v1.py`／`story-flow-proof-v1.json`。中文容量及47頁離線圖沿`preview_story_pages_v1.py`／`preview-story-pages-v1/`，不冒充正式前端或GUI收據。
- 故事接入契約見`docs/spec/014-intro-story.md`；審查保存為同研究根`ready-story-review-v1.json`。原始控制碼仍由原版執行，前端僅保存完整文字區與中文進度。
- 規格014的READY全文另存`ready-story-spec-v1.md`，七張根端離線目視清單存`story-draft-visual-review-v1.json`；與後續正式來源、GUI及收尾身份分開。
- 故事正式接入測試沿`story-test-v1.jsonl`／`story-pair-v1/`，前端候選為`zeliard-preview-story-v1`，vet／建置記於`story-vet-v1.log`／`story-build-v1.log`。正常視窗、原版冷重跑及獨立像素通過後，主入口同步574094bc；舊f733候選保留。
- 故事視窗工具為`gui_story_v1.py`，完整故事輸出`gui-story-v1/`，首頁後正常Space跳過另輸出`gui-story-skip-v1/`。候選身份與80份來源另存`story-candidate-identity-v1.json`／`story-candidate-sources-v1/`；兩路均記原生輸入／週期日誌。
- 正式47頁與獨立離線排版的核對沿`compare_story_pair_pixels_v1.py`／`story-pair-pixels-v1.json`。故事後的六行捲字仍走32C9消費端，與製作名單一起保留待辦，不混入58段正文驗收。
- 故事GUI的原生日誌冷重播工具為`replay_story_gui_v1.go`，完整／跳過輸出`story-gui-replay-v1/`／`story-skip-gui-replay-v1/`；獨立實際畫面核對沿`verify_story_gui_pixels_v1.py`，輸出`story-gui-pixels-v1.json`／`story-skip-gui-pixels-v1.json`。
- 故事收尾核對沿`audit_intro_story_v1.py`／`story-current-audit-v1.json`，根端實際視窗目視清單為`story-root-visual-review-v1.json`。通過後的主入口同步工具及收據為`promote_story_v1.py`／`story-mainalias-sync-v1.json`；提交與私人遠端身份另記`story-repository-sync-v1.json`。
- 三組捲字來源沿同根`probe_scroll_blocks_v1.go`／`original-scroll-blocks-v1/`，獨立原生畫面核對為`prove_scroll_blocks_v1.py`／`scroll-blocks-proof-v1.json`。來源與版面依[規格015](docs/spec/015-intro-scroll-blocks.md)審查，先READY再接正式路徑；本輪已限已列範圍CONFORMED。
- 三組捲字的獨立中文排版／離線圖沿`preview_scroll_blocks_v1.py`／`preview-scroll-blocks-v1/`；來源與容量審查保存為`ready-scroll-blocks-review-v1.json`及`ready-scroll-blocks-spec-v1.md`，與後續正式畫面分開。
- 規格015的READY接入稿另存`scroll-blocks-ready-source-v1.go.txt`，只在READY全文保存後同步到`internal/prototype/overlay/scroll.go`；正式配對、候選與GUI另凍結，不將離線圖當作已接入。
- 捲字正式接入測試沿`scroll-blocks-test-v1.jsonl`／`scroll-blocks-pair-v1/`與既有首章入口的`scroll-legacy-pair-v1/`，候選為`zeliard-preview-scroll-v1`，靜態檢查與建置為`scroll-blocks-vet-v1.log`／`scroll-blocks-build-v1.log`；正常視窗、冷重跑及像素通過後主入口同步fca020a3，舊574094bc候選保留。
- 捲字正常視窗工具為`gui_scroll_blocks_v1.py`，輸出`gui-scroll-blocks-v1/`。候選身份及凍結來源另存`scroll-candidate-identity-v1.json`／`scroll-candidate-sources-v1/`，實際原生輸入日誌供獨立原版冷重播，不還原記憶體。
- 捲字候選凍結工具為`freeze_scroll_candidate_v1.py`，只在正式配對、首章相容入口、vet及建置通過後執行；不覆寫舊574094bc候選與來源快照。
- 捲字GUI獨立冷重播沿`replay_scroll_gui_v1.go`／`scroll-gui-replay-v1/`，逐筆核原生日誌並另存原版來源／完成列事件。實際視窗像素核對沿`verify_scroll_gui_pixels_v1.py`／`scroll-gui-pixels-v1.json`，使用原版事件重建中文行位置，不以正式覆繪程式當預期圖。
- 捲字收尾核對沿`audit_scroll_blocks_v1.py`／`scroll-current-audit-v1.json`，根端實際圖目視清單為`scroll-root-visual-review-v1.json`。通過後以`promote_scroll_v1.py`產生`scroll-mainalias-sync-v1.json`，提交與私人遠端身份另記`scroll-repository-sync-v1.json`。
- 實際視窗的獨立原版重跑與像素核對沿`replay_gui_original_v1.go`／`gui-original-pair-v1/`及`verify_gui_pixels_v1.py`／`gui-pixels-v1.json`。原版只重跑相同冷啟動與週期預算，不還原GUI記憶體。
- 魔王三行前輪收尾為`audit_intro_threat_v1.py`／`current-audit-v1.json`，根端三張圖的目視清單為`root-visual-review-v1.json`。該輪主入口同步與提交後私人遠端身份分別記於`mainalias-sync-v1.json`／`repository-sync-v1.json`。
- 洞窟呈現生命週期與動態格的窄研究沿`workplace/cave-hd-v1/lifecycle_probe_v1.go`與`lifecycle-v1/`。v2另核7114／714C完成及703C／7136開始，輸出`lifecycle-v2/`；`lifecycle-return-v1.commands`沿正常入洞再Up回城，輸出`lifecycle-return-v1/`。只讀觀測，不接正式前端。
- 動態地圖至圖格的獨立解碼沿同根`decode_dynamic_map_v1.py`／`dynamic-map-proof-v1.json`，取前輪已封存的3123來源，不重跑原版或猜物件種類。
- 新完成點的完整重建沿`decode_lifecycle_actor_v1.py`、`prove_lifecycle_composition_v1.py`與`decode_lifecycle_map_v1.py`，各用lifecycle-v2的真實來源；彙總為`lifecycle-composition-proof-v1.json`及`lifecycle-map-proof-v1.json`。
- 隔離Go整畫格解碼試作v1把整個資源段當不可變，169事件因未消費區域變動而拒絕，保留`full-frame-v1/`。v2改為主角與完成時刻各自取真實來源，仍以整區重建驗證，沿`full_frame_v2.go`／`full_frame_probe_v2.go`與`full-frame-v2/`。未審查前不進正式前端。
- 來源與呈現依[規格012](docs/spec/012-cave-frame-sources.md)，審查為`ready-layer-review-v1.json`，回填護欄為`resolution-index-lifecycle-v1.json`。單元測試沿`cave-unit-v2.jsonl`，正常兩機配對輸出集中`layer-pair-v1/`。前端主入口仍為a691，尚未使用新呈現器。
- 地形正式接入前的規格011審查為同根`ready-terrain-review-v1.json`；使用前輪兩張候選及已驗原版圖格身份，不重新生成素材。READY全文與實作／驗證各保存自己的身份。
- 地形接入後的像素與退回驗證沿`cave-terrain-unit-v1.jsonl`，候選前端為`workplace/cave-hd-v1/zeliard-cave-preview-v1`，建置記錄`cave-build-v1.log`。尚未驗候選時不覆寫a691主入口。
- GUI候選另沿`zeliard-cave-preview-v2`與`cave-build-v2.log`，只補完成來源的主角位置觀測。唯讀地形／中文配對在`terrain-pair-v1/`；GUI工具集中`cave-gui-tools-v1/`，輸出`gui-terrain-v1/`，先凍結再執行，不覆寫舊GUI收據。
- 同一GUI候選另以`gui-combat-terrain-v1/`抽首戰／第一寶石，以`gui-return-terrain-v1/`抽正常回城。三路共用唯讀`cave-gui-tools-v1/`，各自獨立程序、輸出及收據；不追加戰鬥重試。
- GUI及收尾使用`candidate-identity-v1.json`／`candidate-sources-v1/`的固定來源。最後觀測欄位加入後的同來源配對為`canonical-pair-v1/`及`cave-canonical-pair-v1.jsonl`。未通過配對前，三路GUI工具維持準備狀態。
- GUI像素核對工具為`cave-gui-tools-v1/cave_pixel_oracle_v1.py`，沿先於Go實作的Python地圖、主角與取樣證據；預檢輸出`pixel-preflight-v1/`。它只產生預期圖與比較收據，不拼造原版記憶體或修改素材。
- 首戰GUI gui-combat-terrain-v1/在第二段前進出劍因洞窟來源撤除而停止，保留failure及66張PNG；較早一張已見第一寶石1與HD。新窄研究沿hurt-normal-v1.commands及hurt-normal-v1/，只接在已驗25點後繼續正常前進／出劍與等待，觀察E8等原始欄位，不猜語意或重擲亂數。
- 回城GUI gui-return-terrain-v1/已切回城鎮模組，但前景／地面守衛拒絕。頭段與mask相同，差異在Town主角未使用的11B1:68A0..8000保留區。隔離驗證沿return-study-v1/，只先改來源守衛副本並檢查原版回城兩點，不直接修改正式守衛。
- 回城使用區的READY審查與回填為return-ready-review-v1.json／return-resolution-index-v1.json，正式守衛依007修訂。診斷候選zeliard-cave-preview-v3新增拒絕候選及live來源擷取；cave-diagnostic-unit-v3.jsonl與cave-build-v3.log記其驗證，既有v1 GUI與73來源快照不覆寫。
- 診斷候選v4另加-input-trace，記真實IRQ輸入邊沿與逐次RunCycles／音訊取出，供原版冷重播，並修正只載洞窟素材時的工具列HD入口。zeliard-cave-preview-v4與cave-build-v4.log不覆寫舊候選；診斷工具／輸出各為cave-diagnostic-tools-v4/及gui-diagnostic-v4/。
- 診斷版固定來源另存diagnostic-candidate-identity-v4.json／diagnostic-sources-v4/。此視窗只收集拒絕來源與guest-inputs.jsonl，不當作首戰HD驗收通過，也不重試到出現特定結果。
- 原生輸入日誌的冷重播工具為同根`input_trace_replay_v1.go`，輸出`input-replay-v4/`。先核對每步暫存器與最終全部記憶體，再以事先固定的正常按鍵延伸觀測；不注入狀態。
- FF39新增來源先依012的DRAFT增量隔離於`full_frame_v3.go`；固定同日誌與延伸的重播輸出`input-replay-v5/`，獨立單一來源檢查為`ff39-proof-v1.json`。未核對前不進正式解碼。
- FF39的READY審查沿`ready-ff39-review-v1.json`；下一候選為`zeliard-cave-preview-v5`，測試與建置沿`cave-current-unit-v5.jsonl`／`cave-build-v5.log`。三條正常GUI工具凍結於`cave-current-tools-v5/`，各自輸出`gui-current-movement-v5/`、`gui-current-combat-v5/`、`gui-current-return-v5/`，保留舊失敗身份。
- v5同來源正常首戰與回城的35點唯讀配對集中`current-pair-v5/`，命令為既有entry-v3及lifecycle-return-v1；不重開鍵序研究。
- v5兩條GUI未完成，原日誌直接冷重播至`combat-replay-v5/`及`return-replay-v5/`，保存真實拒絕來源與最後城鎮完成點。不把沒有最終GUI記憶體擷取的失敗路徑稱為完整RAM對拍；只核每筆暫存器／週期。
- 新圖像分支隔離研究沿`full_frame_v4.go`／`prove_actor_profiles_v2.py`與`actor-profiles-proof-v2.json`；完整固定重播分別輸出`combat-profile-replay-v6/`、`return-profile-replay-v6/`。回城地面相位研究沿`ground-return-proof-v1.json`，保留009舊camera假定與勘誤。
- 回城地面隔離副本為`ground-return-study-v1/`，只在該目錄改寫草案；`ground-return-draft-test-v1.jsonl`、`ready-ground-return-review-v1.json`記其測試與READY審查。正常GUI仍使用各自固定候選。
- 角色增量READY審查為`ready-actor-profiles-review-v2.json`。候選v6及固定來源另存`zeliard-cave-preview-v6`／`current-candidate-identity-v6.json`／`current-sources-v6/`。只重驗受實作變更影響的首戰、回城與正常配對，v5移動收據保留原身份。
- v6固定視窗工具為`cave-current-tools-v6/`，實際首戰及回城分別輸出`gui-current-combat-v6/`／`gui-current-return-v6/`。失敗分支先暫停並擷取真實live／完成來源，再退出，方便對同一筆日誌重播；不重試到通過。正常35點回歸集中`current-pair-v6/`。
- v6回城GUI通過；首戰在第一段回頭拾取出現116像素整區失配，已保留真實拒絕來源。定位沿`combat-rejected-pixels-v6.json`；冷重播原日誌至`combat-rejected-replay-v6/`，不重新跑GUI取另一個結果。
- 增量覆寫只讀研究沿`input_trace_replay_writes_v3.go`與`combat-write-replay-v6/`，只監看已定位時間及原生角色底部VGA，不改寫原版，也不接正式呈現。
- 原版3272靜態格覆寫的隔離解碼為`full_frame_v5.go`，觀測工具`input_trace_replay_paints_v4.go`，固定同日誌輸出`combat-paint-replay-v7/`。以真實3272呼叫參數推導覆寫區，不由最終顏色猜可見性。
- 同一來源另證實3599動態圖格會在舊主角之後覆寫遮罩像素。隔離v6解碼及`input_trace_replay_paints_v5.go`同時記錄兩類真實覆寫，輸出`combat-paint-replay-v8/`。v7剩餘50／51像素與動態圖格逐點一致，舊失敗保留。
- 增量覆寫獨立證據為`prove_native_paints_v1.py`／`native-paints-proof-v1.json`，READY審查為`ready-native-paints-review-v1.json`。候選及測試沿v7身份，後續正式呈現攜帶來源呼叫與遮罩，不改原生日誌。
- 接入候選為`zeliard-cave-preview-v7`，單元／建置為`cave-current-unit-v7.jsonl`／`cave-build-v7.log`。唯讀回歸輸出`current-pair-v7/`；固定舊失敗日誌配對輸出`current-journal-pair-v7/`，正常GUI工具與結果使用`cave-current-tools-v7/`／`gui-current-combat-v7/`。
- 本次接入收尾使用`runtime-index-v7.json`／`current-runtime-audit-v7.json`，腳本為`audit_runtime_v7.py`，與前輪0df離線研究索引分開。資料與收據維持ignored workplace，原件／快照／字型／存檔另核雜湊。
- 已驗候選的主入口同步記於`mainalias-runtime-sync-v7.json`；Git／私人遠端的提交後核對另記`repository-runtime-sync-v7.json`。
- 新日誌測試加入後的同來源單元／靜態檢查另記`cave-current-unit-final-v7.jsonl`／`cave-vet-final-v7.log`。自有檔案清單固定於`current-owned-files-v7.json`，不含任何原版或本地素材。
- v7同一候選另驗移動／跳蹲，輸出`gui-current-movement-v7/`；三路GUI不互換原版收據。最新像素核對沿`verify_current_pixels_v9.py`與`pixel-current-v9/`，使用各自實際完成來源與增量紀錄。
- R-gem-reference：v7首戰工具把`advance-strike-2.indexed`的寶石0當成1，故first_gem_bitmap判定無效，HD與原版配對仍各自有效。正確1參考為原版`pickup-attempt-left-1.indexed`，已人工讀圖。修正工具另凍結於`cave-gem-tools-v8/`，結果`gui-gem-combat-v8/`；候選程式仍v7不變，移除按住中抓圖造成的輸入超時，保留所有舊收據。
- 寶石勘誤的獨立像素核對沿`verify_gem_reference_v8.py`／`gem-reference-review-v8.json`。根端目視的有限樣本另記`root-runtime-visual-review-v7.json`，不把未逐張讀過的圖算作人工驗收。
- 修正寶石GUI的完整地形取樣另記`verify_current_pixels_v10.py`／`pixel-current-v10/`，沿v9方法核對v8首戰自己的原版完成來源，不借用v7首戰畫面。
- 現行像素核對工具`verify_current_pixels_v6.py`輸出`pixel-current-v6/`，分開v5移動、v6首戰診斷最終擷取、v6回城actualX地面，不把不同路徑合成同狀態收據。
- 像素工具v6把含工具列擷取誤當600列而停止，屬驗證座標錯誤。v7依實際PNG的680／600列選畫面起點，輸出`pixel-current-v7/`，不改候選或圖面。
- 像素工具v8依既有ground_check的預乘alpha契約補回原版地面作底，輸出`pixel-current-v8/`。v7未處理alpha的停止保留，不歸類為產品缺陷。

- 前輪洞窟離線研究索引為workplace/cave-hd-v1/research-index-v1.json，該輪收尾為current-audit-v1.json，僅對應0df2c7e的未接入狀態。美術／提示詞為art-index-v1.json，四圖目視為visual-review-v1.json，當輪遠端同步為repository-sync-v1.json。previews-v2沿相同四圖，僅修正receipt的岩面包絡說明；原版零色色號不具通用透明語意。

- 洞窟HD沿[規格011](docs/spec/011-cave-hd.md)，本輪地形接入為限定READY。研究集中`workplace/cave-hd-v1/`，與已封存首戰／移動收據分開；前輪`frames_v1.go`及`frames-v1/`保留早期觀測身份。正式主入口仍為a691，候選已擴至73來源。
- 洞窟地形美術研究同放`cave-hd-v1/`，`terrain-reference-v1.png`從正常原版圖塊解碼；`terrain-prompt-v1.txt`及`terrain-generation-v1.json`保存內建imagegen提示詞、參考身份、產物及權利分類。只作DRAFT離線研究，不接正式路徑。
- 洞窟目前合成證據為同目錄`frames_v3.go`／`frames-v3/`、`decode_body_sources_v3.py`及`composition-proof-v3.json`。正常左向出劍補證使用`frames-left-v1.commands`／`frames-left-v1/`，仍冷F7與正常離門，不注入原版狀態。地形第二候選及完整提示詞沿`terrain-generation-v2.json`／`terrain-prompt-v2.txt`，離線圖沿`terrain-preview-v2.json`。
- 連續岩面材質入口為同目錄`rock-material-generation-v1.json`／`rock-material-prompt-v1.txt`。目前地形離線研究沿`preview_terrain_v5.py`／`terrain-preview-v5.json`，以明示完成來源清除原版角色／武器、繪地形後再還原原貌角色，核對整張原版可重建及中文標籤來源。不當作正式接入或GUI驗收。
- 可重用離線入口為[洞窟預覽工具](tools/cave_hd_preview.py)，參數為`--baseline /baseline --study /study --out /output`。研究輸入及素材唯讀掛載，空白輸出目錄另掛可寫；本輪輸出集中`workplace/cave-hd-v1/previews-v1/`，入口`receipt.json`。三份素材及完整提示詞由同根`art-index-v1.json`索引。

- 移動GUI修正版`workplace/combat-v1/movement-v2/`同43055／65371退出0，原版已證離門鍵序、Up跳／落地／Down蹲／放開恢復、1000／0及五欄HUD已驗；500輸入及存檔mtime不變。根端五圖目視入口為`root-movement-review-v2.json`。首戰GUI保留自己的限定身份，補驗未重跑交戰。

- 首場戰鬥目前來源身份固定於`workplace/combat-v1/current-identity-v1.json`，完整67份執行與測試來源另存`current-sources-v1/`。前端仍使用a691，新增一份測試與第15條鍵序。`pair-review-v1.json`核原版重播與中文配對的25點，`root-pair-review-v1.json`保存根端三張目視；`audit_combat_current_v1.py`及`current-audit-v1.json`是首戰封存收尾，不重跑作本輪洞窟研究現況。私人遠端與乾淨工作樹身份另記`repository-sync-v1.json`。
- 洞窟繪製的窄研究仍使用`workplace/combat-v1/`：`cave-entered.primary`取自正常當次1MiB的01B1段，`ida_cave_v3.py`及`ida-cave-v3.json`保存IDA9.4原始位址／bytes／xref。`cave_actor_v2.py`／`cave-actor-proof-v2.json`核本體與前後手臂，`cave_terrain_v1.py`／`cave-terrain-proof-v1.json`核靜態圖塊。完整身份入口為`cave-research-index-v1.json`。使用既有py312-v1映像，不重建工具鏈；v1匯出編碼失敗保留。本資料只供來源定位，不代表HD已接入。
- 首戰GUI結果集中`workplace/combat-v1/gui-v1/`，29699／31475退出0。首劍、交戰、寶石0→1及五欄HUD已見；低生命後的Up／Down只有送鍵證據，截圖未能清楚辨認跳／蹲。補驗只做入洞後、交戰前的移動，不重跑首戰；工具使用同根的`movement-tools-v1/`，唯一結果使用`movement-v1/`，仍沿a691及目前67來源身份。
- 洞窟完成點的唯讀探針為`workplace/combat-v1/frame_probe_v1.go`，輸出集中同根`frame-v1/`；只觀察IDA已定位的01B1:3123，不改原版。與固定v3鍵序的live狀態另作核對，完成來源不得冒稱當次live。
- 入洞門口與跳躍輸入的情境另以正常原版核對：`workplace/combat-v1/movement-original-v1.commands`及`movement-original-v1/receipt.json`，固定先左移離門再跳／蹲，不交戰、不注入狀態。不將門口Up誤當無條件跳躍。
- 移動GUI初版`movement-v1/`在站姿驗證工具的縮圖幾何停下，未送洞窟跳／蹲；完整960×680視窗誤縮320×200，混入工具列與頁尾。根端與tester正確裁圖各核同ROI零差。修正用`movement-tools-v2/`及唯一`movement-v2/`，並採已驗左移離門鍵序；初版失敗保留，不增加像素容忍。

- 首場戰鬥正常原版入口為`workplace/combat-v1/entry-v3/receipt.json`，中文配對為`pair-v3/receipt.json`，25點完整原版狀態相同；正式[正常鍵序](tools/combat.commands)及[配對測試](internal/prototype/overlay/combat_path_test.go)已接入驗證。實際視窗與移動補驗、洞窟RE來源見本節上方入口。前端a691及原29檔不變；本項不等於戰鬥／洞窟HD完成。

- 主角GUI最小彙總為`foreground-v1/gui-player-v1/final-summary.json`，NAV存相同副本；107PNG含72逐張目視及35完整hash引用。`gui-player-nav-v1/paused-view-pixel-review.json`核20暫停點，`gui-player-v1/final-player-comparison.json`核最終完成來源與面積平均。`pixel-review.json`保留異camera框角18／9／27原始差，不概稱26ROI皆零。315取樣的81保留來源未存主角旗標，不宣稱全真。`audit_player_current_v1.py`與`player-current-audit-v1.json`是本輪唯一現行收尾入口，原29／882快照／font／save及mtime／擁有權／211文件連結已核對；其他歷史audit不重跑作本輪證據。

- 本輪主角沿[規格010](docs/spec/010-town-player-hd.md)。九張本地候選圖、原圖與完整提示詞索引為`coherent-frame-v1/player-art-set-v1.json`；正式只用八個已觀察行走鍵，phase4共用圖不接入。`prepare_player_study_v4.py`／`town_player_v4.go.draft`保存面積平均離線版，`player-preview-v4.json`／`player-pixel-review-v5.json`核十來源；`player-pair-v1/`／`player-pair-test-v1.jsonl`核三條正常鍵序28點完整原版相同。`player-presented-underlay-v1.json`核八可見鍵114688原生像素，`player-presented-preview-v1.json`／`player-presented-pixel-review-v1.json`核八來源下層及新圖各27648高清像素。連同前十來源合計各62208高清像素零差、94居民重疊位置。`player-ready-review-v1.json`固定審查及四圖目視，READY全文為`player-ready-spec-v1.md`。正式40內部pass／vet見`player-production-test-v1.jsonl`／`player-production-vet-v1.log`。正式候選與來源固定於`player-canonical-identity-v1.json`／`player-canonical-v1-sources/`，原版配對輸出為`player-canonical-pair-v1/`；主入口同步a691，19ac保留為ground-v1，見mainalias-sync-player-v1.json。GUI工具沿`foreground-v1/gui-player-nav-tools-v1/`與`gui-player-tools-v1/`，輸出各用`gui-player-nav-v1/`與`gui-player-v1/`；兩路權威40394／3567及30162／43304均退出0，693輸入不變；NAV20暫停點框外NPC與還原0，MAIN最後14336原生／3456主角高清／2889框外NPC像素零差。根端九圖沿root-player-nav-gui-review-v1.json及root-player-main-gui-review-v1.json。完整來源與衛生核對為player-pre-gui-audit-v1.json，現行收尾另記player-current-audit-v1.json。私人遠端同步與乾淨工作樹身份另見repository-sync-player-v1.json。

- 本輪地面接入沿[規格009](docs/spec/009-town-ground-hd.md)與既有`foreground-v1/scene-prototype-v1/coherent-frame-v1/`來源。GUI工具為`foreground-v1/gui-ground-nav-tools-v1/`與`gui-ground-tools-v1/`，結果在`gui-ground-nav-v1/`及`gui-ground-v1/`；主入口及`workplace/bin/zeliard-preview-ground-v1`均19ac2719，`mainalias-sync-ground-v1.json`記同步。`ground-canonical-identity-v1.json`／`ground-canonical-v1-sources/`固定64完整來源與5圖，READY009另存；`ground-canonical-test-v1.jsonl`／`ground-canonical-pair-test-v1.jsonl`與`ground-canonical-pair-review-v1.json`為37內部pass、16完整點及五點地面合成／洞窟拒絕。兩GUI audit退出0、623輸入不變，`visual-review.json`核107PNG，`ground-pixel-review.json`核22暫停點加最終來源，共82432原生RGB格／741888高清直接取樣零差。root九張actualX沿`root-ground-nav-gui-review-v1.json`／`root-ground-main-gui-review-v1.json`；目前完整稽核入口為`ground-current-audit-v1.json`及`audit_ground_current_v1.py`。

- 最新重讀修正候選為 `workplace/bin/zeliard-preview-frames-v2`，SHA-256 `def76c62b558695ff341ebb82ab4b886612c1e8e2666c1113202dbe32dd36f04`。`foreground-v1/scene-prototype-v1/coherent-frame-v1/candidate-identity-v2.json`與`candidate-v2-sources/`固定62來源，READY007／008另存；`source-audit-v2.json`核對99自有檔、52Go、14鍵序、45完整點／11066呈現／315中途保留／523快照及68pass事件。`reload-cave-entry-review-v1.json`另核對三次F7及Malicia模組。GUI短輪v2同38816退出1，重讀後HD已恢復但舊proof角色模板未涵蓋當次player相位；`nav-template-review-v1.json`以全部50資源候選精確找到phase3/x240且角色HD抽樣0差。未送Up且無當次完整RAM，`nav-template-review-v2.json`另驗3528三倍人物像素零差；`alignment-review-v1.json`以正常原版Right500k及8組Left50k／release250k重現並對齊x240到224。新 `foreground-v1/gui-coherent-nav-tools-v3/`與`gui-coherent-tools-v3/`保存50候選及有界正常對齊工具；對應`gui-coherent-nav-v3/`與`gui-coherent-v3/`保存各唯一GUI輸出。nav3唯一83751／9726退出0，三招牌、第二F7、一組短按對齊及正常Up入Malicia已驗；根端五張actualX目視沿`root-nav-gui-review-v2.json`。main3同66607／19124退出0，兩頁及關閉已驗。007／008限定CONFORMED，完整整合沿`current-audit-v1.json`；主入口同步def76，地面保持獨立READY試作。

- 重讀失配的最新證據為 `foreground-v1/scene-prototype-v1/coherent-frame-v1/reload-first-refusal.json`、`reload-scene-proof-v1.json`及 `reload-footprint-v1.json`。已證實group0..4只讀ID0..151，13來源canonical tile一致、三點各14336pixels零差；007／008按此重新READY。原90fbb的61份完整來源另存 `candidate-v1-sources/`並逐檔吻合原identity，失敗可重生。新版驗收見本節最新入口，舊失敗保持原身份。

- [完整畫格規格](docs/spec/008-town-frame-presentation.md) 來源範圍修正已重新READY，舊90fbb失敗證據保留。原版捲動67D8／6852與繪製68AC共用68BA完成點，沿 `foreground-v1/scene-prototype-v1/coherent-frame-v1/review-v4.json`。可撤回候選 `workplace/bin/zeliard-preview-frames-v1` SHA-256為 `90fbb5fa2a04db2e140449ec82a5bfa853a353ef1aae7e852e7986d52522bd2d`，固定51Go／61來源與13鍵序；identity、原READY快照、unit-test-v2、pair-test-v2、pair-v2及source-audit-v1在同工作區。權威20033退出0，7733呈現／263中途保留／未涵蓋0、28點完整1MiB及Regs相同，21內部pass、原件29／snapshot882／存檔mtime／字型／擁有權均通過。`gui-coherent-v1/`同41756退出0、兩頁／切換成立；根端四圖見 `root-gui-review-v1.json`。`gui-coherent-nav-v1/`同12395退出1、三招牌及中途切換成立，但第二次F7後captured221／refused135，洞口未送Up。先前右向缺口及本次重讀失敗均保留；原.indexed維持live，額外presented及source副檔名明示來源時間。

- 行走HD切換與中途暫停的呈現問題沿 `foreground-v1/gui-scenery-nav-v1/`與 `scene-prototype-v1/coherent-frame-v1/`追查。短輪僅通過銀行／魔法招牌，IP35F1中途暫停時原版人物尚未畫完；先查68AC原版繪製開始／完成與完整快照，不放寬像素守衛、不改dosgolem共用來源。

- 前輪主入口與保留別名 `workplace/bin/zeliard-preview-scenery-v3` SHA-256曾均為 `cf763f7e226304064ec7e9bbdfef23237ad6647459f104625d4657d7dd66ba8c`。`foreground-v1/scene-prototype-v1/candidate-identity-v3.json`固定48Go／58來源，與0b4b只差 `overlay/town.go`及 `town_test.go`。`town-ds-pair-v3/receipt.json`與 `town-ds-test-v3.jsonl`記每50k呈現、61次DS非01B1及正常完整配對；`test-v3.jsonl`、同20229 vet／build通過。`pair-v3/`三路徑28點同36364退出0，`source-audit-v3.json`PASS。`gui-scenery-v3/`同595691b9dd08及權威wait1623退出0、兩頁rejected0、241輸入不變、58按放、30圖／26區域、角色／聯集外／還原0；根端四圖入口 `scene-prototype-v1/root-gui-review-v3.json`。規格004限定CONFORMED，READY004／007快照另存。`gui-scenery-nav-v1/`同24169退出1，250輸入不變、74ack、40有效圖；銀行／魔法招牌已驗，武器／洞口未達，449行走樣本328匹配／121拒絕與actualX原貌退回如實保留。

- 新前景GUI第二輪發現居民首頁等待時中文撤除，入口 `foreground-v1/gui-scenery-v2/`及後續 `scene-prototype-v1/mutation-probe-v1/`。先用只開居民中文的原版、每50k呈現與當次pixel差異定位原版producer；未判明前不標GUI通過、不改正式生命週期。

- 後續主角與居民HD的資產盤點入口為 `workplace/town-hd-v1/foreground-v1/actor_inventory_v1.py`、`actor-inventory-v1.json`及 `player-source-sheet-v1.png`。`actor_observations_v1.py`與`scene-prototype-v1/coherent-frame-v1/actor-observations-v1.json`另核對既有12個不同正常狀態，其中11個可見來源共2221個主角不透明像素零差；正常看見三個資源姿態鍵，其餘七個仍是候選。phase2有正常狀態但角色全在畫布外，不能算可見姿態驗收。只核對固定資源及現成來源，不接正式前端；替換不同alpha圖像前須先處理主角下方圖層，避免殘留原版輪廓。
- 主角手繪試樣同放`foreground-v1/`：`player-c2bit0-phase1-handpaint-v1.png`造型棄用，`player-c2bit0-phase1-handpaint-v2.png`為待合成候選。兩份完整提示詞及來源身份為`player-prompt-v1.txt`／`player-prompt-v2.txt`與`player-generation-record-v1.json`／`player-generation-record-v2.json`，目視判讀在`player-art-review-v2.json`。以內建imagegen產生，alpha未改寫，均不入Git。手冊外封面來源為`workplace/manual-terms-v1/珍084-神劍封魔/SCAN0357_001.jpg`及`player-manual-cover-v1.json`；只用右半《神劍封魔》角盔／盔甲輪廓，左半其他遊戲廣告不沿用。此為首輪造型研究；後續八鍵正式接入沿規格010與本節最新入口。
- 主角下層證據為同目錄`player_underlay_v1.py`及`scene-prototype-v1/coherent-frame-v1/player-underlay-v1.json`與逐點indexed。10個已驗Muralla正常畫格以地形→原版居民→原版主角順序重建，143360色號零差，2015個主角不透明像素含94個居民重疊位置。這閉合原版下層來源；後續高清地形與主角alpha合成沿規格010最新入口，不改原版記憶體。
- [主角規格010](docs/spec/010-town-player-hd.md)初版DRAFT研究保留。離線研究沿`coherent-frame-v1/town_player.go.draft`、`player-overlay-v1.json`與`player_preview_v1.go`，十來源underlay及mask與既有獨立Python證據完全一致，兩個已選姿態合成區外0。`verify_player_preview_v2.py`／`player-pixel-review-v2.json`直接核既有原版證據、34560高清下層像素及6912新圖像素零差，94原生居民重疊點保留；v1先前遮罩來源文字不精確，保留原版、以v2為準。`town_player_v2.go.draft`／`player-overlay-v2.json`／`player_preview_v2.go`另試固定裁切對齊腳底，結果入口`player-preview-v2.json`。`player_pose_references_v1.go`／`player-reference-v1.json`僅作十原版資源12倍檢視，不是HD素材或正常姿態驗收。該離線階段保持64正式來源與主入口19ac；後續正式66來源候選另記本節最新入口。
- 下方原生(48,142)-(272,158)畫帶的初步取樣在同工作區 `ground-candidates-v1.json`。10城鎮來源有6個不同色號雜湊，不能當固定背景直接換圖；捲動關係已由後續ground-proof-v1解出，初始化producer仍未知，未接前端。
- 後續 `foreground-v1/ground_proof_v1.py`／`ground-proof-v1.json`已證實10來源共35,840原生地面像素可由112pixels週期及camera重建。上8列每camera移8pixels，下8列移16pixels；IDA捲動原始來源見 `scene-prototype-v1/coherent-frame-v1/ida-scroll-v1.json`。這解決捲動關係，初始化producer仍未知，其他camera先列強推論；尚未接地面HD。來源與待驗接入沿[地面規格](docs/spec/009-town-ground-hd.md)，已由隔離試作證據審查到READY，正式入口尚未接入。`ground-prompt-v1.txt`及`ground-generation-record-v1.json`保存完整內建imagegen提示詞、實際2172×724圖面、來源hash與權利分類；候選crop為0,273,2172,467，alpha未修改，接縫與前端尚未驗。`scene-prototype-v1/coherent-frame-v1/ground_preview_v1.go`及`ground-preview-v1.json`另存三點離線構圖，未接正式入口。`ground-visual-review-v1.json`保存三張離線目視與來源hash，原生公式及區外零差；動態接合另驗。`ground-ready-review-v1.json`保存37內部pass及隔離來源審查；`ground-frontend-build-v1.json`保存49330d68隔離前端、vet與build，正式62來源不變，`ground-pair-review-v1.json`已補七點隔離正常原版配對、兩點地面合成及洞窟拒絕，完整狀態相同；正式接入與GUI仍待完成。既有候選四圖不變。

- 城鎮前景與角色的後續來源沿 `workplace/town-hd-v1/foreground-v1/`，入口 `decode_scene.py`／`scene-proof-v1.json`、`scene-proof-walk-v2.json`、`scene-contract-v3.json`及 `canonical-map-proof-v1.json`。四個原版點加八步行點，14336pixels全相同，restored重複，共11不同狀態；九個Muralla來源穩定，末步資源及C000已切城堡，整層退回。右行及正常Up另存 `scene-cave-v1/receipt.json`，`decode_scene_cave_v1.py`／`scene-proof-cave-v1.json`補第十個城鎮來源。`world_foreground.py`／`world-foreground-v1.json`保存215欄全鎮來源，兩批手繪圖、提示詞及權利分類在 `town-art-generation-v1.json`／`town-art-generation-v2.json`。`town-art-layout-v2.json`保存透明畫帶與六入口校正；`scene-prototype-v1/receipt.json`及 `visual-review.json`為離線10城鎮／3拒絕、角色與區外0及五圖目視。規格已READY，候選 `workplace/bin/zeliard-preview-scenery-v2`為0b4b6282，48Go／58來源；主入口仍為ed01c199，三路徑完整配對已驗，正常視窗待驗，見 [前景規格](docs/spec/007-town-scenery-hd.md)。

- 完整遠山背景集中在 `workplace/town-hd-v1/foreground-v1/`。IDA9.4原名／bytes／xref入口 `ida-foreground-evidence-v3.json`，原版正常重播 `normal-v1/receipt.json`，四個1MiB與舊原版相同；`decode_layers_v2.py`／`layer-proof-v2.json`及 `display-contract.json`證明A000背景快取、E000已畫圖塊、6位元圖像及D000遮罩，四場景下24列零差。v1直接讀地圖的城堡65pixels差異保留，改讀原版已畫快取，不調寬容忍。內建image_gen完整圖 `background-handpaint-v2.png`實際2126×740，提示詞／權利分類各存 `generation-prompt-v2.txt`／`generation-record-v2.json`。`candidate-identity-v1.json`固定ed01c199、54來源／10鍵序，兩項測試無skip、go vet／build通過，`pair-v1/receipt.json`正常9點完整1MiB相同。`gui-tools-v1/`／`gui-v1/`正常冷F7／走路／兩頁中文／結束退出0，200固定輸入與原件不變；獨立decoder用當次E000讀值重建遮罩，前景／聯集外0、原貌還原0、直接HD取樣相符。根端三張配對及三張actualX圖入口為 `root-pair-review-v1.json`／`root-gui-review-v1.json`；`audit.py`／`source-audit.json`與 `current-audit.json`整合收尾，原READY另存。完整遠山已接，但樹木建物角色本身未HD、跳躍／戰鬥未驗，見 [完整背景規格](docs/spec/006-town-background-mask.md)。

- 前輪54列城鎮遠景HD集中在 `workplace/town-hd-v1/`，原版入口 `prepare.py`／`source-inventory.json`／`layer-evidence.json`。內建image_gen圖面、完整實際提示詞與權利分類為 `skyline-handpaint-v1.png`／`generation-prompt.txt`／`generation-record.json`，只留本地。固定候選 `candidate-identity-v1.json`為28ea8cc、52來源／10既有鍵序；`test-v1.jsonl`兩項無skip、`pair-v1/receipt.json`9點完整1MiB相同，四點遠景各108862像素且區外0，兩頁中文保留。`gui-tools-v2/`／`gui-v2/`正常冷F7／走路／中文／結束退出0，184固定輸入、58按放確認、30圖目視；兩組暫停HD／側框切換、原貌還原與直接取樣均通過。`gui-v1/`保留腳本暫停點錯的exit1，沒有執行HD切換。根端另讀配對及實際視窗圖的入口為 `root-pair-review-v1.json`／`root-gui-review-v2.json`；`audit.py`／`source-audit.json`與 `current-audit.json`合併現況，READY規格另存 `ready-spec-v1.md`。該輪只驗上方54列，水平切口已由後續foreground-v1完整背景層修正；前景、角色、其他城鎮與跳躍／戰鬥仍待完成或驗證；見 [遠景規格](docs/spec/005-town-skyline-hd.md)。

- 居民中文及控制碼集中在 `workplace/prose-v1/`，原版入口 `consumer-evidence.json`、`ida-town-evidence-v2.json`、`world-script-inventory.json`。最新 `candidate-identity-v2.json`固定08b3309、50來源／10既有鍵序及town鍵序；`pair-v5/receipt.json`／`pair-test-v5.jsonl`驗9點完整1MiB、30捲動子步及3種原版寫入前呈現，go vet／build-v3通過。`gui-tools-v3/`／`gui-v3/`正常F7兩頁及結束，163輸入／58按放確認／24圖目視，20區域、頁間區外與還原0差；固定參考palette的限制保留。`root-gui-review-v3.json`根端另讀三張actualX圖，`current-audit.json`為整合入口。原06b候選及50來源存於 `candidate-v1-sources/`，GUI1導航與GUI2捲動退回失敗保留。`navigation-v1/verification.json`及兩份相位／圓角比較解釋導航及不同camera差異；`mutation-baseline.jsonl`證實6513中途退回真因。`mixed-control-notes.json`保存5混合入口與6種handler原始定位，不進runtime。113居民中112仍未正常走過，見 [居民規格](docs/spec/004-town-dialogue.md)。
- 其餘場景名稱的共用消費端與來源研究集中在 `workplace/scene-v1/`，入口為 `prepare.py`、`prepare_observer.py`、`analyze_v2.py`與 `consumer-evidence-v2.json`。沿既有國王／武器及長老正常路徑補header證據，21個完成header及10點64KiB配對相同，不重新解交易或存檔規則。44筆擴充候選及字形推導在 `source_contract.py`／`source-contract.json`；新層須當次入口唯一驗來源，不以字面直接覆繪。完整配對入口 `prepare_pair_commands.py`／`pair-commands.json`／`pair-test-v1.jsonl`；正常視窗首輪工具／部分結果在 `gui-tools-v1/`／`gui-v1/`，Marid補驗在 `gui-tools-v2/`／`gui-v2/`。未走過名稱仍保留候選等級。
- 後續批次中文化的字面盤點沿同研究區 `body_candidates.py`／`body-text-candidates.json`及摘要。只盤點固定SAR可讀ASCII，不把切分點當成已驗段落、控制碼或消費端；原文與研究產物仍不加入Git。
- 前輪場景擴充固定候選SHA-256為 `f42ab9da5260cf213d91a210df54f64fe9b6dd5b3c9d38c1136218d52190899e`，別名 `workplace/bin/zeliard-preview-scene-v1`。身份在 `workplace/scene-v1/candidate-identity-v1.json`，47來源／10既有鍵序及2scene鍵序、51／48／53 catalog；建立時GUI待驗狀態保留。兩份測試紀錄六頂層／九子項PASS，13點新路徑及96點回歸通過，go vet／build-v1通過。`freeze.py`／`audit.py`及 `current-audit.json`合併兩份GUI、原件、舊失敗、補證與主入口。`gui-v2/`正常F7補Marid，139固定輸入、52筆按放確認、持鍵空、退出0；根端四種名稱實圖目視在 `root-gui-review-v2.json`。HD關閉、不交易，未重驗其他人物HD或正式封包。
- 場景GUI v1保留部分結果：國王、武器店及小屋已達，長老前停在城鎮而權威exit1。原版門口補證入口同研究區 `navigation_prepare.py`／`navigation-v2/`；既有成功視窗的 `workplace/gui_sage_hd.py:329-362`使用房屋外觀對齊後Up，不能把盲步數累積當成已對齊。actualX的before／ack是觀察區間，不冒稱精確guestedge。
- 狀態列與場景名稱研究集中在 `workplace/hud-v1/`，入口為 `prepare.py`、`prepare_v2.py`、`analyze.py`與 `source-evidence.json`。兩份原版觀察各48點色號、色盤、64KiB段、FileOps與完整收據相同，見 `observation-comparison-v2.json`；九筆來源／23次完成字形已驗。IDA9.4匯出入口 `ida_export.py`／`ida-evidence.json`，輸入為冷啟動01B1段，原名及位址保留；初次編碼失敗及來源斷言錯誤有獨立歷史檔，不視為產品缺陷。首版配對及46來源身份保留於 `pair-test-v1.jsonl`／`pair-v1/`與 `pair-source-v1.json`，最終驗證依下一筆固定候選。
- 原版洞窟名稱勘誤：先前由畫面讀為Cavern of Ialicia，SAR3原始字串與實際22CD來源逐byte證實為Cavern of Malicia。後續以Malicia為準，歷史收據不改名或覆寫。
- 共用字形核對在 `workplace/hud-v1/font_primitive_proof.py`／`font-primitive-proof.json`：01B1:F504指向F856，768-byte執行期字型以IDA2345指令重建24個完成header，全部原生像素零差，48個快照字型相同。原檔內尚未找到同一連續字型，來源producer仍未知，不開硬體driver研究。後續完整場景名稱候選清單由同研究區 `scene_candidates.py`／`scene-candidates.json`保存，只是header形狀與字面候選，消費端未驗者不得直接接入。
- 前輪狀態列固定候選SHA-256為 `3aeca36131bb9ca2192a9d9225ee8593c5f57caed5b5dce80948605a2be32206`，別名 `workplace/bin/zeliard-preview-hud-v1`。同研究區 `candidate-identity-v1.json`固定46來源／10鍵序、51對話／48標籤／9狀態列，保留建立時GUI待驗狀態；最終結果由 `current-audit.json`合併。`pair-test-v2.jsonl`及 `pair-v2/`含48點新層與48點舊使用回歸，三個頂層及七個catalog子項通過。`freeze.py`保留固定流程，`audit.py`／`source-audit.json`核對來源。正常視窗 `gui-v1/`同一handle完成冷F7、買50、物品頁開關、入洞及返回950，持鍵空、capture-tail退出0；根端三圖目視另記 `root-gui-review-v1.json`。47個原版及40個中文區域零差，8個HUD帶安全區外0；保留paragraph18／kind2／first_difference422舊對話安全退回，不稱全程零失配。
- 道具使用與物品頁文字的本輪證據、窄唯讀工具與正常鍵序集中在 `workplace/item-use-v1/`。入口為同目錄 `prepare.py`、`interactive.py`及後續來源收據；先沿已驗首瓶購買至物品頁，不重解買賣。中文手冊 `workplace/save-path-manual/珍084-神劍封魔/SCAN0357_008.jpg`印刷11至12頁已人工核對，道具在裝備頁選取後按Space使用；效果依固定原版實跑，不由手冊推定。
- 前輪洞窟使用固定候選SHA-256為 `b8c1be290043ac525adfbd7687783b07065b8142f8daa755f76d2d9cf2c808c3`，別名 `workplace/bin/zeliard-preview-item-use-v2`。身份 `workplace/item-use-v1/candidate-identity-v2.json`，48點完整配對在 `pair-v4/item-use/receipt.json`與 `pair-test-v4.jsonl`，43來源／10鍵序、51對話／48標籤；go vet／build-v2與五項 `regression-v3.jsonl`通過。正常視窗為 `gui-tools-v3/`及 `gui-v3/`，根端另目視三圖記 `root-gui-review-v3.json`。收尾 `audit.py`／`current-audit.json`合併配對、回歸、GUI3、兩次歷史失敗及主入口；原件與固定輸入不變，擁有權異常0。
- 使用GUI v1／v2保留各自權威exit1與部分收據；同一研究區下的 `navigation-v3/receipt.json`以原版持鍵重現首次失敗角色區域RGB零差，辨完整走路相位與暫態，再正常Up入洞。v2只增相位仍失敗，後續放開補證與 `gui-tools-v3/`、`gui-v3/`改核對先放開再擷取的正常次序，狀態依目前表及WORKLOG，不將準備入口當作通過。`pair-visual-review-v4.json`記根端三張配對圖目視；`next-hud-candidates.json`保存LIFE／PLACE／GOLD／ALMAS及Muralla Town的字面位置，消費端未知，沒有翻譯或掛鉤結論；不得據此直接接正式顯示。
- 本輪道具店其餘分支原版路徑與唯讀文字證據在 `workplace/item-branches-v1/`；HD原圖分組、內建imagegen提示詞、素材及驗證在 `workplace/item-hd-v1/`。固定建置 `workplace/bin/zeliard-preview-item-hd-service-v1` SHA-256為 `881ef8dd27ec6f28ee880e27f557c135f1d40531ea43be589103a64d29951eef`，身份 `candidate-identity-v1.json`，來源稽核 `source-audit.json`；正常視窗工具／輸出為 `gui-tools-v1/`、`gui-v1/`。兩者只作本地研究及中間產物，原件與存檔來源唯讀。
- 捲動清單修正的窄顯示證據、配對及候選保存在 `workplace/item-hd-v1/scroll-fix-v1/`；後續短GUI工具／輸出為同上層 `gui-tools-v2/`、`gui-v2/`。先用既有原版bitmap及固定viewport查明顯示對應，不猜刷新consumer；不重新探索六頁、購買或出售來源。
- 捲動修正候選 `workplace/bin/zeliard-preview-item-hd-service-v2` SHA-256為 `70a3f8e33b5be29d9f710b26eba6f72d300828bbab0c876b0f70f45d3bb7d35b`；身份 `workplace/item-hd-v1/scroll-fix-v1/candidate-identity-v2.json`，來源稽核 `source-audit-v2.json`，配對 `pair-v2/`與 `pair-test-v2.jsonl`。43來源／9鍵序及51／43標籤已固定。與881只差清單刷新及其呼叫／測試四檔，HD與TSV不變；短GUI2不用換名為交易／HD重驗。
- 本輪收尾入口為 `workplace/item-hd-v1/scroll-fix-v1/closeout_v2.py`及 `current-audit-v2.json`，合併來源稽核、主入口身份與兩份不同範圍GUI。`gui-v1/`保留六頁／買賣／HD及英文捲動限制；`gui-v2/`只驗修正後中文捲動、油品兩頁與離店1000。共39張目視、254筆按放，HD只實遇0／1／3，不冒充全四態或完整中文化。
- 銀行HD舊候選 `workplace/bin/zeliard-preview-bank-hd-v1`，SHA-256 `c74e94846937224ed5a116bc107814c0adea876a03191278229e7559be1b5311`；38來源身份、三畫格映射與完整提示詞在 `workplace/bank-hd-v1/`。首輪GUI失敗及補證保留，後續已由dc2696ed整合並通過 `gui-v2/`，不重新開啟舊候選驗證。
- 道具店原版證據入口為 `workplace/item-path-v1/receipt.json`、`source-evidence.json`及 `visual-review.json`。第一瓶Ken’ko Potion選取即扣款1000→950，後接詢問是否繼續購物；離店後物品頁可見藥水圖示。原版32點唯讀重播相同；對應中文及GUI已由dc2696ed完成，不重新探索購買來源。
- 道具店中文準備區為 `workplace/item-overlay-v1/`，保存可撤回適配稿及後續配對收據；銀行HD視窗驗證期間不改動其38個固定來源。
- 銀行HD視窗v1權威終態為exit1，保留 `workplace/bank-hd-v1/gui-v1/failure.json`與 `partial-receipt.json`。迎客pose2切換通過，未交易告別使用另一段原文，舊腳本誤等交易後參考而未完成離店。新分支原版窄證據入口為同目錄 `no-purchase-source-v1/`，補來源後才重驗，不將v1換名為成功。
- 前輪整合建置 `workplace/bin/zeliard-preview-bank-hd-item-v2` SHA-256 `dc2696ed47cf8d3f7d3a2a199f7408fc89fdf00943d298f216cb406ba24c7356`，來源固定於 `workplace/item-overlay-v1/candidate-identity-v2.json`。八個頂層回歸及十一個catalog子項通過，41來源、36對話、43標籤保留當輪身份。銀行HD視窗 `workplace/bank-hd-v1/gui-v2/`與道具店中文視窗 `workplace/item-overlay-v1/gui-v2/`均已通過，主入口已同步；新候選狀態依目前狀態表。
- 本輪稽核入口為 `workplace/item-overlay-v1/audit.py`與 `source-audit.json`，核對自有檔案邊界、原件、固定快照、候選來源、配對、文件連結及擁有權；實際視窗收尾另合併收據。
- 本輪完整整合稽核為 `workplace/item-overlay-v1/current-audit.json`，同一 `audit.py`以 `FINAL_GUI=1`核對銀行／道具店兩份實際視窗、數字／圖示、HD區域及主入口身份；提交後另核對來源符合功能commit。
- 稽核腳本契約更正保存於 `workplace/item-overlay-v1/audit-failure-record.json`；source_digest與manifest JSON檔案雜湊分開，離店／關閉依命名檢查點核對。整合原版狀態與六條配對已通過，不將腳本錯誤列為產品缺陷。
- 道具店視窗v1在門口定位停止，權威exit1、未進店或購買，原件與固定來源不變。實際按鍵步數使門中心由128跳至144，超出腳本範圍；保留 `workplace/item-overlay-v1/gui-v1/failure.json`、`partial-receipt.json`及 `visual-review.json`。最小原版導航補證在同目錄 `navigation-v2/`，先確認144位置能否正常進店，再修驗證腳本，不改遊戲碰撞或候選程式。
- 原版已證實門中心144可正常Up進MAGIC，完整indexed與既有初訪相同、色盤相同，收據 `workplace/item-overlay-v1/navigation-v2/receipt.json`。新道具店視窗工具／輸出為同目錄 `gui-tools-v2/`、`gui-v2/`，只修導航契約，沿同dc2696ed候選驗證；舊失敗保留。
- 本輪兩份GUI已通過並同步主入口dc2696ed。銀行 `workplace/bank-hd-v1/gui-v2/receipt.json`與 `audit.json`驗迎客pose2／告別pose1、區外0及完整還原，60筆按放、退出0、離店1000。道具店 `workplace/item-overlay-v1/gui-v2/receipt.json`與 `audit.json`驗首瓶50／續購否／離店950／物品頁及關閉，88筆按放、11圖目視、原版數字與六個ROI相同、退出0。兩者原件及固定輸入不變；保留各一筆離店舊文字安全退回，不稱零失配。
- 本輪銀行原版入口為 `workplace/bank-path-v1/receipt.json`，文字、數字及手冊映射為同目錄 `source-evidence.json`。銀行標題的原始推論保留，最新底色5／字色9／水平陰影45及位置98的勘誤見 `title-v3/palette-geometry-evidence.json`與 `workplace/bank-overlay-v1/bank-title-style-analysis.json`。
- 銀行中文固定建置 `workplace/bin/zeliard-preview-bank-v4`，SHA-256 `fc7972373703d0124182e2f24331bf3cdba9913ff2364fdb239f903db816837a`，主入口已同步。來源身份為 `workplace/bank-overlay-v1/candidate-identity-v4.json`，28筆對話及39筆標籤。v4銀行35點、catalog與裝備11點配對通過，測試在同目錄 `test-v4.jsonl`及 `v4/`；來源稽核為 `source-audit.json`，目前整合稽核為 `current-audit.json`。v2／v3的標題缺漏失敗保留，不冒稱通過。
- 銀行正常X視窗入口為 `workplace/bank-gui-v1/receipt.json`、`visual-review.json`及 `audit.json`，工具與固定身份在 `workplace/bank-gui-v1-tools/`。同一handle完成冷F7、存100、領50、查50、無alma及離店950；128筆按放確認、最長0.34秒，8個金額欄及3個GOLD區域與原版RGB相同。最後持鍵空、銀行覆繪inactive、capture-tail退出0，未開HD，未另驗Ctrl+Q或銀行ledger新存檔恢復。

- dosgolem 來源 HEAD：`3e17c1eb3f3d58935b1e8a59b978b8c899f4a653`，來源工作樹有其他工作的未提交修改。
- 本專案固定快照：`workplace/dosgolem-source/`，882 個檔案；來源摘要欄位 `source_digest`為 `2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`。`workplace/dosgolem-source-manifest.json`檔案本身SHA-256為 `9f7211ce703d40b27c881651fb3077f10d2649c7b69685ce83c2752d7f09b8aa`。HEAD不能單獨識別這份工具。
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
- 前輪HD與首章稽核：`workplace/current-audit-hd.json`、`current-audit-scroll.json`，保留各自來源與二進位身份。提示、長老／商品、彈窗、報價、裝備、商人及銀行輪另保留 `current-audit-ui.json`、`current-audit-weapon-sage.json`、`modal-audit-v1/current-audit.json`、`quote-audit-v1/current-audit.json`、`inventory-audit-v1/current-audit.json`、`merchant-hd-v1/current-audit.json`與 `bank-overlay-v1/current-audit.json`；目前整合稽核為 `workplace/item-use-v1/current-audit.json`。
- 同狀態區域比較：`workplace/gui-smoke-v4/intro-region-comparison.json`。原版色號完全相同，覆繪區內 3,655 個像素不同，區外 0。
- 初始化工具身份：`workplace/gui-smoke-v4/provenance.json`，記錄當時 13 個來源檔、二進位、dosgolem 快照及映像 ID。該份舊收據沒有 Git 身份，現行程式以本儲存庫 HEAD 為準。
- 最新前端二進位 SHA-256：`b8c1be290043ac525adfbd7687783b07065b8142f8daa755f76d2d9cf2c808c3`，主入口與 `zeliard-preview-item-use-v2`相同。新增洞窟藥水使用中文與不用圖示，未稱正式版本；70a3f8e3及其他別名保留歷史身份。
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
- 武器商人第一訪問六個原版畫格的初始盤點保留 `workplace/inventory-source-v1/merchant-pose-candidates.json`。本輪已增未購物第三姿態並驗HD，現行入口為前述 `merchant-hd-v1/source-mapping-v2.json`，不把初始兩組資料外推為全部動畫。
- 武器商人手繪圖位於 `workplace/merchant-hd-v1/`，三圖及完整內建imagegen提示詞、構圖參考、來源雜湊與本地權利分類各自保存。已接入並通過三姿態GUI；原版衍生圖不入Git。

- 本輪Muralla前景HD正常視窗工具／輸出為 `workplace/town-hd-v1/foreground-v1/gui-scenery-tools-v1/`及 `gui-scenery-v1/`。先沿可靠冷F7／居民兩頁與暫停切換，使用獨立原版角色圖樣核對保留；三段招牌／洞口另以已驗導航短輪補證。候選與結果依各自固定身份及權威終態，準備入口不代表通過。

- 前景主鏈首輪 `gui-scenery-v1/`保留暫停觀察契約失敗；獨立 `workplace/town-hd-v1/foreground-v1/gui-scenery-tools-v2/`與 `gui-scenery-v2/`只修兩筆allow_paused，來源候選／正常鍵序／超時不變，後續以各權威收據判定。

- 居民任意Draw點DS契約修正候選cf763f7固定後，可靠前景主鏈重驗在 `workplace/town-hd-v1/foreground-v1/gui-scenery-tools-v3/`及 `gui-scenery-v3/`；只換候選／固定來源／READY快照，原鍵序不變。前兩輪失敗保留；三段店面與洞口另短輪依正常入口判圖。

- 前景後續短視窗工具與輸出入口為 `workplace/town-hd-v1/foreground-v1/gui-scenery-nav-tools-v1/`、`gui-scenery-nav-v1/`。用途為正常冷F7後分段行走、三段店面招牌與右洞口跨場；使用固定cf763候選，不取代主鏈GUI3或歷史失敗收據。

- 完整城鎮畫格的下一輪正常視窗準備入口為 `workplace/town-hd-v1/foreground-v1/gui-coherent-tools-v1/`、`gui-coherent-v1/`與 `gui-coherent-nav-tools-v1/`、`gui-coherent-nav-v1/`。前者驗居民主鏈與完成來源切換，後者驗行走、原版繪製中途暫停、三招牌及洞口；候選尚未固定，未啟動遊戲。呈現來源與live原版色號分開記錄，沿規格008，不改主狀態表。

- 重讀來源修正版的正常視窗工具入口為 `workplace/town-hd-v1/foreground-v1/gui-coherent-nav-tools-v2/`、`gui-coherent-nav-v2/`及 `gui-coherent-tools-v2/`、`gui-coherent-v2/`。固定def76c62候選，優先三招牌／正常第二F7／洞口，再居民兩頁；只準備及凍結，pair3權威終態放行前不啟動。洞窟末態應撤除Town完成來源，不要求presented副檔名，舊兩輪身份保留。
