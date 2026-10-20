"""Docker／Xvfb內以正常冷讀、入洞、HD切回及回城驗證新來源辨識。"""
import ast
from pathlib import Path

helper = Path("/project/tools/integration_smoke.py")
prefix = []
for node in ast.parse(helper.read_text()).body:
    if isinstance(node, ast.Try):
        break
    prefix.append(node)
exec(compile(ast.Module(body=prefix, type_ignores=[]), str(helper), "exec"))
names[0x39] = "space"
candidate = Path(os.environ.get("ZELIARD_GUI_BINARY", str(R/"workplace/full-hd-v1/cave-catalog-runtime-v1/zeliard-preview")))
ice_art = os.environ.get("ZELIARD_CAVE_ICE_ART", "")
peligro_art = os.environ.get("ZELIARD_CAVE_PELIGRO_ART", "")
mpp3_art = os.environ.get("ZELIARD_CAVE_MPP3_ART", "")
assert candidate.is_file()
original_start = start


def start():
    # 沿既有啟動器，實際命令會保存新程式路徑與追加參數。
    original_popen = subprocess.Popen
    def launch(command, *args, **kwargs):
        if isinstance(command, list) and command[0] == str(B/"zeliard-preview"):
            command[0] = str(candidate)
            command.extend(["-audio", "off", "-realtime"])
            if ice_art:
                command.extend(["-cave-ice-art", ice_art])
            if peligro_art:
                command.extend(["-cave-peligro-art", peligro_art])
            if mpp3_art:
                command.extend(["-cave-mpp3-art", mpp3_art])
        return original_popen(command, *args, **kwargs)
    subprocess.Popen = launch
    try:
        original_start()
    finally:
        subprocess.Popen = original_popen


def ready(s):
    source = s.get("cave_frame_presentation") or {}
    return s["cave_terrain_hd_matched"] and source.get("used") and source.get("source_map_key") == "MP10" and source.get("source_terrain_key") == "MPP1"


def align_town_entrance():
    import struct
    reference = R/"workplace/combat-v1/entry-v3/cave-near.memory"
    raw = reference.read_bytes()[0x1b10:0x11b10]
    target = struct.unpack_from("<H", raw, 0x80)[0]+raw[0x83]
    for _ in range(24):
        value = state()["native_primary_0080"]+state()["native_primary_0083"]
        if value == target:
            return
        scan = 0x4b if value > target else 0x4d
        edge(scan, True)
        try:
            wait("正常步行對齊洞口", lambda s: s["native_primary_0080"]+s["native_primary_0083"] != value, 5)
        finally:
            edge(scan, False)
        advance(100000)
    raise RuntimeError("正常步行未對齊洞口")


def align_cave_exit():
    def delta(s):
        source = s.get("cave_frame_presentation") or {}
        portals = (s.get("cave_portal_hd") or {}).get("placements") or []
        if not portals or "source_actor_x" not in source:
            return None
        # 原生角色24寬、入口40寬，對齊中心需將角色左界放入口左界+8。
        return source["source_actor_x"]-portals[0]["X"]-8
    for _ in range(24):
        value = delta(wait("可見原版入口", lambda s: delta(s) is not None, 15))
        if value == 0:
            return
        scan = 0x4b if value > 0 else 0x4d
        edge(scan,True)
        try:
            wait("正常走回入口中心", lambda s: delta(s) is not None and delta(s) != value, 5)
        finally:
            edge(scan,False)
        advance(100000)
    raise RuntimeError("正常步行未對齊洞內出口")


def frozen(name):
    before = state()["frames"]
    s = wait("暫停畫面更新", lambda s: s["paused"] and s["frames"] > before, 15)
    file = directory/(name+".png")
    invoke(["import", "-window", window, str(file)])
    shots.append({"name": name, "image_sha256": h(file), "state": s})
    return s


pairs = []


def compare(name):
    wait("正常洞窟來源與HD", ready, 30)
    click(644, 24)
    wait("暫停", lambda s: s["paused"] and not held(s), 15)
    hd(False)
    original = frozen(name+"-original")
    hd(True)
    painted = frozen(name+"-hd")
    assert ready(painted)
    hd(False)
    returned = frozen(name+"-returned")
    for key in ("steps", "cycles", "oracle_memory_sha256", "oracle_indexed_sha256", "oracle_palette_sha256"):
        assert original[key] == painted[key] == returned[key], (name, key)
    pairs.append({"name": name, "original": name+"-original.png", "hd": name+"-hd.png", "returned": name+"-returned.png"})
    hd(True)
    click(644, 24)
    wait("繼續", lambda s: not s["paused"], 15)
    print("PAIR", name, flush=True)


try:
    xvfb = subprocess.Popen(["Xvfb", ":96", "-screen", "0", "1280x800x24", "-nolisten", "tcp"], env=env,
                            stdout=(O/"xvfb.log").open("wb"), stderr=subprocess.STDOUT)
    children.append(xvfb)
    end = time.monotonic()+10
    while not Path("/tmp/.X11-unix/X96").exists():
        assert time.monotonic() < end
        time.sleep(.05)
    user = O/"user/game"
    user.mkdir(parents=True)
    shutil.copy2(B/"starter/SAVETEST.usr", user/"SAVETEST.usr")
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
    compare("cave-entry")
    key(0x4b, 500000)
    advance(500000)
    compare("cave-walking")
    key(0x4d, 500000)
    advance(500000)
    align_cave_exit()
    key(0x48, 250000)
    advance(20000000)
    wait("正常回城及撤除洞窟", lambda s: not (s.get("cave_frame_presentation") or {}).get("used") and s["town_scenery_hd_matched"], 30)
    shot("returned-town")
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
    assert h(user/"SAVETEST.usr") == h(B/"starter/SAVETEST.usr")
    report = {"status": "PASS_CAVE_CATALOG_GUI_SCOPED", "binary_sha256": h(candidate),
              "ice_art_directory": ice_art,
              "peligro_art_directory": peligro_art,
              "mpp3_art_directory": mpp3_art,
              "script_sha256": h(Path(__file__)), "helper_sha256": h(helper),
              "boots": boots, "shots": shots, "pairs": pairs, "normal_DOS_exit": 0,
              "limits": "只驗MPP1正常冷讀、入洞、步行、HD切回及回城；未驗其他洞窟或新冰窟美術"}
    (O/"receipt.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n")
    print(report["status"], flush=True)
except Exception:
    if process is not None and process.poll() is None:
        invoke(["import", "-window", window, str(O/"failed-live.png")])
        (O/"failed-state.json").write_text(json.dumps(state(),ensure_ascii=False,indent=2)+"\n")
    raise
finally:
    for child in reversed(children):
        stop(child)
