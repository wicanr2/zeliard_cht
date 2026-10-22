# 大字介面雙版本交付與推廣片

狀態：限定CONFORMED，2026-10-22。先訂READY契約再執行。使用者確認064候選後，授權重新打包完整版與製作推廣影片。沿既有CHT／CHT-HD各三平台範圍，採新版`v.1.0.5-20261022`，不覆寫舊版、tag、Release或影片。

來源為已完成064的正式前端，沿相同原版、中文、3倍手繪圖庫與音源。以既有雙版本工具從乾淨來源副本重建；字型改用`workplace/ui-text-v1/rounded-font`，確保新增字級文字有目標尺寸圓體遮罩。兩版皆提供16／24／32系統字級、預設24與保存，工具列最高24。Linux為AppImage，Windows為x86_64 ZIP，macOS為Intel／Apple Silicon通用ZIP。完整遊戲與HD素材只留本機。

唯一建置工作區為`workplace/release-v105/`。正式六包在`dist-all/v.1.0.5-20261022/full-local/`，影片在同版`promo/`，限定啟動與字級操作收據在同版`smoke/`。中間錄影與剪輯留建置工作區，不另建工作根。

推廣片沿既有60秒風格，新錄正式1.0.5 HD AppImage的字級選擇、特大作弊與速度頁。可以依使用者先前授權沿用前片：城鎮來源1.0.2，對話／綠色平台HD連續證明來源1.0.4，均在片中明列版本，不能冒稱1.0.5遊玩實錄。配樂沿可追溯的原版標題曲並明列非同期聲；證明段保持各自連續，不在區間切換HD或變速。

驗收為六包內容、架構、逐檔manifest、SHA、來源與版本，以及Linux兩版原生／Windows兩版Wine輔助啟動，正式HD包三檔字級與冷重開，影片音畫、字幕、來源及連續段。macOS只核通用結構，實體手把與真機限制照實保留。使用者已接受幀率，不重開30fps、全遊戲或通關閘門。完成後只清本輪明列的重複組裝與解壓層，保留正式包、影片、原始錄影、固定來源與收據。

建置入口：[build_dual_release.sh](../../tools/build_dual_release.sh)，準備時使用`ZELIARD_RELEASE_FONT=/project/workplace/ui-text-v1/rounded-font`。封包入口：[package_dual_release.py](../../tools/package_dual_release.py)，六包核對：[audit_dual_release.py](../../tools/audit_dual_release.py)，限定字級：[ui_text_smoke.py](../../tools/ui_text_smoke.py)。影片與收尾工具沿本規格及WORKLOG登錄。

錄製由字級工具的`ZELIARD_UI_WORK`、`ZELIARD_UI_BUNDLE`、`ZELIARD_UI_OUTPUT`及`ZELIARD_UI_RECORD`指向1.0.5正式HD解壓包、同版smoke/toolbar及本工作區錄影根。製片：[make_readable_promo.py](../../tools/make_readable_promo.py)，音畫：[verify_promo.py](../../tools/verify_promo.py)，來源與重用證明：[verify_readable_promo.py](../../tools/verify_readable_promo.py)，登錄：[finalize_readable_release.py](../../tools/finalize_readable_release.py)。具體容器掛載及結果見WORKLOG與固定工具快照。

清理入口：[cleanup_readable_release.py](../../tools/cleanup_readable_release.py)，先核七份成品SHA，再沿既有封包清理及明列的剪輯／失敗Wine重複層，收據為工作區`cleanup.json`與`cleanup-media.json`。保留原始錄影、舊版與固定建置來源。

六包逐檔manifest、來源、模式與架構通過；Linux兩版原生、Windows兩版Wine輔助啟動通過，macOS只核通用結構。正式HD包三檔字級、14張設定頁與冷重開通過，DOS steps／cycles皆0，存檔不變。根manifest登錄六包與60秒影片共七份產物，登錄收據為`repack-receipt.json`。

影片28–32秒為16字級、32–36秒為24、36–40秒為32，分別取自正式錄影的連續4秒節選。音畫、字幕與來源通過：mean -18.6dB、peak -4.5dB，無長黑幀或兩秒靜音。540個重用HD畫格及720個新錄設定編碼畫格與來源比對通過；X11原錄的時間戳以製片相同`-r30`節奏重取樣，仍不當作遊戲30fps證據。來源版號、生命鎖定與配樂非同期聲標示已目視。

README使用[正式字級頁](../images/text-size-v105.png)及[特大作弊頁](../images/system-text-v105.png)，[圖片來源](../images/system-text-v105-provenance.json)可回查實際包。清理約3.78GiB重複層後，七份成品SHA再核對不變。舊版及完整錄影、固定來源、二進位與收據保留；不宣稱Windows／macOS真機、實體手把、完整遊玩或全遊戲HD已驗。
