# OneMount Development CLI

A modern, unified command-line interface for OneMount development, build, and testing operations. Built with [Typer](https://typer.tiangolo.com/) and [Rich](https://rich.readthedocs.io/) for a beautiful CLI experience.

## Quick Start

```bash
# Install dependencies
pip install -r scripts/requirements-dev-cli.txt

# Make the CLI executable
chmod +x scripts/dev.py

# Show help
./scripts/dev --help

# Check environment
./scripts/dev info

# Install shell completion (optional)
./scripts/install-completion.sh
```

## Features

- 🎨 **Beautiful output** with Rich formatting, tables, and progress indicators
- 🔧 **Modular architecture** with organized command groups
- 🐍 **Native Python** implementations replacing shell script dependencies
- ✅ **Environment validation** with comprehensive checks
- 🧹 **Cleanup operations** for build artifacts and temporary files
- 📊 **Analysis tools** for code quality and project statistics
- 🚀 **Release management** with version bumping and Git integration
- 🔧 **Shell completion** for bash, zsh, and fish shells

## Available Commands

### 📋 Environment Information
```bash
./scripts/dev.py info                    # Show comprehensive environment info
```

### 🔨 Build Commands
```bash
./scripts/dev.py build deb --docker      # Build Debian package with Docker
./scripts/dev.py build deb --native      # Build Debian package natively
./scripts/dev.py build binaries          # Build OneMount binaries
./scripts/dev.py build manifest --target makefile --type user --action install
./scripts/dev.py build status            # Show build status
```

### 🧪 Test Commands
```bash
./scripts/dev.py test unit               # Run unit tests
./scripts/dev.py test integration        # Run integration tests
./scripts/dev.py test system --category comprehensive
./scripts/dev.py test coverage --threshold-line 80
./scripts/dev.py test docker all         # Run tests in Docker
./scripts/dev.py test status             # Show test status
```

### 🚀 Release Commands
```bash
./scripts/dev.py release bump patch      # Bump patch version
./scripts/dev.py release bump minor      # Bump minor version
./scripts/dev.py release status          # Show release status
./scripts/dev.py release history         # Show release history
./scripts/dev.py release check           # Check release configuration
```

### 🐙 GitHub Commands
```bash
./scripts/dev.py github status           # Show GitHub integration status
./scripts/dev.py github workflows        # Show workflow status
./scripts/dev.py github create-issues --file issues.json
```

### 🚢 Deploy Commands
```bash
./scripts/dev.py deploy docker-remote --host example.com
./scripts/dev.py deploy setup-ci         # Setup CI environment
./scripts/dev.py deploy status           # Show deployment status
./scripts/dev.py deploy test-connection --host example.com
```

### 📊 Analyze Commands
```bash
./scripts/dev.py analyze code-quality    # Run code quality checks
./scripts/dev.py analyze dependencies    # Analyze Go dependencies
./scripts/dev.py analyze project-stats   # Show project statistics
./scripts/dev.py analyze test-suite --mode analyze
```

### 🧹 Clean Commands
```bash
./scripts/dev.py clean list              # List cleanable artifacts
./scripts/dev.py clean build             # Clean build artifacts
./scripts/dev.py clean coverage          # Clean coverage files
./scripts/dev.py clean temp              # Clean temporary files
./scripts/dev.py clean all               # Clean everything
```

## Global Options

- `--verbose, -v`: Enable verbose output for debugging
- `--help, -h`: Show help for any command

## Shell Completion

The CLI supports shell completion for bash, zsh, and fish shells:

### Automatic Installation
```bash
# Auto-detect shell and install completion
./scripts/install-completion.sh

# Install for specific shell
./scripts/install-completion.sh bash
./scripts/install-completion.sh zsh
./scripts/install-completion.sh fish
```

### Manual Installation
```bash
# Generate completion script
./scripts/dev completion bash > ~/.local/share/bash-completion/completions/dev
./scripts/dev completion zsh > ~/.local/share/zsh/site-functions/_dev
./scripts/dev completion fish > ~/.config/fish/completions/dev.fish
```

After installation, you can use tab completion:
```bash
./scripts/dev <TAB>          # Show all commands
./scripts/dev build <TAB>    # Show build subcommands
./scripts/dev test c<TAB>    # Complete to 'coverage'
```

## Architecture

The CLI is built with a modular architecture:

```
scripts/
├── dev.py                    # Main CLI entry point
├── commands/                 # Command modules
│   ├── build_commands.py     # Build and packaging
│   ├── test_commands.py      # Testing and coverage
│   ├── release_commands.py   # Version and release management
│   ├── github_commands.py    # GitHub integration
│   ├── deploy_commands.py    # Deployment operations
│   ├── analyze_commands.py   # Code analysis
│   └── clean_commands.py     # Cleanup operations
├── utils/                    # Shared utilities
│   ├── environment.py        # Environment validation
│   ├── paths.py             # Path management
│   ├── shell.py             # Shell command execution
│   └── git.py               # Git operations
└── requirements-dev-cli.txt  # Dependencies
```

### Key Features

- **Rich Output**: Beautiful tables, progress indicators, and colored output
- **Error Handling**: Graceful error handling with helpful messages
- **Environment Validation**: Checks for required tools, paths, and configurations
- **Progress Indication**: Visual feedback for long-running operations
- **Modular Commands**: Organized into logical groups for better usability

## Migration from Legacy Scripts

The new CLI replaces the monolithic `onemount-dev.py` and provides better organization:

### Before (Old Way)
```bash
./scripts/onemount-dev.py build deb --docker
./scripts/onemount-dev.py test coverage --threshold-line 80
./scripts/onemount-dev.py release bump patch
```

### After (New Way)
```bash
./scripts/dev.py build deb --docker
./scripts/dev.py test coverage --threshold-line 80
./scripts/dev.py release bump patch
```

## Requirements

The CLI requires the following dependencies:

- `typer[all]>=0.9.0` - Modern CLI framework
- `rich>=13.0.0` - Rich text and beautiful formatting
- `gitpython>=3.1.0` - Git operations

Optional dependencies for enhanced functionality:
- `requests>=2.28.0` - GitHub API integration
- `matplotlib>=3.5.0` - Coverage trend plotting

## Development

### Adding New Commands

1. **Create a new command in the appropriate module**:
   ```python
   @app.command()
   def my_command(
       ctx: typer.Context,
       option: str = typer.Option(..., help="Description"),
   ):
       """Command description."""
       verbose = ctx.obj.get("verbose", False) if ctx.obj else False
       # Implementation here
   ```

2. **Use the utilities**:
   ```python
   from utils.environment import ensure_environment
   from utils.shell import run_command_with_progress
   from utils.paths import get_project_paths
   ```

3. **Follow the patterns**:
   - Use Rich for output formatting
   - Handle errors gracefully
   - Provide verbose mode support
   - Add progress indicators for long operations

### Testing the CLI

```bash
# Test basic functionality
./scripts/dev.py info

# Test with verbose output
./scripts/dev.py --verbose build status

# Test error handling
./scripts/dev.py build deb  # Should show help
```

## Contributing

When adding new functionality:

1. **Use the CLI framework** - Add commands to appropriate modules
2. **Follow the patterns** - Use Rich for output, handle errors gracefully
3. **Add help text** - Provide clear descriptions for commands and options
4. **Test thoroughly** - Ensure commands work in different environments
5. **Update documentation** - Add new commands to this README

## Troubleshooting

### Common Issues

1. **Import errors**: Make sure you're running from the project root
2. **Permission denied**: Run `chmod +x scripts/dev.py`
3. **Missing dependencies**: Run `pip install -r scripts/requirements-dev-cli.txt`
4. **Command not found**: Ensure you're in the OneMount project directory

### Getting Help

```bash
# Show all available commands
./scripts/dev.py --help

# Show help for a specific command
./scripts/dev.py build --help
./scripts/dev.py test coverage --help

# Check environment setup
./scripts/dev.py info
```
