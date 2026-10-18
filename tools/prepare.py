"""在 Docker 內盤點原始 ZIP，建立 dosgolem 隔離來源快照。"""

import argparse
import hashlib
import json
import pathlib
import stat


def digest(data):
    return hashlib.sha256(data).hexdigest()


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def prepare(project, source, source_head):
    work = project / "workplace"
    work.mkdir(exist_ok=True)
    originals = work / "orig"
    originals.mkdir(exist_ok=True)
    import zipfile

    archive = project / "神劍封魔.jsdos.zip"
    archive_data = archive.read_bytes()
    expected = "f7fc8bd3ea02e56ef01f170b9a00faa921634dc96a6f9bf6a6a736514e0bb6ac"
    if digest(archive_data) != expected:
        raise ValueError("來源 ZIP 雜湊與初始盤點不符，請先核對版本")
    files = []
    with zipfile.ZipFile(archive) as bundle:
        members = bundle.infolist()
        if sum(item.file_size for item in members) > 64 * 1024 * 1024:
            raise ValueError("解包總量超過本次來源的安全上限")
        seen = set()
        for item in members:
            relative = pathlib.PurePosixPath(item.filename)
            if (
                relative.is_absolute()
                or ".." in relative.parts
                or "\\" in item.filename
                or ":" in item.filename
                or not relative.parts
                or stat.S_ISLNK(item.external_attr >> 16)
            ):
                raise ValueError(f"不安全的來源路徑：{item.filename}")
            key = str(relative).casefold()
            if key in seen:
                raise ValueError(f"重複來源路徑：{item.filename}")
            seen.add(key)
            target = originals / relative
            if not target.resolve().is_relative_to(originals.resolve()):
                raise ValueError(f"來源路徑跳脫輸出目錄：{item.filename}")
            if item.is_dir():
                target.mkdir(parents=True, exist_ok=True)
                continue
            data = bundle.read(item)
            target.parent.mkdir(parents=True, exist_ok=True)
            if target.exists():
                if target.read_bytes() != data:
                    raise ValueError(f"原始解包副本已變動：{item.filename}")
            else:
                target.write_bytes(data)
            files.append(
                {"name":item.filename, "bytes":len(data), "sha256":digest(data)}
            )
    write_json(work / "source-manifest.json", {
        "archive":archive.name, "sha256":digest(archive_data), "files":files,
    })

    snapshot = work / "dosgolem-source"
    snapshot.mkdir(exist_ok=True)
    top_dirs = (
        "apps", "audio", "bootroot", "cmd", "frontend", "host", "internal",
        "oracle", "presentation", "runtime", "session", "xlate",
    )
    paths = [source / name for name in ("go.mod", "go.sum", "dosgolem.go")]
    for directory in top_dirs:
        paths.extend(p for p in (source / directory).rglob("*") if p.is_file())
    if sum(p.stat().st_size for p in paths) > 64 * 1024 * 1024:
        raise ValueError("來源快照超過上限，先核對新增檔案")
    records = []
    for path in sorted(paths):
        if path.is_symlink():
            raise ValueError(f"來源快照含符號連結：{path}")
        relative = path.relative_to(source)
        data = path.read_bytes()
        target = snapshot / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        if target.exists() and target.read_bytes() != data:
            raise ValueError(f"來源已變動，不覆寫既有快照：{relative}")
        if not target.exists():
            target.write_bytes(data)
        records.append({"file":str(relative), "sha256":digest(data)})
    expected_paths = {r["file"] for r in records}
    actual_paths = {str(p.relative_to(snapshot)) for p in snapshot.rglob("*") if p.is_file()}
    if actual_paths != expected_paths:
        raise ValueError("快照含有本次來源清單以外的檔案，請先核對既有快照")
    raw = json.dumps(records, ensure_ascii=False, sort_keys=True).encode()
    write_json(work / "dosgolem-source-manifest.json", {
        "source":"/home/anr2/cht/dosgolem", "head":source_head,
        "source_kind":"working-tree snapshot; HEAD alone is not a content identity",
        "files":records, "source_digest":digest(raw),
    })
    workspace = "go 1.24.0\n\nuse .\nuse ./workplace/dosgolem-source\n"
    workspace_path = project / "go.work"
    if workspace_path.exists() and workspace_path.read_text() != workspace:
        raise ValueError("現有 go.work 未指向已盤點快照，不自行覆寫")
    workspace_path.write_text(workspace)
    if not (project / "go.sum").exists():
        (project / "go.sum").write_bytes((snapshot / "go.sum").read_bytes())
    (work / "bin").mkdir(exist_ok=True)
    print(f"遊戲：{len(files)} 檔；dosgolem：{len(records)} 檔；快照 {digest(raw)}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--project", type=pathlib.Path, default="/project")
    parser.add_argument("--dosgolem", type=pathlib.Path, default="/source")
    parser.add_argument("--source-head", required=True)
    args = parser.parse_args()
    prepare(args.project, args.dosgolem, args.source_head)
