# 全洞窟地形HD增量

狀態：DRAFT。承接[034](034-full-game-hd.md)，擴充第一洞窟以外的地形家族。沿既有手繪奇幻、3倍畫布與dosgolem原版覆繪，不改碰撞、地形、時間、玩法或存檔。

## 原版來源與研究入口

- 固定原版EXE與SAR雜湊沿CONTEXT及[覆蓋台帳](../hd-coverage.json)。SAR3記錄74..84的名稱候選為MPP1..MPPB；解壓身份已由84筆RLE原版常式對照固定。
- [地形來源工具](../../tools/hd_cave_sources.py)在Docker內逐組保留48-byte header，再將其後每48bytes用已驗三平面變換解出8×8原圖。不能將header當圖格，或把兩個圖格的相鄰bytes誤當寬圖。
- 第一洞窟正常來源為`workplace/prose-v1/progression-v1/dungeon-entry.memory`及同名palette，正常按鍵及原始存檔身份保存在該目錄receipt。MPP1的header應與11B1:8000相同，全部25圖格與11B1:8030相同；核對後才列為已證實。其餘十組在正常載入前只稱結構候選。
- 擷取輸出集中`workplace/full-hd-v1/cave-terrain-sources-v1/`，含逐格原圖、分組總覽及來源／工具／地址空間／SHA收據。原版與完整衍生圖均留本機，不入Git。

## 下一閘門

1. 核對11組原圖與實際圖格數，明列原生輪廓、結構、色盤及header用途的已知與未知。
2. 以既有正常存檔冷讀及原版按鍵尋找代表性家族載入，保存完整原版狀態與檔案讀取。不得注入位置或把直接呼叫載入常式當正常路徑。
3. 已查明來源的家族先產生完整手繪圖集，保留原版構圖與相鄰邊界。未知來源與未畫圖格不能靠換色、像素放大或安全退回計為HD完成。
4. 證據足以定義來源、圖格／圖面對應、遮擋、轉場生命週期及驗收後升READY，再擴充正式呈現。保留012完整32256原生像素重建與原子退回。
5. 各新增家族做代表性正常進出及原貌／HD切回；不要求完整通關，但資產覆蓋不得縮成第一洞窟。

本規格不授權尚未證實的格式或來源進入正式前端。其他敵人、首領、角色、武器、效果與介面沿034另外追蹤。

## 本輪來源結果

| 來源 | 已核結果 | 限制 |
|---|---|---|
| MPP1..MPPB | 圖格數依序25、34、43、21、52、36、61、43、38、20、5，共378格；前48bytes為header | MPP1全部header及25格與既有正常RAM相同；其餘十組正常載入未驗 |
| 31份地圖 | 64列、寬42..320欄，共306048格；四種RLE分支及最終SI均與原始6CED..6D6D常式相同 | 隔離常式驗格式，不當正常到達證據 |
| 圖集選擇 | 地圖尾端+2經7E9B複製至9EF7；7ECD以11-byte descriptor選9C43表的MPP資源 | 11項descriptor與SAR3記錄74..84一一對應；MP73使用MPPB，不能只按檔名推斷 |
| 正常來源 | 四份原版USR冷讀各三點，另SAVEM1三點，兩個獨立dosgolem完整狀態相同；全部仍是MPP1 | 存檔前史未知，未走到後續家族 |
| 原版地圖初始化 | 真實01B1:6CEC的E000環形地圖36×64共2304格與離線來源相同 | 後續門口、敵人與事件的寫入不混作壓縮地圖來源 |

[離線地圖工具](../../tools/hd_cave_maps.py)保存31份原始payload、逐欄資料與靜態研究圖；[原版常式對照](../../tools/hd_cave_map_oracle.go)保留原位址6CED..6D6D，執行真實資料，31份皆正常DOS退出0。這個退出是隔離常式的測試程式，不是遊戲通關或GUI驗收。

原始正常段65536bytes SHA-256為`12d91918212c4688a575bc31d395c133c7232f99ead769c7cfbc680333545c93`。IDA9.4地址基準為binary base0對應該段內偏移。[分級索引](../cave-terrain-source-index.json)保存6C98、6CFA、6D57、7E9B及7ECD；[匯出工具](../../tools/ida_cave_resource_evidence.py)自動合併索引，其他指令仍附未知警示，不用函式名當證據。

核對入口為`workplace/full-hd-v1/cave-source-stage-audit-v1.json`，由[來源稽核](../../tools/hd_cave_source_audit.py)重生至新檔。原圖為`cave-terrain-sources-v1/`，地圖為`cave-map-sources-v1/`，原版常式為`cave-map-oracle-v1/`，正常存檔調查為`cave-save-survey-v1/`及`cave-save-survey-v2/`，IDA匯出為`cave-map-ida-v2/`。均位於同一研究根。

## 冰窟手繪候選

已用內建image_gen依MPP4原圖、MP40靜態構圖及既有手繪材質產生2048×768、8欄3列的圖集。兩版與完整提示詞保存在`cave-ice-art-v1/generation.json`及`cave-ice-art-v2/generation.json`，輸入圖片與輸出SHA另存`cave-ice-art-v2/provenance-v2.json`；v2 SHA-256為`16fb57a71cf5ec0fe377afec7fc972e1d8b51ec4cdb83f6f52b05454e9ce21fc`。已目視冰壁、攀爬片、門框與兩個光球。

兩版都有真實alpha，空白格仍有殘留alpha；圖格位置、裁切、相鄰邊界與正常路徑尚未驗完。保留為候選，不接正式前端，不計HD完成。後續先依原生輪廓註冊前21格及核對接縫，再補其餘家族美術；31份地圖解壓成功不將本規格升為READY。
