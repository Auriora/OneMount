#!/usr/bin/env python3
"""
Standalone installation manifest script for Docker builds.
This is a simplified version that doesn't require external dependencies.
"""

import argparse
import json
import os
import sys
from pathlib import Path


def load_manifest():
    """Load the installation manifest."""
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    manifest_path = project_root / "packaging" / "install-manifest.json"
    
    if not manifest_path.exists():
        print(f"Error: Manifest file not found at {manifest_path}", file=sys.stderr)
        sys.exit(1)
    
    with open(manifest_path, 'r') as f:
        return json.load(f)


def get_all_files(manifest, install_type):
    """Get all files for the specified installation type."""
    files = []
    dest_key = f'dest_{install_type}'

    # Add binaries
    for binary in manifest.get('binaries', []):
        files.append({
            'type': 'binary',
            'source': binary['source'],
            'dest': binary[dest_key],
            'mode': binary['mode']
        })

    # Add icons
    for icon in manifest.get('icons', []):
        files.append({
            'type': 'icon',
            'source': icon['source'],
            'dest': icon[dest_key],
            'mode': icon['mode']
        })

    # Add desktop files
    for desktop in manifest.get('desktop', []):
        # For packages, use the pre-built desktop file
        if install_type == 'package' and 'source_package' in desktop:
            source = desktop['source_package']
        else:
            source = desktop['source_template']

        files.append({
            'type': 'desktop',
            'source': source,
            'dest': desktop[dest_key],
            'mode': desktop['mode']
        })

    # Add systemd files
    for systemd in manifest.get('systemd', []):
        # For packages, use the pre-built systemd file
        if install_type == 'package' and 'source_package' in systemd:
            source = systemd['source_package']
        else:
            source = systemd['source_template']

        files.append({
            'type': 'systemd',
            'source': source,
            'dest': systemd[dest_key],
            'mode': systemd['mode']
        })

    # Add documentation
    for doc in manifest.get('documentation', []):
        files.append({
            'type': 'documentation',
            'source': doc['source'],
            'dest': doc[dest_key],
            'mode': doc['mode'],
            'process': doc.get('process')
        })

    # Add nemo extensions
    for nemo_ext in manifest.get('nemo_extensions', []):
        files.append({
            'type': 'nemo_extension',
            'source': nemo_ext['source'],
            'dest': nemo_ext[dest_key],
            'mode': nemo_ext['mode']
        })

    return files


def generate_debian_install(manifest):
    """Generate Debian rules install commands."""
    commands = []

    files = get_all_files(manifest, 'package')
    for file_info in files:
        source = file_info['source']
        dest = file_info['dest']
        mode = file_info['mode']

        # Replace $(OUTPUT_DIR) with build directory
        source = source.replace('$(OUTPUT_DIR)', 'build')

        if file_info['type'] == 'documentation' and file_info.get('process') == 'gzip':
            # Documentation is already gzipped in build phase
            commands.append(f"install -D -m {mode} docs/man/onemount.1.gz $(pwd)/debian/onemount/{dest}")
        else:
            commands.append(f"install -D -m {mode} {source} $(pwd)/debian/onemount/{dest}")

    return commands


def generate_rpm_install(manifest):
    """Generate RPM spec install commands."""
    commands = []

    files = get_all_files(manifest, 'package')
    for file_info in files:
        source = file_info['source']
        dest = file_info['dest']
        mode = file_info['mode']

        # Replace $(OUTPUT_DIR) with build directory
        source = source.replace('$(OUTPUT_DIR)', 'build')

        if file_info['type'] == 'documentation' and file_info.get('process') == 'gzip':
            commands.append(f"install -D -m {mode} docs/man/onemount.1.gz $RPM_BUILD_ROOT/{dest}")
        else:
            commands.append(f"install -D -m {mode} {source} $RPM_BUILD_ROOT/{dest}")

    return commands


def generate_rpm_files(manifest):
    """Generate RPM spec files list."""
    files = []
    
    all_files = get_all_files(manifest, 'package')
    for file_info in all_files:
        dest = file_info['dest']
        files.append(dest)
    
    return files


def main():
    parser = argparse.ArgumentParser(description='Generate installation commands from manifest')
    parser.add_argument('--target', required=True, choices=['debian', 'rpm'], 
                       help='Target packaging system')
    parser.add_argument('--action', required=True, choices=['install', 'files'],
                       help='Action to perform')
    
    args = parser.parse_args()
    
    manifest = load_manifest()
    
    if args.target == 'debian' and args.action == 'install':
        commands = generate_debian_install(manifest)
    elif args.target == 'rpm' and args.action == 'install':
        commands = generate_rpm_install(manifest)
    elif args.target == 'rpm' and args.action == 'files':
        commands = generate_rpm_files(manifest)
    else:
        print(f"Error: Unsupported combination: target={args.target}, action={args.action}", file=sys.stderr)
        sys.exit(1)
    
    for command in commands:
        print(command)


if __name__ == '__main__':
    main()
