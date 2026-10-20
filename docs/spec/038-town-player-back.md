# 城鎮主角背向姿態

狀態：READY。2026-10-21核對正常原版來源與獨立像素重建後接入。沿全遊戲HD目標補城鎮與城堡的主角phase4；不改輸入、方向、地圖、規則或存檔。

## 原版來源

- 既有正常Muralla快照`workplace/music-release-v1/release-build/data/assets/town/restored.memory`的SHA-256為`9dd3657107f25c097debf5f6be68137d055d98bda2e22a59f6b854d88be5fdd9`。TMAN全部46格2208bytes及368遮罩bytes已由034的原版來源核對，雜湊分別為`045d4246ae0be72df3ac0201bc9ac8e5e1b50829783d6c79727fce7af0ef66c7`及`15f70e4b15a5887babfdf1c9154d788e6a37cd2e6a196e24cece90b1d6ac98d0`。
- 01B1:6A59與6A3B的五組主角表各六個零基圖格ID。phase4使用6A71與6A53，兩者皆為20、22、24、21、23、25，來源在11B1:6000起，遮罩在21B1:8000起。這些位址是dosgolem段內偏移，不和SAR檔案偏移混用。
- 兩個朝向重建的16×24 RGBA完全相同，SHA-256為`70ca7bed8492e9f7c4a49a951ba602782c57cc94ba62d3b07ee4b1ddd5fbf2d8`。完整圖是背向鏡頭，非另一個側向行走姿態。原圖、對照及輸入雜湊在`workplace/full-hd-v1/town-neutral-source-v1/receipt.json`。
- 舊TownScenery雖接受phase0..4，實際正常進門畫格仍因多畫居民而失配。TownPlayer.prepare也只接受0..3。不能把接受姿態編號當成完整畫格已驗。
- 已證實：正常冷新局到城堡，從兩方向走到主門再按Up，可在01B1:68BA讀到phase4。來源分別為`town-neutral-normal-v3/observed-back-bit0.memory`的`50ba7bef235d3c9ac684feffd86b2b318a9e9f43b9b45f1bdeb8184d9e19be0d`，及`town-neutral-normal-v4/observed-back-bit1.memory`的`368b1ce2e672bfe5c48c88537401c3fbe5b3917c8a8216c60659cb5379ad9d00`。兩次觀察器的完整RAM、畫面、色盤、暫存器、步數、cycles、檔案操作及鍵盤狀態均與另一原版相同。
- 已證實：這兩張進門來源的四筆居民記錄仍存在，但原版地圖該欄第5列已由253還原為記錄的底圖值。IDA9.4的sub_6C4E於6C5A讀[SI+3]，6C67以C01C加上欄號乘8，6C6B以MOV [BX],AL還原。工具與位址空間及原始bytes在`town-back-ida-v3/evidence.json`，來源索引見[town-back-source-index.json](../town-back-source-index.json)。
- 獨立Python重建底圖與背向主角，兩張各14336色號像素與原版完全相同，不含居民。證據為`town-back-source-audit-v1/native-check.json`。這只證明已觀察的正常進門形狀，不推論其他phase4居民配置。

## 美術準備

使用既有主角側向手繪圖作身份與材質參考，原版phase4作背向姿態參考。保留雙角頭盔、橘紅長髮、紅色軀幹、藍色護甲、金色邊與棕色靴，雙手空、雙腳同一落腳線。以內建imagegen產生一張背面圖，兩個朝向共用，不鏡射或另畫成面向鏡頭。完整提示詞及圖面在`town-neutral-art-v1/generation.json`，完整素材只留本機。

## 接入與退回

- `-town-player-back-art`選擇註冊目錄，需先有城鎮主角。`town-player-back-v1.png`固定48×72及SHA-256 `d84525e69589decf0045cc32b35032eaea862c967e9876df5294cf208d3096f4`。兩朝向槽4及9原子共用同一張圖。缺檔、雜湊或尺寸不符時不變更既有八姿態，也不開放phase4。
- phase0..3的居民與主角路徑保留。phase4只接受全部居民標記已還原成各自底圖的已驗形狀，居民不畫；任一不符即拒絕。底圖還原仍使用完整記錄，不刪改來源。
- 每張仍檢查原有程式、資源、地圖、色盤與14336原生像素。只新增背面圖及已驗居民呈現條件，不縮小像素守衛、不改原版資料。
- 未載入背面圖時，人物層沿舊行為退回；未知姿態仍拒絕。高清直接在3倍畫布合成，位置、16×24範圍及落腳線不改。

## 驗收

核對固定素材身份、兩正常來源完整像素、背向共同合成與居民不被重畫、來源及素材失配原子拒絕、既有行走與城堡／Muralla正常往返。另由實際視窗正常進門捕捉背向HD及原貌切回。未驗的城鎮與姿態不稱完成；存檔、輸入與原版狀態須保持相同。

2026-10-21來源、實作與正常原版回呼合成已通過，GUI仍未驗，保持READY。獨立原版28672像素完全相同，HD6912像素以每色2階容差核整數alpha取捨；完整原版狀態不使用此容差。兩條正常冷新局各捕捉一次背向完成回呼，原版狀態相同。三項新測試、十項回歸、19個正常往返點、vet及建置通過。GUI v1／v2／v3皆未捕捉短暫姿態，不冒稱PASS；相同時機不反覆重試，也不因擷取更改原版時序。

候選`town-back-runtime-v4/`、收尾`town-back-audit-v1.json`、獨立核對`town-back-source-audit-v2/`及正常配對`town-back-regression-v1/`均在`workplace/full-hd-v1/`。由[背向稽核](../../tools/hd_town_back_audit.py)重生到新檔。實際視窗是尚未完成的驗收項，其他城鎮及美術持續依034補齊。

全量目標見[034](034-full-game-hd.md)。此增量不處理其他城鎮、交談背景、敵人／首領或新封包，不將一個姿態當全人物完成。
