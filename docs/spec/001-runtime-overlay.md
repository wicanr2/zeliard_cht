# 001：原版執行與覆繪架構

狀態：架構 READY；前端版面、預設綁定及 HD 美術 DRAFT。

## 使用者定案

2026-10-18 選定直接用 dosgolem 執行原版，加上中文、HD 與手把前端；另指定參考 `/home/anr2/cht/psychic-war` 的執行期覆繪做法。不另建玩法重製引擎。

## 證據與邊界

- 原始 ZIP／EXE 雜湊見 [目前狀態](../../CONTEXT.md)。本規格不引用遊戲內線性位址。
- dosgolem 啟動至標題及故事片頭的收據見 [研究紀錄](../../RESEARCH-LOG.md)。只能證明該段已執行，不外推完整遊戲。
- 即時 API 採 `/home/anr2/cht/dosgolem/docs/spec/199-live-session-api.md` 的 READY 契約：`RunCycles`、`KeyDown`、`KeyUp`、`NewAudio`；檔案寫入使用 `SetScratch`。
- 對照 `/home/anr2/cht/psychic-war/cmd/psychicwar/main.go` 的 Draw／Update 與 `/home/anr2/cht/psychic-war/AGENTS.md` 第 3、5、6 節。遊戲專屬位址、字庫與文字資料不能直接抄用。
- 手把 API 核對本機 Ebitengine v2.9.9 來源與[同版公開文件](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2@v2.9.9)。裝置列舉不保證提供 USB／藍牙連線類型。

## 已審查的架構契約

1. 原版 EXE、資料、規則與記憶體保持原樣，由 dosgolem 執行。所有寫入另落在 scratch。
2. 合成順序為原版畫格 → HD 圖面 → 中文疊字 → 前端工具列。原貌能獨立擷取；未辨識素材保留原圖。
3. 翻譯以來源事件與已驗證畫面識別為鍵，不影響比較、查找或存檔。
4. Zeliard 專屬適配留本專案。共用 dosgolem 只讀；目前使用逐檔固定的本地來源快照。
5. 手把先映射到前端動作，再經原版鍵盤掃描碼送入。此映射不宣稱原版 gameport 硬體相容。
6. 按下／放開事件同一執行緒送入。鍵盤、手把同時按同一鍵時，任一來源仍按住就不放開。
7. 工具列或綁定捕捉輸入時，遊戲暫停推進並釋放已按住動作；原版 Esc 在遊戲畫面仍送入原版。
8. 裝置綁定以裝置識別字串及名稱分組，當次列舉索引不寫入設定。相同識別裝置不自動猜選。

## 可撤回的前端雛形

`cmd/zeliard-preview` 及 `internal/prototype/` 用於檢驗以上接入方式。固定工具列版面、死區初值、設定檔格式與裝置介面仍為雛形，不作正式發行契約。

雛形不包含 HD 素材，按鈕必須標示待製作。中文含前端工具列與可選的片頭單句 prototype，不能稱為遊戲中文化完成。正常新局至城鎮右移已驗；實體手把、對話、存讀檔、戰鬥與完整音訊另驗。

片頭試作採 dosgolem `202-translation-overlay.md`、`203-baked-text-watchers.md` 的既有 API。只在本地原版參考畫格的 `(88,136,160,16)` 區域完整相符時，覆繪「兩千年前，」。目前固定時鐘基線的文字墨跡範圍為 `(89,143)` 至 `(236,151)`。原文字面已定位至 `ZELRES1.SAR` 的檔案偏移 `0x109F`；控制碼與正式繪字契約未解，不能把此 prototype 升格為正式字串擷取方式。

GUI 診斷可用 `-capture-request` 等待指定訊號檔，再走原有 Draw 擷取與退出流程。訊號只控制擷取時機，不送按鍵或改寫原版狀態。正常玩家路徑驗證仍經 X11 鍵盤事件。

參考畫格 SHA-256 為 `de520c43e8b0098258cf5938f9afb42b60e7f04a8742d3eee5e897fcbcbcad5d`，來源是本地 `workplace/probe-intro-clock3000/screen.bin`，由 3000 cycles／毫秒、10,729,940 指令、無輸入重生，與診斷 GUI 的原版色號畫格完全相同。早期無固定 cycles 的 40m 探針保留在研究紀錄，不作此即時前端的基線。全匹配、原文消失後移除、不匹配拒絕覆繪及區外不寫入均須驗證。

## 驗證

- 建置、靜態檢查與輸入層測試：合併輸入、斷線釋放、軸中立值、越界裝置及設定存讀。
- GUI：Docker／Xvfb 正常啟動，保留完整原版畫面，開關設定頁不誤送遊戲按鍵。
- 對拍：同一原版初始狀態及指令預算的畫面／色盤不因前端載入而改變。
- 實體 USB 與藍牙驗收分開記錄，合成輸入不取代實機。
