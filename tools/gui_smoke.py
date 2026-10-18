"""在具 Xvfb／xdotool 的 Docker 映像中驗證前端，結束時回收所有子程序。"""

import hashlib
import json
import os
import pathlib
import subprocess
import time


ROOT = pathlib.Path("/project")
OUTPUT = ROOT / "workplace/gui-smoke-v4"
FONT = pathlib.Path("/fonts/unifont.hex.gz")
ORIGINAL = pathlib.Path("/orig")


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def stop(process):
    if process is not None and process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait(timeout=5)


def invoke(args, env):
    return subprocess.run(args, env=env, check=True, timeout=5, capture_output=True)


def scenario(name, env, controller=False, intro=False, scroll=False):
    capture = OUTPUT / f"{name}.png"
    with (OUTPUT / f"{name}.log").open("wb") as log:
        command = [
            str(ROOT / "workplace/bin/zeliard-preview"),
            "-orig", str(ORIGINAL), "-font", str(FONT),
            "-text", str(ROOT / "text/host.json"),
            "-user-dir", str(OUTPUT / "user"),
            "-capture", str(capture), "-capture-frames", "240",
        ]
        if intro:
            command.extend(["-intro-reference", str(ROOT / "workplace/probe-intro-clock3000/screen.bin"),
                            "-intro-text", str(ROOT / "text/intro-prototype.json")])
        if scroll:
            command.extend(["-intro-scroll", str(ROOT / "text/intro-scroll-prototype.tsv"),
                            "-capture-frames", "600"])
        process = subprocess.Popen(command, cwd=ROOT, env=env, stdout=log, stderr=log)
        try:
            deadline = time.monotonic() + 20
            window = None
            while time.monotonic() < deadline:
                if process.poll() is not None:
                    raise RuntimeError(f"前端提早結束：{(OUTPUT / (name + '.log')).read_text()}")
                result = subprocess.run(
                    ["xdotool", "search", "--onlyvisible", "--pid", str(process.pid)],
                    env=env, capture_output=True, timeout=3,
                )
                if result.returncode == 0 and result.stdout.strip():
                    window = result.stdout.decode().splitlines()[-1]
                    break
                time.sleep(.1)
            if window is None:
                raise RuntimeError("找不到前端視窗")
            invoke(["xdotool", "windowfocus", window], env)
            time.sleep(.3)
            if controller:
                invoke(["xdotool", "mousemove", "--window", window, "420", "22", "click", "1"], env)
                time.sleep(.3)
            deadline = time.monotonic() + 30
            metadata = pathlib.Path(str(capture) + ".json")
            while not metadata.exists():
                if process.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("未取得完成繪製訊號")
                time.sleep(.05)
            state = json.loads(metadata.read_text())
            assert state["panel"] == controller, state
            if intro:
                assert state["intro_stamps"] == 1, state
            if scroll:
                assert state["intro_scroll_sources"] >= 3 and state["intro_scroll_rows"] >= 20, state
            time.sleep(.15)
            # 等 Draw 完成後擷取實際 X 視窗，避免把尚未處理滑鼠事件的畫面當成果。
            invoke(["import", "-window", window, str(OUTPUT / f"{name}-window.png")], env)
            code = process.wait(timeout=30)
            if code != 0 or not capture.exists():
                raise RuntimeError(f"GUI 驗證失敗：{code}，{(OUTPUT / (name + '.log')).read_text()}")
            return {"command":command, "exit_code":code, "draw_png_sha256":digest(capture),
                    "window_png_sha256":digest(OUTPUT / f"{name}-window.png"),
                    "controller_opened_by_mouse":controller, "frame":state}
        finally:
            stop(process)


def town_scenario(env, hd_reference=None, hd_art=None, king=False, dialogue=False, king_hd=False):
    """僅用 X 視窗鍵盤事件走冷開機；語意結果須另行目視核對。"""
    capture = OUTPUT / "town-final.png"
    request = OUTPUT / "capture.request"
    command = [
        str(ROOT / "workplace/bin/zeliard-preview"),
        "-orig", str(ORIGINAL), "-font", str(FONT),
        "-text", str(ROOT / "text/host.json"),
        "-user-dir", str(OUTPUT / "user"),
        "-capture", str(capture), "-capture-request", str(request),
        "-observe",str(OUTPUT / "observe.json"),
    ]
    if hd_reference and hd_art:
        command.extend(["-hd-reference", str(hd_reference), "-hd-art", str(hd_art)])
    if dialogue:
        command.extend(["-dialogue-text",str(ROOT / "text/dialogue-prototype.tsv")])
    if king_hd:
        command.extend(["-king-reference-a",str(ROOT / "workplace/probe-king-memory-84m/screen.bin"),
                        "-king-reference-b",str(ROOT / "workplace/probe-king-pose-96m/screen.bin"),
                        "-king-art",str(ROOT / "workplace/hd-king-prototype.png"),
                        "-king-speaking-art",str(ROOT / "workplace/hd-king-speaking-prototype.png")])
    events = []
    reference = ROOT / "workplace/probe-newgame-60m"
    indexed = (reference / "screen.bin").read_bytes()
    palette = (reference / "palette.bin").read_bytes()
    # 只辨認已目視核對的城鎮 LIFE／PLACE 與地名區，不推論座標或規則。
    offsets = [3 * (y * 320 + x) + c for y in range(162, 184)
               for x in range(48, 184) for c in range(3)]
    expected = bytes(palette[indexed[i // 3] * 3 + i % 3] for i in offsets)

    def town_visible(path):
        pixels = invoke(["convert", str(path), "-crop", "960x600+0+48", "+repage",
                         "-sample", "320x200!", "-depth", "8", "rgb:-"], env).stdout
        assert len(pixels) == 320 * 200 * 3, len(pixels)
        return bytes(pixels[i] for i in offsets) == expected

    def display_pixels(path):
        data = invoke(["convert", str(path), "-crop", "960x600+0+48", "+repage",
                       "-depth", "8", "rgb:-"], env).stdout
        assert len(data) == 960 * 600 * 3, len(data)
        return data
    with (OUTPUT / "town.log").open("wb") as log:
        process = subprocess.Popen(command, cwd=ROOT, env=env, stdout=log, stderr=log)
        try:
            deadline = time.monotonic() + 15
            window = None
            while time.monotonic() < deadline:
                if process.poll() is not None:
                    raise RuntimeError("城鎮流程的前端提早結束")
                result = subprocess.run(
                    ["xdotool", "search", "--onlyvisible", "--pid", str(process.pid)],
                    env=env, capture_output=True, timeout=3,
                )
                if result.returncode == 0 and result.stdout.strip():
                    window = result.stdout.decode().splitlines()[-1]
                    break
                time.sleep(.1)
            if window is None:
                raise RuntimeError("找不到城鎮流程視窗")
            invoke(["xdotool", "windowfocus", window], env)
            started = time.monotonic()

            def wait_until(seconds):
                while time.monotonic() - started < seconds:
                    if process.poll() is not None:
                        raise RuntimeError("城鎮流程提前退出")
                    time.sleep(.02)

            def shot(name):
                invoke(["import", "-window", window, str(OUTPUT / (name + ".png"))], env)
                events.append({"seconds":round(time.monotonic()-started, 3), "capture":name})

            # 先看畫面再送鍵。片頭轉場會忽略輸入，不能把牆鐘秒數當成畫面就緒。
            for number in range(1, 29):
                wait_until(number * 2)
                name = f"checkpoint-{number}"
                shot(name)
                if town_visible(OUTPUT / (name + ".png")):
                    events.append({"seconds":round(time.monotonic()-started, 3),
                                   "town_hud_matches_reference":True})
                    break
                invoke(["xdotool", "keydown", "space"], env)
                time.sleep(.12)
                invoke(["xdotool", "keyup", "space"], env)
                events.append({"seconds":round(time.monotonic()-started, 3),
                               "key":"Space", "held_seconds":.12})
            else:
                raise RuntimeError("未在時限內辨認到城鎮 HUD；不可把截圖成功算成新局成功")
            shot("town-before-right")
            invoke(["xdotool", "keydown", "Right"], env)
            events.append({"seconds":round(time.monotonic()-started, 3), "keydown":"Right"})
            time.sleep(.65 if king else 1)
            invoke(["xdotool", "keyup", "Right"], env)
            events.append({"seconds":round(time.monotonic()-started, 3), "keyup":"Right"})
            time.sleep(1)
            shot("town-after-right")
            time.sleep(2)
            shot("town-released")
            if king:
                king_ref = ROOT / "workplace/probe-king-0-84m"
                kr = (king_ref / "screen.bin").read_bytes()
                kp = (king_ref / "palette.bin").read_bytes()
                name_offsets = [3*(y*320+x)+c for y in range(174,184) for x in range(54,184) for c in range(3)]
                expected_name = bytes(kp[kr[i//3]*3+i%3] for i in name_offsets)

                def native_shot(name):
                    shot(name)
                    return invoke(["convert", str(OUTPUT / (name + ".png")), "-crop", "960x600+0+48", "+repage",
                                   "-sample", "320x200!", "-depth", "8", "rgb:-"], env).stdout

                # 視窗更新速度不等於探針的指令預算。小步前進、嘗試進門，再核對實際地名。
                for attempt in range(30):
                    invoke(["xdotool", "keydown", "Up"], env)
                    time.sleep(.15)
                    invoke(["xdotool", "keyup", "Up"], env)
                    time.sleep(.25)
                    pixels = native_shot(f"king-entry-{attempt}")
                    if bytes(pixels[i] for i in name_offsets) == expected_name:
                        events.append({"king_room_name_matches_reference":True,"entry_attempt":attempt})
                        break
                    invoke(["xdotool", "keydown", "Right"], env)
                    time.sleep(.1)
                    invoke(["xdotool", "keyup", "Right"], env)
                    events.append({"entry_attempt":attempt,"key_sequence":["Up","Right"],"held_seconds":[.15,.1]})
                else:
                    raise RuntimeError("未從視窗正常進入國王房間")
                # 只辨認已目視核對的國王首段文字，避免把空對話框當成對話完成。
                text_offsets = [3*(y*320+x)+c for y in range(98,138) for x in range(54,262) for c in range(3)]
                expected_text = bytes(kp[kr[i//3]*3+i%3] for i in text_offsets)
                if dialogue:
                    expected_pixels = invoke(["convert",str(ROOT / "workplace/dialogue-v2/084000000.png"),
                                              "-sample","320x200!","-depth","8","rgb:-"],env).stdout
                    expected_text = bytes(expected_pixels[i] for i in text_offsets)
                for attempt in range(24):
                    time.sleep(.5)
                    name = f"king-checkpoint-{attempt}"
                    pixels = native_shot(name)
                    if bytes(pixels[i] for i in text_offsets) == expected_text:
                        events.append({"king_first_dialogue_matches_reference":True, "reference_sha256":digest(king_ref / "screen.bin")})
                        break
                else:
                    raise RuntimeError("未辨認到國王首段對話；不以空框或肖像當作交談成功")
                shot("king-dialogue")
                if king_hd:
                    invoke(["xdotool","mousemove","--window",window,"620","22","click","1"],env)
                    time.sleep(.3)
                    shot("king-original-paused")
                    original = display_pixels(OUTPUT / "king-original-paused.png")
                    invoke(["xdotool","mousemove","--window",window,"250","22","click","1"],env)
                    time.sleep(.3)
                    shot("king-hd-paused")
                    enhanced = display_pixels(OUTPUT / "king-hd-paused.png")
                    invoke(["xdotool","mousemove","--window",window,"80","22","click","1"],env)
                    time.sleep(.3)
                    shot("king-original-restored")
                    restored = display_pixels(OUTPUT / "king-original-restored.png")
                    portrait_changed, borders_changed, outside = 0,0,0
                    for y in range(600):
                        for x in range(960):
                            i=3*(y*960+x)
                            if original[i:i+3] == enhanced[i:i+3]:
                                continue
                            if 339<=x<621 and 72<=y<258:
                                portrait_changed+=1
                            elif x<144 or x>=816:
                                borders_changed+=1
                            else:
                                outside+=1
                    assert portrait_changed>50000 and outside==0,(portrait_changed,borders_changed,outside)
                    assert original==restored,"國王畫格切回原貌後不同"
                    comparison={"portrait_pixels_changed":portrait_changed,"border_pixels_changed":borders_changed,
                                "outside_changed":outside,"original_restored_exactly":True,
                                "source_state":"normal dialogue paused through toolbar; no core execution during switches",
                                "closed_art_sha256":digest(ROOT / "workplace/hd-king-prototype.png"),
                                "speaking_art_sha256":digest(ROOT / "workplace/hd-king-speaking-prototype.png")}
                    (OUTPUT / "king-hd-region-comparison.json").write_text(json.dumps(comparison,indent=2)+"\n")
                    events.append({"king_hd_same_state_comparison":comparison})
                    invoke(["xdotool","mousemove","--window",window,"250","22","click","1"],env)
                    invoke(["xdotool","mousemove","--window",window,"620","22","click","1"],env)
                    time.sleep(.2)
                invoke(["xdotool", "keydown", "space"], env)
                time.sleep(.12)
                invoke(["xdotool", "keyup", "space"], env)
                events.append({"seconds":round(time.monotonic()-started, 3), "key":"Space", "held_seconds":.12})
                time.sleep(2)
                shot("king-after-advance")
                if dialogue:
                    second = invoke(["convert",str(ROOT / "workplace/dialogue-v2/106000000.png"),
                                     "-sample","320x200!","-depth","8","rgb:-"],env).stdout
                    second_text = bytes(second[i] for i in text_offsets)
                    marker = invoke(["convert",str(ROOT / "workplace/dialogue-v2/096000000.png"),
                                     "-sample","320x200!","-depth","8","rgb:-"],env).stdout
                    marker_offsets = [3*(y*320+x)+c for y in range(139,147) for x in range(156,164) for c in range(3)]
                    expected_marker = bytes(marker[i] for i in marker_offsets)
                    second_seen = False
                    for attempt in range(60):
                        time.sleep(.5)
                        name = f"king-progress-{attempt}"
                        pixels = native_shot(name)
                        if town_visible(OUTPUT / (name + ".png")):
                            assert second_seen, "離開對話前未顯示完整第二段中文"
                            events.append({"king_dialogue_completed_normally":True,"second_translation_seen":True})
                            shot("king-returned-town")
                            break
                        text_complete = bytes(pixels[i] for i in text_offsets) == second_text
                        if text_complete:
                            second_seen = True
                        if text_complete or bytes(pixels[i] for i in marker_offsets) == expected_marker:
                            invoke(["xdotool","keydown","space"],env)
                            time.sleep(.12)
                            invoke(["xdotool","keyup","space"],env)
                            events.append({"advance_after_observation":True,"second_translation_complete":text_complete})
                    else:
                        raise RuntimeError("國王中文對話未正常結束")
            hd_comparison = None
            if hd_reference and hd_art:
                # 正常工具列暫停後，同一原版畫格切主題，不以載入測試狀態代替。
                invoke(["xdotool", "mousemove", "--window", window, "620", "22", "click", "1"], env)
                time.sleep(.3)
                # 國王驗證可能沿用已開啟的 HD；每次比較先明確切到原貌。
                invoke(["xdotool", "mousemove", "--window", window, "80", "22", "click", "1"], env)
                time.sleep(.2)
                shot("town-original-paused")
                original = display_pixels(OUTPUT / "town-original-paused.png")
                invoke(["xdotool", "mousemove", "--window", window, "250", "22", "click", "1"], env)
                time.sleep(.3)
                shot("town-hd-paused")
                hd = display_pixels(OUTPUT / "town-hd-paused.png")
                invoke(["xdotool", "mousemove", "--window", window, "80", "22", "click", "1"], env)
                time.sleep(.3)
                shot("town-returned-original-paused")
                restored = display_pixels(OUTPUT / "town-returned-original-paused.png")
                changed_inside, changed_outside = 0, 0
                for y in range(600):
                    for x in range(960):
                        i = 3 * (y * 960 + x)
                        if original[i:i+3] != hd[i:i+3]:
                            if x < 144 or x >= 816:
                                changed_inside += 1
                            else:
                                changed_outside += 1
                assert changed_inside > 0 and changed_outside == 0, (changed_inside, changed_outside)
                assert original == restored, "原貌切回後畫格不同"
                hd_comparison = {
                    "source_state":"frontend paused through normal toolbar; no core execution during switches",
                    "display_rectangles":[[0,48,144,600],[816,48,144,600]],
                    "pixels_changed_inside":changed_inside, "pixels_changed_outside":changed_outside,
                    "original_restored_exactly":True,
                    "original_rgb_sha256":hashlib.sha256(original).hexdigest(),
                    "hd_rgb_sha256":hashlib.sha256(hd).hexdigest(),
                    "art_sha256":digest(hd_art), "reference_sha256":digest(hd_reference),
                    "scope":"decorative side borders only; not full-game HD or RNG parity",
                }
                (OUTPUT / "hd-region-comparison.json").write_text(json.dumps(hd_comparison,indent=2)+"\n")
                invoke(["xdotool", "mousemove", "--window", window, "250", "22", "click", "1"], env)
                time.sleep(.3)
            request.touch(exist_ok=False)
            metadata = pathlib.Path(str(capture) + ".json")
            deadline = time.monotonic() + 25
            while not metadata.exists():
                if process.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("城鎮流程未取得 Draw 收據")
                time.sleep(.05)
            frame = json.loads(metadata.read_text())
            assert frame["mode"] == 0x13 and not frame["panel"], frame
            assert not frame["held_keys"], frame
            if dialogue:
                assert frame["dialogue_glyphs"]==241 and frame["dialogue_paragraphs"]==2,frame
            if hd_comparison:
                assert frame["hd_enabled"] and frame["hd_matched"], frame
            time.sleep(.15)
            shot("town-final-window")
            code = process.wait(timeout=10)
            assert code == 0, code
            return {"command":command, "exit_code":code, "events":events, "frame":frame,
                    "input_method":"X11 keyboard events, no state loading or memory writes",
                    "town_reference_sha256":digest(reference / "screen.bin"),
                    "town_hud_rect":[48,162,136,22],
                    "hd_comparison":hd_comparison,
                    "semantic_result":"requires visual inspection; no full-game completion claim"}
        finally:
            stop(process)


def main(gameplay=False, hd_reference=None, hd_art=None, scroll=False, king=False, dialogue=False, king_hd=False):
    OUTPUT.mkdir(exist_ok=False)
    binary_before = digest(ROOT / "workplace/bin/zeliard-preview")
    before = {p.name:digest(p) for p in ORIGINAL.iterdir() if p.is_file()}
    env = {**os.environ, "DISPLAY":":91", "LIBGL_ALWAYS_SOFTWARE":"1", "GOMAXPROCS":"2", "LP_NUM_THREADS":"2"}
    xvfb = None
    try:
        with (OUTPUT / "xvfb.log").open("wb") as log:
            xvfb = subprocess.Popen(["Xvfb", ":91", "-screen", "0", "1280x800x24", "-nolisten", "tcp"], stdout=log, stderr=log)
            deadline = time.monotonic() + 10
            while not pathlib.Path("/tmp/.X11-unix/X91").exists():
                if xvfb.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("Xvfb 未啟動")
                time.sleep(.1)
            scenarios = ([scenario("intro-scroll", env, scroll=True)] if scroll else
                         [town_scenario(env, hd_reference, hd_art, king, dialogue, king_hd)] if gameplay else
                         [scenario("original", env), scenario("controller", env, True),
                          scenario("intro-zh", env, intro=True)])
        after = {p.name:digest(p) for p in ORIGINAL.iterdir() if p.is_file()}
        assert before == after, "原版檔案變動"
        assert binary_before == digest(ROOT / "workplace/bin/zeliard-preview"), "GUI 執行途中二進位被替換"
        receipt = {"kind":"prototype GUI smoke; not complete gameplay or physical controller validation",
                   "binary_sha256":binary_before,
                   "font_sha256":digest(FONT), "originals_unchanged":True, "scenarios":scenarios,
                   "script_sha256":digest(pathlib.Path(__file__)),
                   "outputs":{p.name:digest(p) for p in OUTPUT.iterdir() if p.is_file()}}
        (OUTPUT / "receipt.json").write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
        print(json.dumps(receipt,ensure_ascii=False,indent=2))
    finally:
        stop(xvfb)


if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=pathlib.Path, default=OUTPUT)
    parser.add_argument("--gameplay", action="store_true", help="用實際視窗按鍵驗證新局與移動")
    parser.add_argument("--hd-reference", type=pathlib.Path)
    parser.add_argument("--hd-art", type=pathlib.Path)
    parser.add_argument("--intro-scroll", action="store_true", help="自然片頭的多行中文覆繪")
    parser.add_argument("--king", action="store_true", help="新局正常進城堡與國王交談，須配合 --gameplay")
    parser.add_argument("--dialogue", action="store_true", help="國王兩段中文與正常對話結束，須配合 --king")
    parser.add_argument("--king-hd",action="store_true",help="國王中文與 HD 圖面正常切換，須配合 --dialogue")
    args = parser.parse_args()
    if bool(args.hd_reference) != bool(args.hd_art) or args.hd_art and not args.gameplay:
        parser.error("HD 驗證須同時指定 --gameplay、--hd-reference、--hd-art")
    OUTPUT = args.output
    if args.intro_scroll and args.gameplay:
        parser.error("--intro-scroll 與 --gameplay 分開驗證")
    if args.king and not args.gameplay:
        parser.error("--king 須配合 --gameplay")
    if args.dialogue and not args.king:
        parser.error("--dialogue 須配合 --king")
    if args.king_hd and not args.dialogue:
        parser.error("--king-hd 須配合 --dialogue")
    main(args.gameplay, args.hd_reference, args.hd_art, args.intro_scroll, args.king, args.dialogue, args.king_hd)
