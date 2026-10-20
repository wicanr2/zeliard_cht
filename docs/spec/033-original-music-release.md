# 原版配樂音色與三平台本機封包

狀態：限定CONFORMED，2026-10-20。音源接入與實際時間排程已獲授權；使用者選定最多追趕100毫秒，暫停及失焦不追趕。

使用者要求在原版基礎改善音色，接著製作 Windows、macOS、Linux AppImage 完整素材封包及推廣影片。遊戲內容沿已驗第一洞窟，不宣稱全遊戲翻譯或美術完成。

## 已知來源

- 原版 `RESOURCE.CFG` 選用 MSCSTD。只在可寫隔離層改選包內 MSCADLIB，音效保留 SNDSTD。原始29檔不改。
- `workplace/music-release-v1/` 的正常冷開機與三次空白鍵探針，在相同30秒虛擬時間記錄5630次OPL寫入、270次起音，未實作清單為空。無鍵版本保持靜音，不能用它誤判無配樂。
- Nuked-OPL3 固定來源為 [765ec962](https://github.com/nukeykt/Nuked-OPL3/tree/765ec962e473aeb767e4cba74ffdc8f588ffbfe8)，原始檔與SHA保存在 `third_party/nuked-opl3/provenance.json`，授權為同目錄LICENSE。使用獨立可替換動態函式庫，程式自身授權不變。
- 使用者選擇的B手繪美術及既有中文保持。音樂不另作曲、不換旋律、不重新分配音符；只由成熟音源核心合成原版寫入。單聲道44100 Hz及標準規格近似沿029，不宣稱逐週期硬體一致。

## 接入契約

1. dosgolem採隔離副本，只增加作品中立 `FMSynth` 介面與首次Render前的設定方法。機器、事件時間、驅動及遊戲流程不加條件。
2. C橋接程式只配置音源、提交暫存器與取得取樣，不持有原版執行器。音源關閉後解除動態函式庫。橋接程式與核心一併提供來源及LGPL條款，允許替換與所需除錯。
3. 原生封包直接定位鄰近資產，macOS定位Contents/Resources/game。首次將自建SAVETEST複製到系統使用者設定目錄；正式存檔與鍵盤／手把設定不寫入安裝目錄或AppImage。
4. Linux交付x86_64 AppImage，Windows交付x86_64 ZIP，macOS交付Intel／Apple Silicon通用.app ZIP。全部含本機原版及現有中文／HD素材，不提交或公開散布。
5. 每平台從固定來源建置；驗版本、架構、檔案與SHA。Linux抽測實際AppImage，Windows以Wine作輔助啟動；macOS交叉建置與結構驗證不稱真機驗收。
6. 推廣影片使用實際遊戲錄影與原版音樂事件，來源、時間範圍及SHA可回查。核對視訊／音訊、時長、非靜音、字幕邊界、黑幀與凍結。本機保存，不自動上傳。

## 入口

- `tools/prepare_release.py`：固定來源與原版資料、音源及各平台資產整理。
- `tools/build-release.sh`：受限Docker建置、三平台與AppImage打包。
- `tools/dosgolem-fm.patch`：通用音源介面的可回查差異。
- `internal/prototype/fm/`：可替換動態音源載入。
- `cmd/zeliard-preview/bundle.go`：跨平台資料與存檔位置。
- 中間產物為 `workplace/music-release-v1/`；正式封包與影片依 `dist-all/<版本>/` 契約，不覆寫前版。

## 本輪驗收

使用者確認後的v.0.2.0-20261020已完成Clock／音源／輸入測試、vet、三平台建置及正常第一洞窟視窗。Linux AppImage與Windows ZIP的Wine輔助啟動、macOS雙架構與60秒本機推廣片均通過已列檢查，入口為CONTEXT的同版smoke／promo。兩個4096音框裝置片段相符，0.5秒起點跨度比1.0；不以此宣稱全音訊零丟樣或逐取樣時序。Windows／macOS真機、實體手把及喇叭維持未驗。
