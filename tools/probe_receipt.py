"""在 Docker 內執行固定預算的原版啟動探針並保存完整收據。"""

import argparse
import hashlib
import json
import os
import pathlib
import re
import subprocess
import time


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(args):
    segment = None
    if args.segment:
        match = re.fullmatch(r"([0-9A-Fa-f]{1,4}):([0-9A-Fa-f]{1,4}):([0-9]+)", args.segment)
        if not match or not 1 <= int(match[3]) <= 65536:
            raise ValueError("segment 格式須為 SEG:OFF:十進位長度，最多 65536 bytes")
        segment = (int(match[1], 16), int(match[2], 16), int(match[3]))
    if args.registers_at and not re.fullmatch(
        r"[0-9A-Fa-f]{1,4}:[0-9A-Fa-f]{1,4}(,[0-9A-Fa-f]{1,4}:[0-9A-Fa-f]{1,4})*",
        args.registers_at,
    ):
        raise ValueError("registers-at 須為逗號分隔的 SEG:OFF")
    output = args.output
    output.mkdir(parents=True, exist_ok=False)
    scratch = output / "scratch"
    scratch.mkdir()
    command = [
        str(args.probe), "-exe", str(args.original / "ZELIARD.EXE"),
        "-root", str(args.original), "-scratch", str(scratch),
        "-steps", str(args.steps), "-log-calls",
        "-dump-vram", str(output / "screen.bin"),
        "-dump-palette", str(output / "palette.bin"),
        "-trace", "40", "-trace-file", str(output / "trace.txt"),
    ]
    if args.hold:
        command.extend(["-hold", args.hold, "-hold-typematic=false"])
    if args.cycles:
        command.extend(["-cycles", str(args.cycles)])
    if segment:
        command.extend(["-dump-seg", args.segment + "=" + str(output / "memory")])
    if args.registers_at:
        command.extend(["-regs-at", args.registers_at])
    before = {p.name:sha256(p) for p in args.original.iterdir() if p.is_file()}
    started = time.monotonic()
    timed_out = False
    with (output / "stdout.txt").open("wb") as stdout, (output / "stderr.txt").open("wb") as stderr:
        try:
            result = subprocess.run(
                command, stdout=stdout, stderr=stderr,
                timeout=args.timeout, check=False,
                env={**os.environ, "GOTRACEBACK":"single"},
            )
            code = result.returncode
        except subprocess.TimeoutExpired:
            timed_out = True
            code = None
    after = {p.name:sha256(p) for p in args.original.iterdir() if p.is_file()}
    memory_complete = True
    if segment:
        memory = output / f"memory-{segment[0]:04X}_{segment[1]:04X}.bin"
        memory_complete = memory.is_file() and memory.stat().st_size == segment[2]
    receipt = {
        "kind":"original executable startup probe; not gameplay completion",
        "command":command, "steps_budget":args.steps,
        "wall_timeout_seconds":args.timeout,
        "elapsed_wall_seconds":round(time.monotonic() - started, 3),
        "exit_code":code, "timed_out":timed_out,
        "probe_sha256":sha256(args.probe),
        "source":json.loads(args.source_manifest.read_text()),
        "tool_image":"dosgolem-go-ebiten:1.24-r1",
        "build_go_version":"go1.24.13 linux/amd64",
        "originals_unchanged":before == after,
        "memory_dump_complete":memory_complete,
        "originals":before,
        "outputs":{
            p.name:{"bytes":p.stat().st_size, "sha256":sha256(p)}
            for p in output.iterdir() if p.is_file()
        },
    }
    (output / "receipt.json").write_text(
        json.dumps(receipt, ensure_ascii=False, indent=2) + "\n"
    )
    print(json.dumps({k:receipt[k] for k in (
        "exit_code", "timed_out", "elapsed_wall_seconds", "originals_unchanged"
    )}, ensure_ascii=False))
    for name in ("stdout.txt", "stderr.txt"):
        text = (output / name).read_text(errors="replace")
        lines = text.splitlines()
        print(name, "\n".join(lines[:24] + (["…完整內容保存在收據目錄…"] if len(lines) > 48 else []) + lines[-24:]))
    return 1 if timed_out or code != 0 or before != after or not memory_complete else 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--probe", type=pathlib.Path, default="/probe")
    parser.add_argument("--original", type=pathlib.Path, default="/orig")
    parser.add_argument("--output", type=pathlib.Path, default="/project/workplace/probe-startup-mcga")
    parser.add_argument("--source-manifest", type=pathlib.Path, default="/project/workplace/dosgolem-source-manifest.json")
    parser.add_argument("--steps", type=int, default=2_000_000)
    parser.add_argument("--timeout", type=int, default=45)
    parser.add_argument("--hold", default="", help="dosgolem 定時按下／放開序列，單位為指令數")
    parser.add_argument("--cycles", type=int, default=0, help="可選的 DOSBox 相容 cycles；0 保留舊基線時鐘")
    parser.add_argument("--segment", default="", help="可選：唯讀傾印 SEG:OFF:十進位長度，不修改記憶體")
    parser.add_argument("--registers-at", default="", help="可選：於 SEG:OFF 記錄暫存器，以逗號分隔")
    raise SystemExit(run(parser.parse_args()))
