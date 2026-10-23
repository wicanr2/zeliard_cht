# 目前狀態

更新日期：2026-10-23，Asia/Taipei。

本輪入口為[067開局揭開、F7選單與結局的無閃爍呈現及補譯](docs/spec/067-start-menu-ending-presentation.md)，工作區`workplace/release-v107/`。使用者在1.0.6推送後要求：結局敘述、製作名單與版權翻譯；開局逐格揭開與第一次進城堡、F7都要維持HD；中文出現後又變回英文要確定處理好。決定另出1.0.7，結局以直接載入結局模組驗證，HUD場景名、房間標題與alma改「中文（English）」，日籍人名查得到漢字就寫漢字。

現行本機版本為`v.1.0.7-20261023`，完整包位於`dist-all/v.1.0.7-20261023/full-local/`，`bash tools/play.sh`啟動HD版、`--cht`選原貌中文版。正式建置在`workplace/release-v107/package/`，Linux二進位SHA為`3eeb4b7a9ebdeeea7cbc9e826fdc84bec401aeef48df4aae37c52eb87552bed6`，出貨來源與工作樹逐檔相同。第一次1.0.7打包未發布即作廢（說話沿用原本不分模式，可能蓋住遊戲過場背景上的角色）；作廢產物已於2026-10-23依使用者指示刪除。1.0.7沒有新推廣片。

## 開發中：068片頭與結局插畫HD覆蓋（未打包）

1.0.7以`6830358`推送後，使用者要求清理作廢打包並修正結局與片頭陽台插畫；第一輪以`ad77357`推送。入口[068片頭與結局插畫的HD覆蓋修正](docs/spec/068-picture-coverage.md)，工作區`workplace/picture-coverage-v1/`。[補登錄工具](tools/hd_picture_complete.py)把只有人工圖格的插畫登錄為整張插畫與片段，FIN兩張遮去第0列，國王擁抱圖背景透明；[插畫覆繪](internal/prototype/hd/complete_art.go)在片頭／結局模式下以特徵點過半相符補畫被局部蓋住的插畫。

2026-10-23使用者指示以正確色序重畫，由Codex內建生圖處理。[參考圖工具](tools/hd_picture_refs.py)依實機截圖的點陣與色盤做參考圖，生成圖由[色相閘門](tools/hd_hue_check.py)檢查；[補丁工具](tools/hd_portrait_patch.py)處理人像說話與眨眼、魔王與`new1`的說話嘴型及yuu3灰光剪影畫格。story四張、完整插畫三張、`new1`與以實機點陣為準的`sei`、`dmaou`都已重畫並登錄；`complete-art.json` `b4103bf6…`（12,138項），開發執行檔仍是`d32af63c…`，未打包。片頭每次比對約多3到4毫秒，遊戲中約多1.6毫秒。

未處理：片尾捲字期間捲字層蓋過yuu3剪影畫格的HD，主角進光時溶解沿用誤判（兩者都是前端既有行為，改動須動規格066的契約）；`dmaou-cycle-all`、`credits-scroll`第96列以後、`ttl1`、`ttl2`、`nec`沒有實機證據；片頭單獨的魔王雙眼沒有來源。

使用者決定（2026-10-23）：本輪只commit推送，不打包、不出新版號。


## 目前狀態表

| 項目 | 1.0.7現況 |
|---|---|
| 結局中文 | 敘述27段、製作名單與登場魔物及版權60段覆繪；探針以`poke 01B1:0049=1`直接觸發，屬結構驗證 |
| 開局柱子 | 揭開期間逐格畫HD，眼睛暗色畫格列入允許；邊框HD隨選單模組載入提前備妥 |
| 第一次進城堡 | 原版畫完前沿用上一張城鎮畫帶，完成即HD，上限1M指令 |
| F7選單 | 城鎮沿用最後一張已驗完成畫面；洞窟以存框追蹤沿用；選單框內原版樣式加中文 |
| 結局插畫 | 片頭／結局模式下，說話小圖時大插畫沿用整張HD；對話肖像框沒有整張肖像時整框原版 |
| 譯名 | HUD場景名、房間標題「賈辛（Jashiin）殿」、「靈魂寶石（alma）」、片頭故事與捲字專名、名單漢字 |
| 探針 | 正式binary 17條（HD 12、CHT 5）英文先於中文、中文後變回英文、寫回後可見、狀態列未畫全為0；不追蹤色號0字形，字形可見須同格原版底色照原版顯示 |
| 保留限制 | 結局部分插畫HD對不上（加蘭離城、公主對話姿態、FIN）；陽台棋盤格；後段內容只靜態通過。加蘭離城、FIN與陽台棋盤格的處理見068（未打包）；公主對話姿態仍無HD |

1.0.7正式包驗證在同版`smoke/`與`SHA256SUMS.json`；探針輸出在`workplace/release-v107/probe/out/rel*`。

## 前次交付（1.0.6）

1.0.6入口為[066原版字形改寫與無閃爍呈現](docs/spec/066-glyph-patch-stable-presentation.md)，工作區`workplace/release-v106/`，以`4b884ba`推送。1.0.6版本為`v.1.0.6-20261023`，完整包位於`dist-all/v.1.0.6-20261023/full-local/`。正式建置在`workplace/release-v106/package/`，Linux二進位SHA為`5a73a5d59a44076cec75b1dd13c797746adfb4fb2f2e136fb2466aa36b4138cb`。前兩次打包未發布即作廢：第一次在片頭跳過與房間離場會露出英文，第二次在居民對話開框有一個合成點HD與原版拼接；兩次作廢產物已於2026-10-23依使用者指示刪除。1.0.6沒有新推廣片，1.0.5影片保留原版號。

## 1.0.6狀態表

| 項目 | 1.0.6現況 |
|---|---|
| Windows啟動 | 改以檔案身份比對祖先目錄，遊戲與使用者目錄在不同磁碟也能啟動；Wine跨磁碟預設目錄兩包通過，使用者回報的C槽案例未重現 |
| 字形改寫 | 啟動時由原版`GMMCGA.BIN`另存patch版到使用者目錄；已認領文字畫到`B000`不顯示，原版與patch版同步差分零差 |
| 英文露出 | 正式binary探針：英文先於中文0；中文之後退回英文只剩黑底上的黑色陰影字；從未覆蓋的只有黑字與續頁、填空圖形 |
| 人像 | 國王、道具店、銀行、教堂進場揭開期間顯示HD；離場清黑時保留HD人像同步溶解 |
| 轉場清黑 | 片頭按Space與房間離場改由保存的合成畫面同步溶解，見066「轉場逐點清黑」 |
| 居民開框 | 對話框內只有原版已畫到的列換回原版，畫框前保留HD |
| 補譯 | 對話、商店說明、數字句、洞窟訊息、首領台詞與標題版權補齊，專有名詞寫「中文（English）」；部分只靜態通過 |
| 保留限制 | 開局外框柱子逐格組合時顯示原版；F7讀檔選單期間城鎮背景退原版；陽台插畫棋盤格；結局未中文化；HUD專名未改 |

1.0.6正式包驗證在同版`smoke/`與`SHA256SUMS.json`；探針輸出在`workplace/release-v106/probe/out/v106rel3-*`，摘要在`smoke/probe-release/`。

## 前次交付（1.0.5）

前次入口為[065大字版重包與推廣片](docs/spec/065-readable-release-promo.md)，工作區`workplace/release-v105/`。使用者確認064後已授權正式重包及錄影，取代候選階段「不重打」安排。

065已完成六包與60秒推廣片，入口為工作區`repack-receipt.json`。影片音畫、字幕及來源通過，540個重用HD畫格與720個新錄設定編碼畫格比對通過；原版標題曲配樂非同期聲。清理約3.78GiB後七份SHA不變，明細在`cleanup.json`與`cleanup-media.json`。完整遊玩與30fps未重驗。

已完成[064系統介面字級](docs/spec/064-readable-system-text.md)並收錄1.0.5六包。沿既有圓體提供16／24／32像素，設定本文預設24，工具列及底部最高24。正式HD AppImage的三檔切換、四種設定頁特大字與冷重開通過，原版steps／cycles為0、存檔不變。收據在同版`smoke/toolbar/receipt.json`；舊`development-ui-text-v1`候選及收據保留歷史身份。

1.0.5版本完整包位於`dist-all/v.1.0.5-20261022/full-local/`。CHT與CHT-HD各提供Linux AppImage、Windows x86_64 ZIP及macOS通用ZIP。舊六包、影片與已發布tag保留原身份；最近GitHub Release仍為1.0.2，完整版與影片只留本機。

最新影片仍是065新錄的正式1.0.5字級、作弊與速度頁，混剪具名舊遊玩來源。1.0.4的[063連續防閃爍證明](docs/spec/063-feature-promo-stability-proof.md)保留原版號；新版主片10–17秒對話、17–28秒平台沿用該版來源，各自連續。28–40秒展示16／24／32字級，各4秒。配樂為原版標題曲、非同期聲，使用者已接受幀率，不重開30fps或通關。

1.0.4基礎修正入口為[060作弊與難度](docs/spec/060-player-cheats.md)、[061啟動與按需HD](docs/spec/061-startup-on-demand-hd.md)、[062穩定呈現](docs/spec/062-stable-hd-presentation.md)。該輪研究與候選在`workplace/runtime-controls-v1/`，正式來源副本與建置在`workplace/release-v104/`。1.0.5沿用這些功能，不重開完整通關、全遊戲逐幀或實體手把驗收。

## 1.0.5狀態表

| 項目 | 現況 |
|---|---|
| 系統字級 | 1.0.5兩版皆可選16／24／32，預設24並保存；正式HD包14張設定視窗及冷重開已核對 |
| 工具列作弊 | 獨立鎖滿生命與MP，每次啟動關閉；MP為各法術剩餘次數，保留原學習與選擇規則 |
| 站立回血 | 預設沿DOS版開啟，可獨立關閉並保存；道具與教堂回血仍有效，不稱PC-88完整規則重現 |
| 啟動 | 先開準備視窗與錯誤日誌；大量圖片按需核SHA及解碼，城鎮／室內／洞窟與其餘地形按場景準備 |
| 正式包啟動 | 1.0.5兩個Linux包與兩個Windows包的Wine輔助啟動通過，原版steps／cycles皆0；macOS只核通用架構及內容 |
| 1.0.4啟動抽樣 | 當時正式HD AppImage約0.33秒映射視窗、2.03秒設定就緒，15812個圖片註冊而0張載入；不是1.0.5或所有電腦的時間保證 |
| 對話HD | 正式1.0.4新增連續實錄，完整證明327個影片畫格及161筆去重來源均保留HD；片中10–17秒另核210個畫格 |
| 洞窟HD | 正式1.0.4綠色平台站立／升降／移動連續實錄，完整證明857個影片畫格及248筆去重來源均保留HD；主片17–30秒另核390個畫格 |
| 效能 | 使用者已接受前次幀率結果，本輪不再調速或以30fps為交付閘門；錄影封裝30fps不代表遊戲每秒30個新畫格 |
| 原版來源 | 固定EXE、原始遊戲與存檔不覆寫；作弊開啟是明示規則差異，不列原版oracle證據 |
| 保留限制 | 本次音訊關閉，未重驗播放品質、Windows／macOS真機、實體手把或完整通關 |

1.0.5正式包驗證在同版`smoke/linux-CHT*`、`windows-CHT*`及`toolbar`，六包內容稽核為`smoke/release-audit.json`。下方1.0.4的遊玩與來源收據保持原身份，不冒稱1.0.5正常遊玩。前次候選路徑在`workplace/runtime-controls-v1/gameplay-v6/`，二進位SHA為`45f74ac1a963e5083291847041aa30e23180096b6a1388bb8edc3d74e284f926`；來源測試為`targeted-tests-v5.jsonl`、`controls-tests-v6.jsonl`，原始平台補證為`platform-source-v2/`。

README已加入1.0.5正式字級與特大作弊頁，來源為`docs/images/system-text-v105-provenance.json`。舊1.0.4作弊、1.0.3速度與1.0.2遊玩圖片／影片保留版本標示。

1.0.4重包當時清除約3.08GiB重複組裝、解包與Wine層。六包、原版、字型、來源與二進位、截圖及驗證收據保留；歷史入口為`workplace/release-v104/repack-receipt.json`與`cleanup.json`。

1.0.4影片當時清除約279.2MiB解包與剪輯重複層，三片及六包SHA再核對不變。同版`promo/media-manifest.json`索引該輪成品與QA，`workplace/promo-features-v104/cleanup.json`保存清理結果。

1.0.4證明片由當時正式AppImage擷取，Intel MTL／VirtualGL EGL與h264_vaapi硬體編碼，兩段保留完整960×680視窗，不剪接、不加速、不手動切換HD。生命鎖定開啟、原版標題曲非同期聲均標明。該版的音畫與HD核對仍在1.0.4的`promo/quality.json`、`hd-stability-proof.json`、`visual-review.json`；原貌／HD參考只在證明區間之外取得，沒有混入補充證明。

## 正式交付與證據入口

- 1.0.3打包入口為[build_dual_release.sh](tools/build_dual_release.sh)，單CPU與Go平行度1。[audit_dual_release.py](tools/audit_dual_release.py)的`--cpu-speed-repack`核六包，結果為內容通過、執行待驗；同版`SHA256SUMS.json`保存大小與SHA。來源與工具收據在`workplace/release-v103/`，[cleanup_dual_release.py](tools/cleanup_dual_release.py)的同名選項只清本版組裝重複層，保留Go快取、來源與獨立二進位。
- 1.0.3六包內容稽核及清理後SHA核對均通過。清除本輪重複組裝約1.14GiB，550個文件連結有效，root-owned與誤建.md目錄為零，專案容器已移除。入口為同工作區`repack-receipt.json`、`cleanup.json`與`hygiene.json`；Linux二進位SHA為`ad37d074b06c54115c8144662970a33650e09944f29bd52e45b106fab1276e9d`。來源同步回讀保存在`git-publication.json`。
- 先前CPU候選在`workplace/cpu-speed-v1/preview/`，沿既有1.0.2圖庫連結，保留診斷來源及`build-receipt.json`。新版打包也以`tools/dosgolem-bulk-read.patch`在獨立dosgolem副本重生，不改共用執行器。前期固定倍率資料只代表高載下的可丟棄量測，不作最終自動調速驗收。

- 1.0.3六包在`dist-all/v.1.0.3-20261022/full-local/`；`bash tools/play.sh`依VERSION啟動目前版本。1.0.2六包與影片仍在`dist-all/v.1.0.2-20261021/`，保持原身份。所有完整版與影片只留本機。
- 本機新片來源為`workplace/ada-promo-v102/received/`，保留原始錄影、PCM／WAV、輸入及十張視窗；`local-input-manifest.json`記錄正式包及固定腳本。139筆HD過場中途整幅回退0，正常DOS退出0，存檔內容與mtime不變；過場角色沿057的局部HD與原貌保留邊界，不宣稱每個像素皆重繪。
- 本次錄影清理約473MiB重複傳輸、解包及剪輯中間檔，原始錄影與八個正式成品保留。清理後八份SHA及遊戲來源相同，root-owned與誤建.md目錄均零；入口為同工作區`cleanup-local.json`、`hygiene.json`及`git-publication.json`。
- 1.0.2固定來源為`workplace/release-v102/identity.json`；舊六包核對在`dist-all/v.1.0.2-20261021/smoke/release-audit.json`，正常入洞在`smoke/integration/receipt.json`。舊Linux執行檔SHA為`031cfe6f45e6d3f48f588e1e31b0c09950106ef512dfcfb1bf474a83c1814037`。
- [057過場規格](docs/spec/057-cave-transition-hd.md)與`workplace/hd-cave-transition-v1/`保存原版重生、來源、連續畫面及mtime勘誤。[重生工具](tools/hd_transition_probe.go)保留存檔mtime，兩個具名觀察點完整RAM與GUI相同，不外推整局parity。
- [058首領規格](docs/spec/058-coherent-boss-hd.md)與`workplace/hd-boss-coherent-v1/`保存完整圖、舊圖、登錄、提示詞及來源。`art-notes.json`是選用美術與提示詞索引。正式圖組在`workplace/full-hd-v1/manual-hd-v1/registered/`。
- 圖庫入口為[重組配置](docs/images/creatures-layout.json)、[來源清單](docs/images/creatures-provenance.json)與[製圖工具](tools/readme_creatures.py)。手冊、原始檔名與SHA見[RESEARCH-LOG](RESEARCH-LOG.md)。原始ZIP／RAR留在根目錄，固定遊戲資料在`workplace/orig/`。
- [雙版本重建入口](tools/build_dual_release.sh)與[包裝工具](tools/package_dual_release.py)讀取[VERSION](VERSION)，在Docker內使用固定輸入，拒絕覆寫既有正式輸出。1.0.3六包核對使用`--cpu-speed-repack`；1.0.2的`--transition-fix`與[正常過場檢查](tools/hd_transition_smoke.py)保留歷史重現入口。
- GitHub維持私人儲存庫`wicanr2/zeliard_cht`。只推來源、文件與已授權文件縮圖，完整包、完整素材及影片不上傳。發布與清理結果見[WORKLOG](WORKLOG.md#2026-10-21-057058-入洞過場與五組首領構圖修正)。歷次tag、Release、封包與影片保持原身份。
- [1.0.2正式修正版](https://github.com/wicanr2/zeliard_cht/releases/tag/v.1.0.2-20261021)已發布，來源提交`6f87a98`，非草稿／非預發行，assets為空，儲存庫回讀PRIVATE。回讀在`workplace/release-v102/github-publication.json`。清理約3.95GiB後，六包SHA及固定遊戲來源相同，root-owned與誤建.md目錄為零；詳見同目錄`hygiene.json`。

## 既有實作與驗收邊界

維持dosgolem直接執行原版、3倍手繪HD與繁體中文唯讀覆繪。中文使用粉圓2.1。工具列可設定鍵盤與系統已辨識USB／藍牙手把，藍牙由系統配對。配樂使用原版AdLib與Nuked-OPL3，實際時間推進最多追趕100毫秒，暫停與失焦不追趕。

未完整通關、未逐幀核全遊戲或逐一到達後段；未知來源及Pureza191保留原貌。Windows／macOS真機及實體USB／藍牙手把仍未驗，Wine與合成輸入不能代替實機。這些是明示限制，不自動重開舊待辦。

既有全圖整合見[050](docs/spec/050-manual-hd-delivery.md)，加速與受傷HD見[052](docs/spec/052-hd-runtime-damage.md)，Ada方法見[053](docs/spec/053-ada-gameplay-promo.md)，洞窟誤貼根因見[055](docs/spec/055-cave-artifact-context.md)。[README](README.md)保存穩定使用入口，[VERIFICATION-MATRIX](VERIFICATION-MATRIX.md)保存實際驗證範圍，歷程只追加至[WORKLOG](WORKLOG.md)。
