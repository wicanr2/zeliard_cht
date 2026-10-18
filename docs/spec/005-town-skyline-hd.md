# 城鎮遠山與天空 HD

狀態：CONFORMED，2026-10-19。限城堡／Muralla共用54列遠景及已驗正常路徑的可撤回 prototype；正式倍率、完整城鎮美術與預設仍待確認。

## 來源與證據

- ZELIARD.EXE SHA-256 `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`；ZELRES2.SAR SHA-256 `bf7c792a58d04d241bde1545bf04d22b3a9386471b22b6c9e501bfecb23f14c1`。
- dosgolem來源摘要 `2a5ad11a6c67b1e1d8a16d61830d8565f262ab38566f49d6ff076a721be58ada`，Go1.24.13／Ebitengine2.9.9，工具鏈及正常來源見 [CONTEXT](../../CONTEXT.md)。
- 原版實驗入口 `workplace/town-hd-v1/source-inventory.json`及 `layer-evidence.json`；實際畫格在前輪正常冷啟動／F7／走路的 `prose-v1/npc-normal-v2/`與 `prose-v1/navigation-v1/`，不是注入座標或測試專用畫面。
- 已證實：原生320×200畫布中 `(48,24)-(272,78)`遠山／天空在城堡、Muralla讀回、兩個走路位置與離開對話後完全相同。224×54色號SHA-256 `22fb614b02b3d9ac436c85b7d56c62ce9004d1254f6c52160de2cf62edeb7cf9`。
- 已證實：y78之後開始有前景建物／樹木。此層只處理上述遠景矩形，不重畫前景、人物、文字、碰撞或地圖。
- 已證實：C000載入的SAR2 payload038A1A、0392DC前綴分別為 `968819bb6c4906514b4fce3e2ce99764b73cfa003f48d7d57d4cf8b12da8ffb6`、`025bf2982660c942cbefed303ccdd0523ff824756a0fe0bb2587b99e0931e7ec`。SAR檔案偏移、dosgolem段內位址及VGA座標分開，不互當位址。
- 未知：遠景內部producer及其他城鎮適用性。不為靜態顯示替換逐行追原版driver。

## 圖面與顯示契約

- 內建image_gen生成手繪圖，完整實際提示詞在 `workplace/town-hd-v1/generation-prompt.txt`；原構圖參考與權利分類在 `generation-record.json`。
- 圖面 `skyline-handpaint-v1.png`，實際1973×797，SHA-256 `275f14de1b18ea6dea297999cdf76e2574bcdc9a4a975782d5c01901480058de`。輸出尺寸與提示不同，以實際PNG為準。畫帶 `(0,207)-(1973,589)`直接合成到672×162，不先縮成224×54。
- 只改遠景的明示顯示差異；不新增建物、角色、道路或敘事物件。黑色上下製作留白不畫入遊戲。
- typed input為固定原版SAR2、完整原版indexed參考、768-byte色盤及固定PNG。768-byte色盤SHA-256 `6dc0b8c0cf558eb91b0582a80b6d5970b701b235b9964adbff3d2dedfad3d08c`。
- 每次呈現須模式13、64-byte C000前綴屬兩個已證實模組、遠景整區色號與RGB完全相符。未知來源、色盤淡入、短畫格或區內任一像素變化則整層不畫。對話蓋住遠景時自然保留原版對話，再由既有中文層呈現。
- 映射為 `target(x,y)=art(0+x*1973/672,207+y*382/162)`，最近鄰直接取高解析度來源。圖面與整個目標矩形先驗證再交換；不可畫半組或半個越界矩形。
- 畫面合成順序：原版 → HD背景／既有HD圖面 → 中文 → 工具列／設定頁。開關只換顯示，不重開局、不改存檔、不重擲亂數或送遊戲按鍵。
- 獨立 `-town-sky-reference`、`-town-sky-palette`、`-town-sky-art`載入。與現有HD工具列共用開關，缺省仍關閉；錯誤輸入保留原貌。

## 驗收與停止線

1. 正常冷啟動／F7／走路／對話兩頁／結束完整配對，1MiB、色號、色盤、CS:IP、steps、cycles、FileOps及持鍵均與原版相同。
2. 五個已知城堡／城鎮畫格接入；兩個對話畫格必須退回。逐像素確認遠景區外0差、HD直接取樣相符、原貌完整還原，不覆蓋人物、建物、HUD或中文。
3. 模式、模組前綴、原文像素、RGB、短畫格、畫布越界、圖面雜湊／尺寸等失配拒絕；合成事件不稱正常遊戲驗收。
4. 正常X視窗另走冷F7、HD原貌切換、走路、居民交談、結束；暫停切換的steps／cycles不變，最後持鍵空、退出0。

素材及含原版畫面的收據只留本機ignored工作區。此層完成不代表樹木、建物、角色、其他城鎮或全遊戲HD完成；跳躍及戰鬥未驗，不由像素拒絕測試外推。

## 限定驗收

- `workplace/town-hd-v1/test-v1.jsonl`兩項通過，沒有skip；`pair-v1/receipt.json`正常9點完整原版狀態相同。四點遠景各108862像素改變，區外0，兩頁中文維持；五個靜態來源及八類拒絕通過。
- 固定候選SHA-256 `28ea8cc2b6dbe1a5a29eee735516d94f1e25ebd60d5dd48e86b6b3dc44e4f455`，52來源及10既有鍵序，go vet與建置通過。
- `gui-v2/`實際X視窗冷F7、走路、居民兩頁及結束退出0。兩次暫停原貌／HD切換steps與cycles不變，遠景各108862像素、既有側框172017像素改變，兩層聯集外0，原貌還原0，直接高解析度取樣相符。對話時天空退回，結束後恢復HD；最後持鍵空。
- `gui-v1/`保留驗證腳本點錯暫停按鈕的exit1，沒有執行HD切換。v2只修正點擊幾何，不改候選、超時或原版鍵序。
- 根端實圖核對入口為 `root-pair-review-v1.json`及 `root-gui-review-v2.json`。下緣手繪與原版前景的水平接縫可見，後續須接完整背景／前景，不能宣稱全場景HD。
- 固定READY規格保留於 `ready-spec-v1.md`。其餘收據與來源統一由 [CONTEXT](../../CONTEXT.md) 的town-hd-v1入口導覽。
