# 1.0.8插畫重畫版交付、推廣片與結案清理

狀態：CONFORMED，2026-10-23。使用者在068推送`d8c54a0`後指示：錄製推廣影片、打包三平台完整版號，打包完清理工作區節省空間，這樣就算結案。

## 範圍

- 版號`v.1.0.8-20261023`：同日1.0.7之後的重新發行，依版號規則提高修訂號。
- 來源為`d8c54a0`的工作樹，HD圖庫為068登錄後的`complete-art.json` `b4103bf6…`；字模沿1.0.6的`workplace/release-v106/rounded-font`。CHT／CHT-HD各三平台：Linux AppImage、Windows x86_64 ZIP、macOS Intel／Apple Silicon通用ZIP。
- 唯一工作區`workplace/release-v108/`：`package/`為打包，`edit/`為剪輯。正式產物只放`dist-all/v.1.0.8-20261023/`的`full-local/`、`promo/`、`smoke/`與`SHA256SUMS.json`。
- 不建tag或GitHub Release，不公開上傳；含原版資料、音樂與衍生美術，只留本機。
- 不改068「未處理」所列的前端行為（捲字層蓋過剪影畫格、溶解沿用誤判），照實寫入README與交付紀錄。

## 打包與驗收

- 沿[build_dual_release.sh](../../tools/build_dual_release.sh) prepare／compile／package，`ZELIARD_RELEASE_WORK=workplace/release-v108/package`、`ZELIARD_RELEASE_FONT=/project/workplace/release-v106/rounded-font`。[package_dual_release.py](../../tools/package_dual_release.py)的包內README加1.0.8說明，CHT版註明遊戲內容與1.0.7相同。
- [audit_dual_release.py](../../tools/audit_dual_release.py)新增`--picture-repack`：六包內容、架構、逐檔manifest與SHA；另核包內README版號、結局譯表參數、字模SHA，HD包的`complete-art.json`與identity相同且含`story-dmaou`、`story-new1`、`story-sei`、`story-yuu3-dark`與兩組說話補丁，CHT包不含`assets/manual-hd/`。
- [platform_smoke.py](../../tools/platform_smoke.py)：Linux兩版原生、Windows兩版Wine輔助啟動，Windows另核跨磁碟啟動；macOS只核通用架構與結構。
- `package_dual_release.py finish`寫入六包與影片的`SHA256SUMS.json`。

## 推廣片

- 60秒，1280×720／30fps，沿065的字卡風格。配樂是原版MSCAdLib標題曲經Nuked-OPL3的錄音（`workplace/music-release-v1/title-music/source.wav` `44699af8…`），自第8秒取28秒循環並調整響度，片中明列非同期聲。
- 畫面是1.0.8正式HD執行檔經探針合成路徑存下的截圖，片中每張字卡標示「1.0.8正式執行檔合成畫面」。[make_picture_promo.py](../../tools/make_picture_promo.py)先核探針基底的執行檔與`workplace/release-v108/package/linux/zeliard-preview`同SHA，再把用到的畫格連同SHA複製到`workplace/release-v108/edit/frames/`。
- 來源三條探針：冷開機不按鍵到魔王宣言（`rel-boot-threat`）、片頭故事（`rel-story-title`）、結局直接觸發（`rel-ending-poke`，`poke 01B1:0049=1`）。結局字卡另標「結局以探針結構觸發」，不冒稱通關或遊玩實錄。
- 字卡背景只取插畫範圍（原生第8到144列），不帶入下方故事字列與版權字。
- [verify_promo.py](../../tools/verify_promo.py)核音畫、音量、靜音、黑幀與凍結；靜態字卡都列為預定靜止。分鏡、來源畫格SHA與音樂身份在同版`promo/provenance.json`。

原定以正式HD AppImage的實際X11視窗錄製片頭。實錄時主機平均負載約45（14核，來自其他專案），工具列顯示的實際CPU倍率只有0.083，片頭在錄影時間內走不完，所以中止，未留任何錄影；使用者選擇改用正式執行檔的合成畫面。

## 驗證結果

| 項目 | 結果 |
|---|---|
| 出貨來源 | 與工作樹逐檔相同（cmd、internal、text、third_party、go.mod、go.sum、VERSION、LICENSE）；Linux二進位`1424b1201942397e68ad907987b2af0c6095f7155935c9c3b65a7d04046f4559`，`complete-art.json` `b4103bf6…ef41` |
| 平台啟動 | Linux兩包原生、Windows兩包Wine輔助（含跨磁碟）均為`PASS_NATIVE_PACKAGE_LAUNCH`；macOS只核通用架構與.app結構 |
| 六包稽核 | `audit_dual_release.py --picture-repack`為`PASS_LOCAL_DUAL_EDITION_PICTURE_RELEASE`；清理後重解AppImage再跑，`release-audit.json`與清理前逐位元相同 |
| 正式執行檔探針 | 三條`rel-*`與068開發執行檔同步數的21張截圖逐位元相同；中文覆繪的英文先現、英文未覆蓋、與英文同框均為0 |
| 推廣片 | `PASS_PROMO_TECHNICAL_QA`，mean -18.6／peak -4.5dB，無長黑幀或靜音，凍結區間都是預定靜止；字卡與聯絡表已目視 |

| 產物 | SHA-256 |
|---|---|
| `Zeliard-v.1.0.8-20261023-CHT-Linux-x86_64.AppImage` | `d059afd55b07847123cda74c1a9f00ce4880a2d005a58d695aed937d684ab3d2` |
| `Zeliard-v.1.0.8-20261023-CHT-HD-Linux-x86_64.AppImage` | `e3edc667b9fb4aa3ad0a47b1f7b0194995c0d74707137d65fdd844ca9b5c1c6b` |
| `Zeliard-v.1.0.8-20261023-CHT-Windows-x86_64.zip` | `cf985a55e14655a2bb87ef5ee9f3bc148f3c272b7207842f5529931c84bd0bb2` |
| `Zeliard-v.1.0.8-20261023-CHT-HD-Windows-x86_64.zip` | `6e80c2d75e9a42d10005545280868c09d89e4078a7ff66639458db6d5e61354d` |
| `Zeliard-v.1.0.8-20261023-CHT-macOS-universal.zip` | `0b926d9b66b1a4eb5b629b77423c962cc000f7e5e834e81d86c12011a819ffa9` |
| `Zeliard-v.1.0.8-20261023-CHT-HD-macOS-universal.zip` | `e12215516ffa2179e420515f34a9ec58cf40eb6a3700c8304be607f274a702c5` |
| `promo/Zeliard-v.1.0.8-20261023-promo.mp4` | `9b69f5144b44009c052d12d098de175fe122935ccedc90e9ae51a3cce8328f2d` |

限制：音訊關閉的啟動檢查；不是Windows／macOS真機、實體手把、完整通關或實際視窗遊玩錄影。068「未處理」的前端行為照舊。

## 結案清理

[cleanup_picture_release.py](../../tools/cleanup_picture_release.py)先核`SHA256SUMS.json`七份產物，再刪除明列範圍；設`ZELIARD_CLEANUP_DRY=1`只列清單不刪。效能基準`BenchmarkPictureCoverageDraw`讀的開發截圖先複製成`workplace/picture-coverage-v1/evidence/g-story-long_probe.shot-468050000.indexed`，測試路徑同步改過。

清除範圍：

- 1.0.6、1.0.7、1.0.8打包的`data-*`、`AppDir-*`與`Zeliard-*-Windows`、`Zeliard-*-macOS`組裝層。
- 同三版`smoke/`下的AppImage解壓層、Windows解壓層與Wine prefix。
- 探針基底`workplace/release-v106/probe-base/`與`workplace/picture-coverage-v1/probe-base/`。
- 1.0.6、1.0.7與068探針輸出的截圖與記憶體快照（`.png`、`.indexed`、`.mem`），jsonl、日誌與探針存檔保留；1.0.8推廣片來源`rel-*`整組保留。
- `workplace/gocache`的內容，目錄保留給建置腳本。

保留：六包與推廣片、各版正式交付與啟動收據、來源副本、獨立dosgolem、各平台二進位、identity.json、推廣片畫格、068實機截圖證據、唯一生成圖、參考圖、提示詞、字模、原版與存檔。

逐檔配置bytes合計19,670,798,336，硬連結的檔案在兩個基底間重複計入。第一次執行刪完組裝層與解壓層（10,112,786,432）後，停在凍結成唯讀的探針基底；腳本改成只在要刪的目錄樹內恢復目錄寫入權限，第二次刪除其餘9,558,011,904。清單在`workplace/release-v108/cleanup-dry-run.txt`，第二次的收據在同目錄`cleanup.json`。主機可用空間由90G增為108G。沒有root擁有的殘留，沒有本專案容器在執行。

清理後要重跑探針，先以`--appimage-extract`把正式AppImage解到`workplace/picture-coverage-v1/probe-base/<名稱>`再用`probe/docker.sh`；`mkbase.sh`依賴的1.0.7基底已刪除。`make_picture_promo.py`會核對探針基底的執行檔，重跑前須先以同法重建`probe-base/v108-rel-CHT-HD`；三條`rel-*`輸出仍在。
