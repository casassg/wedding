#!/usr/bin/env python3
"""
i18n Translation Key Validator for Hugo multilingual sites.
Checks for missing keys, unused keys, empty values, and inconsistencies across languages.

Usage:
    ./scripts/check-i18n.py              # Run from project root
    ./scripts/check-i18n.py /path/to/wedding  # Specify path
"""

import os
import re
import sys
from pathlib import Path
from typing import Set, Dict, Tuple

# Use current directory as base if run from project root
if len(sys.argv) > 1:
    BASE_DIR = Path(sys.argv[1])
else:
    BASE_DIR = Path(__file__).parent.parent

LAYOUTS_DIR = BASE_DIR / "layouts"
I18N_DIR = BASE_DIR / "i18n"


def parse_yaml(file_path: Path) -> Tuple[Set[str], Set[str]]:
    """Parse a YAML i18n file and extract all top-level keys and empty keys."""
    keys = set()
    empty_keys = set()
    if not file_path.exists():
        print(f"Warning: File not found: {file_path}", file=sys.stderr)
        return keys, empty_keys
    
    with open(file_path, "r", encoding="utf-8") as f:
        for line in f:
            line_stripped = line.strip()
            if not line_stripped or line_stripped.startswith("#"):
                continue
            match = re.match(r'^([a-zA-Z0-9_.-]+):\s*(.*)$', line_stripped)
            if match:
                key = match.group(1)
                value = match.group(2).strip()
                keys.add(key)
                # Check if value is empty or just quotes
                if not value or value == '""' or value == "''":
                    empty_keys.add(key)
    return keys, empty_keys


def extract_template_keys(walk_dir: Path) -> Dict[str, Set[str]]:
    """Extract all i18n keys used in Hugo templates."""
    keys_by_file = {}
    pattern = re.compile(r'(?:i18n|T)\s+"([^"]+)"')
    
    for file_path in walk_dir.rglob("*.html"):
        file_keys = set()
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
            matches = pattern.findall(content)
            file_keys.update(matches)
        if file_keys:
            keys_by_file[str(file_path.relative_to(BASE_DIR))] = file_keys
    
    all_keys = set()
    for keys in keys_by_file.values():
        all_keys.update(keys)
    
    return {"_all": all_keys, **keys_by_file}


def main():
    # Parse all i18n files
    en_keys, en_empty = parse_yaml(I18N_DIR / "en.yaml")
    es_keys, es_empty = parse_yaml(I18N_DIR / "es.yaml")
    ca_keys, ca_empty = parse_yaml(I18N_DIR / "ca.yaml")
    
    # Extract keys from templates
    template_data = extract_template_keys(LAYOUTS_DIR)
    template_keys = template_data["_all"]
    
    print("=" * 70)
    print("i18n Translation Key Validation Report")
    print("=" * 70)
    print()
    
    # Check for missing keys in each language
    missing_en = template_keys - en_keys
    missing_es = template_keys - es_keys
    missing_ca = template_keys - ca_keys
    
    has_issues = False
    
    if missing_en:
        has_issues = True
        print(f"❌ MISSING KEYS IN en.yaml ({len(missing_en)}):")
        for key in sorted(missing_en):
            print(f"   - {key}")
        print()
    
    if missing_es:
        has_issues = True
        print(f"❌ MISSING KEYS IN es.yaml ({len(missing_es)}):")
        for key in sorted(missing_es):
            print(f"   - {key}")
        print()
    
    if missing_ca:
        has_issues = True
        print(f"❌ MISSING KEYS IN ca.yaml ({len(missing_ca)}):")
        for key in sorted(missing_ca):
            print(f"   - {key}")
        print()
    
    # Check for empty values (warning only if used in templates)
    empty_and_used_en = en_empty & template_keys
    empty_and_used_es = es_empty & template_keys
    empty_and_used_ca = ca_empty & template_keys
    
    if empty_and_used_en:
        print(f"⚠️  EMPTY VALUES IN en.yaml ({len(empty_and_used_en)}):")
        print("   (These keys are used in templates but have empty values)")
        for key in sorted(empty_and_used_en):
            print(f"   - {key}")
        print()
    
    if empty_and_used_es:
        print(f"⚠️  EMPTY VALUES IN es.yaml ({len(empty_and_used_es)}):")
        print("   (These keys are used in templates but have empty values)")
        for key in sorted(empty_and_used_es):
            print(f"   - {key}")
        print()
    
    if empty_and_used_ca:
        print(f"⚠️  EMPTY VALUES IN ca.yaml ({len(empty_and_used_ca)}):")
        print("   (These keys are used in templates but have empty values)")
        for key in sorted(empty_and_used_ca):
            print(f"   - {key}")
        print()
    
    # Check for unused keys
    unused_keys = en_keys - template_keys
    if unused_keys:
        print(f"⚠️  POTENTIALLY UNUSED KEYS ({len(unused_keys)}):")
        print("   (These keys exist in i18n files but aren't used in templates)")
        for key in sorted(unused_keys):
            print(f"   - {key}")
        print()
    
    # Check for inconsistencies across languages
    all_keys_union = en_keys | es_keys | ca_keys
    inconsistent_keys = []
    
    for key in sorted(all_keys_union):
        missing_in = []
        if key not in en_keys:
            missing_in.append("en")
        if key not in es_keys:
            missing_in.append("es")
        if key not in ca_keys:
            missing_in.append("ca")
        
        if missing_in:
            inconsistent_keys.append((key, missing_in))
    
    if inconsistent_keys:
        has_issues = True
        print(f"❌ INCONSISTENCIES ACROSS LANGUAGES ({len(inconsistent_keys)}):")
        print("   (Keys that exist in some languages but not others)")
        for key, missing_in in inconsistent_keys:
            print(f"   - {key} missing in: {', '.join(missing_in)}")
        print()
    
    # Summary
    print("=" * 70)
    if not has_issues and not unused_keys:
        print("✅ All translations are consistent!")
    else:
        print(f"Summary:")
        print(f"  - Template keys used: {len(template_keys)}")
        print(f"  - Keys in en.yaml: {len(en_keys)}")
        print(f"  - Keys in es.yaml: {len(es_keys)}")
        print(f"  - Keys in ca.yaml: {len(ca_keys)}")
        if missing_en or missing_es or missing_ca:
            print(f"  - Missing translations: YES ❌")
        if empty_and_used_en or empty_and_used_es or empty_and_used_ca:
            print(f"  - Empty values in use: YES ⚠️")
        if unused_keys:
            print(f"  - Unused keys: {len(unused_keys)} ⚠️")
        if inconsistent_keys:
            print(f"  - Inconsistencies: {len(inconsistent_keys)} ❌")
    print("=" * 70)
    
    sys.exit(1 if has_issues else 0)


if __name__ == "__main__":
    main()
