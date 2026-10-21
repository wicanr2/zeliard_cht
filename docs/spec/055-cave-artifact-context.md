# 洞窟誤貼結局圖塊

狀態：CONFORMED，限定於本頁顯示修正與原生來源。2026-10-21使用者指出推廣片主角右上方的HD方塊，授權修正顯示錯誤。此任務不重開已結案的遊戲規則、通關及完整後段驗收。

## 原因及證據

已證實：Ada的cave-played-HD.png在視窗座標477,318出現24×24綠色圖塊，對應原生座標159,90的8×8區域。complete-yuu3-final-000-x120-y072.png來自結局YUU3／YUU4合成圖，原生圖樣只有一個色號2、其餘為0；yuu3-final資源組沒有記憶體綁定。現行CompleteArtwork.draw逐像素匹配局部圖樣後直接覆繪，沒有核對母圖上下文，因此將洞窟主角附近的偶合圖樣誤認成結局。

用原Ada輸入日誌、相同AdLib設定與正式版固定dosgolem來源重播5,064筆，原版暫存器、時鐘、觀察RAM及完成畫面SHA均相同。完成來源step247835674，目標觀察step247916034。artifact-placements.json確認該結局片段置於159,90，bindings=0；不是原版地形或新增角色姿態。原生地址與資源資料不變，本輪證據位址空間是320×200畫面座標及清單內母圖座標。

第一次研究工具沿舊預設MSCSTD，第二筆暫存器即不同，分類為重播設定差異。新增明確-adlib旗標，在獨立scratch建立MSCADLIB的RESOURCE.CFG，並啟用同前端AdLib契約後原紀錄重生通過；保留失敗目錄，不改日誌或原始資料。

## 修正契約

- 沿既有schema1、derivative_of與登錄檔名的-xNNN-yNNN座標，不改資產格式或重新生成美術。讀取時核母圖存在、來源組相同、子圖尺寸／座標有效，且子圖原生像素與母圖對應區完全相同。
- 沒有資源綁定的插畫局部圖塊，除了原有自身原生圖樣，還須核對同一偏移母圖在320×200內的可見原生上下文。全部可見且未遮罩排除的來源像素相同才可覆繪，任何不符即保留既有下層。
- 有資源綁定的來源沿原契約。整幅插畫、原版角色、專用HD、座標、碰撞、音樂、輸入及存檔均不變。
- 失配安全回退，不用黑色方塊遮住問題，也不從影片中塗改；正確母圖及畫面邊界裁切情境保留可匹配能力。

## 最小驗證與停止線

使用真實Ada重生來源驗證該結局片段被拒絕、來源RAM／畫面不變；獨立正例驗證正確母圖仍接入，破壞母圖其他像素時拒絕，畫面邊界裁切仍核可見上下文。以正常第一洞窟實際視窗確認方塊移除及主角HD保留。只做修正所需抽樣，不重跑通關。

工作區workplace/hd-cave-artifact-v1，重播入口[hd_gui_frame_probe.go](../../tools/hd_gui_frame_probe.go)，覆繪入口[complete_art.go](../../internal/prototype/hd/complete_art.go)。原始紀錄與素材保持本機，不上傳。

## 固定輸入

工具：dosgolem-go-ebiten:1.24-r1、Go1.24.13，dosgolem使用workplace/release-v100/dosgolem固定副本；影像分析eob-audio-capture:20261009-r3。輸入如下：

| 檔案 | SHA-256 |
|---|---|
| workplace/hd-cave-artifact-v1/user-report.png | 54ed5c0fed945989eab35c6372b611a3c5e60232982b12c1ea5619bb952db94c |
| workplace/ada-promo-v1/received/capture/boot1/cave-played-HD.png | 83099c9dcfc0471d8551b935d041d9d8b35c42c765ad2fe6051bd2e6bb1b7e39 |
| workplace/hd-cave-artifact-v1/replay-input/boot1/inputs.jsonl | 1e026d4694d56b53bfcec269412c7dde11772ee5c140e39da79468e6685014cc |
| workplace/hd-cave-artifact-v1/replay-source/source.memory | 04a839b3a2f343a29f6a0c3469dd2cf92288eca704987943292c964bc77d9338 |
| workplace/hd-cave-artifact-v1/replay-source/source.indexed | 09731a92bdbab893c8d3a82bb0a7f183c8123e0a2be2a9d2760f75ee5a4768f9 |
| workplace/full-hd-v1/manual-hd-v1/registered/complete-art.json | 5aa16ac9a8d3d12b19eb778a3f5983fd7b96e83c278b88a4e9af0a9be351758d |

## 實作與實際視窗

[母圖上下文回歸](../../internal/prototype/hd/complete_art_context_test.go)使用原Ada同狀態重生來源，原版來源SHA固定；誤貼片段拒絕，正確母圖、上下文失配及邊界裁切正反例通過，原版RAM不變。HD補圖層移動基準20次約7.62ms，不宣稱整體FPS。

修正版v.1.0.1-20261021已重新建置CHT／CHT-HD各三平台完整包，正常冷F7、進洞、接敵變色及特殊受傷HD、原貌切回及正常DOS退出通過，USR bytes／mtime不變。Linux啟動、Windows Wine輔助啟動與macOS通用架構／內容核對，沿既有工具產生新版收據。

Ada以修正後HD AppImage在NVIDIA RTX 2000 Ada重新錄製。新cave-played-HD.png已目視，主角右上方與左側誤貼片段消失，主角、洞窟及門框HD保留。新舊實錄為各次正常遊玩，不宣稱相同整局狀態；同狀態根因證據由先前原日誌重生提供。錄影原檔、輸入、同步音訊及十張代表畫面保留workplace/ada-promo-v101/received，原版存檔不變，正常退出。

新影片及封包交付dist-all/v.1.0.1-20261021；1.0.0及以前版本不覆寫。只修來源辨識，沒有改圖、碰撞、規則或音樂，不擴大完整通關、逐幀全遊戲及實體手把驗收範圍。
