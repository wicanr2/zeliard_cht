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
| 首次武器交易中文與裝備查看 | 部分通過，限已列正常路徑 | 最新 `modal-overlay-v2/receipt.json` 九點完整原版狀態相同、區外 0；`shop-path-gui-modal-v1/` 商品框上方及報價旁舊中文保留，確認框差異 0，1000→800、告別及裝備查看通過。報價與裝備仍英文，保留一筆離店舊對話退回 |
| 原版清框與中文遮擋 | 通過，限 prototype | `shop-path-modal-trace-v1/` 購買9點／取消7點完整狀態相同；26次矩形操作完整返回、區外0。`modal-overlay-v2/` 驗舊中文未寫入彈窗，`modal-save-v2/` 驗來源鍵、遮擋及未知變化撤除。原版框仍可能遮掉部分中文字 |
| 戰鬥 | 未驗證 | 後續正常玩家切片 |
| 中文工具列與片頭單句 | 通過，限雛形 | 最新二進位重跑 `workplace/gui-smoke-v6/` 三條流程，原始色號仍相同。原文消失／不匹配／不改原圖測試通過 |
| 原版／中文區域對拍 | 通過，限片頭單句 | `intro-region-comparison.json`，原生色號一致；覆繪區外差異 0 |
| 首章 30 行捲動中文 | 通過，限 prototype | `intro-scroll-v2/receipt.json`，160m 自然片頭六點抽樣的 1 MiB 記憶體、CS:IP、cycles、色號及色盤均與無掛鉤原版相同；覆繪區外 0，結尾及換章移除通過 |
| 多行中文實際視窗 | 通過，限首章 | `gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`；冷開機自然片頭，擷取時 7 來源行／62 完成列；正式字級與術語待校訂 |
| 完整中文顯示與換頁 | 未完成 | 正式字串來源與其餘場景待 Z04 |
| HD 與原貌切換 | 通過，限雙側裝飾邊框試作 | `workplace/gui-hd-frame-v1/`，區內 172,017 像素不同、區外 0，原貌完整還原；完整左右失配、淡入色盤與短畫格退回測試通過 |
| 國王 HD 肖像與口部姿態 | 通過，限兩個已觀察姿態 | `gui-dialogue-hd-v5/king-hd-region-comparison.json`；肖像內部 52,452 像素改變，外框及區外 0，原貌完整還原。兩個姿態只改口部圖面，未知姿態／淡入／缺圖退回測試通過 |
| 長老 HD 肖像與動畫像素 | 通過，限兩個已觀察畫格 | `gui-sage-hd-v1/` 正常新局進房，同狀態切換內部 50,926 像素不同、區外 0、原貌完整還原；四個原版動畫像素不變、steps／cycles 不變。缺圖、未知畫格、色盤及短畫格退回測試通過 |
| 其餘場景／角色 HD | 未實作 | B 手繪奇幻已選定，不把邊框與兩個人物外推為整個遊戲 HD |
| 工具列與手把映射 | 部分通過，限雛形 | 實際滑鼠開啟設定頁；合併輸入／斷線釋放、軸中立與越界、設定存讀三項測試通過 |
| USB／藍牙實機 | 未執行 | 不以合成事件代替 |
| GitHub PRIVATE | 通過 | 主機 gh 登入成功，建立 `wicanr2/zeliard_cht` 並回讀 `visibility=PRIVATE` |

前輪十二項具名測試與兩次不同建置的長老 HD／交易 GUI 保留原有身份，未改名為新收據。本輪只重驗受影響五項：國王、長老、遮擋與商店在 `modal-overlay-v2/test.jsonl` 通過；該檔保留存讀檔舊數量斷言失敗，更新為必要來源各一次與遮擋檢查後，`modal-save-v2/test.jsonl` 通過。新存讀檔22點完整狀態、新檔及冷讀回相同。離線五張圖另驗輸入框、是／否及速度3；`sage-outside.png` 仍為房內告別，不能當已回城鎮的證據。go vet與建置通過，最新實際交易GUI使用 `c355ade5`。以上不計為完整中文、實體手把、完整遊玩或全場景HD完成。
