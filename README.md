# 神劍封魔 Zeliard

使用 dosgolem 直接執行 DOS《神劍封魔》，在原版畫面上疊加繁體中文與 HD，並加入可綁定藍牙、USB 手把的工具列。執行與覆繪參考《銀河超能力戰記》中文化專案，工具列參考《三國演義》重製版。

目前前端雛形包含片頭首章 30 行、國王與長老對話中文、存讀提示、武器店選單與兩個商品名，以及手繪 HD 邊框、國王與長老肖像。手把設定頁可選裝置並綁定按鍵。正常領賞、長老存讀檔與第一筆武器交易已有原版對照；其餘中文、場景與角色 HD、實體手把及戰鬥仍待完成或驗證，尚未提供正式遊玩版。最新進度見 [CONTEXT.md](CONTEXT.md)，驗證範圍見 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md)。

## 故事背景

公主被邪惡的魔龍化為石像，王國的繁榮也受到威脅。為了拯救公主，玩家接受邪靈的挑戰，帶著人民的期待與神靈的祝福，踏入未知的地下迷宮。城鎮裡的人們會提供消息，商店則讓旅人補充裝備與魔法，為接下來的探索作準備。

以上改寫自本地中文手冊《神劍封魔》第 1 頁「前言」及第 7–8 頁「開始遊戲」。來源為 `珍084-神劍封魔.rar` 內的 `珍084-神劍封魔/SCAN0357_003.jpg` 與 `SCAN0357_006.jpg`，已人工核對掃描圖。手冊來源與雜湊見 [RESEARCH-LOG.md](RESEARCH-LOG.md)。掃描檔不隨儲存庫散布。

## 開發方向

- 以固定版本的原版程式為行為依據，使用 dosgolem 建立可重播的啟動、操作與畫面證據。
- 中文譯名參考當年中文手冊；文字來源與排版先驗證再接入。
- 原貌保留獨立基準，HD 採手繪奇幻，以獨立圖面覆繪，不改遊戲規則。正式倍率、構圖與預設尚待確認。
- 工具列提供手把裝置選擇、按鍵綁定、設定保存與斷線處理。藍牙配對由作業系統完成。

## 本地資料

遊戲與手冊保留在本機，未納入 Git。遊戲包來自使用者指定的[線上 DOS 遊戲頁面](https://legacy.dos.zczc.cz/games/%E7%A5%9E%E5%89%91%E5%B0%81%E9%AD%94/)。執行與測試需要合法持有的原版資料，來源清單見 [CONTEXT.md](CONTEXT.md)。

分析、建置與測試使用 Docker。開工規則與文件入口見 [AGENTS.md](AGENTS.md)，實際命令與結果記在 [WORKLOG.md](WORKLOG.md)。

## 前端雛形

入口為 [cmd/zeliard-preview](cmd/zeliard-preview/main.go)。提供本地手繪試作圖時，工具列可切換原貌與 HD 圖面；沒有素材時標示待製作。手把需先由系統完成配對或辨識，再到工具列選擇裝置並逐項綁定；尚未進行實體手把驗收。

建置後在 Docker 內執行：

```sh
workplace/bin/zeliard-preview \
  -orig /orig -font /fonts/unifont.hex.gz \
  -text /project/text/host.json \
  -user-dir /project/workplace/frontend-user
```

`/orig` 須唯讀掛載完整原版資料，`/fonts` 為使用者提供的字型目錄。原版存檔另落在使用者目錄下的 `game/`，手把設定獨立保存。雛形尚未接音效輸出。

首章捲動中文試作加上 `-intro-scroll /project/text/intro-scroll-prototype.tsv`，會沿原版事件顯示 30 行試譯。專名與正式字級尚待校訂，其餘片頭及遊戲文字保留原文。舊單句試作仍可用 `-intro-reference` 指向本地重生的原版畫格。方法及限制見[架構與試作規格](docs/spec/001-runtime-overlay.md)。

HD 試作另加 `-hd-reference /project/workplace/probe-newgame-60m/screen.bin -hd-art /project/workplace/hd-border-prototype.png`。參考畫格同目錄須有 `palette.bin`。試作圖只保存在本機；雛形檢查固定圖像身份後，僅替換兩側裝飾，不遮擋中間場景或資訊。缺圖、換場或色盤不符時保留原貌。這些選項與正式資產格式仍在試作階段。

國王與長老對話中文另加 `-dialogue-text /project/text/dialogue-prototype.tsv`，選單及提示加 `-labels-text /project/text/labels-prototype.tsv`。出字、等待按鍵、賞金與存讀檔仍由原版處理。存檔到長老房屋選「記錄歷程」，讀檔使用 F7；名稱輸入提示、是／否、退出與讀取確認已有中文試譯，檔名及 `Re-Start` 保留原值。商店選項已有試譯，尚未逐店驗證。

國王 HD 肖像使用 `-king-reference-a`、`-king-reference-b`、`-king-art`、`-king-speaking-art` 載入本地兩個原版姿態及兩張手繪圖。它保留黃色外框，未知姿態退回原貌；完整參數與限制見[試作規格](docs/spec/001-runtime-overlay.md)。

長老 HD 另用 `-sage-reference-a`、`-sage-reference-b`、`-sage-palette`、`-sage-art` 載入本地兩個原版畫格、色盤及手繪圖。它保留黃色外框與四個原版動畫像素。武器店目前已接初訪、交易後詢問、告別、五列選單與「練習劍／賢者之劍」試譯。商品框上方及確認框旁的舊中文可保留，原版彈窗仍會遮住部分字。動態報價、裝備頁及其他商品仍待中文化，價格與交易規則由原版處理。

## 授權

自有程式與文件採 [RRSAL-1.0](LICENSE)，屬 source-available 專案。非商業用途免費，允許實況與影片分潤，商業使用另洽。原版遊戲、手冊、音樂、美術及字型不在此授權範圍內。第三方元件依各自授權使用。
