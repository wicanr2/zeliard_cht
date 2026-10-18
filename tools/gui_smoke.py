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


def scenario(name, env, controller=False, intro=False):
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


def town_scenario(env, hd_reference=None, hd_art=None):
    """僅用 X 視窗鍵盤事件走冷開機；語意結果須另行目視核對。"""
    capture = OUTPUT / "town-final.png"
    request = OUTPUT / "capture.request"
    command = [
        str(ROOT / "workplace/bin/zeliard-preview"),
        "-orig", str(ORIGINAL), "-font", str(FONT),
        "-text", str(ROOT / "text/host.json"),
        "-user-dir", str(OUTPUT / "user"),
        "-capture", str(capture), "-capture-request", str(request),
    ]
    if hd_reference and hd_art:
        command.extend(["-hd-reference", str(hd_reference), "-hd-art", str(hd_art)])
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
            time.sleep(1)
            invoke(["xdotool", "keyup", "Right"], env)
            events.append({"seconds":round(time.monotonic()-started, 3), "keyup":"Right"})
            time.sleep(1)
            shot("town-after-right")
            time.sleep(2)
            shot("town-released")
            hd_comparison = None
            if hd_reference and hd_art:
                # 正常工具列暫停後，同一原版畫格切主題，不以載入測試狀態代替。
                invoke(["xdotool", "mousemove", "--window", window, "620", "22", "click", "1"], env)
                time.sleep(.3)
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


def main(gameplay=False, hd_reference=None, hd_art=None):
    OUTPUT.mkdir(exist_ok=False)
    before = {p.name:digest(p) for p in ORIGINAL.iterdir() if p.is_file()}
    env = {**os.environ, "DISPLAY":":91", "LIBGL_ALWAYS_SOFTWARE":"1", "GOMAXPROCS":"2"}
    xvfb = None
    try:
        with (OUTPUT / "xvfb.log").open("wb") as log:
            xvfb = subprocess.Popen(["Xvfb", ":91", "-screen", "0", "1280x800x24", "-nolisten", "tcp"], stdout=log, stderr=log)
            deadline = time.monotonic() + 10
            while not pathlib.Path("/tmp/.X11-unix/X91").exists():
                if xvfb.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("Xvfb 未啟動")
                time.sleep(.1)
            scenarios = ([town_scenario(env, hd_reference, hd_art)] if gameplay else
                         [scenario("original", env), scenario("controller", env, True),
                          scenario("intro-zh", env, intro=True)])
        after = {p.name:digest(p) for p in ORIGINAL.iterdir() if p.is_file()}
        assert before == after, "原版檔案變動"
        receipt = {"kind":"prototype GUI smoke; not complete gameplay or physical controller validation",
                   "binary_sha256":digest(ROOT / "workplace/bin/zeliard-preview"),
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
    args = parser.parse_args()
    if bool(args.hd_reference) != bool(args.hd_art) or args.hd_art and not args.gameplay:
        parser.error("HD 驗證須同時指定 --gameplay、--hd-reference、--hd-art")
    OUTPUT = args.output
    main(args.gameplay, args.hd_reference, args.hd_art)
