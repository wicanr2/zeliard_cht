# Muralla 地面手繪覆繪

狀態：CONFORMED，限固定版本Muralla已列正常路徑的可撤回三倍地面試作。其他地面、角色、正式倍率／預設與發行平台未在此驗收。

回城增量另列下一節，已通過限定驗收；不外推其他城鎮。

## 回城地面相位勘誤

增量狀態：CONFORMED，限候選0227c040的正常Up回城，索引`R009-ground-phase`。正常Up回城的camera179與地面相位不符合原先全域推論。`return-replay-v5/last-town.*`的主段、資源、extra、色號雜湊均與失敗GUI已呈現來源相同。上／下八列用原camera公式分別差970／926像素，但各自唯一相符的112像素循環位移皆為0。相同camera數字不能單獨決定地面相位。

修訂契約：仍先驗完整城鎮來源及前景，再各自搜尋上下八列的0..111位移。只有全部色號及RGB恰好符合一個位移才接受；零個或多個均整層退回。高清取樣直接使用該原生位移乘3，不改素材、原版畫面、碰撞或存檔。舊camera公式保留為已列行走路徑的證據，不再外推回城。

`ground-return-proof-v1.json`保存原版／GUI來源身份及獨立相位核對。隔離`ground-return-study-v1/`的39個pass事件無skip／fail，包含既有原貌／高清取樣、來源失配及新真實回城來源、四個邊界色號／RGB失配與區外不寫入；審查為`ready-ground-return-review-v1.json`。正式實作後仍須正常Up回城GUI及同來源高清取樣才可把增量標為CONFORMED。

接入後`current-pair-v7/`回城十點完整原版相同，`gui-current-return-v7/`正常Up回城退出0，洞窟撤除且城鎮前景／地面／主角恢復。`pixel-current-v9/receipt.json`按當次原版唯一相位及素材預乘alpha核對32256地面高清像素零差，actualX完整遊戲畫布與程式擷取相同。輸入、原件及自建存檔／mtime不變；收尾見`workplace/cave-hd-v1/runtime-index-v7.json`。

## 範圍與來源

只覆繪 Muralla 原生 `(48,142)-(272,158)` 地面畫帶。維持已選定的手繪奇幻，沿用可撤回的三倍顯示試作。正式倍率、預設與發行平台沒有在本規格定案。

- 原版 EXE SHA-256：`fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`。
- 原版來源、色號、色盤及逐點雜湊見 `workplace/town-hd-v1/foreground-v1/ground-proof-v1.json`。十個正常來源共35,840原生像素零差。
- 地面原版參考須為`foreground-v1/normal-v1/restored.indexed`的camera153完整64000-byte畫格，SHA-256為`dd777f4940d2ffe32d5fae6a8274c0f3a0878fb84e3f7fd3f425a4af92d6d224`。它同時可作`-town-sky-reference`：遠山實際矩形`(48,24)-(272,78)`與舊camera149參考逐byte相同，crop SHA-256為`22fb614b02b3d9ac436c85b7d56c62ce9004d1254f6c52160de2cf62edeb7cf9`。遠山crop相同不代表其他完整畫格可作地面參考。
- 已證實：原生每列112像素重複兩次。以camera153為參考，上8列每camera移8像素，下8列移16像素，超出以112取模。
- 已證實：IDA 9.4 的 `3628`／`36A4` 原始 `rep movsw` 分別搬移8／16像素。原名、bytes及工具輸入見 `scene-prototype-v1/coherent-frame-v1/ida-scroll-v1.json`。IDA線性0對應執行期01B1段偏移，不是EXE檔案偏移。
- 強推論：相同關係適用其他正常camera。每次覆繪仍須核對當次全部3,584原生色號與RGB，不能由camera數字單獨接受。
- 未知：首次建立地面參考畫帶的producer。使用固定原版參考及完整當次像素守衛，不為初始化流程另開逆向切片。

## 圖面與離線預覽

內建imagegen產物 `ground-handpaint-v1.png` 實際2172×724，SHA-256 `ea2b08001d76d6ed739e48a6ec6695db2bda05a4c843c9e573977b8ffcc45074`。完整提示詞、來源圖與權利分類保存在 `ground-prompt-v1.txt`、`ground-generation-record-v1.json`。這些圖面及原版衍生資料只留本機，不加入Git。

直接取原圖矩形 `(0,273)-(2172,467)` 到336×48高清紋理，保留原alpha，不先縮至原生大小。兩個24列畫帶分別按原版速度取樣，橫向以336取模。此裁切是顯示取樣契約，原圖不改寫。

三點離線構圖入口為 `scene-prototype-v1/coherent-frame-v1/ground_preview_v1.go`、`ground-preview-v1.json`及 `ground-visual-review-v1.json`。camera153、61、179的原生公式及高清區外差異均為0，目視底線齊平，沒有蓋住角色或狀態列。這不等於行走或實際視窗驗收；重複紋理與上下帶動態接合仍待確認。

## 接入契約

1. 獨立、可選素材入口。缺圖、損壞、尺寸或雜湊不符時保留原貌；不改無素材啟動方式。
2. 沿用規格007已驗的Muralla來源範圍與人物group上限，另驗規格008的捲動程式碼與固定色盤。
3. 驗全部原生地面色號及RGB後才寫畫布。短來源、未知場景、未知camera關係或任一像素不符時整層不寫入。
4. 有規格008完整快照時，與前景共用同一來源時間、camera、色號及RGB。原版 `.indexed` 繼續記當次原版；呈現來源時間另記。
5. 只寫三倍畫帶 `(144,426)-(816,474)` 加畫面起點。角色、碰撞、地形高度、動畫、原版時間、輸入與存檔都不改動。
6. 原貌／HD切換不額外執行原版。對話、F7及跨場沿既有快照撤除契約處理，不保留舊城鎮地面。

## READY 與驗收條件

先核對固定參考畫帶、圖面取樣、全部原生公式、上下帶邊界與失配退回，再審查到READY。未達READY不接正式入口。

接入後用既有正常冷F7、左右行走、居民對話、重讀與Up入洞鍵序，確認完整原版記憶體、CPU、時間、檔案及輸入保持相同。另驗缺圖、色盤、來源、短畫格、地面單像素與RGB差異的原子退回。實際X視窗至少看行走接縫、暫停切換／還原與跨場撤除，通過範圍才標CONFORMED。

## 證據審查

`coherent-frame-v1/ground-overlay-v1.json`只在編譯器測試環境加入`town_ground.go.draft`及`town_ground_test.go.draft`，沒有修改候選def76的62份來源或正式入口。權威45520退出0，`ground-draft-test-v1.jsonl`三頂層及34子項共37pass事件，無skip／fail。十五個正常原版來源包含十個地面接受點與五個對話／跨場拒絕；每個接受點驗完整高清直接取樣、源記憶體不變及區外0。十九種模式、來源長度、色盤、程式碼、地面色號／RGB、前景、人物group及畫布失配均原子退回；缺圖、損壞圖／參考與缺前景也拒絕。這些只證明離線試作契約，不當作GUI或完整遊戲一致。

接入採獨立`-town-ground-art`選項，搭配已驗前景來源及上述camera153的`-town-sky-reference`。不提供素材則維持原貌。每次繪製驗實際完整色盤、規格008來源以及規格007完整前景，再核對全部地面色號與RGB；有完整快照時全部共用快照來源。

正式來源候選`19ac2719`及64份完整來源由`coherent-frame-v1/ground-canonical-identity-v1.json`固定，原READY全文另存`ready-ground-spec-v2.md`。37內部pass、vet及build通過；兩條正常路徑16點完整1MiB、14欄Regs、原版色號／色盤、時間、檔案及輸入相同，五個完成點合成地面、洞窟拒絕，見`ground-canonical-pair-review-v1.json`。5157次呈現屬完整畫格層，不稱全部地面逐幀對拍。

實際X視窗入口為`foreground-v1/gui-ground-nav-v1/`及`gui-ground-v1/`。兩路唯一執行皆退出0，317加306共623項固定輸入、原件29檔、字型及存檔雜湊與mtime不變。正常三招牌、第二次F7、短鍵對齊及Up入Malicia、居民兩頁及關閉恢復通過；根端九張實際圖的目視另存`root-ground-nav-gui-review-v1.json`／`root-ground-main-gui-review-v1.json`。

兩份`visual-review.json`核對107張PNG，88張本輪逐張目視、19張以完整雜湊沿用已讀舊圖。兩份`ground-pixel-review.json`核對22個暫停點及一個最終完成來源，共82432原生RGB格及741888高清直接取樣像素零差；最終3584原生色號也相同。nav的19組聯集外差異為0，角色樣本共76770高清像素零差，其中17組同時覆蓋主角與居民；main兩組暫停9351及最終4653角色像素零差、26項中文與數字ROI及完整還原通過。302筆有限移動取樣中77筆保留較早完成畫格並匹配前景與遠山HD；這些移動樣本未保存地面旗標，地面驗證採上述暫停與最終來源取樣。不外推連續影片的每一幀。洞窟原版模組與可見名稱相符，城鎮地面撤除且沒有舊presented副檔名。184筆按下／放開確認，末持鍵空，正常capture-tail退出；沒有在本輪驗Ctrl＋Q退出路徑或新增存檔。
