# 工具列CPU速度與HD幀率

狀態：READY，2026-10-22。使用者明確要求2x、4x、8x、16x、32x及HD 30fps。只新增前端速度設定及等價效能改善，不改遊戲資料、地形、規則或存檔。

## 已確認來源

- 已證實：現行main以`-cycles=3000`呼叫dosgolem的`Oracle.SetDOSBoxCycles`，每次Update以固定預算或sound.Clock推進RunCycles。倍率不是幀率。
- 已證實：固定執行器`workplace/release-v102/dosgolem/oracle/live.go`提供SetDOSBoxCycles、DOSBoxCycles、RunCycles；`internal/machine/dosbox_cycles.go`將CPUHz設為cycles/ms乘1000，並重算IRQ0週期。這是既有執行器API，不另猜硬體driver。
- 已證實：Audio.Render依前後cycle差及當下cycles/ms取樣。切速前須先消化舊倍率音訊，不能用新倍率換算舊區段。
- 已證實：工具列遊戲畫布960×680，頂部48像素，底部650至678為狀態文字。設定頁原有暫停與放開輸入閘門。
- HD成本與各倍率實際幀率待本機量測。CPU倍率提高會增加每秒要求執行的指令，不能直接推論呈現必達30fps。

## 顯示與輸入

上方六個入口依序為原貌、HD、速度、手把、暫停、鍵盤。速度入口開啟CPU設定頁，提供1x、2x、4x、8x、16x、32x；1x保留原速入口。滑鼠選擇或方向鍵選擇後Enter確認，Esc取消。設定頁暫停原版，輸入不得同時送入原版。

工具列矩形為[8,8,112,40)、[120,8,256,40)、[264,8,400,40)、[408,8,592,40)、[600,8,736,40)、[744,8,952,40)。文字各留左右10像素。速度選項六列位於[300,170+i×50,660,212+i×50)，不能侵入標題與說明區。保留既有遊戲畫面尺寸。

## 狀態與保存

使用者已確認優先HD 30fps，所選倍率作CPU上限。3000基準下六檔上限分別3000、6000、12000、24000、48000、96000 cycles/ms。依負載自動降低實際CPU速率，右下角同時顯示實際倍率及完成合成幀率；不把上限稱為實際速度。保存於既有前端controller.json的可選cpu_speed欄位，舊版忽略新欄位，缺欄位或無效倍率回到1x，不修改原版USR。命令列`-cpu-speed`可覆蓋當次啟動上限。

切換時先放開原版按鍵、Render舊音訊，再SetDOSBoxCycles、更新前端速率並重設實際時間基準。暫停與失焦不追趕。切換不重新讀檔或注入遊戲狀態。日誌須明列倍率與生效cycles，不能把設定倍率稱作實際達成速度。

自動調速保留已累積虛擬時間，不對每段換算套用新速率；先Render舊區段，才更新Clock的分段時間及SetDOSBoxCycles。自動調速不清空播放佇列。每次Draw之間共享CPU執行額度，依33.33ms減去上次合成成本及3ms餘裕，額度限制1至12ms；每次RunCycles最多2048 cycles，單次原版呼叫或回呼仍可超出額度，不宣稱硬即時。每200ms按已執行cycles與耗時更新速率，向上最多增加約12.5%，不超過所選上限；最低60 cycles/ms。冷啟動由1x開始，不直接灌入32x負載。

原生日誌逐段保存實際RunCycles，另存audio及cpu_speed事件；重生工具沿相同段界與速率，不能把CPU上限當執行預算。批次讀取以普通RAM區塊複製，VGA、讀取監看、A20／HMA及跨界維持Read8次序與副作用。

## 驗證與停止線

核對六檔倍率、舊設定相容與保存、UI取消／輸入隔離，以及正常城鎮與第一洞窟的HD幀率。量測Update與Draw成本，幀率以實際完成合成計，不以錄影編碼60fps作證明。只做有證據的等價優化，保留同輸入像素。若主機無法兼得指定CPU倍率與30fps，必須列出實測並讓使用者決定取捨，不能暗中改成自動限速或宣稱所有平台皆能達標。

工作區`workplace/cpu-speed-v1/`保存基準及新程式來源、CPU profile、幀率、正常輸入與限制。原1.0.2封包及tag保持原身份。使用者接著告知目前CPU負載過重，無法完整測試；停止遊戲壓測，僅做必要編譯及低成本邊界核對。自動調速整合與30fps實測仍待低載環境驗收，不升CONFORMED、不宣稱30fps已達成。

## 本機候選入口

執行`bash tools/play-cpu-preview.sh`可開啟`workplace/cpu-speed-v1/preview/zeliard-preview`。候選沿既有本機正式包的素材連結，不能單獨搬移；原版資料與1.0.2包均保留。版本標為development，不覆寫正式版。

重建時在既有Go工具容器、單CPU及GOMAXPROCS=1下使用`GOWORK=/project/workplace/cpu-speed-v1/build.work`、唯讀模組快取及本專案可寫Go快取，執行`go build -buildvcs=false -o /project/workplace/cpu-speed-v1/preview/zeliard-preview ./cmd/zeliard-preview`。build.work指定本專案與新獨立執行器，後者由既有正式來源套用[批次讀取差異](../../tools/dosgolem-bulk-read.patch)取得；後續[雙版本包裝](../../tools/package_dual_release.py)也會核對並套用該差異。

`build-receipt.json`記錄候選二進位SHA及來源。設定保存、倍率上限、負載降速、時鐘換速保留已累積時間、普通／平面／跨界／監看副作用及差異重生核對通過；這些是內部邊界證據。最終候選依使用者指示未啟動遊戲，實際HD 30fps、音樂節奏與操作尚待後續確認。
