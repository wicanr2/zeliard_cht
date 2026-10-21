# 洞窟FF41非零來源

狀態：CONFORMED，限定`FF41=1、FF40=0`。其他FF41非零情況維持拒絕。

玩家可見問題是正常GUI長按攻擊途中，完成候選的`01B1:FF41=1`被拒，洞窟退回原貌。尚未證實此旗標的遊戲語意，不以攻擊或傷害名稱取代原始定位。

## 固定來源與研究入口

- 原始`ZELIARD.EXE` SHA-256為`fe1c00b3c95f516f01266acaa94fbb04fe7436492926d89ba4e522b66012a341`。IDA使用固定`primary.segment`，SHA-256為`5e7f76080cc542eb06ca90ec349ad2ec0c5ab374c32acf59bc4c4307899a0a5b`。IDA base0對應dosgolem `01B1`段內偏移，不是EXE檔案偏移。
- [日誌重播工具](../../tools/hd_gui_rejected_probe.go)以Go1.24.13與專案既有dosgolem重播`cave-saved-player-gui-v2/boot1/inputs.jsonl`。2,590筆後到達原觀察點202514118步，完整RAM與原GUI相同，USR內容及mtime未變。
- 本機輸出`workplace/full-hd-v1/cave-ff41-source-v1/receipt.json`保存202440421步候選與202428860步主體。完整RAM SHA-256為`c43f08361e64e6abaa0b9f0f1818d42305556a89c5652937e28e60cbc7b52a6c`；主體為`0fa4da56ec7b063d0185e449410f04a562d224283e3a29ba5f4aa2856472dffc`。
- 已證實：原GUI來源的畫面、色盤、主體及完成RAM完全重生，拒絕理由仍是`unobserved actor profile`。此結果不代表HD通過。
- [IDA窄範圍工具](../../tools/ida_cave_ff41.py)對一次性資料庫匯出原始指令、bytes、函式名、xref與未知警示。語意待核對。

## 研究停止線

先核原版繪製consumer，再以真實保存來源完整重建原生遊戲區。只有證據能描述輸入、圖層、遮擋與邊界後，才接入正式解碼及HD；不能只移除`FF41`守衛。規則、輸入、數值、亂數與存檔不得改動。原資料與派生圖留本機，不進Git或Release。

## 已核證據與READY契約

- 已證實：IDA9.4的`01B1:3ACF..3AD4`及`3C36..3C3B`在`FF40=0`時，分別跳至`3B18`及`3C7F`。兩個FF41 consumer僅位於被跳過的非零分支：`3B00..3B11`及`3C67..3C78`。consumer以1及2選不同模板位移，此次不擴充這兩個非零FF40分支。
- 已證實：`70A8`將`FF45`複製至`FF41`；另一支`70C5`寫入1。`7094`先清FF40，未同時清FF41。僅記資料流，不推定旗標的玩家術語或清除時機。
- 窄範圍匯出為`workplace/full-hd-v1/cave-ff41-ida-v4/evidence.json`，460條原始指令。原DB的`3A85`三byte JMP止於`3A88`，一次性副本由此補分析。正式DB與原名未修改。輸入程式的`302C..4F8C`與`6F9B..7183`均和真實重生RAM相同。
- 已證實：`cave-ff41-decode-study-v1/receipt.json`僅複製解碼輸入並將FF41由1改為0，以隔離既有守衛。原版機器與原始檔未改。重建224×144遊戲區共32256像素零差；ActorRaster鍵`fca3d24d4f7337fb31cd034a1ed667824aaf482cf03b41a3a463450eefe78a4c`已存在原55張圖面。
- 正式typed input仍為完整／主體RAM、indexed畫面、色盤及paint時間線。只允許主體`FF41=1`且`FF40=0`通過原profile守衛。其餘裝備、組合、來源雜湊及整幅原生像素驗證照舊。不存在新的模板、動作或持續狀態。
- 垂直鏈為固定EXE與USR→原GUI日誌正常鍵序→原版完成／主體來源→既有圖層解碼→既有高清圖面。USR與設定路徑不變，呈現器維持唯讀。
- 實作驗收：直接使用未改動的真實保存來源完整重建；比較FF41=0與1在FF40=0時的圖層及HD像素；未知值、非零FF40與既有profile限制仍拒絕；重播原GUI日誌核對同一觀察RAM及完成來源SHA；正常平台路徑回歸。新的實際視窗尚待抽樣，不以前一失敗GUI冒稱新GUI通過。
- 限制：此修正沒有完成其他FF41模式、其他敵人／首領或後段場景，也沒有新增圖面。其作用是避免已知、原版未使用的值阻擋既有HD。

## 實作與限定驗收

[來源解碼](../../internal/prototype/hd/cave_sources.go)只補READY所列情況。[正式檢查](../../internal/prototype/hd/cave_ff41_test.go)直接讀未改動的正常保存來源，原生32256像素零差；0／1邊界的完整圖層及576000個高清畫布像素相同，來源未寫入。原55主角及兩個動態物件已有對應圖面。其他FF41值、非零FF40及未知裝備／組合均拒絕。

44項圖面／來源頂層檢查、四項focused檢查及三項正常／來源檢查無skip／fail。兩路13點完整原版狀態與USR bytes／mtime相同；普通平台完成來源由211增為213、拒絕由42減為40，魔法石路徑474／92不變。未將剩餘拒絕計成HD。

[視窗抽樣](../../tools/hd_cave_ff41_smoke.py)的v3正常冷F7、步行與出劍後放開，擷取FF41=1且FF40=0來源。原貌／HD／切回的CPU、RAM、indexed與palette相同，576000遊戲區像素切回相同，正常DOS退出0、存檔不變，根端已目視。v1抽樣時段不足、v2無關姿態暫停錯過來源，失敗driver與日誌保留，不冒稱通過。

候選`workplace/full-hd-v1/cave-ff41-runtime-v3/zeliard-preview` SHA-256為`6383ec84d59bcd1da1694ecc29d165dd6e976fd1d545f8d2b386ebb8090a3e15`，vet通過且乾淨重建binary相同。收尾入口`workplace/full-hd-v1/cave-ff41-final-audit-v1.json`由[稽核工具](../../tools/hd_ff41_audit.py)重生至新檔。封包、版號、影片及README縮圖未更新。下一步轉向仍被拒的組合來源及未接城鎮，不重做此來源或既有圖面。
