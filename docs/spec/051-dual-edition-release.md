# 原貌中文與HD中文：三平台完整包及新版推廣片

狀態：限定CONFORMED。v.0.4.0-20261021六包、正常第一洞窟錄影及60秒推廣片已完成；平台身份與限制見VERIFICATION-MATRIX。使用者要求的「原版＋中文」及「原版＋中文＋HD」分開打包，不覆寫歷史發布。原版規則、存檔及輸入不改。

## 版本與資料

| 版本 | 內容與預設 |
|---|---|
| CHT | 原版遊戲、既有繁中、粉圓字型、原版AdLib與Nuked-OPL3音源、鍵盤與手把工具列。不附HD圖庫，預設原貌 |
| CHT-HD | 同上，附全部已接HD圖庫與050完整補圖，預設開啟HD，可切回原貌 |

兩版使用相同程式及相容存檔。作業系統的Zeliard-CHT設定目錄不變，安裝目錄保持唯讀。完整資料只放本機，不推送GitHub或公開服務。

## 交付與製作

輸出唯一入口為`dist-all/v.0.4.0-20261021/`，六個正式包平放`full-local/`，檔名附CHT或CHT-HD。Linux為x86_64 AppImage，Windows為x86_64 ZIP，macOS為Intel／Apple Silicon通用.app ZIP。SHA256SUMS.json記錄版本、內容、輸入、輸出雜湊與權利分類；smoke/保存封包啟動與結構收據，promo/保存新影片及來源。

工作區`workplace/release-v040/`只放固定來源、編譯、封包暫存、錄影原檔與剪輯中間產物。[build_dual_release.sh](../../tools/build_dual_release.sh)使用現有固定Docker工具鏈；[package_dual_release.py](../../tools/package_dual_release.py)按明列來源從乾淨目錄重建兩版。050生成原圖及來源唯讀使用，既有AppDir不改寫。字型沿已完成的粉圓2.1遮罩與原字距。

推廣片使用新版實際遊戲視窗，展示原貌中文、HD中文、代表對話、第一洞窟與工具列。音樂沿原版MSCADLIB事件及成熟Nuked-OPL3音源，不重新作曲。不以未實際到達的後段素材圖集冒充遊玩畫面。

正常錄影與模式切換由[record_dual_release.py](../../tools/record_dual_release.py)沿既有玩家鍵序擷取；[make_dual_promo.py](../../tools/make_dual_promo.py)合成新版對照與60秒影片，[verify_promo.py](../../tools/verify_promo.py)檢查訊號與抽幀。[audit_dual_release.py](../../tools/audit_dual_release.py)是六包、原版資料、模式、平台、影片與來源雜湊的整批收尾入口。

交付後由[cleanup_dual_release.py](../../tools/cleanup_dual_release.py)核正式檔案SHA再清理明列重複產物。保留原版、唯一美術、提示詞、錄影、源音訊、固定原始碼與收據。需要重跑內容稽核時，Linux解壓層可在容器暫存目錄重建，不依賴已清除的建置副本。

## 完成條件

- 六包各自驗證原版資料、中文檔、字型、音源、版本、平台架構、預設顯示模式及檔案清單。CHT包不得帶HD圖庫或HD參數，CHT-HD須包含050的complete-art.json及其引用圖檔。
- Linux實際由AppImage解出後啟動；Windows以Wine輔助啟動；macOS核通用架構、權限及.app結構，不宣稱macOS真機。
- 重新錄影同時檢視正常第一洞窟、代表中文、原貌／HD切換與工具列。只做本次交付所需的一次整合，不重跑完整原版對拍或通關。
- 推廣片驗視訊、音訊、時長、非靜音、黑幀、凍結及字幕邊界，另目視抽幀。原始擷取、音源及合成範圍保留來源雜湊。
- 實體USB／藍牙手把與Windows／macOS真機仍列未驗，不以合成事件或Wine代替。
