# 神劍封魔 Zeliard

使用 dosgolem 直接執行 DOS《神劍封魔》，在原版畫面上疊加繁體中文與 HD，並加入可綁定藍牙、USB 手把的工具列。執行與覆繪參考《銀河超能力戰記》中文化專案，工具列參考《三國演義》重製版。

目前前端雛形已驗證從冷開機進入城鎮並移動，另有中文手把設定頁與片頭單句覆繪試作。完整中文、HD 美術、實體手把、對話、存讀檔與戰鬥仍待完成或驗證，尚未提供正式遊玩版。最新進度見 [CONTEXT.md](CONTEXT.md)，驗證範圍見 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md)。

## 故事背景

公主被邪惡的魔龍化為石像，王國的繁榮也受到威脅。為了拯救公主，玩家接受邪靈的挑戰，帶著人民的期待與神靈的祝福，踏入未知的地下迷宮。城鎮裡的人們會提供消息，商店則讓旅人補充裝備與魔法，為接下來的探索作準備。

以上改寫自本地中文手冊《神劍封魔》第 1 頁「前言」及第 7–8 頁「開始遊戲」。來源為 `珍084-神劍封魔.rar` 內的 `珍084-神劍封魔/SCAN0357_003.jpg` 與 `SCAN0357_006.jpg`，已人工核對掃描圖。手冊來源與雜湊見 [RESEARCH-LOG.md](RESEARCH-LOG.md)。掃描檔不隨儲存庫散布。

## 開發方向

- 以固定版本的原版程式為行為依據，使用 dosgolem 建立可重播的啟動、操作與畫面證據。
- 中文譯名參考當年中文手冊；文字來源與排版先驗證再接入。
- 原貌保留獨立基準，HD 以獨立圖面覆繪，不改遊戲規則。畫風與倍率尚待確認。
- 工具列提供手把裝置選擇、按鍵綁定、設定保存與斷線處理。藍牙配對由作業系統完成。

## 本地資料

遊戲與手冊保留在本機，未納入 Git。遊戲包來自使用者指定的[線上 DOS 遊戲頁面](https://legacy.dos.zczc.cz/games/%E7%A5%9E%E5%89%91%E5%B0%81%E9%AD%94/)。執行與測試需要合法持有的原版資料，來源清單見 [CONTEXT.md](CONTEXT.md)。

分析、建置與測試使用 Docker。開工規則與文件入口見 [AGENTS.md](AGENTS.md)，實際命令與結果記在 [WORKLOG.md](WORKLOG.md)。

## 前端雛形

入口為 [cmd/zeliard-preview](cmd/zeliard-preview/main.go)。畫面中的 HD 選項目前標示待製作，不把原版放大稱為 HD。手把需先由系統完成配對或辨識，再到工具列選擇裝置並逐項綁定；尚未進行實體手把驗收。

建置後在 Docker 內執行：

```sh
workplace/bin/zeliard-preview \
  -orig /orig -font /fonts/unifont.hex.gz \
  -text /project/text/host.json \
  -user-dir /project/workplace/frontend-user
```

`/orig` 須唯讀掛載完整原版資料，`/fonts` 為使用者提供的字型目錄。原版存檔另落在使用者目錄下的 `game/`，手把設定獨立保存。雛形尚未接音效輸出。

可選的片頭單句試作用 `-intro-reference` 指向本地重生的原版畫格，方法及限制見[架構與試作規格](docs/spec/001-runtime-overlay.md)。此選項只驗證覆繪機制，不代表全遊戲已翻譯。

## 授權

自有程式與文件採 [RRSAL-1.0](LICENSE)，屬 source-available 專案。非商業用途免費，允許實況與影片分潤，商業使用另洽。原版遊戲、手冊、音樂、美術及字型不在此授權範圍內。第三方元件依各自授權使用。
