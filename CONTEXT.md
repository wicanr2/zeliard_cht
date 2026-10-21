# 目前狀態

更新日期：2026-10-21，Asia/Taipei。

## 目前狀態表

現行修正版為 **v.1.0.2-20261021**。本次處理入洞過場與五組首領HD造型，並補上README遊戲簡介、手冊故事與完整角色總覽。既有結案範圍維持。

| 項目 | 現況 |
|---|---|
| 入洞過場 | 原版完成回呼提供城鎮擦除來源；洞壁核對兩側錨點與逐8×8圖格，未完成單格保留原版，其餘維持HD |
| 五組首領 | 章魚、飛鳥、藍白魔像、獨眼魔及披風魔王改用完整構圖，再裁切回原生元件；遊戲與README使用同一圖面 |
| 覆蓋邊界 | 五張代表姿態及430個模板已更新；930個未涵蓋非空象限保留原貌，未驗完整首領動畫或後段正常出場 |
| README | 已補中文手冊故事背景，30組怪物及10組首領代表姿態以完整角色呈現，來源及雜湊可回查 |
| 完整包 | CHT與CHT-HD分開，各有Linux x86_64 AppImage、Windows x86_64 ZIP、macOS Intel／Apple Silicon通用ZIP |
| 封包驗證 | Linux兩版實際啟動、Windows兩版Wine輔助啟動、macOS兩版架構及.app結構；六包內容、SHA、原版資料、模式、字型、音源及新增素材核對通過 |
| 正常入洞 | 正式Linux包冷F7、入洞及DOS退出0；16筆擦除、29筆HD過場，中途整幅回退0，USR內容與mtime不變 |
| 影片 | 本版附10.434秒無音軌入洞檢查片段；Ada推廣片保留1.0.1原身份，本輪未重錄 |

## 正式交付與證據入口

- 唯一現行交付為`dist-all/v.1.0.2-20261021/`。六個含遊戲完整版在`full-local/`，驗證與入洞片段在`smoke/`，大小及SHA-256在`SHA256SUMS.json`。完整版只留本機。
- 固定來源為`workplace/release-v102/identity.json`；六包核對在`dist-all/v.1.0.2-20261021/smoke/release-audit.json`，正常入洞在`smoke/integration/receipt.json`。Linux正式執行檔SHA為`031cfe6f45e6d3f48f588e1e31b0c09950106ef512dfcfb1bf474a83c1814037`。
- [057過場規格](docs/spec/057-cave-transition-hd.md)與`workplace/hd-cave-transition-v1/`保存原版重生、來源、連續畫面及mtime勘誤。[重生工具](tools/hd_transition_probe.go)保留存檔mtime，兩個具名觀察點完整RAM與GUI相同，不外推整局parity。
- [058首領規格](docs/spec/058-coherent-boss-hd.md)與`workplace/hd-boss-coherent-v1/`保存完整圖、舊圖、登錄、提示詞及來源。`art-notes.json`是選用美術與提示詞索引。正式圖組在`workplace/full-hd-v1/manual-hd-v1/registered/`。
- 圖庫入口為[重組配置](docs/images/creatures-layout.json)、[來源清單](docs/images/creatures-provenance.json)與[製圖工具](tools/readme_creatures.py)。手冊、原始檔名與SHA見[RESEARCH-LOG](RESEARCH-LOG.md)。原始ZIP／RAR留在根目錄，固定遊戲資料在`workplace/orig/`。
- [雙版本重建入口](tools/build_dual_release.sh)與[包裝工具](tools/package_dual_release.py)讀取[VERSION](VERSION)，在Docker內使用固定輸入，拒絕覆寫既有正式輸出。[六包核對](tools/audit_dual_release.py)以`--transition-fix`驗本版；[正常過場檢查](tools/hd_transition_smoke.py)及[清理工具](tools/cleanup_dual_release.py)保留重現入口。
- GitHub維持私人儲存庫`wicanr2/zeliard_cht`。只推來源、文件與已授權文件縮圖，完整包、完整素材及影片不上傳。發布與清理結果見[WORKLOG](WORKLOG.md#2026-10-21-057058-入洞過場與五組首領構圖修正)。歷次tag、Release、封包與影片保持原身份。

## 既有實作與驗收邊界

維持dosgolem直接執行原版、3倍手繪HD與繁體中文唯讀覆繪。中文使用粉圓2.1。工具列可設定鍵盤與系統已辨識USB／藍牙手把，藍牙由系統配對。配樂使用原版AdLib與Nuked-OPL3，實際時間推進最多追趕100毫秒，暫停與失焦不追趕。

未完整通關、未逐幀核全遊戲或逐一到達後段；未知來源及Pureza191保留原貌。Windows／macOS真機及實體USB／藍牙手把仍未驗，Wine與合成輸入不能代替實機。這些是明示限制，不自動重開舊待辦。

既有全圖整合見[050](docs/spec/050-manual-hd-delivery.md)，加速與受傷HD見[052](docs/spec/052-hd-runtime-damage.md)，Ada方法見[053](docs/spec/053-ada-gameplay-promo.md)，洞窟誤貼根因見[055](docs/spec/055-cave-artifact-context.md)。[README](README.md)保存穩定使用入口，[VERIFICATION-MATRIX](VERIFICATION-MATRIX.md)保存實際驗證範圍，歷程只追加至[WORKLOG](WORKLOG.md)。
