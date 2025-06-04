#!/usr/bin/env python3
"""
OneMount Installation Manifest Parser

This module parses the installation manifest and generates installation/uninstallation
commands for different packaging systems (Makefile, RPM, Debian, etc.).
"""

import os
import json
from pathlib import Path

# ANSI color codes for better output
class Colors:
    GREEN = '\033[92m'
    BLUE = '\033[94m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    BOLD = '\033[1m'
    END = '\033[0m'

def colored_echo(message, color=Colors.BLUE):
    """Generate echo command with color"""
    return f'echo -e "{color}{message}{Colors.END}"'

class InstallManifestParser:
    def __init__(self, manifest_path):
        with open(manifest_path, 'r') as f:
            self.manifest = json.load(f)

    def expand_variables(self, text):
        """Expand environment variables in text"""
        # Expand common variables
        text = text.replace('$(OUTPUT_DIR)', os.environ.get('OUTPUT_DIR', 'build'))
        text = text.replace('$(HOME)', os.environ.get('HOME', os.path.expanduser('~')))
        return text
    
    def get_all_files(self, install_type):
        """Get all files to be installed for a given install type (user/system/package)"""
        files = []
        
        # Add binaries
        for binary in self.manifest.get('binaries', []):
            files.append({
                'source': self.expand_variables(binary['source']),
                'dest': self.expand_variables(binary[f'dest_{install_type}']),
                'mode': binary['mode'],
                'type': 'binary'
            })
        
        # Add icons
        for icon in self.manifest.get('icons', []):
            files.append({
                'source': self.expand_variables(icon['source']),
                'dest': self.expand_variables(icon[f'dest_{install_type}']),
                'mode': icon['mode'],
                'type': 'icon'
            })
        
        # Add desktop files
        for desktop in self.manifest.get('desktop', []):
            if install_type == 'package':
                source = desktop['source_package']
            else:
                source = desktop['source_template']

            files.append({
                'source': self.expand_variables(source),
                'dest': self.expand_variables(desktop[f'dest_{install_type}']),
                'mode': desktop['mode'],
                'type': 'desktop',
                'template': install_type != 'package',
                'substitutions': desktop.get(f'substitutions_{install_type}', {}) if install_type != 'package' else {}
            })
        
        # Add systemd files
        for systemd in self.manifest.get('systemd', []):
            if install_type == 'package':
                source = systemd['source_package']
            else:
                source = systemd['source_template']

            files.append({
                'source': self.expand_variables(source),
                'dest': self.expand_variables(systemd[f'dest_{install_type}']),
                'mode': systemd['mode'],
                'type': 'systemd',
                'template': install_type != 'package',
                'substitutions': systemd.get(f'substitutions_{install_type}', {}) if install_type != 'package' else {}
            })

        # Add documentation
        for doc in self.manifest.get('documentation', []):
            files.append({
                'source': self.expand_variables(doc['source']),
                'dest': self.expand_variables(doc[f'dest_{install_type}']),
                'mode': doc['mode'],
                'type': 'documentation',
                'process': doc.get('process')
            })
        
        return files
    
    def generate_makefile_install(self, install_type, dry_run=False):
        """Generate Makefile install commands"""
        commands = []

        # Add header message
        install_desc = "system-wide" if install_type == "system" else "user"
        if dry_run:
            commands.append(f"{Colors.BOLD}{Colors.BLUE}Would install OneMount ({install_desc}):{Colors.END}")
        else:
            commands.append(colored_echo(f"Installing OneMount ({install_desc})...", Colors.BOLD + Colors.GREEN))

        # Show file list (only for dry run, otherwise integrate into installation steps)
        if dry_run:
            commands.append("")
            files = self.get_all_files(install_type)
            files_by_type = {}
            for file_info in files:
                file_type = file_info['type']
                if file_type not in files_by_type:
                    files_by_type[file_type] = []
                files_by_type[file_type].append(file_info)

            type_descriptions = {
                'binary': 'Binaries',
                'icon': 'Icons',
                'desktop': 'Desktop Files',
                'systemd': 'Systemd Service Files',
                'documentation': 'Documentation'
            }

            for file_type in ['binary', 'icon', 'desktop', 'systemd', 'documentation']:
                if file_type in files_by_type:
                    type_files = files_by_type[file_type]
                    commands.append(f"{Colors.YELLOW}  {type_descriptions[file_type]}:{Colors.END}")
                    for file_info in type_files:
                        source = file_info['source']
                        dest = file_info['dest']
                        commands.append(f"    {source} → {dest}")
                    commands.append("")

        if dry_run:
            return commands

        # Create directories
        directories = self.manifest['directories'][install_type]
        sudo_prefix = "sudo " if install_type == "system" else ""

        commands.append(colored_echo("Creating directories...", Colors.BLUE))
        for directory in directories:
            expanded_dir = self.expand_variables(directory)
            commands.append(colored_echo(f"{sudo_prefix}mkdir -p {expanded_dir}", Colors.YELLOW))
            commands.append(f"{sudo_prefix}mkdir -p {expanded_dir}")

        # Get files and group by type for installation
        files = self.get_all_files(install_type)
        files_by_type = {}
        for file_info in files:
            file_type = file_info['type']
            if file_type not in files_by_type:
                files_by_type[file_type] = []
            files_by_type[file_type].append(file_info)

        # Install each type with progress messages
        install_type_descriptions = {
            'binary': 'Installing binaries...',
            'icon': 'Installing icons...',
            'desktop': 'Installing desktop files...',
            'systemd': 'Installing systemd service files...',
            'documentation': 'Installing documentation...'
        }

        for file_type, type_files in files_by_type.items():
            if type_files:
                commands.append(colored_echo(install_type_descriptions.get(file_type, f"Installing {file_type} files..."), Colors.BLUE))

                for file_info in type_files:
                    source = file_info['source']
                    dest = file_info['dest']

                    if file_info['type'] == 'desktop' and file_info.get('template'):
                        # Handle template substitution
                        substitutions = file_info['substitutions']
                        sed_args = []
                        for key, value in substitutions.items():
                            expanded_value = self.expand_variables(value)
                            sed_args.append(f"-e 's|{key}|{expanded_value}|g'")
                        sed_cmd = " ".join(sed_args)
                        commands.append(colored_echo(f"{sudo_prefix}sed {sed_cmd} {source} → {dest}", Colors.YELLOW))
                        commands.append(f"{sudo_prefix}sed {sed_cmd} {source} > {dest}")

                    elif file_info['type'] == 'systemd' and file_info.get('template'):
                        # Handle template substitution
                        substitutions = file_info['substitutions']
                        sed_args = []
                        for key, value in substitutions.items():
                            expanded_value = self.expand_variables(value)
                            sed_args.append(f"-e 's|{key}|{expanded_value}|g'")
                        sed_cmd = " ".join(sed_args)
                        commands.append(colored_echo(f"{sudo_prefix}sed: {sed_cmd} {source} → {dest}", Colors.YELLOW))
                        commands.append(f"{sudo_prefix}sed {sed_cmd} {source} > {dest}")

                    elif file_info['type'] == 'documentation' and file_info.get('process') == 'gzip':
                        # Handle gzipped documentation
                        commands.append(colored_echo(f"{sudo_prefix}gzip -c {source} → {dest}", Colors.YELLOW))
                        commands.append(f"{sudo_prefix}gzip -c {source} > {dest}")

                    else:
                        # Regular file copy
                        commands.append(colored_echo(f"{sudo_prefix}cp {source} → {dest}", Colors.YELLOW))
                        commands.append(f"{sudo_prefix}cp {source} {dest}")
        
        # Post-install commands
        post_install = self.manifest['post_install'][install_type]
        if post_install:
            commands.append(colored_echo("Running post-install tasks...", Colors.BLUE))
            for cmd in post_install:
                commands.append(colored_echo(f"{cmd}", Colors.YELLOW))
                commands.append(f"{cmd}")

        # Completion message
        commands.append(colored_echo(f"OneMount ({install_desc}) installation completed successfully!", Colors.BOLD + Colors.GREEN))

        return commands

    def generate_makefile_uninstall(self, install_type, dry_run=False):
        """Generate Makefile uninstall commands"""
        commands = []

        # Add header message
        install_desc = "system-wide" if install_type == "system" else "user"
        if dry_run:
            commands.append(f"{Colors.BOLD}{Colors.BLUE}Would uninstall OneMount ({install_desc}):{Colors.END}")
        else:
            commands.append(colored_echo(f"Uninstalling OneMount ({install_desc})...", Colors.BOLD + Colors.YELLOW))

        # Show files that would be removed (only for dry run)
        if dry_run:
            commands.append("")
            files = self.get_all_files(install_type)
            commands.append(f"{Colors.YELLOW}  Files to be removed:{Colors.END}")
            for file_info in files:
                dest = file_info['dest']
                commands.append(f"    {dest}")
            commands.append("")
            return commands

        # Remove files
        sudo_prefix = "sudo " if install_type == "system" else ""
        files = self.get_all_files(install_type)

        file_paths = []
        icon_dirs = set()

        for file_info in files:
            dest = file_info['dest']
            file_paths.append(dest)

            # Track icon directories for removal
            if file_info['type'] == 'icon':
                icon_dirs.add(os.path.dirname(dest))

        # Remove individual files
        if file_paths:
            commands.append(colored_echo("Removing installed files...", Colors.BLUE))
            file_list = " ".join(file_paths)
            commands.append(colored_echo(f"{sudo_prefix}rm -f {file_list}", Colors.YELLOW))
            commands.append(f"{sudo_prefix}rm -f {file_list}")

        # Remove icon directories
        if icon_dirs:
            commands.append(colored_echo("Removing icon directories...", Colors.BLUE))
            for icon_dir in icon_dirs:
                commands.append(colored_echo(f"{sudo_prefix}rm -rf {icon_dir}", Colors.YELLOW))
                commands.append(f"{sudo_prefix}rm -rf {icon_dir}")

        # Post-uninstall commands
        post_uninstall = self.manifest['post_uninstall'][install_type]
        if post_uninstall:
            commands.append(colored_echo("Running post-uninstall tasks...", Colors.BLUE))
            for cmd in post_uninstall:
                commands.append(colored_echo(f"{cmd}", Colors.YELLOW))
                commands.append(f"{cmd}")

        # Completion message
        commands.append(colored_echo(f"OneMount ({install_desc}) uninstallation completed successfully!", Colors.BOLD + Colors.GREEN))

        return commands

    def generate_rpm_install(self):
        """Generate RPM spec install commands"""
        commands = []

        # Create directories
        directories = self.manifest['directories']['package']
        for directory in directories:
            commands.append(f"mkdir -p %{{buildroot}}/{directory}")

        # Install files
        files = self.get_all_files('package')
        for file_info in files:
            source = file_info['source']
            dest = file_info['dest']

            if file_info['type'] == 'documentation' and file_info.get('process') == 'gzip':
                # Documentation is already gzipped in build phase
                commands.append(f"cp docs/man/%{{name}}.1.gz %{{buildroot}}/{dest}")
            else:
                commands.append(f"cp {source} %{{buildroot}}/{dest}")

        return commands

    def generate_rpm_files(self):
        """Generate RPM spec files section"""
        commands = []
        commands.append("%defattr(-,root,root,-)")

        files = self.get_all_files('package')
        icon_dirs = set()

        for file_info in files:
            dest = file_info['dest']
            mode = file_info['mode']

            if file_info['type'] == 'icon':
                icon_dirs.add(os.path.dirname(dest))

            if file_info['type'] in ['binary']:
                commands.append(f"%attr({mode}, root, root) /{dest}")
            else:
                commands.append(f"%attr({mode}, root, root) /{dest}")

        # Add icon directories
        for icon_dir in sorted(icon_dirs):
            commands.append(f"%dir /{icon_dir}")

        return commands

    def generate_debian_install(self):
        """Generate Debian rules install commands"""
        commands = []

        files = self.get_all_files('package')
        for file_info in files:
            source = file_info['source']
            dest = file_info['dest']
            mode = file_info['mode']

            if file_info['type'] == 'documentation' and file_info.get('process') == 'gzip':
                # Documentation is already gzipped in build phase
                commands.append(f"install -D -m {mode} docs/man/onemount.1.gz $(pwd)/debian/onemount/{dest}")
            else:
                commands.append(f"install -D -m {mode} {source} $(pwd)/debian/onemount/{dest}")

        return commands

    def generate_validation(self):
        """Generate validation commands for required source files"""
        commands = []

        commands.append(colored_echo("Validating source files...", Colors.BLUE))

        # Check all source files exist
        all_sources = set()

        for binary in self.manifest.get('binaries', []):
            # Skip built binaries for validation
            pass

        for icon in self.manifest.get('icons', []):
            all_sources.add(icon['source'])

        for desktop in self.manifest.get('desktop', []):
            all_sources.add(desktop['source_template'])

        for systemd in self.manifest.get('systemd', []):
            all_sources.add(systemd['source_template'])

        for doc in self.manifest.get('documentation', []):
            all_sources.add(doc['source'])

        for source in sorted(all_sources):
            commands.append(f'test -f {source} || (echo -e "{Colors.RED}Error: {source} not found{Colors.END}" && exit 1)')

        commands.append(colored_echo("All source files validated successfully!", Colors.BOLD + Colors.GREEN))

        return commands
