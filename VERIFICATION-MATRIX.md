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
| 存讀檔／戰鬥 | 未驗證 | 尚待 Z09 及後續切片 |
| 中文工具列與片頭單句 | 通過，限雛形 | 最新二進位重跑 `workplace/gui-smoke-v6/` 三條流程，原始色號仍相同。原文消失／不匹配／不改原圖測試通過 |
| 原版／中文區域對拍 | 通過，限片頭單句 | `intro-region-comparison.json`，原生色號一致；覆繪區外差異 0 |
| 首章 30 行捲動中文 | 通過，限 prototype | `intro-scroll-v2/receipt.json`，160m 自然片頭六點抽樣的 1 MiB 記憶體、CS:IP、cycles、色號及色盤均與無掛鉤原版相同；覆繪區外 0，結尾及換章移除通過 |
| 多行中文實際視窗 | 通過，限首章 | `gui-intro-scroll-v2/receipt.json` 與 `intro-scroll-window.png`；冷開機自然片頭，擷取時 7 來源行／62 完成列；正式字級與術語待校訂 |
| 完整中文顯示與換頁 | 未完成 | 正式字串來源與其餘場景待 Z04 |
| HD 與原貌切換 | 通過，限雙側裝飾邊框試作 | `workplace/gui-hd-frame-v1/`，區內 172,017 像素不同、區外 0，原貌完整還原；完整左右失配、淡入色盤與短畫格退回測試通過 |
| 國王 HD 肖像與口部姿態 | 通過，限兩個已觀察姿態 | `gui-dialogue-hd-v5/king-hd-region-comparison.json`；肖像內部 52,452 像素改變，外框及區外 0，原貌完整還原。兩個姿態只改口部圖面，未知姿態／淡入／缺圖退回測試通過 |
| 場景／角色 HD | 未實作 | B 手繪奇幻已選定，尚未將邊框外推為整個遊戲 HD |
| 工具列與手把映射 | 部分通過，限雛形 | 實際滑鼠開啟設定頁；合併輸入／斷線釋放、軸中立與越界、設定存讀三項測試通過 |
| USB／藍牙實機 | 未執行 | 不以合成事件代替 |
| GitHub PRIVATE | 通過 | 主機 gh 登入成功，建立 `wicanr2/zeliard_cht` 並回讀 `visibility=PRIVATE` |

本輪八項具名 Go 測試、建置與 go vet 通過，真實原版、字型及生成圖有掛載，無測試 skip。`bitmap` 套件沒有獨立測試。新建置通過完整國王初訪中文與 HD 的 X 視窗流程；前輪收據保留原二進位身份。以上不計為實體手把、完整遊玩或全場景 HD 完成。
