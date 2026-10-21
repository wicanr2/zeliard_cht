#!/usr/bin/env python3
"""正式版沿用已驗影片，保留實際錄製版號，並核對遊戲來源沒有改動。Docker內執行。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil


def sha(path):
    digest = hashlib.sha256()
    with path.open('rb') as stream:
        for chunk in iter(lambda: stream.read(1 << 20), b''):
            digest.update(chunk)
    return digest.hexdigest()


def inventory(root, excluded=()):
    assert root.is_dir(), root
    return {str(f.relative_to(root)): sha(f) for f in sorted(root.rglob('*'))
            if f.is_file() and str(f.relative_to(root)) not in excluded}


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('capture_version')
    args = parser.parse_args()
    project = Path('/project')
    version = (project / 'VERSION').read_text().strip()
    prior = args.capture_version
    for value in (version, prior):
        if not re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', value):
            raise ValueError('版號格式不符')
    if version == prior:
        raise ValueError('來源與交付版本必須不同')
    work = lambda v: project / 'workplace' / ('release-v' + v[2:].split('-')[0].replace('.', ''))
    old, new = work(prior), work(version)
    for folder, excluded in (('sources', ('VERSION',)), ('dosgolem', ())):
        assert inventory(old / folder, excluded) == inventory(new / folder, excluded), folder
    before = json.loads((old / 'identity.json').read_text())
    after = json.loads((new / 'identity.json').read_text())
    for key in ('original_files', 'complete_art_sha256', 'asset_sources'):
        assert before[key] == after[key], key
    source = project / 'dist-all' / prior / 'promo'
    target = project / 'dist-all' / version / 'promo'
    assert target.is_dir() and not any(target.iterdir())
    source_manifest = json.loads((source.parent / 'SHA256SUMS.json').read_text())
    movies = []
    for file in sorted(source.glob('*.mp4')):
        expected = source_manifest['artifacts']['promo/' + file.name]
        assert sha(file) == expected['sha256'] and file.stat().st_size == expected['bytes']
        filename = file.name.replace(prior, version)
        shutil.copy2(file, target / filename)
        assert sha(target / filename) == expected['sha256']
        movies.append({'file': filename, 'sha256': expected['sha256'],
                       'capture_version': prior, 'source': str(file.relative_to(project))})
    assert len(movies) == 2
    for file in sorted(source.iterdir()):
        if file.is_file() and file.suffix != '.mp4':
            shutil.copy2(file, target / file.name)
    provenance = json.loads((source / 'provenance.json').read_text())
    provenance['version'] = version
    provenance['captured_program_version'] = prior
    provenance['source_provenance_sha256'] = sha(source / 'provenance.json')
    provenance['promoted_without_reencoding'] = True
    provenance['gameplay_clip']['file'] = provenance['gameplay_clip']['file'].replace(prior, version)
    write(target / 'provenance.json', provenance)
    for report, filename in (('quality.json', f'Zeliard-{version}-promo.mp4'),
                             ('gameplay-quality.json', f'Zeliard-{version}-ada-gameplay.mp4')):
        assert json.loads((target / report).read_text())['sha256'] == sha(target / filename)
    write(new / 'formal-promotion.json', {
        'status': 'PASS_UNCHANGED_RUNTIME_FORMAL_PROMOTION', 'version': version,
        'source_release': prior, 'runtime_sources_identical_except_VERSION': True,
        'dosgolem_identical': True, 'original_inputs_and_art_sources_identical': True,
        'movies': movies, 'media_quality_reports_reused_for_identical_bytes': True,
        'note': '正式版重新編譯及打包；影片保留原錄製身份，不宣稱重新錄製。'})
    print('PASS_UNCHANGED_RUNTIME_FORMAL_PROMOTION', version, flush=True)


if __name__ == '__main__':
    main()
