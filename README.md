# 神劍封魔 Zeliard

使用 dosgolem 直接執行 DOS《神劍封魔》，在原版畫面上疊加繁體中文與 HD，並加入可綁定藍牙、USB 手把的工具列。執行與覆繪參考《銀河超能力戰記》中文化專案，工具列參考《三國演義》重製版。

目前前端雛形包含片頭首章 30 行與魔王三行宣言、國王與長老對話中文、存讀提示、武器店、銀行存提款、道具店買賣與三種商品介紹，以及部分裝備名稱、藥水使用訊息與居民交談。生命、寶石、金幣、地點四個狀態欄，以及城堡、城鎮、洞窟、國王、武器店、小屋與Marid長老名稱也已接入。手繪 HD 包含邊框、國王、長老、武器商人、銀行與道具店肖像、城堡／Muralla鎮的完整遠山，以及Muralla樹木、建物與岩壁前景試作。Muralla地面與主角八個行走姿態也已接入，居民保留原貌；武器、魔法與銀行招牌可顯示中文。手把設定頁可選裝置並綁定按鍵。正常領賞、存讀檔、首筆武器交易、裝備開關、銀行存提款、首瓶藥水買賣與使用、六頁商品介紹及狀態列切片已有原版對照及正常視窗驗證，HD 切換只驗已觀察姿態。首場交戰、第一顆寶石及洞窟中文狀態列也已通過正常視窗驗證。其餘中文、場景與角色 HD、音效的穩定播放、實體手把及後續戰鬥仍待完成或驗證，尚未提供正式遊玩版。最新進度見 [CONTEXT.md](CONTEXT.md)，驗證範圍見 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md)。

故事正文58片段、47頁中文，以及51行製作名單與六行收尾試譯已接入，正常播放、捲動與撤除通過原版配對、實際視窗及獨立排版核對。故事正文另已驗Space跳過；片頭美術仍為原貌，正式字級與專名待校訂。

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

Malicia 洞窟地形試作另加 `-cave-art-dir /project/workplace/cave-hd-v1`，目錄須含身份相符的 `terrain-handpaint-v2.png` 與 `rock-material-v1.png`。目前替換已確認的地形1..14。正常入洞、首戰拾取、跳蹲、原貌切換及回城已抽驗；原貌與HD共用原版完成畫格。未知來源、缺圖或整區重建不符時保留原貌。素材僅存本機，範圍見[洞窟地形規格](docs/spec/011-cave-hd.md)及[完成畫格規格](docs/spec/012-cave-frame-sources.md)。

洞窟主角另加 `-cave-player-art-dir /project/workplace/cave-hd-v1`，搭配上述地形選項。目錄需含八張本地固定圖集：

- `player-atlas-handpaint-v1.png`
- `player-motion-atlas-handpaint-v1.png`
- `player-gui-atlas-handpaint-v1.png`
- `player-left-atlas-handpaint-v2.png`
- `player-motion-extra-handpaint-v1.png`
- `player-expedition-atlas-handpaint-v1.png`
- `player-shield-atlas-handpaint-v2.png`
- `player-shield-transition-handpaint-v1.png`

已接55個來源姿態，包含17個一般持盾姿態及雙朝向共用的背面離洞姿態。直接高清合成保留原版位置與後畫遮擋；未知來源、配色與未驗特殊組合保留原貌。正常離洞、回鎮撤除與重入已抽驗，原始遊戲與存檔不改，來源存檔取得過程未知。圖集只留本機，範圍見[主角規格](docs/spec/017-cave-player-hd.md)、[探索補圖](docs/spec/023-cave-expedition-player-hd.md)、[一般持盾](docs/spec/027-cave-shield-profile-hd.md)與[持盾離洞](docs/spec/028-cave-shield-transition-hd.md)。

洞窟武器另加 `-cave-weapon-art-dir /project/workplace/cave-hd-v1`，搭配地形選項，載入本地 `weapon-atlas-handpaint-v1.png`。十個已確認來源直接以高清合成，保留原版包絡、方向與最後圖層；收劍即撤除，未知來源維持原貌。主角可用原貌或HD，未提供物件圖組時動態物件仍為原貌。範圍見[洞窟武器規格](docs/spec/018-cave-weapon-hd.md)，不代表全部武器或洞窟已完成。

洞窟動態物件另加 `-cave-object-art-dir /project/workplace/cave-hd-v1`，搭配地形選項，載入 `object-red-handpaint-v1.png`、`object-green-handpaint-v1.png`、`object-effects-handpaint-v1.png`、`object-green-extra-handpaint-v1.png` 四張本地固定圖。29個已確認圖像來源按原版順序合成，保留清黑、部分出界裁切及主角／武器前後層。缺圖、未知鍵、不同配色或非零底圖保留原貌。這些圖不入儲存庫，範圍見[洞窟物件規格](docs/spec/019-cave-object-hd.md)，其餘物件與洞窟仍待完成。

Malicia通往Muralla鎮的入口另加 `-cave-portal-art /project/workplace/cave-hd-v1/portal-handpaint-v1.png`，搭配地形選項。拱門直接以高清合成，人物、武器與物件仍依原版順序遮擋。搭配 `-hud-text /project/text/hud-prototype.tsv` 時，圖牌沿用「Muralla鎮」；未載譯表時顯示MURALLA。完整入口矩陣與資源不符時保留原圖，不套用其他入口。圖面只留本地，範圍見[入口規格](docs/spec/020-cave-portal-hd.md)。

建置後在 Docker 內執行：

```sh
workplace/bin/zeliard-preview \
  -orig /orig -font /fonts/unifont.hex.gz \
  -text /project/text/host.json \
  -user-dir /project/workplace/frontend-user
```

`/orig` 須唯讀掛載完整原版資料，`/fonts` 為使用者提供的字型目錄。原版存檔另落在使用者目錄下的 `game/`，手把設定獨立保存。音效雛形可加 `-audio ebiten`，預設為 `-audio off`。暫停、失焦及手把設定頁會關閉當次播放，恢復時建立新佇列。現有視窗仍有欠載與佇列滿載，聲音可能插入空白或丟樣；穩定節奏、人耳與實體裝置仍待驗證，見[音訊規格](docs/spec/029-runtime-audio.md)。

研究用 `-audio-wav /project/workplace/<新檔名>.wav` 保存原版生成的單聲道取樣，拒絕覆寫已有檔案。這份檔案不代表裝置實際播放的聲音，須另做裝置監聽。

片頭中文可依需要載入對應的試譯TSV：

| 內容 | Docker內參數 | 範圍 |
|---|---|---|
| 首章捲字 | `-intro-scroll /project/text/intro-scroll-prototype.tsv` | 30行 |
| 魔王宣言 | `-intro-threat /project/text/intro-threat-prototype.tsv` | 三行逐字與淡出 |
| 故事正文 | `-intro-story /project/text/intro-story-prototype.tsv` | 58片段、47頁 |
| 製作名單 | `-intro-credits /project/text/intro-credits-prototype.tsv` | 51行，保留人名、公司、年份及空白行 |
| 收尾捲字 | `-intro-outro /project/text/intro-outro-prototype.tsv` | 六行，使用原版較高視窗 |

五組可以同時載入。原版負責出字、捲動、字色、清框、等待及Space輸入，中文只改顯示。來源、畫面或字型不符時保留原貌；專名與正式字級仍待校訂。方法見[魔王宣言規格](docs/spec/013-intro-narration.md)、[故事覆繪規格](docs/spec/014-intro-story.md)及[三組捲字規格](docs/spec/015-intro-scroll-blocks.md)。其餘遊戲文字仍待完成。舊單句試作仍可用 `-intro-reference` 指向本地重生的原版畫格，方法及限制見[架構與試作規格](docs/spec/001-runtime-overlay.md)。

HD 試作另加 `-hd-reference /project/workplace/probe-newgame-60m/screen.bin -hd-art /project/workplace/hd-border-prototype.png`。參考畫格同目錄須有 `palette.bin`。試作圖只保存在本機；雛形檢查固定圖像身份後，僅替換兩側裝飾，不遮擋中間場景或資訊。缺圖、換場或色盤不符時保留原貌。這些選項與正式資產格式仍在試作階段。

國王與長老對話中文另加 `-dialogue-text /project/text/dialogue-prototype.tsv`，選單及提示加 `-labels-text /project/text/labels-prototype.tsv`。出字、等待按鍵、賞金與存讀檔仍由原版處理。存檔到長老房屋選「記錄歷程」，讀檔使用 F7；名稱輸入提示、是／否、退出與讀取確認已有中文試譯，檔名及 `Re-Start` 保留原值。商店選項已有試譯，尚未逐店驗證。

國王 HD 肖像使用 `-king-reference-a`、`-king-reference-b`、`-king-art`、`-king-speaking-art` 載入本地兩個原版姿態及兩張手繪圖。它保留黃色外框，未知姿態退回原貌；完整參數與限制見[試作規格](docs/spec/001-runtime-overlay.md)。

長老 HD 另用 `-sage-reference-a`、`-sage-reference-b`、`-sage-palette`、`-sage-art` 載入本地兩個原版畫格、色盤及手繪圖。它支援已驗左側及昏迷送返時的置中位置，保留黃色外框與四個原版動畫像素。送返的四段中文、前文保留、待續、捲動及回鎮撤除已驗，見[送返覆繪規格](docs/spec/021-exhaustion-return-overlay.md)。送返後可恢復城鎮山景、前景、地面與主角HD，見[城鎮來源規格](docs/spec/022-return-town-hd.md)。武器店目前已接初訪、交易後詢問、告別、五列選單與「練習劍／賢者之劍」試譯。商品框上方及確認框旁的舊中文可保留，原版彈窗仍會遮住部分字。

同時載入對話與標籤 TSV 時，第一筆武器報價可顯示中文，售價與折抵保留原版數字圖形。各段等原版完成後才出現，未完成時暫留白。選第二件武器的資金不足對話與翻頁也已接入。Return開啟的裝備頁已有分類標題、「無」及練習劍兩行試譯，保留原版標題顏色與陰影。其他報價、物品、魔法與場景名稱仍待中文化，交易與裝備規則由原版處理。

武器商人HD需三個原版畫格、色盤與三張本地手繪圖，入口見[商人覆繪規格](docs/spec/001-runtime-overlay.md#武器商人三姿態手繪肖像)。目前支援迎客、交叉雙臂與未購物告別姿態，跟隨原版畫格切換；缺圖、未知畫格或色盤不符時保留原貌。未購物離店另有中文告別，金幣與離店條件由原版處理。

銀行已有標題、選單、金額欄、存錢、領錢、查詢、無alma寶石及交易後／未交易告別試譯。金額沿原版方向鍵調整：上加1、下減1、左加10、右減10，以Space確認。存提款問句會保留，數字使用原版字形。存款50金幣的新存檔冷讀已通過原版與中文配對，這段存讀流程仍待實際視窗驗證；其餘銀行分支未驗，見[銀行試作規格](docs/spec/001-runtime-overlay.md#銀行交易與標題-prototype)。

銀行HD使用三組本地原版畫格與手繪圖，入口為 `-bank-reference-a`、`-bank-reference-b`、`-bank-reference-c`、`-bank-palette`及 `-bank-art-a`、`-bank-art-b`、`-bank-art-c`。它依已知畫格切換，保留外框；未知姿態或缺圖保留原貌。素材與參數對應見[銀行肖像規格](docs/spec/001-runtime-overlay.md#銀行未交易告別與手繪肖像-prototype)，不代表完整銀行動畫。

魔法物品店已有標題、四選項、三個商品名稱、六頁商品介紹及首瓶藥水買賣中文。選取第一瓶即由原版扣除50金幣，後面的「是／否」是是否繼續購物；出售的25金幣由原版確認後入帳，物品頁可查看買入及賣出結果。價格保留原版數字圖形，商品名為試譯，見[道具店規格](docs/spec/001-runtime-overlay.md#道具店介紹出售與手繪畫格-prototype)。

物品頁另加入「不用」文字、小圖示及Ken’ko藥水兩行名稱。正常進入洞窟後可按Enter開物品頁、Right選藥水、Space使用；原版處理消耗與使用訊息。48點原版配對與正常視窗已驗使用、重開空欄及返回城鎮。該次原版生命已滿，只驗消耗與顯示；藥水受傷恢復量及未列物品分支尚未驗，見[道具使用規格](docs/spec/001-runtime-overlay.md#洞窟道具使用與物品頁文字-prototype)。

共用物品文字已依原版指標表分類，新增60筆名稱及使用訊息候選。Magia石的正常買入、選用、使用、重開空欄及關閉已通過原版配對、獨立中文像素與實際視窗驗證。原包存檔正常F7後的刀劍油使用，以及已列刀劍、盾牌名稱也已驗證，原版數字保持原貌；來源存檔的取得過程未知。其他物品、裝備與魔法仍待驗，譯名仍待校訂，見[共用物品文字](docs/spec/016-shared-inventory-text.md)與[刀劍文字及刀劍油](docs/spec/025-sword-inventory-label.md)規格。

道具店 HD 使用 `-item-reference-0`至 `-item-reference-3`、`-item-palette`及 `-item-art-0`至 `-item-art-3`載入本地四組原圖與手繪圖。人物與蒸氣依已知原版畫格合成，保留黃色外框；未知畫格或缺圖退回原貌。四組抽樣來源不代表完整動畫，素材與對應見[道具店覆繪規格](docs/spec/001-runtime-overlay.md#道具店介紹出售與手繪畫格-prototype)。

狀態列中文另加 `-hud-text /project/text/hud-prototype.tsv`。已驗四欄、Felishika城堡、Muralla鎮、Malicia洞窟，以及國王、武器防具店、小屋、Marid長老名稱，保留原版數字、生命條與圖示。53筆來源中，其餘40筆新增名稱仍待正常路徑抽樣；只有實際入口、模組、原文與完成字形都相符才覆繪。專名保留原文，譯文仍待校訂，見[狀態列規格](docs/spec/001-runtime-overlay.md#場景來源擴充契約)。

城鎮居民試譯另加 `-town-text /project/text/town-prototype.tsv`。已接113筆純文字來源，其中Muralla鎮七筆來源的完整交談、續頁與結束已通過原版配對及正常視窗驗證。其餘106筆待正常抽樣，含特殊控制碼的5筆保留原文。來源筆數不等於獨立居民人數。對話框沿原版行數變化，續頁箭頭及Space操作不變，來源或畫面不符時保留原文。試作邊界見[居民對話規格](docs/spec/004-town-dialogue.md)。

城堡與Muralla鎮的完整遠山 HD 使用 `-town-sky-reference`、`-town-sky-palette`、`-town-background-memory`、`-town-background-art`載入本地原版畫格、色盤、正常記憶體及手繪圖。背景依原版樹梢遮罩合成，保留前景輪廓；正常走路、居民中文及原貌切換已驗，矩形切口已消除，見[完整背景規格](docs/spec/006-town-background-mask.md)。舊 `-town-sky-art`可單獨重現上方54列的[前輪試作](docs/spec/005-town-skyline-hd.md)。

Muralla前景另加 `-town-scenery-art-dir`，目錄內須有本地 `town-left-art-v2.png`、`town-middle-art-v2.png`及 `town-right-art-v2.png`，並提供上述正常記憶體與色盤。加 `-town-signs-text /project/text/town-signs-prototype.tsv`顯示三個中文招牌。圖面沿原版地圖位置合成，保留主角與居民；換場、對話或來源失配時整層退回。行走時以原版完成點的完整畫格合成，重讀後恢復、三招牌、正常入洞與居民兩頁已驗。原貌與HD共用同一呈現來源；當次原版畫格及來源時間另存，方法見[完整畫格規格](docs/spec/008-town-frame-presentation.md)。部分原版像素邊緣仍可見，居民保留原貌，主角另有下述可選覆繪。素材身份、入口映射與驗證範圍見[前景規格](docs/spec/007-town-scenery-hd.md)。

Muralla地面另加 `-town-ground-art /project/workplace/town-hd-v1/foreground-v1/ground-handpaint-v1.png`，並將上述 `-town-sky-reference`指定為本地 `foreground-v1/normal-v1/restored.indexed`的完整camera153畫格。圖面按原版上下兩層的速度捲動；正常行走、原貌切換與還原、重讀、居民交談後恢復及進洞撤除已有視窗驗證。缺圖或來源失配時保留原貌。固定參考身份與驗證邊界見[地面規格](docs/spec/009-town-ground-hd.md)；其餘地面與洞窟尚未HD。

Muralla主角試作新增可選 `-town-player-art-dir /project/workplace/town-hd-v1/foreground-v1`，搭配上述前景HD與八張固定身份的本地手繪姿態圖。先還原原版主角下層，再直接合成高清圖面，保留原版位置與規則。範圍限已觀察的兩方向行走八姿態，其他姿態保留原貌；正常行走、再次讀檔、切換還原、居民交談後恢復及入洞撤除已有視窗驗證。素材與範圍見[主角規格](docs/spec/010-town-player-hd.md)。

## 授權

自有程式與文件採 [RRSAL-1.0](LICENSE)，屬 source-available 專案。非商業用途免費，允許實況與影片分潤，商業使用另洽。原版遊戲、手冊、音樂、美術及字型不在此授權範圍內。第三方元件依各自授權使用。
