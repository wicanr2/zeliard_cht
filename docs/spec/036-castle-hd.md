# 城堡前景與原版快取變化

狀態：CONFORMED，限下列城堡前景、四姿態旗幟與Castle／Muralla正常往返。屬全遊戲HD的增量，不代表全城鎮、全洞窟或全遊戲完成。既有三平台封包尚未包含此增量。

## 固定來源

- 原版EXE及SAR沿CONTEXT。CPAT來源為SAR2記錄33，解壓SHA-256 `831bf7b9a3cb0d31870d57d23f06e69a56d420fe17f8c27dc0c2323ddbce1339`，256bytes header後為157個48-byte圖格。
- 正常冷新局來源 `workplace/town-hd-v1/foreground-v1/normal-v1/cold-town.memory`，1MiB SHA-256 `5c5b610c97bd2cd5266ae25b3da3764d0b46f02b30e8392a3b09c286a0db45a2`。正常鍵序與檔案讀取沿同目錄receipt，色盤沿既有6dc0b8c0。
- [GRP來源工具](../../tools/hd_grp_sources.py) 的 `grp-sources-v3/receipt.json` 證明157圖格7536bytes與11B1:8100一致，非零flag的232遮罩bytes與11B1:D000一致，三個header指標加8000後與執行中相同。這是來源byte證據，不是HD或每個圖格正常可見的證據。
- [全景來源工具](../../tools/hd_castle_sources.py) 的 `castle-source-v3/receipt.json` 沿007城鎮契約：01B1:C017，114欄各8格，原生912×64；FD底圖取C00F居民表。正常視窗以E000已畫快取、A000背景、MMAN居民與TMAN主角重建，14336像素與原版相同。

## 快取差異與推論等級

已證實：固定來源的世界欄37／43第1列、欄47／53／59第0列，地圖ID為4而已畫快取ID為3。兩圖格是旗幟差異，逐點記錄在來源收據。靜態全圖直接與當次畫面比較會失敗，不能將這些差異忽略或加入像素容忍。

此項已由後續正常來源解決：原版CPAT的80A6表有四對2→3、3→4、4→5、5→2。既有IDA9.4的原始函式sub_31D8在32BD讀已畫ID，32EA讀表下一ID、32EE寫入DS:[SI-1]，32F9完成後恢復ES／DI。正常觀察754次更新，保存的128項涵蓋四對，寫後值均等於表值。原始位址空間仍為IDA binary base0／dosgolem01B1段內，不更名函式或運算元。

`castle-normal-v1/receipt.json`沿 [正常鍵序](../../tools/castle-hd.commands) 取得13點完整原版狀態相同，包含冷新局、左右步行、正常離場及返回。八個城堡最近完成來源各14336像素與原版相同，位於 `castle-normal-sources-v1/`。來源檔保存自己的步數／週期，不冒稱當次CPU收據。HD按已畫快取ID選旗幟，不自建動畫計時。

已證實：CPAT、MPAT非零flag的遮罩與前處理有差別。flag1／2／3分別清除三平面色4／2／1，成對子像素同為該色時透明；flag4在兩個正常profile的八個遮罩bytes均為FF。flag0圖格不採該遮罩。完整公式與逐byte結果見GRP來源收據；這不是對所有遊戲圖像格式的概括。

## 首版手繪候選紀錄

- 輸出在 `workplace/full-hd-v1/castle-art-v1/castle-panels-handpaint-v1.png`，SHA-256 `cd6af931976ab0d55b5ea9ab8e4fca76ed70d7ebf8575664907b6006f01a29a0`，1214×1295 RGBA，有真實alpha。
- 內建imagegen以原版四段佈局為構圖參考、既有Muralla前景為畫風參考。完整提示詞、兩個輸入雜湊與來源路徑在同目錄 `generation.json`。
- 四段原生x範圍為0..240、224..464、448..688、672..912，各有16pixels重疊。圖面包含左森林／塔、城牆／主門、右城牆／塔及森林／亭台。已目視四段存在、無角色及文字。
- 首版生成圖的四段高度不等，不能直接平均切成四份。當時尚待裁切與校準，不計HD覆蓋；後續採下列READY契約的無旗幟底圖及四張動畫圖，保留首版身份。

## 進入READY的條件

1. 以正常新局、左右步行及離開城堡再返回，核對資源、地圖、相機與完成畫格生命週期。新增狀態須由同一原版執行器重生。
2. 查清已畫ID3／地圖ID4的有限變體與更新端；以最小證據描述旗幟來源、遮罩與時點，不追查無關玩法helper。
3. 固定四段圖的裁切、世界座標映射、重疊接縫、原版門口及可通行開口。直接3倍解析度合成，不改地圖、熱區、碰撞、原版輸入或存檔。
4. Castle profile與Muralla profile按來源原子選擇；不得放寬既有Muralla守衛。人物仍按原版居民先、主角後，不能被新前景覆蓋。
5. 先有READY契約，再實作來源拒絕、動畫變體、邊界與原貌還原測試，最後做正常原版配對及實際視窗抽樣。未驗證的後續城鎮與DPAT不借此升格。

全遊戲範圍仍見 [034](034-full-game-hd.md)。原圖、完整手繪、原版bytes及正常收據只留本機，不加入Git或公開交付。

## READY固定契約

- `castle-art-v2/`的無旗幟底圖與`castle-animation-v1/`四姿態均由內建imagegen產生。完整提示詞與來源各在generation.json。底圖7ffb17e6、旗幟圖aa38d78c，完整SHA與註冊座標由 [編譯工具](../../tools/hd_castle_register.py) 及 `castle-registration-v1/manifest.json`固定。直接預乘alpha取樣到2736×192世界圖與四張24×24旗幟，不先縮為原版。
- 世界圖SHA-256 `288ac2c28d625582a401343dedbc108ad8e50a7c78e1956cbe92b32ae9c63bb5`。四段交接世界x232／456／680。主門開口x421..427及亭台開口x765..774使用明列分段映射，不能移動原版位置或輸入熱區。註冊圖及實際視窗均已目視，驗收見末節。
- Castle只接受114欄、157地形圖格與固定程式／色盤；八個正常來源的相同使用區：3000..371C、6026..6C72、C000..C040、A000..B500，資源8000..9E70、4000..5D80、6000..68A0、居民遮罩1216bytes、主角遮罩368bytes。地形只驗非零flag實際採用的232遮罩bytes，不把無效保留區當資料來源。
- 可使用既有TownScenery的明示Castle profile；未載入時Muralla行為不變。每個profile仍獨立驗來源。完成快照遇profile切換立即撤除，不把城堡舊畫格蓋到Muralla；兩者共用的玩家、居民與地面仍須通過當次完整畫格驗證。
- 地圖與已畫快取只在已確認旗幟位置接受ID2..5的原版變化；其餘圖格與固定底圖一致。先重建全部14336原版像素及RGB，失配不寫任何目的像素。
- 原版旗幟8×8區先畫同位置的已驗HD遠山，再畫24×24指定旗幟，使用已畫ID而非已更新地圖ID。靜態底圖跳過旗幟格，其餘只在原版前景不透明且人物透明處畫入。人物仍按居民、主角的原版順序，不改動畫或碰撞。
- 地面與既有Muralla的224×16來源相同，使用009已驗雙帶相位比對，不另猜地面捲動。玩家及居民共用TMAN／MMAN且固定bytes相同；不藉此擴張未畫的phase4或交談背景HD。
- 五張註冊PNG須整組驗SHA、尺寸、alpha及來源；缺少、損壞、未知資料或畫布不足原子退回。規則、亂數、輸入、存檔與原版檔案不改。
- 實作後驗八個正常來源、四姿態、失配與跨profile撤除；再做正常原版配對及實際視窗左右步行、旗幟變化、原貌還原與離場／返回。達成後才升CONFORMED。

## 主角右邊界勘誤與READY增量

連續正常鍵序在88800000步、01B1:07BF遇完整畫格(221,40)失配。原始資料保留於castle-pair-v2/uncovered，前輪兩個失敗不改稱通過。原因是主角欄位83為27時，原版沒有畫半個主角；舊重建卻按224邊界裁切後畫入左半部。

已證實：IDA9.4的新匯出 `castle-edge-ida-v1/evidence.json` 保留原始sub_6950。6950讀byte_83、6953與1Bh比較、6955的JB只讓0..26執行；其餘在6957返回。通過時才把E000對應兩欄各三格標為FF。sub_68AC依序呼叫6B1C、6975、6950及3004，與正常失配來源的完整原版畫面一致。輸入段SHA為aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d，IDA base0對應該快照01B1段內位址。

增量契約：主角只在83小於27時覆繪，不畫裁切的半個主角。共用合成仍可畫可見居民，不以主角隱藏為由撤除整個城鎮HD。主角mask與背景還原也遵守此條件；不改原版欄位或移動。6950..6975加入精確來源守衛；舊Muralla守衛保留。先以這個真實邊界來源核14336像素及右側無虛構主角，再重跑原正常鍵序與受影響回歸。

跨profile驗收按每次場景進入的首次完整來源分別開始，返回已訪問場景也重新計算。profile改變或來源尚未完整時結束前次呈現生命週期。新場景第一個完整畫格尚未完成時仍顯示當次原版，不能沿用另一場景或前次訪問的capture計數。bootstrap另計；已初始化場景的中途畫格仍不允許未涵蓋，穩定檢查點須有正確profile及當次完整呈現，不能只靠不報錯通過。此修正只校正驗證生命週期，不放寬像素或來源守衛。

同一固定鍵序的13點實際分類為八個Castle完整點、三個Muralla完整點及兩個轉場點。right-3在新場景第一個完整來源之前，left-return-2在返城資料切換期間，兩者必須沒有沿用快照；其餘十一點必須有正確profile。v5保留將right-3誤當穩定點的失敗，不用修改鍵序避開轉場。

## 限定驗收與入口

目前程式為 `workplace/full-hd-v1/castle-runtime-v3/zeliard-preview`，SHA-256 `6670698f72fbd6d2c537693db0149e09553f280124a09de5c04deefe3d543d7a`。啟動時在既有城鎮素材參數之外指定 `-castle-memory` 固定冷新局參考及 `-castle-art-dir` 註冊圖目錄，另載入既有完整遠山與居民圖集。完整實際參數保存於 `castle-gui-v2/receipt.json`。

| 驗收 | 結果 | 證據 |
|---|---|---|
| 正常來源與動畫更新 | 13點完整原版狀態相同，754次更新；保存128項涵蓋2→3→4→5→2，寫後值等於表值 | `castle-normal-v2/receipt.json`；自動合併[來源索引](../castle-source-index.json) |
| 合成、邊界與退回 | 四項Castle及四項舊居民測試通過，無skip或fail；八個真實完成來源及右邊界均保留14336原生像素守衛 | `castle-runtime-v3/final-unit-tests-v2.jsonl` |
| 正常往返 | 13點原版RAM、VGA、色盤、暫存器、時間、檔案及鍵盤狀態相同；11穩定點、2轉場點，2940次呈現、293次保留完成畫格、8次首次來源等待、0未涵蓋 | `castle-pair-v6/receipt.json`及`test.jsonl` |
| 舊城鎮回歸 | 舊居民正常路徑與城鎮原圖／人物隔離通過 | `town-regression-after-castle-v1/test.jsonl` |
| 實際視窗 | 正常冷新局、左右步行、主門／亭台、四旗幟與兩地往返；10組原貌切回、5760000遊戲區像素完全還原，DOS退出0 | `castle-gui-v2/receipt.json` |
| 指定圖面核對 | 1911個不透明旗幟像素及11649個全透明背景像素獨立相同；7752個中間alpha像素另依Go合成與目視驗證 | `castle-final-audit-v2.json` |

以上路徑均位於 `workplace/full-hd-v1/`。最終稽核由 [hd_castle_audit.py](../../tools/hd_castle_audit.py) 在既有Pillow／ffprobe容器重生至新的輸出檔；它固定程式、工具、索引及收據雜湊，拒絕覆寫。IDA原始函式及分級索引匯出為 `castle-edge-ida-v2/evidence.json`，使用 [ida_castle_evidence.py](../../tools/ida_castle_evidence.py)。

12秒實際視窗影片 `castle-gui-v2/castle-animation.mkv` 只驗畫面，音訊關閉，不當成新版推廣片。原版轉場及首次完整來源前顯示原貌；主角phase4、交談背景、其他城鎮與後續洞窟仍是已知HD缺口。版本、配樂及v.0.2.0封包未變；034仍為DRAFT。
