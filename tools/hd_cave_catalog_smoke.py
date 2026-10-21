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
mpp5_art = os.environ.get("ZELIARD_CAVE_MPP5_ART", "")
mpp6_art = os.environ.get("ZELIARD_CAVE_MPP6_ART", "")
mpp7_art = os.environ.get("ZELIARD_CAVE_MPP7_ART", "")
remaining_art = {f: os.environ.get("ZELIARD_CAVE_"+f+"_ART", "") for f in ("MPP1", "MPP8", "MPP9", "MPPA", "MPPB")}
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
            if mpp5_art:
                command.extend(["-cave-mpp5-art", mpp5_art])
            if mpp6_art:
                command.extend(["-cave-mpp6-art", mpp6_art])
            if mpp7_art:
                command.extend(["-cave-mpp7-art", mpp7_art])
            for family, directory in remaining_art.items():
                if directory:
                    command.extend(["-cave-"+family.lower()+"-art", directory])
            if common_art := os.environ.get("ZELIARD_CAVE_COMMON_ART", ""):
                command.extend(["-cave-common-art", common_art])
            if magia_art := os.environ.get("ZELIARD_CAVE_MAGIA_ART", ""):
                command.extend(["-cave-magia-art", magia_art])
            if door_art := os.environ.get("ZELIARD_CAVE_DOOR_ART", ""):
                command.extend(["-cave-door-art", door_art])
            if cyan_art := os.environ.get("ZELIARD_CAVE_CYAN_WEAPON_ART", ""):
                command.extend(["-cave-weapon-cyan-art", cyan_art])
            if player_extra := os.environ.get("ZELIARD_CAVE_PLAYER_EXTRA_ART", ""):
                command.extend(["-cave-player-extra-art", player_extra])
            if player_profiles := os.environ.get("ZELIARD_CAVE_PLAYER_PROFILE_ART", ""):
                command.extend(["-cave-player-profile-art", player_profiles])
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


def align_cave_exit(target):
    # 以正常入洞時的原版觀察值作輸入目標，不依較早圖面推算門框命中區。
    # 只送方向鍵，不寫原版座標、場景或記憶體。
    for _ in range(24):
        s = wait("正常洞窟來源", ready, 15)
        value = s["native_primary_0080"]+s["native_primary_0083"]
        # 即時觀察可能跨過單一座標；鄰近值只用來嘗試Up，是否離洞由原版轉場判定。
        # v1輸入紀錄顯示控制器反覆左右跨過目標，最後因持續等候位移而逾時。
        if abs(value-target) <= 2:
            exit_alignment.append({"target":target,"observed":value,"steps":s["steps"],"action":"try-original-up"})
            return
        exit_alignment.append({"target":target,"observed":value,"steps":s["steps"],"action":"left" if value>target else "right"})
        scan = 0x4b if value > target else 0x4d
        edge(scan,True)
        try:
            wait("正常走回入洞位置", lambda s: s["native_primary_0080"]+s["native_primary_0083"] != value, 5)
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
exit_alignment = []


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
    entry_state = wait("記錄正常入洞位置", ready, 30)
    exit_target = entry_state["native_primary_0080"]+entry_state["native_primary_0083"]
    compare("cave-entry")
    key(0x4b, 500000)
    advance(500000)
    compare("cave-walking")
    key(0x4d, 500000)
    advance(500000)
    align_cave_exit(exit_target)
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
              "mpp5_art_directory": mpp5_art,
              "mpp6_art_directory": mpp6_art,
              "mpp7_art_directory": mpp7_art,
              "remaining_family_art_directories": remaining_art,
              "exit_target_primary_sum": exit_target,
              "exit_alignment": exit_alignment,
              "script_sha256": h(Path(__file__)), "helper_sha256": h(helper),
              "boots": boots, "shots": shots, "pairs": pairs, "normal_DOS_exit": 0,
              "limits": "只驗MPP1正常冷讀、入洞、步行、HD切回及回城；未驗其他洞窟或新冰窟美術"}
    (O/"receipt.json").write_text(json.dumps(report, ensure_ascii=False, indent=2)+"\n")
    print(report["status"], flush=True)
except Exception:
    (O/"failed-observations.json").write_text(json.dumps({"shots":shots,"pairs":pairs,"exit_alignment":exit_alignment},ensure_ascii=False,indent=2)+"\n")
    if process is not None and process.poll() is None:
        invoke(["import", "-window", window, str(O/"failed-live.png")])
        (O/"failed-state.json").write_text(json.dumps(state(),ensure_ascii=False,indent=2)+"\n")
    raise
finally:
    for child in reversed(children):
        stop(child)
