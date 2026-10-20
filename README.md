# 神劍封魔 Zeliard

使用 dosgolem 直接執行 DOS《神劍封魔》，在原版畫面上疊加繁體中文與 HD，並加入可綁定藍牙、USB 手把的工具列。執行與覆繪參考《銀河超能力戰記》中文化專案，工具列參考《三國演義》重製版。

目前正擴充全遊戲 HD，尚未覆蓋所有城鎮與洞窟。候選版已新增城鎮居民、城堡前景、四姿態旗幟與小屋石像，旅館圖面已接入但正常到達仍待驗。完整範圍見 [規格034](docs/spec/034-full-game-hd.md)，已驗與待核項目見 [覆蓋清單](docs/hd-coverage.json)。以下封包與截圖仍代表已交付的第一洞窟試玩範圍。

已提供 Windows x86_64、macOS Intel／Apple Silicon 通用版與 Linux x86_64 AppImage。完整資料包包含原版遊戲及目前完成的繁體中文、手繪 HD 素材。驗收依使用者指定，以第一洞窟可玩及代表性對話、HD、工具列操作抽測為準，不宣稱全遊戲翻譯或美術完成。

配樂改用原版 AdLib 驅動與可替換的 Nuked-OPL3 音源，保留原曲。封包採實際時間推進，卡頓時最多追趕100毫秒，暫停及失焦不追趕。實體USB／藍牙手把、Windows與macOS真機尚未驗收；Windows已通過Wine輔助啟動，macOS已驗通用架構及封包結構。詳見 [VERIFICATION-MATRIX.md](VERIFICATION-MATRIX.md)。

## 故事背景

公主被邪惡的魔龍化為石像，王國的繁榮也受到威脅。為了拯救公主，玩家接受邪靈的挑戰，帶著人民的期待與神靈的祝福，踏入未知的地下迷宮。

冒險從城鎮開始。向居民打聽消息，到商店補充裝備與魔法，再進入迷宮探索。城鎮裡的準備與地下的戰鬥，構成這段救援旅程。

故事與探索描述改寫自當年中文手冊《神劍封魔》第 1 頁「前言」及第 7、8 頁「開始遊戲」。來源為 `珍084-神劍封魔.rar` 內的 `珍084-神劍封魔/SCAN0357_003.jpg` 與 `SCAN0357_006.jpg`，已人工核對掃描圖。來源與雜湊見 [手冊證據](RESEARCH-LOG.md#2026-10-18輸入盤點與中文手冊初讀)，掃描檔留在本機。

## 原版與 HD 畫面

![Muralla 城鎮與教堂的原貌、手繪 HD 對照](docs/images/original-hd-comparison.jpg)

左側為原版美術，右側為手繪 HD。每組取自同一暫停狀態的實際視窗，兩側皆開啟繁體中文覆繪。教堂圖展示當時的肖像覆繪，邊框與狀態列仍保留原貌；這些截圖只呈現已驗證的場景。

![目前封包的城鎮與第一洞窟 HD 實際畫面](docs/images/gameplay-hd.jpg)

目前封包的 Muralla 城鎮與 Malicia 第一洞窟實際截圖，保留工具列與狀態列。可從工具列切換原貌／HD、選擇手把及設定鍵盤。

## 素材總覽

以下兩張圖以相同的 16 個類別排列，展示目前接入的原版來源與 HD 圖面。人物各取一個姿態，主角、武器及動態物件採圖集抽樣；城鎮地面與洞窟岩壁的原版圖取自場景。這是目前成果的總覽，不代表全遊戲素材已完成。

### 原版素材總覽

![原版素材總覽：人物、城鎮、主角、武器、動態物件與洞窟](docs/images/original-assets-overview.jpg)

### HD 素材總覽

![手繪 HD 素材總覽：與原版總覽相同的 16 個類別](docs/images/hd-assets-overview.jpg)

圖面來源、裁切範圍與 SHA-256 保存於 [圖片來源清單](docs/images/provenance.json)，由 [製圖工具](tools/readme_images.py) 從本機素材重建。本頁圖片是依使用者要求加入的文件預覽；原版與衍生美術不適用自有程式的授權，完整解析度素材仍留在本機。

## 開始遊玩

本機封包放在 `dist-all/v.0.2.0-20261020/full-local/`，尚未上傳GitHub：

| 平台 | 檔案 |
|---|---|
| Linux x86_64 | `Zeliard-v.0.2.0-20261020-Linux-x86_64.AppImage` |
| Windows x86_64 | `Zeliard-v.0.2.0-20261020-Windows-x86_64.zip` |
| macOS 通用版 | `Zeliard-v.0.2.0-20261020-macOS-universal.zip` |

Windows解壓縮後開啟 `Zeliard.exe`，macOS解壓縮後開啟 `Zeliard.app`。Linux可直接執行AppImage，也可在專案目錄執行：

```sh
bash tools/play.sh
```

原生封包不需要Docker。若Linux缺少FUSE，可在新目錄以AppImage的 `--appimage-extract` 解開後執行 `squashfs-root/AppRun`，這條啟動路徑已抽測。

工具列可切換原貌／HD，並開啟「搖桿／手把」及「鍵盤設定」。藍牙先在系統配對，USB手把先插入，再選裝置及綁定。鍵盤可重綁、解除、還原預設，兩種設定都會保存。

| 預設按鍵 | 操作 |
|---|---|
| 方向鍵 | 移動、跳躍、蹲下及選單 |
| 空白鍵 | 攻擊／交談 |
| Enter | 裝備頁 |
| Alt | 法術 |
| Esc | 原版暫停；設定頁內取消／返回 |
| F1／F2 | 原版音樂／音效開關 |
| F12 | 鍵盤設定 |
| Ctrl+Q，再按Y | 正常離開 |

包內附自建存檔SAVETEST，可按F7讀取後從Muralla鎮向右進入第一洞窟。存檔與操作設定保存在系統使用者設定目錄的 `Zeliard-CHT`：Linux為 `~/.config/Zeliard-CHT/`，Windows為 `%AppData%\Zeliard-CHT\`，macOS為 `~/Library/Application Support/Zeliard-CHT/`。可用 `ZELIARD_USER_DIR` 指定其他目錄。

60秒推廣影片在 `dist-all/v.0.2.0-20261020/promo/`。檔案大小與SHA-256見同版 `SHA256SUMS.json`，抽測見 `smoke/`。封包與影片包含原版資料、音樂、字型及研究圖面，只供本機使用，不得上傳或散布。

建置入口為 [tools/build-release.sh](tools/build-release.sh)，使用Docker與已固定的本機原版、字型及HD素材，拒絕覆寫既有輸出。版號由 [VERSION](VERSION) 指定；目前狀態見 [CONTEXT.md](CONTEXT.md)。

## 開發方向

- 以固定版本的原版程式為行為依據，使用 dosgolem 建立可重播的啟動、操作與畫面證據。
- 中文譯名參考當年中文手冊；文字來源與排版先驗證再接入。
- 原貌保留獨立基準，HD沿既有3倍畫布，以手繪圖面覆繪，不改遊戲規則。
- 工具列提供手把裝置選擇、按鍵綁定、設定保存與斷線處理。藍牙配對由作業系統完成。

## 本地資料

遊戲與手冊保留在本機，未納入 Git。遊戲包來自使用者指定的[線上 DOS 遊戲頁面](https://legacy.dos.zczc.cz/games/%E7%A5%9E%E5%89%91%E5%B0%81%E9%AD%94/)。執行與測試需要合法持有的原版資料，來源清單見 [CONTEXT.md](CONTEXT.md)。

分析、建置與測試使用 Docker。開工規則與文件入口見 [AGENTS.md](AGENTS.md)，實際命令與結果記在 [WORKLOG.md](WORKLOG.md)。

## 進階啟動選項

Linux封包可在啟動命令後加 `-audio off` 關閉播放。`-realtime` 保留本次已確認的100毫秒追趕模式；原始命令列仍可用固定更新方式供研究。

原生資料定位見 [bundle.go](cmd/zeliard-preview/bundle.go)，音源及封包契約見 [規格033](docs/spec/033-original-music-release.md)。Nuked-OPL3依其LGPL條款提供獨立可替換函式庫與來源，自有程式仍採RRSAL-1.0。

## 授權

自有程式與文件採 [RRSAL-1.0](LICENSE)，屬 source-available 專案。非商業用途免費，允許實況與影片分潤，商業使用另洽。原版遊戲、手冊、音樂、美術及字型不在此授權範圍內。第三方元件依各自授權使用。
