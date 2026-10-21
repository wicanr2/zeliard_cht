# 目前狀態

更新日期：2026-10-21，Asia/Taipei。

## 目前狀態表

使用者已要求結案，正式版為 **v.1.0.0-20261021**。沿用已確認的第一洞窟、代表中文對話、HD、工具列及音樂範圍，不重新開啟通關或全遊戲逐幀驗收。下表是唯一現行狀態；舊版候選、歷史缺口與歷程不能單獨重新開工。

| 項目 | 結案結果 |
|---|---|
| 執行架構 | dosgolem直接執行原版，3倍手繪HD與繁體中文唯讀覆繪；沒有另建玩法引擎 |
| 中文與輸入 | 粉圓2.1圓體、中文選單／對話／狀態列；工具列可設定鍵盤與系統已辨識手把，設定獨立保存 |
| HD | 已盤點美術完成製作與整合；43組來源、30張新增圖、8,392筆完整／局部來源及1,221個主角／刀光組裝鍵；未知來源安全退回 |
| 效能及受傷 | 沿052修正；圖樣搜尋基準44.32ms降至6.62ms，持盾受擊變色及特殊姿態維持HD |
| 音樂 | 原版AdLib曲目與Nuked-OPL3音源；實際時間推進，卡頓最多追趕100毫秒，暫停及失焦不追趕 |
| 正式封包 | CHT／CHT-HD各有Linux x86_64 AppImage、Windows x86_64 ZIP、macOS Intel／Apple Silicon通用ZIP，共六個含遊戲本機完整包 |
| 新包檢查 | 六包內嵌版號、逐檔內容／SHA、原版29檔、語言與HD模式、字型／音源及存檔目錄通過；Linux兩版啟動、Windows兩版Wine輔助啟動、macOS雙架構與.app結構通過 |
| 正常玩家路徑 | 新1.0程式正常冷F7進第一洞窟、接敵受傷、原貌／HD同狀態切換、正常DOS退出0；USR內容與mtime不變 |
| 影片 | 附55.45秒Ada連續遊玩與60秒推廣片，保留v.0.4.1錄製身份。遊戲來源與素材除版號外相同，影片位元組與已驗品質收據相同，沒有宣稱重新錄製 |
| workspace | 清除重複組裝／解壓、Wine、Go快取及已合併影片片段，釋放5,335,515,136配置bytes，約4.97GiB。唯一素材、原版、存檔、錄影、證據及歷次正式交付保留 |

## 正式交付與重建入口

- 本機唯一現行交付：`dist-all/v.1.0.0-20261021/`。完整版在`full-local/`，影片在`promo/`，驗收在`smoke/`，八份成品的大小與SHA-256在`SHA256SUMS.json`。
- 工作區：`workplace/release-v100/`；`identity.json`固定來源，`formal-promotion.json`證明此次來源未改，`program-version.txt`是實際程式版號，`cleanup.json`保存逐項清理。
- [雙版本建置](tools/build_dual_release.sh)讀取[VERSION](VERSION)，以Docker執行prepare、compile及package；先建立同版本的workplace/release-v100空目錄，再呼叫`bash tools/build_dual_release.sh all`。既有版本輸出拒絕覆寫，重現已發布版本請使用隔離工作樹及空輸出根。
- [影片沿用核對](tools/promote_release_media.py)在Docker內接受來源版號`v.0.4.1-20261021`；只在遊戲、dosgolem、原版與圖面來源相同時複製，保留原錄製身份及SHA。
- [封包啟動](tools/platform_smoke.py)、[正常受傷HD](tools/hd_runtime_smoke.py)、[六包內容核對](tools/audit_dual_release.py)與[清理](tools/cleanup_dual_release.py)保存在專案，命令與結果見[WORKLOG](WORKLOG.md#2026-10-21-054正式版100結案)。
- GitHub維持私人儲存庫`wicanr2/zeliard_cht`，[正式來源Release](https://github.com/wicanr2/zeliard_cht/releases/tag/v.1.0.0-20261021)已發布，非草稿、非預發行；同名tag固定來源提交`556a113`，未上傳任何封包附件。含原版資料與衍生美術的六包及影片只留本機。

## 驗收邊界

使用者接受的交付範圍已完成。未完整通關、未逐幀核全遊戲、未逐一正常到達所有後段場景。Pureza第316欄ID191及未知來源仍保留原貌；此版號不提升這些來源的證據等級。Windows／macOS真機與實體USB／藍牙手把仍未驗，Wine及合成輸入不能代替實機。這些是明示限制，不是未結案的自動工作。

## 來源與證據入口

| 來源 | 位置與用途 |
|---|---|
| 原始遊戲與手冊 | 根目錄神劍封魔.jsdos.zip、珍084-神劍封魔.rar；固定遊戲輸入workplace/orig。手冊故事與檔案雜湊見[RESEARCH-LOG](RESEARCH-LOG.md) |
| 完整HD、原圖與提示詞 | workplace/full-hd-v1/manual-hd-v1/generated、registered與各來源目錄；整合入口[050](docs/spec/050-manual-hd-delivery.md) |
| 圓體字型 | 同工作區rounded-font，含OFL全文、來源、manifest與尺寸字形；[轉製工具](tools/prepare_rounded_font.py) |
| HD效能與受傷 | workplace/hd-runtime-v1；[052](docs/spec/052-hd-runtime-damage.md)與[實際驗證矩陣](VERIFICATION-MATRIX.md) |
| Ada錄影 | workplace/ada-promo-v1/received、原始PCM／影片、10張視窗及輸入日誌；[053](docs/spec/053-ada-gameplay-promo.md) |
| 文件圖庫 | [README](README.md)內的原版／HD、素材、怪物及首領縮圖；docs/images兩份provenance清單 |

舊CONTEXT與WORKLIST已由本次結案狀態取代，原文可回查Git提交`3af5b79`。完整時間序列留在唯一[WORKLOG](WORKLOG.md)，不把舊候選描述當成現行待辦。
