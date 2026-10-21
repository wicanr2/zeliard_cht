# 洞窟剩餘主角組合來源

狀態：CONFORMED，限定下列兩個繪製分支與五個圖面鍵；未列組合維持拒絕。

## 來源與研究入口

- 原始`ZELIARD.EXE` SHA-256為`fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`，`SAVEM1.USR`為`03df9abcf62b1712d6b46dd1fc5b23335a4b8711b3dc84875a1498260d37735c`。包內USR取得歷史未知，只由原版F7冷讀其未改副本，不賜予裝備或注入狀態。
- [既有拒絕來源探針](../../tools/hd_cave_rejected_probe.go)新增可選`-profile-frames`，依13個主角繪製欄位去重，只記錄真實完成回呼中被拒的候選。最多128組，來源保存於本機`workplace/full-hd-v1/`。不更動正常指令預算、鍵序或正式解碼。
- 兩個獨立原版機器走[cave-platform.commands](../../tools/cave-platform.commands)與[cave-platform-magia.commands](../../tools/cave-platform-magia.commands)，每個具名命令觀察點及結束時核完整原版狀態。回呼內的來源只有後者正在執行，不冒稱每個回呼皆做雙機同步配對。
- 原平台／魔法石路徑有40／92次來源拒絕。先核分類、固定來源SHA與原版consumer，再決定解碼及圖面的缺口。不將拒絕次數當成素材數。

## READY前停止線

必須以未改動來源完整重建原生畫面，逐項列出模板、圖層順序、遮擋、裝備分類與邊界。未知欄位沿用原始位址，標示證據等級；不以動畫外觀推定玩法語意。只在證據充分後擴充正式解碼與高清圖面，不改原版規則、數值、輸入、亂數或存檔。原始RAM及派生圖片不進Git或Release。

## 已核證據與READY契約

- 兩個來源收據為`cave-profile-source-v1/receipt.json`與`cave-profile-magia-source-v1/receipt.json`，Go1.24.13／專案dosgolem，所有輸出位於`workplace/full-hd-v1/`。普通路徑15個欄位組合，魔法石路徑31個；13個觀察點的steps、cycles、完整RAM與indexed SHA均與047正常配對相同，雙原版狀態及存檔不變。完成回呼中有效profile拒絕事件為20／45，其餘拒絕包含撤除後沒有主體的候選，不能把40／92當素材數。
- IDA9.4輸入`primary.segment` SHA-256為`5e7f76080cc542eb06ca90ec349ad2ec0c5ab374c32acf59bc4c4307899a0a5b`。IDA base0對應01B1段內偏移，不是EXE檔案偏移。[九筆分級索引](../cave-profile-source-index.json)由[IDA工具](../../tools/ida_cave_profiles.py)合併至`cave-profile-ida-v1/evidence.json`，保留原始函式名、運算元、bytes與未知警示。
- 已證實：`01B1:3AAE..3ABB`將非零E8／FF39／FF3A導向主體模板，跳過後手；`3B8A..3B92`及`3BE7..3BF0`以`60EA+9×(E7&3)`選FF39主體。`3C16..3C33`的FF39／FF3A特殊前手經`3D22`裝備分類；原0093=2時分類1，位移7E。正式只擴充FF39=255、E8=0、FF3A=0、FF38=0、FF40=0的已觀察持盾情況，保留既有FF3A分支與多profile拒絕。
- 已證實：`3B00..3B14`與`3C67..3C7B`在FF40非零、FF41=1時，對其中一手額外加24h模板位移。朝向由00C2最低位決定，另一手仍沿0093裝備分類。FF38非零維持六格、跳過首三格。只擴充FF41=1及FF40=255；其他FF41值與非法FF40仍拒絕。
- 原始consumer位址`01B1:3B08`及`3C6F`的加24h，與`01B1:3C27`的7E特殊前手分別回填047與012／027／028。歷史規格及收據不改名，回填台帳為[re-resolution-backlinks.tsv](../re-resolution-backlinks.tsv)。
- [研究編譯](../../tools/hd_profile_study.py)只覆蓋編譯器讀取的四段來源，不改正式檔案或原版機器。`cave-combined-profile-study-v1/receipt.json`的46份未改動來源，完整224×144原生遊戲區均零差，去重為六個ActorRaster鍵。這是來源重建證據，不是正式前端或GUI驗收。
- 五個缺圖鍵已由內建imagegen補圖，另一鍵沿用046原姿態57。六次完整請求、三份初始引用與後續修改引用SHA、原cache及生成檔在`cave-combined-profile-art-v1/generation.json`。側向盾牌初版畫錯側，保留v1但只選修正v2；黃藍紫變體另有一次生成，不自行改色。
- [圖面註冊](../../tools/hd_profile_art.py)依原生包絡直接取樣至三倍，五張72×72與固定SHA見`cave-combined-profile-complete-v1/manifest.json`，已目視。沿既有alpha96墨跡及兩像素邊界，不重畫RGBA，不經原解析度縮圖。
- 正式輸入維持完整／主體RAM、indexed、palette及paint時間線。只擴充上述來源判定與模板，仍須整幅原生重建零差；五圖以完整來源鍵及PNG SHA原子載入，保留原67圖。可選`-cave-player-profile-art`須有洞窟主角呈現器；缺圖、壞圖或未知鍵安全退回，不留下半套。
- 垂直鏈為固定原版資料→原F7正常鍵序→完成／主體來源→圖層重建→高清選圖→前端；位置、原生包絡、裁切、遮擋、動畫節點、輸入、亂數與USR均不改。繪圖層只讀來源。
- 驗收包含46份真實來源、五圖直接高清像素與混色、原67圖不變、未知來源及載入原子性、兩路13點完整原版狀態與缺圖／拒絕盤點，以及代表背向姿態的正常GUI切回及退出。五圖逐一GUI不列閘門；未捕捉者列保存來源驗證，不外推其他場景或全遊戲HD。

## 實作與限定驗收

[來源解碼](../../internal/prototype/hd/cave_sources.go)依READY模板接入，[五圖載入](../../internal/prototype/hd/cave_player_profiles.go)新增67至71姿態，保留原67圖。46份未改動來源的1483776原生像素零差，五圖25920註冊RGBA及8057個獨立混色像素相同；未知鍵、缺圖、末張損壞與原67圖保留通過。46項圖面／來源及三項正常／來源檢查無skip／fail，vet與build通過。

正常配對為`cave-profile-normal-pair-v2/`：普通路徑253次完成、零拒絕；魔法石564次完成、兩次拒絕。後者為`magia unobserved record`一次，接著`source or missing body`一次，不放寬守衛。兩路主角缺圖為零，13點完整原版狀態及USR bytes／mtime相同；刀光仍缺四個來源鍵，物件跨兩路尚有13個不同來源鍵，均有原始來源保存，不混稱全HD。

[GUI抽樣](../../tools/hd_cave_profile_smoke.py)正常冷F7、走到繩索並向下，已擷取FF39=255、主角新姿態。原貌／HD／切回的CPU、RAM、indexed與palette相同，576000遊戲區像素切回相同，正常DOS退出0、USR不變，根端已目視。其餘四圖只稱保存正常來源驗證。

候選`workplace/full-hd-v1/cave-profile-runtime-v1/zeliard-preview` SHA-256 `458a8149658c958e11cd9ffc0e12c240b2a9c82135c4e5532a5aa1a45c71da4a`。收尾`workplace/full-hd-v1/cave-profile-final-audit-v1.json`由[稽核工具](../../tools/hd_profiles_audit.py)重生至新檔。刀光與物件來源入口為兩個正常收據的weapon_coverage.missing_sources及object_coverage.unknown_sources；不重跑已存來源或重新繪製這五圖。封包、影片與版號未更新，全遊戲目標仍進行中。

## 049接續範圍

[049](049-cave-unmatched-assets.md)已補四張刀光、四張物件及九個非零底圖的精確遮罩對應，原20刀光／29物件不變。選載新增物件組才建立純靜態HD下層，依原四格背景與順序還原；舊合成不改。兩條平台路徑主角／刀光／物件缺圖零，來源拒絕仍0／2。17保存來源、13正常點、舊物件像素回歸及代表遮罩GUI已驗，不外推其他城鎮、首領、後段或全遊戲HD。收尾為`workplace/full-hd-v1/cave-unmatched-final-audit-v1.json`，舊證據保留原身份。
