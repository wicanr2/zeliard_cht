# 後續城鎮來源與HD接入

狀態：DRAFT。目標包含城堡、Muralla以外八份城鎮地圖、其前景／背景、居民及正常進出；沿既有手繪奇幻、3倍畫布與dosgolem唯讀覆繪。不改地圖、碰撞、路徑、交易或存檔。

## 原始輸入與範圍

- 固定SAR2身份沿[全量規格034](034-full-game-hd.md)及[覆蓋台帳](../hd-coverage.json)。record36至45為十份未壓縮MDT候選，已驗城堡與Muralla分別為36、37；後續38至45仍須核載入及圖像家族。
- 既有[城堡規格036](036-castle-hd.md)與[回鎮規格022](022-return-town-hd.md)的原版來源、畫帶及ID守衛保留。其他地圖不套用Muralla固定地形或居民外觀。
- 原版消費端沿固定城堡段 `workplace/full-hd-v1/castle-edge-ida-v1/primary.segment`，SHA-256 `aed45060ee02eafa07b41891d326ac0217a8afdd2957d79b401369a8bb2fc11d`。IDA base0對應該dosgolem01B1段內偏移，SAR檔案位址另列。
- [分級來源索引](../town-resources-source-index.json)與[IDA匯出](../../tools/ida_town_resources.py)保存原始位址與未知項。研究輸出位於既有 `workplace/full-hd-v1/` 的 `town-resources-*` 目錄。

## READY閘門

1. 核原版地圖descriptor選擇、C000表頭、八列地形、居民底圖、動畫及出口／室內入口記錄。
2. 核各地圖實際選用的前景、背景與居民圖像家族，從SAR解碼到正常RAM及原生畫面保持可追溯。
3. 逐來源產生完整原圖與手繪圖，保留尺寸、透明、門口、人物錨點、動畫鍵及雜湊。正常未知來源完整退回，不能以同類城鎮猜圖。
4. READY後才接正式前端；每個新增家族至少有正常進出、步行／捲動、代表交談或室內、原貌切回及存讀恢復抽樣。直接進入測試狀態僅供結構核對。

不因解出八份地圖或生成靜態圖就稱全部城鎮完成。正式封包、README圖片及影片待全量目標的交付閘門更新，原版與完整衍生圖維持本機。

## 已核來源與美術候選

[來源工具](../../tools/hd_town_sources.py)重建十份地圖共17784格，原始descriptor、表頭、人物記錄、地形與背景選擇保存於 `workplace/full-hd-v1/town-resources-source-v1/receipt.json`。IDA9.4以原始位址及函式邊界核九函式、五表，合併分級索引的匯出為 `town-resources-ida-v2/evidence.json`。

| 地圖 | 原始名稱 | 地形 | 背景程式 | 前景候選 |
|---|---|---|---|---|
| STMP | Satono Town | DPAT | CKPD | 已生成，待校準與正常來源 |
| BSMP | Bosque village | MPAT | YMPD | 已生成，待校準與正常來源 |
| HLMP | Helada Town | DPAT | CKPD | 已生成，待校準與正常來源 |
| TMMP | Tumba Town | DPAT | CKPD | 已生成，待校準與正常來源 |
| DRMP | Dorado Town | DPAT | CKPD | 已生成，待校準與正常來源 |
| LLMP | Llama Town | MPAT | YMPD | 已生成，待校準與正常來源 |
| PRMP | Pureza Town | DPAT | CKPD | 已生成，圖格191用途仍未知 |
| ESMP | Esco village | MPAT | YMPD | 已生成，待校準與正常來源 |

原版6D9E以7C46乘11選CPAT／MPAT／DPAT，6AAF以7C45低位選YMPD／CKPD。CPAT157、MPAT242格的7536／11616轉換bytes與既有正常來源一致，DPAT191格仍僅靜態來源。Pureza第316欄八列均用ID191，超出已解DPAT的0至190，不補成空白或相鄰圖樣。

八次內建imagegen的完整提示詞、16份參考記錄及八份原始產物SHA保存於 `town-world-art-v1/generation-v1.json`，候選PNG在同目錄。八圖聯絡表及八張完整候選已目視，逐圖紀錄為同目錄 `review-v1.json`。Esco第一片右端的局部武器店誤寫為INN，Pureza未知標記仍存在，Dorado金黃柱的材質也須核對。接縫、門洞、招牌與透明邊緣尚未校準。這些圖沒有註冊為正式圖組，也沒有新增前端旗標。

[正常路徑探針](../../tools/hd_town_path_probe.go)只送鍵盤，四份冷開機收據保存原版檔案／USR不變、完整快照及按鍵序列。正常購買武器、第一洞窟移動與攀爬已到Satono出口附近，但沒有取得新城鎮完成點。生命0的後段快照不作可玩或到達證據，不把按鍵無移動當執行器故障。

[階段稽核](../../tools/hd_town_stage_audit.py)目前收尾為 `town-resources-stage-audit-v4.json`，舊收據保留身份，狀態仍為 `SOURCE_AND_EIGHT_ART_CANDIDATES_NORMAL_PATH_PENDING`。Esco的第二版已修局部武器店招牌及原版單排四長窗，請求與SHA在 `town-world-art-v1/ESMP-correction-v2.receipt.json`，未完成接縫／門洞校準或前端接入。

正常探針另支援固定雜湊包內USR，SAVEM1正常讀回生命200及魔法石。`town-path-savem1-v3/`已正常升起第一綠色平台並離開到右側通道；v4獨立重播首點RAM／暫存器／畫面／色盤／指令及週期相同，再正常降下及升起。平台來源另由[043](043-cave-common-scenery.md)追蹤。沒有取得Satono或其他新城鎮完成點；原版USR前史未知，不當新局全流程或HD驗收。

目前不達READY，先取得正常新城鎮的圖像、背景與人物來源，再核完整畫面、動畫及美術校準。前輪武器店正式候選與舊封包維持原身份。
