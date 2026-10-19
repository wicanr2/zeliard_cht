# 正常洞窟探索的主角補圖

狀態：CONFORMED，限本文七個來源與已列正常路徑。架構、三倍試作畫布與原版規則沿017，不變更正式預設。實作前READY收據為`expedition-build-v2/ready-review.json`；當時97份來源與主入口身份保持，候選為fab8f646／98來源。

## 原版證據

`workplace/cave-hd-v1/expedition-source-v2/receipt.json`記錄冷啟動、F7讀自建存檔、正常左移、跳躍、攀繩與下層交戰。全程只送鍵盤及推進原版，不注入、還原或重擲。18個擷取點跑到146110000指令；未走到下一洞窟，不宣稱探索通關。

工具為Go1.24.13、凍結dosgolem882檔與`exploration_probe_v1.go`。首次從模組外`/study`建置遭Go internal限制；改在同一容器從`/project/workplace/cave-hd-v1`建置成功，未修改產品。來源段為dosgolem實模式`01B1`及資源段`11B1`，不是SAR檔案偏移。原始檔雜湊沿CONTEXT。二進位及逐檔身份另存本地收據。

`expedition-inventory-v2/index.json`由既有獨立解碼器重建18幅完整原生遊戲區，每幅32256像素，差異0。遮擋前24×24、透明255的整陣列SHA與正式解碼器相同。取得12鍵，其中4鍵已有HD、7鍵需新圖、1鍵是新增來源的配色變化。

| 新鍵SHA-256 | 原生包絡，右下不含 | 正常來源 |
|---|---|---|
| `c03571efe91aea6dc91514cf7b89a7f22752b34ecd6335cc1a9bef9aa687ade0` | `(2,0)-(23,24)` | `01-left-attack-1` |
| `acde2d3ef27ff5aca995429c975fc8452a4fc29b8601cc32c3e2553aa8a7c7a0` | `(1,8)-(19,24)` | `05-vine-down` |
| `f3c8eae54a8414476c042f57a5e390e93596da2d10954aec8a27357e4e7c25f2` | `(1,0)-(20,24)` | `06-vine-align` |
| `e7ca0dc3722c9b053e578791f62f04f904d2a168aef8f82dd5ba4767a68c4685` | `(5,0)-(19,24)` | `07-lower-corridor` |
| `19fabad81a43edd6108a46f0dd091137bfcda2a9e04497277714af52ca002201` | `(4,2)-(20,24)` | `10-lower-left-strike` |
| `55adeb9c9dd4bcdf75ae503ca91f564b325fd4331202f73fea8e224efed4987a` | `(2,8)-(20,24)` | `12-lower-west` |
| `eb7307bb6a7270d82ca058afc46655c942c3d00dce2a3bd8bb16f61a8c100b6a` | `(0,16)-(20,24)` | `14-west-open` |

以上為已證實的圖像身份、幾何與正常鍵序。來源名字只供導覽，不單憑欄位或外觀推論受傷、無敵、死亡、硬體時序或規則。

`39ca530247f1bc39cf0a2f1ac03aa080fc351d31d6a38c9dedd61f0ac09dd9a4`與`acde2d3e`的576bytes透明及位置完全相同，原色號具有唯一雙向映射，`FF36=255`。此來源保留原版色號，不借用正常HD顏色，不稱已HD。

## READY合成契約

- 七鍵沿017既有來源、完整原生重建、可見像素、後畫遮擋與直接高清取樣流程。未知鍵、配色變化、缺圖及失配保持原貌。
- 原30登錄與五張PNG不覆寫。追加一張透明圖集及七組固定裁切；整組身份與直接取樣先核對，再交出可用圖組。
- 圖面依`expedition-inventory-v2/new-art-native-sheet.png`七個佔用格，參考既有`player-motion-atlas-handpaint-v1.png`的角色與手繪奇幻畫風。底列中、右保持空白。精確提示、原圖及SHA留本地。
- 每張直接取樣到原包絡三倍尺寸，置回72×72；保留頭頂、腳底與位置。不鏡射、不推定新規則、不改原版狀態。
- READY需完整原生證據、七張圖人工核對、固定圖面與裁切身份、獨立高清合成。正式接入後驗既有30取樣不變、拒絕原子性、正常鍵序原版完整狀態配對及實際視窗。
- 本項只補已見姿態，不代表全動畫、洞窟探索完成、全遊戲HD或中文完成。

## 實作前圖面審查

採用`player-expedition-atlas-handpaint-v1.png`，1254×1254 RGBA，SHA-256 `7691d0862c5a98e49112ef2022ec5ac47529afc16a07221dc44e3a661925a0c3`。由內建imagegen編修七格原始來源，角色身份參考既有動作圖集。完整提示為`player-expedition-atlas-prompt-v1.txt`，SHA-256 `b1b37f231131223333c063887ba7e249303ae48eab427f90fa5331a42d13b9db`；美術及參考留本地，不加入Git或發行。

`expedition-art-preview-v2/registration.json`固定七個互不重疊窗口，以alpha≥128主體外擴4像素登錄，保留裁切內全部原alpha。首次按欄偵測受到底列腳尖跨欄影響而拒絕，未修改圖面或把腳尖切掉。37張直接取樣中，原30登錄及RGBA不變。

`expedition-composition-preview-v1/`獨立重建18份原圖並合成地形及主角HD。16份可見HD、1份配色保留、1份全遮擋；七個新鍵各至少一份有可見HD改動。`14-west-open`全遮擋不計可見補圖證據，同鍵`15-leave-vine-left`有131個原生主角可見像素。舊報表的`unknown_pose_keys`欄含全遮擋鍵，真正未登錄來源只有`39ca`配色，不將欄名直接當缺圖清單。

根端已目視完整圖集、七張72×72目標圖，以及攀繩、蹲姿、倒地與後畫敵人保留的完整圖。這是離線獨立圖面審查，正常原版配對與實際視窗仍待完成。

## 限定驗收

- `expedition-unit-v1/tests.jsonl`的45具名測試全通過，無fail／skip，vet及build通過。37個直接RGBA取樣相同，既有30個保持原值；第六圖缺失或損壞整組拒絕。
- 新18份完整遊戲區5225472高清像素與獨立Python一致，區外不寫、原版來源不改。16份可見HD、1份全遮擋、1份配色保留。既有22份獨立合成6386688像素亦通過，不能把全遮擋計為可見補圖。
- `expedition-unit-v1/pair-receipt.json`：113命令，18點的完整1MiB原版RAM、Regs、VGA、色盤、時間、完成／本體來源與後畫均相同。正常F7從自建存檔開始，存檔SHA與mtime不變。
- `expedition-gui-v1/receipt.json`：實際X冷F7、左行、攀繩、原貌／HD切換及退出0。根端目視新姿態33及原貌。`expedition-gui-pair-v2/receipt.json`冷重播3494筆實際日誌，每筆Regs／時間與最終GUI完整RAM／VGA／Regs／時間相同。
- `expedition-gui-pixels-v2/receipt.json`：實際X攀繩全遊戲區290304高清像素與獨立完整合成零差；原貌290304像素零差、兩次HD全畫布576000像素相同、中文HUD68544像素保留。實際X與程式擷取相同。完成來源166861689、本體166770124、暫停CPU166889374分開記錄，來源落後27685指令，零洞窟來源拒絕。

首版日誌工具誤讀最終metadata的`cycles`，實際欄位是`elapsed_cycles`，修正v2後以同一原始日誌重跑。首版像素工具錯取helper文字範圍，漏載`paint_portal`；v2只改正切片終點，同一來源通過。工具失敗與原資料保留，不改產品或重跑GUI挑結果。GUI互動工具的stdin緩衝需另送空行喚醒，造成首段多等候的原版時間；3494筆日誌包含實際等待，不把它改稱固定113命令的同一路徑。

本項未驗七姿態的每幀GUI、其他洞窟、未訪居民、未取得物品／法術、音效、實體手把或正式發行。未知來源仍保留原貌。
