"""
Path management utilities for OneMount development CLI.
"""

from pathlib import Path
from typing import Dict


def get_project_paths() -> Dict[str, Path]:
    """
    Get all important project paths.
    
    Returns:
        Dictionary mapping path names to Path objects
    """
    # Determine project root (parent of scripts directory)
    scripts_dir = Path(__file__).parent.parent
    project_root = scripts_dir.parent.resolve()
    
    paths = {
        # Core directories
        "project_root": project_root,
        "scripts_dir": scripts_dir,
        "build_dir": project_root / "build",
        "internal_dir": project_root / "internal",
        "cmd_dir": project_root / "cmd",
        "pkg_dir": project_root / "pkg",
        "tests_dir": project_root / "tests",
        "docs_dir": project_root / "docs",
        "assets_dir": project_root / "assets",
        "deployments_dir": project_root / "deployments",
        "packaging_dir": project_root / "packaging",
        
        # Build subdirectories
        "binaries_dir": project_root / "build" / "binaries",
        "packages_dir": project_root / "build" / "packages",
        "deb_dir": project_root / "build" / "packages" / "deb",
        "rpm_dir": project_root / "build" / "packages" / "rpm",
        "source_dir": project_root / "build" / "packages" / "source",
        "docker_dir": project_root / "build" / "docker",
        "temp_dir": project_root / "build" / "temp",
        
        # Important files
        "go_mod": project_root / "go.mod",
        "go_sum": project_root / "go.sum",
        "makefile": project_root / "Makefile",
        "bumpversion_cfg": project_root / ".bumpversion.cfg",
        "gitignore": project_root / ".gitignore",
        "aiignore": project_root / ".aiignore",
        
        # Coverage and testing
        "coverage_dir": project_root / "coverage",
        "coverage_file": project_root / "coverage" / "coverage.out",
        "coverage_html": project_root / "coverage" / "coverage.html",
        "coverage_json": project_root / "coverage" / "coverage.json",
        "coverage_history": project_root / "coverage" / "coverage_history.json",
        
        # Legacy script locations (for remaining shell scripts)
        "legacy_scripts": {
            "build_deb_native": scripts_dir / "build-deb-native.sh",
            "run_system_tests": scripts_dir / "run-system-tests.sh",
            "run_tests_docker": scripts_dir / "run-tests-docker.sh",
            "deploy_docker_remote": scripts_dir / "deploy-docker-remote.sh",
            "setup_personal_ci": scripts_dir / "setup-personal-ci.sh",
            "manifest_parser": scripts_dir / "manifest_parser.py",
        }
    }
    
    return paths


def ensure_build_directories():
    """Ensure all build directories exist."""
    paths = get_project_paths()
    
    build_dirs = [
        paths["build_dir"],
        paths["binaries_dir"],
        paths["packages_dir"],
        paths["deb_dir"],
        paths["rpm_dir"],
        paths["source_dir"],
        paths["docker_dir"],
        paths["temp_dir"],
    ]
    
    for directory in build_dirs:
        directory.mkdir(parents=True, exist_ok=True)


def ensure_coverage_directories():
    """Ensure coverage directories exist."""
    paths = get_project_paths()
    paths["coverage_dir"].mkdir(parents=True, exist_ok=True)


def get_binary_paths() -> Dict[str, Path]:
    """Get paths to built binaries."""
    paths = get_project_paths()
    binaries_dir = paths["binaries_dir"]
    
    return {
        "onemount": binaries_dir / "onemount",
        "onemount_headless": binaries_dir / "onemount-headless",
        "onemount_launcher": binaries_dir / "onemount-launcher",
    }


def get_package_paths() -> Dict[str, Path]:
    """Get paths where packages are built."""
    paths = get_project_paths()
    
    return {
        "deb_dir": paths["deb_dir"],
        "rpm_dir": paths["rpm_dir"],
        "source_dir": paths["source_dir"],
    }


def find_files_by_pattern(directory: Path, pattern: str) -> list[Path]:
    """Find files matching a pattern in a directory."""
    if not directory.exists():
        return []
    
    return list(directory.glob(pattern))


def get_cleanable_paths() -> Dict[str, list[Path]]:
    """Get paths that can be cleaned up, organized by category."""
    paths = get_project_paths()
    project_root = paths["project_root"]
    
    cleanable = {
        "build_artifacts": [
            paths["build_dir"],
        ],
        "coverage_files": [
            paths["coverage_dir"],
        ],
        "go_cache": [
            # Go build cache files
            *find_files_by_pattern(project_root, "*.test"),
            *find_files_by_pattern(project_root, "*.out"),
        ],
        "temp_files": [
            # Temporary files
            *find_files_by_pattern(project_root, "*.tmp"),
            *find_files_by_pattern(project_root, "*.temp"),
            *find_files_by_pattern(project_root, ".*.swp"),
            *find_files_by_pattern(project_root, ".*.swo"),
        ],
        "log_files": [
            *find_files_by_pattern(project_root, "*.log"),
            *find_files_by_pattern(project_root, "*.fa"),
        ],
        "package_files": [
            *find_files_by_pattern(project_root, "*.deb"),
            *find_files_by_pattern(project_root, "*.rpm"),
            *find_files_by_pattern(project_root, "*.dsc"),
            *find_files_by_pattern(project_root, "*.changes"),
            *find_files_by_pattern(project_root, "*.build*"),
            *find_files_by_pattern(project_root, "*.upload"),
            *find_files_by_pattern(project_root, "*.xz"),
            *find_files_by_pattern(project_root, "*.gz"),
        ],
        "database_files": [
            *find_files_by_pattern(project_root, "*.db"),
        ],
        "auth_files": [
            *find_files_by_pattern(project_root, ".auth_tokens.json"),
        ],
        "python_cache": [
            *find_files_by_pattern(paths["scripts_dir"], "__pycache__"),
            *find_files_by_pattern(paths["scripts_dir"], "*.pyc"),
        ],
    }
    
    return cleanable
