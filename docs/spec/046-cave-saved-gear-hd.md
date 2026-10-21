# 已存平台路徑的主角與武器缺圖

狀態：青藍武器十姿態及主角12來源鍵限定CONFORMED。其餘來源分支仍待補。直接使用045普通／魔法石配對已保存的正常來源，不重新蒐集相同畫格，不改裝備、動作或原版規則。

入口為 `workplace/full-hd-v1/cave-door-normal-pair-v1/`兩份收據的player_coverage及weapon_coverage。主角去重12個來源鍵，武器去重十個；原畫格數不當素材數。先重解各保存來源並核SHA、原生整幅像素及來源鍵，再比對既有55主角／十武器的幾何與色彩。尚未核對前不把來源鍵相似當作可共用高清圖面的依據。

研究工具為 [已存裝備來源匯出](../../internal/prototype/hd/cave_saved_gear_study_test.go)。輸出只放本機workplace，來源檔與派生圖不進Git。READY契約須列明採用的新圖面或完全相同的既有圖面，以及註冊、邊界、拒絕、正常原版配對與視窗驗收範圍。

## 武器來源與READY契約

- 原始EXE SHA-256沿045的固定版本 `fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`；保存來源各有1MiB完整／主體RAM、畫面、色盤與時鐘。Go1.24.13重解22個去重來源，整幅原生畫面零差、完整來源鍵相同，收據為 `cave-saved-gear-source-v1/receipt.json`。
- 已證實：十武器各與既有十姿態之一有完全相同的1024像素透明輪廓及原生包絡，RGBA則有16至54像素不同。原生顏色為青藍，不因輪廓相同沿用銀白素材。這是圖面事實，不推定裝備名稱、傷害或效果。
- 一次內建imagegen編輯現有刀光圖集，保留十姿態方向與布局，改為青藍材質並保留透明。完整提示詞、兩引用SHA、原cache及輸出SHA為 `cave-cyan-weapon-art-v1/generation.json`。十張96×96圖面在 `cave-cyan-weapon-complete-v1/manifest.json`，已目視。
- 以十個完整WeaponRaster雜湊及固定PNG雜湊註冊，保留原來十個姿態。載入全部成功才交換新map；缺圖、損圖、尺寸錯誤或沒有原武器呈現器時拒絕，不留下半套。
- 直接高清取樣至原包絡乘三，仍用原WeaponX／WeaponY及既有最後圖層、透明混色、裁切與未知鍵退回。新增可選 `-cave-weapon-cyan-art`；不改來源、ActorRaster、動作時序、CPU、輸入、數值或USR。
- 驗收包含十個已存正常來源及獨立PNG像素、未知鍵與載入原子性、既有十圖回歸、正常平台配對後武器缺圖為零，以及實際GUI至少一個青藍刀光與原貌切回。所有刀光正常GUI逐姿態出現不列完成閘門。

主角來源比對及READY契約見下節。原版來源及其派生圖留本機，不進Git、Release或公開包。

## 主角來源與READY契約

- 原始EXE、Go1.24.13、22份保存來源及完整原生重建沿本規格前節。12個主角來源鍵是576byte ActorRaster，含色號及255透明值；不是12種已解出的遊戲動作，也不推定色彩變化原因。
- 已證實：對照027已驗17個持盾參考，七個新鍵輪廓不同；五個輪廓相同但RGBA不同。六個來源有明顯黃／藍／紫配色，不能把它們正規化成普通紅色姿態。來源及比較為 `cave-saved-player-source-v1/receipt.json`，註冊工具另核全部12個輪廓與原圖SHA。
- 一次內建imagegen製作3欄4列的12姿態圖集，保留左右朝向、蹲下及六個色彩變體。完整提示詞、兩份引用SHA、原cache及產物SHA在 `cave-saved-player-art-v1/generation.json`。原圖SHA-256 `c2779d1d775eb89baa0be91de5bf567ea205b7fba502fd7d56f6a9ae899ca803`。
- [註冊工具](../../tools/hd_saved_player_art.py)沿既有NPC工具的alpha96墨跡界定，再保留兩像素抗鋸齒邊界；先以實際透明間隔分列，不以等分格線截斷腳尖。原RGBA不重畫、不改色，直接取樣至原生包絡乘三後置入72×72。12張結果與列間空隙已目視，manifest為 `cave-saved-player-complete-v1/manifest.json`。
- 以12個完整來源鍵、固定PNG雜湊及原生包絡註冊，全部成功才交換新map。原55姿態不變，新圖編號55至66；六個色彩變體依原版當次ActorRaster選圖，不自訂色彩週期或修改原記憶體。
- 新增可選 `-cave-player-extra-art`，須已有原主角圖組。缺圖、壞圖或未知鍵時保留既有圖組與原版退回；仍走現有玩家位置、裁切、後畫遮擋及圖層順序，不改CPU、數值、亂數、動作、輸入或USR。
- 驗收涵蓋12個已存正常來源、獨立PNG／混色及後畫遮擋、原55圖不變、載入原子性及未知鍵、兩條正常平台路徑主角缺圖為零，以及至少一個新姿態的實際GUI與原貌切回。六種變色逐一GUI捕捉不當預設閘門；未捕捉者明列為已存來源驗證。

## 青藍刀光實作與驗收

實作為 [cave_weapon_cyan.go](../../internal/prototype/hd/cave_weapon_cyan.go)，檢查為 [cave_weapon_cyan_test.go](../../internal/prototype/hd/cave_weapon_cyan_test.go)。十個保存來源全部匹配；92160個註冊RGBA與5152個帶透明度合成像素獨立相同，原始完整／主體記憶體未變。內部筆觸、輪廓與alpha屬明示美術差異，原生外包絡、位置與動作維持原版。

38項圖面／來源回歸、當前原子載入與三項正常／來源測試無skip／fail，vet及build通過。兩條原版正常平台路徑觀察到全部十個新鍵，可見武器缺圖均為零；已匹配畫格分別18與30。13點完整原版狀態、USR bytes及mtime相同。主角缺圖仍為18／100畫格，未混稱已完成。

實際GUI冷F7、出劍後暫停擷取，再正常使用魔法石、門框與平台升降，六組原貌切回3456000像素相同，原版狀態、存檔及正常DOS退出0通過。根端已目視青藍刀光。GUI不是逐像素獨立混色證據，5152像素限已存來源測試。

前輪候選 `workplace/full-hd-v1/cave-cyan-weapon-runtime-v2/zeliard-preview` SHA-256 `587e53e5ca885ccdfbf3ee91115153446d5b0a18981baac3c7568128c815233a`。收尾 `workplace/full-hd-v1/cave-cyan-weapon-final-audit-v1.json`由[正式稽核](../../tools/hd_cyan_weapon_runtime_audit.py)重生至新檔，完整GUI命令在同根 `cave-cyan-weapon-gui-v1/receipt.json`。主角來源由下節補齊；其他武器及後段仍待補，封包和影片未更新。

## 主角保存姿態實作與驗收

實作為 [cave_player_extra.go](../../internal/prototype/hd/cave_player_extra.go)，驗證為 [cave_player_extra_test.go](../../internal/prototype/hd/cave_player_extra_test.go)。12個保存來源與六個變色全數匹配；62208個註冊RGBA、31376個未受後畫遮擋的混色像素獨立相同，原始完整／主體記憶體未改。未知鍵不借圖，缺圖及最後一圖損壞時全67個pose指標不變，原55圖的RGBA雜湊相同。

42項圖面／來源與三項正常／來源測試無skip／fail，vet及build通過。兩條正常路徑均觀察到新鍵，合起來包含全部12鍵；主角已匹配畫格為156／419，已選完成來源的主角及武器缺圖均為零。13點完整原版狀態與USR bytes／mtime相同。42／92次來源拒絕仍未消除，不稱每畫格HD。

GUI v3從冷F7正常擷取主角55、短按青藍刀光、物品後門框及平台升降，七組原貌切回4032000像素相同，正常DOS退出0、存檔不變，根端已目視。六個變色全有已存正常來源驗證，沒有逐一GUI捕捉。v1的平台站位偏到46，v3以正常方向鍵站穩已驗44；v2長按途中出現FF41=1的未支援來源，仍退回原貌，不把這個缺口列通過，也不推定旗標含義。

目前候選 `workplace/full-hd-v1/cave-saved-player-runtime-v1/zeliard-preview` SHA-256 `473dddb1c21b057d345db531c4a822000bf26cc7e488a1324980e1b6e83be0d4`。收尾 `workplace/full-hd-v1/cave-saved-player-final-audit-v1.json`由[主角稽核](../../tools/hd_saved_player_runtime_audit.py)重生至新檔，完整命令為同根 `cave-saved-player-gui-v3/receipt.json`。下一步沿v2保留日誌重生FF41被拒候選，另核來源；其他美術及後段仍待補，封包與影片未更新。
