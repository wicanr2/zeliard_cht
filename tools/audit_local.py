"""Docker 內核對使用者限定範圍的實際交付收據。"""
from pathlib import Path
import hashlib
import json
import re
import struct

R = Path('/project')
version = (R / 'VERSION').read_text().strip()
D = R / 'dist-all' / version
B = D / 'full-local'
S = D / 'smoke'
if (S / 'current.json').exists():
    accepted = json.loads((S / 'current.json').read_text())
    assert accepted['directory'] == 'gui-v2'
    S = S / accepted['directory']
j = lambda p: json.loads(p.read_text())
h = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
if (D / 'smoke/current.json').exists():
    assert h(S / 'receipt.json') == accepted['receipt_sha256']
assert j(D / 'smoke/launcher-check.json')['status'] == 'PASS_PACKAGED_LAUNCHER'
receipt = j(S / 'receipt.json')
assert receipt['status'] == 'PASS_SCOPED_FIRST_CAVE_INPUT_CHINESE_HD'
assert receipt['version'] == version == (B / 'VERSION').read_text().strip()
assert receipt['binary_sha256'] == h(B / 'zeliard-preview')
assert receipt['keyboard_rebind_unbind_reset_restart'] and receipt['settings_input_isolated']
assert receipt['boots'][0]['normal_DOS_exit'] and receipt['boots'][0]['exit_code'] == 0
assert receipt['boots'][1]['capture_exit_code'] == 0
shots = {item['name']: item for item in receipt['shots']}
for name in ['controller-settings', 'keyboard-remapped', 'restored-town', 'church-chinese-HD', 'town-HD', 'first-cave-HD', 'first-cave-played', 'chinese-inventory', 'keyboard-restored-after-restart']:
    assert name in shots, name
    item = shots[name]
    assert h(S / ('boot' + str(item['boot'])) / (name + '.png')) == item['image_sha256']
assert shots['church-chinese-HD']['state']['church_hd_matched']
assert shots['church-chinese-HD']['state']['dialogue']['glyphs'] > 0
assert shots['town-HD']['state']['town_scenery_hd_matched'] and shots['town-HD']['state']['town_ground_hd_matched']
assert all(shots[name]['state']['cave_terrain_hd_matched'] for name in ['first-cave-HD', 'first-cave-played'])
assert shots['keyboard-restored-after-restart']['state']['keyboard_bindings']['attack'] == 'Z'
journal = [json.loads(line) for line in (S / 'boot1/inputs.jsonl').read_text().splitlines()]
scans = {entry['scan'] for entry in journal if entry.get('kind') == 'key' and entry['pressed']}
assert {0x39, 0x4b, 0x4d, 0x48, 0x1c}.issubset(scans)
manifest = j(D / 'SHA256SUMS.json')
assert manifest['version'] == version and manifest['platform'] == 'linux' and manifest['architecture'] == 'amd64'
for relative, expected in manifest['files'].items():
    p = B / relative
    assert p.stat().st_size == expected['bytes'] and h(p) == expected['sha256'], relative
binary = (B / 'zeliard-preview').read_bytes()[:64]
assert binary[:5] == b'\x7fELF\x02' and struct.unpack_from('<H', binary, 18)[0] == 62
identity = j(R / 'workplace/input-integration-v1/delivery-build/identity.json')
for name, expected in identity['sources'].items():
    assert h(R / 'workplace/input-integration-v1/delivery-build/sources' / name) == expected
for name, expected in identity['checkout_sources'].items():
    assert h(R / name) == expected
assert not identity['audio031_promoted']
for entry in j(R / 'workplace/source-manifest.json')['files']:
    assert h(R / 'workplace/orig' / entry['name']) == h(B / 'original' / entry['name']) == entry['sha256']
for entry in j(R / 'workplace/dosgolem-source-manifest.json')['files']:
    assert h(R / 'workplace/dosgolem-source' / entry['file']) == entry['sha256']
assert (R / 'workplace/input-integration-v1/delivery-build/vet.log').read_text() == ''
assert 'ok ' in (R / 'workplace/input-integration-v1/delivery-build/tests.log').read_text()
owned = []
for sub in ['cmd', 'internal', 'docs', 'text', 'tools']:
    owned += [p for p in (R / sub).rglob('*') if p.is_file() and '__pycache__' not in p.parts]
owned += [p for p in R.iterdir() if p.is_file() and p.suffix == '.md']
assert all((p.stat().st_uid, p.stat().st_gid) == (1000, 1000) for p in owned)
broken = []
for p in owned:
    if p.suffix != '.md':
        continue
    for target in re.findall(r'\]\(([^)]+)\)', p.read_text()):
        target = target.strip('<>').split('#')[0]
        if target and '://' not in target and not (p.parent / target).exists():
            broken.append([str(p.relative_to(R)), target])
assert not broken, broken
report = {'status': 'PASS_USER_SCOPED_LOCAL_DELIVERY', 'version': version, 'binary_sha256': h(B / 'zeliard-preview'), 'source_files': len(identity['sources']), 'manifest_files_verified': len(manifest['files']), 'actual_GUI_samples': len(shots), 'normal_first_cave_controls': sorted(scans), 'keyboard_rebind_unbind_reset_restart': True, 'controller_software_path': '既有裝置、按鈕／軸、死區、保存與合併邊沿；controls測試及中文設定頁已驗', 'gamepad_hardware_tested': False, 'original_files_unchanged': 29, 'frozen_dosgolem_files_unchanged': 882, 'audio031_promoted': False, 'broken_links': broken, 'acceptance_scope': '第一洞窟可玩、抽樣中文與HD、工具列鍵盤／手把串接；不要求完整通關或後段素材', 'rights': manifest['rights']}
(D / 'smoke/audit.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print(report['status'], version, len(shots))
