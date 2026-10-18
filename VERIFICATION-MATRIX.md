# 驗證範圍

只記錄實際驗過的範圍。skip、未執行與環境受阻不得計入通過。

| 項目 | 狀態 | 證據與限制 |
|---|---|---|
| 原始 ZIP CRC | 通過 | 30 個項目，含遊戲程式與 js-dos 設定 |
| ZIP／RAR SHA-256 | 通過 | 見 [CONTEXT.md](CONTEXT.md) |
| 中文手冊故事來源 | 通過 | 人工核對第 1、7–8 頁，見 [RESEARCH-LOG.md](RESEARCH-LOG.md) |
| 安全解包與逐檔清單 | 通過 | 29 檔，`workplace/source-manifest.json`；原始包及解包檔案雜湊不變 |
| dosgolem 啟動 | 通過，限標題與故事片頭 | `workplace/probe-startup-mcga/`、`probe-space-40m/`；未證明新局／城鎮可玩 |
| 固定輸入重播 | 通過，限 200 萬指令 | 畫面色號、色盤、PNG 與末端軌跡相同；`workplace/startup-replay-comparison.json` |
| 城鎮右移重播 | 通過，限 6600 萬指令 | 冷開機、3000 cycles、固定 IRQ1 序列；畫面色號、色盤、PNG 與末端軌跡相同；`workplace/town-replay-comparison.json` |
| 前端建置及 go vet | 通過 | Go 1.24.13、Ebitengine 2.9.9、本地 dosgolem 快照；GOMAXPROCS=2 |
| 正常新局／城鎮右移／放開 | 通過，限此切片 | `workplace/gui-town-v2/`，X11 實際視窗按鍵；狀態列完整相符後才右移，放開後錨點不再平移且持鍵集合為空 |
| 國王初訪兩段中文與結束 | 通過，限此切片 | `dialogue-v3/receipt.json` 與 `gui-dialogue-hd-v5/`；兩段中文、241 原文字形、1000 賞金與正常返回城鎮，未注入狀態 |
| 對話顯示與語意隔離 | 通過，限國王初訪 | 九個抽樣點的完整 1 MiB 記憶體、色號、色盤、CS:IP 及 cycles 與無覆繪原版相同；未知來源退回，CL 座標回歸通過 |
| 長老初訪中文與四列選單 | 通過，限 prototype | `labels-ui-v2/receipt.json` 與 `save-ui-gui-v2/`；初訪 114 字形，四列標籤與對話區外差異 0，完整原版狀態相同 |
| 新存檔、退出與冷讀檔 | 通過，限此正常路徑 | `save-ui-v3/receipt.json` 22 點完整 1 MiB 記憶體、色號、色盤、CS:IP、steps、cycles 相同；兩側各自建立相同的 256-byte 新檔，再冷重開讀取。實際 X 視窗 `save-ui-gui-v2/` 通過 UIVIEW2 保存、退出碼 0、F7 讀回 Muralla Town／1000 金幣 |
| 保存訊息中文與 UI 退回 | 通過，限已列來源 | 實際視窗 `record-saved.png` 及配對收據；原版名稱輸入框改變舊對話畫格時安全移除中文，保留一次 kind=2 診斷，不宣稱全程零失配 |
| 存讀檔、退出與速度提示中文 | 通過，限 prototype | `save-ui-v3/receipt.json`：逐行前綴核對、數字區保留、22 點覆繪區外 0。`save-ui-gui-v2/` 名稱／是／否／退出；提示修正建置 `save-ui-modal-v2/` 完整 F7 兩行、F9／3／Enter／Right，角色實際移動且提示清除。其他商店九列與彈窗下舊對話中文待完成 |
| 首次武器交易中文與裝備查看 | 部分通過，限已列正常路徑 | `quote-overlay-v2/receipt.json` 九點完整原版狀態相同、區外0；`shop-path-gui-quote-v3/` 完整中文報價、1000→800、告別及裝備查看通過。該次裝備仍英文；裝備中文另見下列項，未宣稱商店完整中文 |
| 原版清框與中文遮擋 | 通過，限 prototype | `shop-path-modal-trace-v1/` 購買9點／取消7點完整狀態相同；26次矩形操作完整返回、區外0。`modal-overlay-v2/` 驗舊中文未寫入彈窗，`modal-save-v2/` 驗來源鍵、遮擋及未知變化撤除。原版框仍可能遮掉部分中文字 |
| 第一筆武器報價中文 | 通過，限此正常購買 | `quote-overlay-v2/receipt.json` 九點完整原版狀態相同、區外0，400／200原版數字圖形逐像素相同，買後800；`shop-path-gui-quote-v3/` 實際視窗兩行中文、是／否與購後結果通過。四個模板來源及順序已驗；未知畫面退回。只驗此購買，不外推所有價格／商品 |
| 第二件武器資金不足 | 通過，限配對與離線目視 | `quote-overlay-v2/insufficient/receipt.json` 十一點完整原版狀態相同；不足兩頁、可操作選單、告別與回城鎮仍1000。離線圖中文清楚，沒有冒充實際視窗不足分支 |
| 裝備分類與練習劍兩行中文 | 通過，限已列配對與冷讀檔GUI | `inventory-overlay-v3/inventory/receipt.json` 十一點完整原版狀態相同、區外0；保留原版顏色與陰影，關閉整張覆繪差0，未知像素撤除。`shop-path-gui-inventory-v5/`冷F7讀自己新建存檔，初開／重開九個中文ROI相符，關閉城鎮靜態區相符，1000金幣、存檔不變、持鍵空、退出0；本次GUI沒有新交易、未開HD，不稱全畫面同狀態對拍。其他物品及操作尚未驗 |
| 戰鬥 | 未驗證 | 後續正常玩家切片 |
| 中文工具列與片頭單句 | 通過，限雛形 | 最新二進位重跑 `workplace/gui-smoke-v6/` 三條流程，原始色號仍相同。原文消失／不匹配／不改原圖測試通過 |
| 原版／中文區域對拍 | 通過，限片頭單句 | `intro-region-comparison.json`，原生色號一致；覆繪區外差異 0 |
| 首章 30 行捲動中文 | 通過，限 prototype | `intro-scroll-v2/receipt.json`，160m 自然片頭六點抽樣的 1 MiB 記憶體、CS:IP、cycles、色號及色盤均與無掛鉤原版相同；覆繪區外 0，結尾及換章移除通過 |
| 多行中文實際視窗 | 通過，限首章 | `gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`；冷開機自然片頭，擷取時 7 來源行／62 完成列；正式字級與術語待校訂 |
| 完整中文顯示與換頁 | 未完成 | 正式字串來源與其餘場景待 Z04 |
| HD 與原貌切換 | 通過，限雙側裝飾邊框試作 | `workplace/gui-hd-frame-v1/`，區內 172,017 像素不同、區外 0，原貌完整還原；完整左右失配、淡入色盤與短畫格退回測試通過 |
| 國王 HD 肖像與口部姿態 | 通過，限兩個已觀察姿態 | `gui-dialogue-hd-v5/king-hd-region-comparison.json`；肖像內部 52,452 像素改變，外框及區外 0，原貌完整還原。兩個姿態只改口部圖面，未知姿態／淡入／缺圖退回測試通過 |
| 長老 HD 肖像與動畫像素 | 通過，限兩個已觀察畫格 | `gui-sage-hd-v1/` 正常新局進房，同狀態切換內部 50,926 像素不同、區外 0、原貌完整還原；四個原版動畫像素不變、steps／cycles 不變。缺圖、未知畫格、色盤及短畫格退回測試通過 |
| 武器商人 HD | 通過，限三個已觀察姿態 | `merchant-hd-gui-v2/region-comparison.json`與獨立 `merchant-hd-v1/pixel-verification-v2.json`，三姿態52409／52396／52419像素改變，遊戲畫布內框外0差、原貌完整還原、暫停steps／cycles不變；原貌RGB與來源相同，HD與指定圖取樣相同。離店HD開啟但商人不匹配。來源、區外不寫入、缺圖／未知畫格／色盤／短畫格退回測試通過，不稱完整動畫 |
| 銀行 HD | 通過，限三個抽樣來源及GUI實遇兩姿態 | `bank-hd-v1/hd-test-v1.jsonl`三來源／退回測試通過；`bank-hd-v1/gui-v2/`正常冷F7進銀行，迎客pose2與告別pose1各52450像素改變，區外0、原貌完整還原、直接圖面取樣與來源RGB相符、paused steps／cycles不變。正常離店1000後HD開但不匹配，持鍵空、退出0；未驗完整動畫 |
| 武器店未購物離店中文 | 配對及實際視窗通過 | `merchant-hd-v1/no-purchase/receipt.json` 三點完整1MiB、色號、色盤、CS:IP、steps與cycles相同，區外0；18A17中文完整、金幣1000，離店整張覆繪差0。`merchant-hd-gui-v2/`驗B23E待鍵後離店1000，再正常進店交易及B1FD離店800。保留實際兩筆換場舊對話退回，不稱零失配 |
| 其餘場景／角色 HD | 未實作 | B 手繪奇幻已選定，不把邊框與已驗人物外推為整個遊戲 HD |
| 銀行存提款中文 | 配對及正常X視窗通過 | `bank-overlay-v1/v4/bank/receipt.json`35點完整1MiB、色號、色盤、CS:IP、steps／cycles相同，區外0、原版數字圖形保留，離店950且覆繪差0。`bank-gui-v1/`同handle冷F7、存100、領50、查50、無alma及正常離店950；128筆按放確認、最後持鍵空、capture-tail退出0，8個金額欄與3個GOLD區域同原版RGB。銀行ledger存檔持久性、餘額1、提空、有alma兌換及其他分支未驗 |
| 銀行未交易告別中文 | 配對及正常X視窗通過 | `item-overlay-v1/v2/bank-no-purchase/receipt.json`12點完整原版狀態相同、區外0，AC9B待鍵及離店1000、覆繪差0。`bank-hd-v1/gui-v2/`完整新中文及正常離店，60筆按放確認、退出0；保留離店舊文字安全退回，不稱零失配 |
| 道具店首瓶購買中文 | 配對及正常X視窗通過 | `item-overlay-v1/v2/item/receipt.json`32點完整1MiB、色號、色盤、CS:IP、steps／cycles相同、區外0，第一瓶即扣50、續購否、告別、離店950、物品頁及關閉。`item-overlay-v1/gui-v2/`11張實圖目視、88筆按放確認，50原版45×24數字圖形及六個金額／價格／藥水圖示ROI相同；持鍵空、退出0。NO USE仍原文，未使用藥水、未測其他商品／介紹／賣出；名稱為試譯 |
| 工具列與手把映射 | 部分通過，限雛形 | 實際滑鼠開啟設定頁；合併輸入／斷線釋放、軸中立與越界、設定存讀三項測試通過 |
| USB／藍牙實機 | 未執行 | 不以合成事件代替 |
| GitHub PRIVATE | 通過 | 主機 gh 登入成功，建立 `wicanr2/zeliard_cht` 並回讀 `visibility=PRIVATE` |

裝備輪的三項測試與35個配對抽樣點保留原有身份，紀錄為 `inventory-overlay-v3/test.jsonl`及 `inventory-regression-v1/test.jsonl`；該輪 `52826ff5`建置通過冷讀檔裝備GUI，不換名為本輪重跑。

前輪商人三姿態及未購物離店兩項測試通過，沒有具名skip；紀錄為 `merchant-hd-v1/hd-test-v2.jsonl`與 `no-purchase-test-v2.jsonl`。未購物3點完整原版狀態相同、區外0，go vet及建置通過；該輪48e9c0bc已通過三姿態GUI與兩條離店。未重跑未改動的其他人物HD與手把測試。以上不計為完整中文、實體手把、完整遊玩或全場景HD完成。

銀行存提款切片的fc797237建置保留歷史身份。`bank-overlay-v1/test-v4.jsonl`三個頂層測試與六個catalog拒絕子項全部通過，無具名skip；銀行35點、裝備11點配對及正常銀行GUI通過，go vet及建置通過。武器交易9點回歸取自v2的具名PASS，不稱該檔整批通過；v2／v3標題失敗與來源／顏色勘誤保留於WORKLOG及RESEARCH。銀行標題只在已驗來源使用底色5、字9及水平陰影45，未知畫格退回已驗。該次GUI未開HD、未另驗原版Ctrl+Q；ledger的新存檔恢復仍未驗。

目前主入口為dc2696ed整合候選。`item-overlay-v1/test-v2.jsonl`八個頂層及十一個catalog子項全部PASS、無skip，銀行35點、未交易12點、道具店32點、武器9點、裝備11點及商人未購物3點配對通過；go vet／建置通過。兩份新GUI以同一固定候選依序驗證，完整稽核在 `item-overlay-v1/current-audit.json`，70自有檔案、41來源與8鍵序、29原件、882快照及文件入口均核對。失敗GUI v1與後續來源補證保留原身份，不冒充成功。未重跑未改動的其他人物HD、手把或存讀檔GUI，未驗新平台或正式發行包。
