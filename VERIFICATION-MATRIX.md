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
| 正常進門／國王首段對話 | 通過，限此切片 | `workplace/probe-king-0-84m/` 與 `gui-king-v2/`；正常新局與實際視窗進門，原文逐像素相符，按 Space 後金額更新 1000 |
| 對話結束／存讀檔／戰鬥 | 未驗證 | 尚待 Z09 及後續切片；首段文字與賞金確認不代替完整對話流程 |
| 中文工具列與片頭單句 | 通過，限雛形 | 最新二進位重跑 `workplace/gui-smoke-v6/` 三條流程，原始色號仍相同。原文消失／不匹配／不改原圖測試通過 |
| 原版／中文區域對拍 | 通過，限片頭單句 | `intro-region-comparison.json`，原生色號一致；覆繪區外差異 0 |
| 首章 30 行捲動中文 | 通過，限 prototype | `intro-scroll-v2/receipt.json`，160m 自然片頭六點抽樣的 1 MiB 記憶體、CS:IP、cycles、色號及色盤均與無掛鉤原版相同；覆繪區外 0，結尾及換章移除通過 |
| 多行中文實際視窗 | 通過，限首章 | `gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`；冷開機自然片頭，擷取時 7 來源行／62 完成列；正式字級與術語待校訂 |
| 完整中文顯示與換頁 | 未完成 | 正式字串來源與其餘場景待 Z04 |
| HD 與原貌切換 | 通過，限雙側裝飾邊框試作 | `workplace/gui-hd-frame-v1/`，區內 172,017 像素不同、區外 0，原貌完整還原；完整左右失配、淡入色盤與短畫格退回測試通過 |
| 場景／角色 HD | 未實作 | B 手繪奇幻已選定，尚未將邊框外推為整個遊戲 HD |
| 工具列與手把映射 | 部分通過，限雛形 | 實際滑鼠開啟設定頁；合併輸入／斷線釋放、軸中立與越界、設定存讀三項測試通過 |
| USB／藍牙實機 | 未執行 | 不以合成事件代替 |
| GitHub PRIVATE | 通過 | 主機 gh 登入成功，建立 `wicanr2/zeliard_cht` 並回讀 `visibility=PRIVATE` |

本輪六項 Go 測試、建置及 go vet 通過，真實原版、字型、生成圖均有掛載，未 skip。新建置通過首章中文與國王進門兩條 GUI 流程。前輪三條 GUI 啟動與城鎮 HD 切換收據保留原二進位身份，不當作本輪重跑。以上不計為實體手把、完整遊玩或全場景 HD 完成。
