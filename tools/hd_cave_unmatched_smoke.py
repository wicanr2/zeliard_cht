"""Docker／Xvfb正常冷F7，抽測049有地形遮罩的物件及原貌切回。"""
import ast
from pathlib import Path

platform_helper = Path("/project/tools/hd_cave_platform_smoke.py")
prefix = []
for node in ast.parse(platform_helper.read_text()).body:
    if isinstance(node, ast.Try):
        break
    prefix.append(node)
exec(compile(ast.Module(body=prefix, type_ignores=[]), str(platform_helper), "exec"))


def masked_ready(s):
    objects = s.get("cave_objects_hd") or {}
    return (ready(s) and objects.get("terrain_underlay")
            and objects.get("masked_known", 0) > 0)


try:
    xvfb = subprocess.Popen(["Xvfb", ":96", "-screen", "0", "1280x800x24", "-nolisten", "tcp"], env=env,
                            stdout=(O / "xvfb.log").open("wb"), stderr=subprocess.STDOUT)
    children.append(xvfb)
    end = time.monotonic() + 10
    while not Path("/tmp/.X11-unix/X96").exists():
        assert time.monotonic() < end
        time.sleep(.05)
    user = O / "user/game"
    user.mkdir(parents=True)
    shutil.copy2(save, user / "SAVETEST.usr")
    start()
    advance(12000000)
    for _ in range(6):
        key(0x39, 50000)
        advance(7950000)
    key(0x41, 500000)
    advance(2000000)
    key(0x15, 150000)
    advance(1500000)
    for _ in range(8):
        key(0xe, 100000)
        advance(200000)
    for scan in (0x1f, 0x1e, 0x2f, 0x12, 0x14, 0x12, 0x1f, 0x14):
        key(scan, 100000)
        advance(200000)
    key(0x1c, 100000)
    advance(16000000)
    key(0x4d, 8000000)
    advance(250000)
    align_town_entrance()
    key(0x48, 250000)
    advance(20000000)
    hd(True)
    wait("正常入洞", ready, 30)
    move(0x4b, lambda s: xvalue(s) <= 41, "正常走到繩索")
    compare_action("masked-object", 0x50, masked_ready, "正常物件地形遮罩HD")
    edge(0x1d, True)
    key(0x10, 500000)
    edge(0x1d, False)
    advance(1000000)
    invoke(["xdotool", "keydown", "y"])
    try:
        process.wait(timeout=60)
    finally:
        invoke(["xdotool", "keyup", "y"])
    assert process.returncode == 0
    assert h(user / "SAVETEST.usr") == h(save)
    assert (user / "SAVETEST.usr").stat().st_mtime_ns == save.stat().st_mtime_ns
    report = {"status": "PASS_CAVE_UNMATCHED_ASSETS_GUI_SCOPED", "binary_sha256": h(candidate),
              "script_sha256": h(Path(__file__)), "platform_helper_sha256": h(platform_helper),
              "catalog_helper_sha256": h(catalog), "helper_sha256": h(helper),
              "source_save_sha256": h(save), "boots": boots, "shots": shots, "pairs": pairs, "actions": actions,
              "normal_DOS_exit": 0, "save_unchanged": True,
              "limits": "只驗正常冷F7與049代表物件遮罩；八張補圖另沿17份保存正常來源，不當逐圖GUI或全遊戲HD。"}
    (O / "receipt.json").write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n")
    print(report["status"], flush=True)
except Exception:
    (O / "failed-observations.json").write_text(json.dumps({"shots": shots, "pairs": pairs, "actions": actions}, ensure_ascii=False, indent=2) + "\n")
    if process is not None and process.poll() is None:
        invoke(["import", "-window", window, str(O / "failed-live.png")])
        (O / "failed-state.json").write_text(json.dumps(state(), ensure_ascii=False, indent=2) + "\n")
    raise
finally:
    for child in reversed(children):
        stop(child)
