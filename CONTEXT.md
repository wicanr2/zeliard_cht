# 目前狀態

更新日期：2026-10-21，Asia/Taipei。

## 目前狀態表

現行正式修正版為 **v.1.0.1-20261021**。使用者回報推廣片洞窟主角右上方異常方塊，[055](docs/spec/055-cave-artifact-context.md)已完成修正。先前1.0.0結案範圍維持，沒有重新開啟完整通關或後段逐幀驗收。

| 項目 | 現況 |
|---|---|
| 洞窟方塊 | 已證實為無資源綁定的結局YUU3局部插畫，原生8×8只有單一色點，誤匹配到主角旁；現已核對母圖可見上下文才覆繪 |
| 原版來源 | 原Ada日誌重生5,064筆，暫存器／時鐘／RAM／原生完成畫面SHA相同；沒有改規則、資料、碰撞、座標或存檔 |
| 修正檢查 | 真實來源拒絕、正確母圖及邊界裁切正例、外圍像素失配反例通過；補圖層基準約7.62ms，不宣稱全遊戲FPS |
| 實際畫面 | 新1.0.1在Ada重新錄製；主角右上方及左側誤貼片段消失，主角、門框及洞窟HD保留，已目視 |
| 完整包 | CHT／CHT-HD各有Linux x86_64 AppImage、Windows x86_64 ZIP、macOS Intel／Apple Silicon通用ZIP，共六包 |
| 封包驗證 | Linux兩版實際啟動、Windows兩版Wine輔助啟動、macOS兩版架構及.app結構；六包逐檔SHA、29個原版檔案、模式、字型／音源與存檔目錄核對通過 |
| 正常玩家路徑 | 新1.0.1正常冷F7、入洞、接敵變色／特殊受傷HD、原貌切回、正常DOS退出0，USR內容及mtime不變 |
| 新影片 | Ada新實錄約55.48秒，960×680／60fps及同期聲；新推廣片60秒、1280×720，以原版標題曲配樂。音畫、非靜音、黑幀／凍結檢查通過，字幕、結尾及洞窟畫面已目視 |
| 清理 | 本次本機清理4,101,898,240配置bytes，約3.82GiB，明細以release-v101/cleanup.json及ada-promo-v101/cleanup-local.json為準；原版、素材、存檔、錄影、證據及舊正式版均保留 |

## 正式交付與重建入口

- 唯一現行交付：`dist-all/v.1.0.1-20261021/`。六個含遊戲完整版在`full-local/`，兩片新影片在`promo/`，驗證在`smoke/`，八份成品的大小及SHA-256在`SHA256SUMS.json`。全部只留本機。
- 固定來源及清理：`workplace/release-v101/identity.json`、`cleanup.json`及`hygiene.json`。原1.0.0、0.4.1及以前tag／交付保持原身份。
- 根因及同狀態來源：`workplace/hd-cave-artifact-v1/`，含原使用者截圖、重播輸入／來源、誤貼清單、修正前後診斷圖、回歸及效能日誌。入口為[055規格](docs/spec/055-cave-artifact-context.md)、[覆繪層](internal/prototype/hd/complete_art.go)及[回歸檢查](internal/prototype/hd/complete_art_context_test.go)。
- 新Ada原始錄影：`workplace/ada-promo-v101/received/`，含GPU renderer、原始影片、PulseAudio PCM、原版WAV、輸入與十張代表視窗。新舊實錄是各次正常遊玩，不宣稱整局同狀態；同狀態根因證據另由舊日誌重生。
- 重建用[雙版本入口](tools/build_dual_release.sh)及[包裝工具](tools/package_dual_release.py)，讀取[VERSION](VERSION)。先建立同版本空workplace/release-v101再由Docker重建，拒絕覆寫既有正式輸出。影片用[錄影工具](tools/record_ada_promo.py)及[剪輯工具](tools/make_ada_promo.py)，剪輯時設`ZELIARD_PROMO_WORK=workplace/ada-promo-v101`。
- [封包啟動](tools/platform_smoke.py)、[正常受傷HD](tools/hd_runtime_smoke.py)、[六包核對](tools/audit_dual_release.py)、[影片檢查](tools/verify_promo.py)及[清理](tools/cleanup_dual_release.py)保留可重現入口。實際命令及結果見[WORKLOG](WORKLOG.md#2026-10-21-055洞窟誤貼結局圖塊修正)。
- GitHub維持私人儲存庫`wicanr2/zeliard_cht`，只推來源、文件及原已授權文件縮圖；完整包與影片不上傳。

## 既有實作與驗收邊界

維持dosgolem直接執行原版、3倍手繪HD與繁體中文唯讀覆繪。已整合43組來源、30張新增圖、8,392筆完整／局部來源及1,221個主角／刀光組裝鍵；中文使用粉圓2.1。工具列可設定鍵盤與系統已辨識USB／藍牙手把，藍牙由系統配對。配樂用原版AdLib與Nuked-OPL3，實際時間推進最多追趕100毫秒，暫停／失焦不追趕。

未完整通關、未逐幀核全遊戲或逐一到達後段；未知來源及Pureza191保留原貌。Windows／macOS真機及實體USB／藍牙手把仍未驗，Wine及合成輸入不能代替實機。這些是明示限制，不自動重開舊待辦。

## 素材與文件索引

原始ZIP／RAR保留根目錄，固定遊戲在workplace/orig。手冊故事、來源及雜湊見[RESEARCH-LOG](RESEARCH-LOG.md)。完整HD原圖、提示詞、registered及rounded-font在workplace/full-hd-v1/manual-hd-v1，入口[050](docs/spec/050-manual-hd-delivery.md)。HD加速及受傷來源見[052](docs/spec/052-hd-runtime-damage.md)，初次Ada方法見[053](docs/spec/053-ada-gameplay-promo.md)。[README](README.md)保留故事、原版／HD、素材、怪物及首領總覽；[VERIFICATION-MATRIX](VERIFICATION-MATRIX.md)保留實際範圍。舊CONTEXT／WORKLIST可回查各版Git，歷程留在唯一[WORKLOG](WORKLOG.md)。
