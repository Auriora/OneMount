#!/bin/bash
# OneMount Development CLI Shell Completion Installer
# This script installs shell completion for the OneMount development CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Detect shell
detect_shell() {
    if [ -n "$ZSH_VERSION" ]; then
        echo "zsh"
    elif [ -n "$BASH_VERSION" ]; then
        echo "bash"
    elif [ -n "$FISH_VERSION" ]; then
        echo "fish"
    else
        # Fallback to checking $SHELL
        case "$SHELL" in
            */zsh) echo "zsh" ;;
            */bash) echo "bash" ;;
            */fish) echo "fish" ;;
            *) echo "unknown" ;;
        esac
    fi
}

# Install completion for bash
install_bash_completion() {
    local completion_dir
    
    # Try different completion directories
    if [ -d "$HOME/.local/share/bash-completion/completions" ]; then
        completion_dir="$HOME/.local/share/bash-completion/completions"
    elif [ -d "/usr/local/share/bash-completion/completions" ]; then
        completion_dir="/usr/local/share/bash-completion/completions"
        echo -e "${YELLOW}Warning: Installing to system directory. You may need sudo.${NC}"
    elif [ -d "/etc/bash_completion.d" ]; then
        completion_dir="/etc/bash_completion.d"
        echo -e "${YELLOW}Warning: Installing to system directory. You may need sudo.${NC}"
    else
        # Create user completion directory
        completion_dir="$HOME/.local/share/bash-completion/completions"
        mkdir -p "$completion_dir"
    fi
    
    echo -e "${BLUE}Installing bash completion to: $completion_dir${NC}"
    
    # Generate and install completion
    "$SCRIPT_DIR/dev" completion bash > "$completion_dir/dev"
    
    echo -e "${GREEN}✅ Bash completion installed!${NC}"
    echo -e "${YELLOW}Restart your shell or run: source ~/.bashrc${NC}"
}

# Install completion for zsh
install_zsh_completion() {
    local completion_dir
    
    # Try different completion directories
    if [ -d "$HOME/.local/share/zsh/site-functions" ]; then
        completion_dir="$HOME/.local/share/zsh/site-functions"
    elif [ -n "$ZDOTDIR" ] && [ -d "$ZDOTDIR/completions" ]; then
        completion_dir="$ZDOTDIR/completions"
    elif [ -d "$HOME/.zsh/completions" ]; then
        completion_dir="$HOME/.zsh/completions"
    else
        # Create user completion directory
        completion_dir="$HOME/.local/share/zsh/site-functions"
        mkdir -p "$completion_dir"
        
        # Add to fpath if not already there
        if ! grep -q "$completion_dir" "$HOME/.zshrc" 2>/dev/null; then
            echo "fpath=($completion_dir \$fpath)" >> "$HOME/.zshrc"
            echo -e "${YELLOW}Added $completion_dir to fpath in ~/.zshrc${NC}"
        fi
    fi
    
    echo -e "${BLUE}Installing zsh completion to: $completion_dir${NC}"
    
    # Generate and install completion
    "$SCRIPT_DIR/dev" completion zsh > "$completion_dir/_dev"
    
    echo -e "${GREEN}✅ Zsh completion installed!${NC}"
    echo -e "${YELLOW}Restart your shell or run: autoload -U compinit && compinit${NC}"
}

# Install completion for fish
install_fish_completion() {
    local completion_dir="$HOME/.config/fish/completions"
    
    # Create completion directory if it doesn't exist
    mkdir -p "$completion_dir"
    
    echo -e "${BLUE}Installing fish completion to: $completion_dir${NC}"
    
    # Generate and install completion
    "$SCRIPT_DIR/dev" completion fish > "$completion_dir/dev.fish"
    
    echo -e "${GREEN}✅ Fish completion installed!${NC}"
    echo -e "${YELLOW}Completion will be available in new fish sessions${NC}"
}

# Main installation function
install_completion() {
    local shell="$1"
    
    case "$shell" in
        bash)
            install_bash_completion
            ;;
        zsh)
            install_zsh_completion
            ;;
        fish)
            install_fish_completion
            ;;
        *)
            echo -e "${RED}Error: Unsupported shell: $shell${NC}"
            echo "Supported shells: bash, zsh, fish"
            exit 1
            ;;
    esac
}

# Show usage
show_usage() {
    echo "OneMount Development CLI Shell Completion Installer"
    echo ""
    echo "Usage: $0 [SHELL]"
    echo ""
    echo "SHELL can be: bash, zsh, fish"
    echo "If no shell is specified, auto-detection will be attempted."
    echo ""
    echo "Examples:"
    echo "  $0 bash    # Install bash completion"
    echo "  $0 zsh     # Install zsh completion"
    echo "  $0 fish    # Install fish completion"
    echo "  $0         # Auto-detect and install"
}

# Main script
main() {
    echo -e "${BLUE}OneMount Development CLI Shell Completion Installer${NC}"
    echo ""
    
    # Check if dev script exists
    if [ ! -f "$SCRIPT_DIR/dev" ]; then
        echo -e "${RED}Error: OneMount dev CLI not found at $SCRIPT_DIR/dev${NC}"
        echo "Make sure you're running this from the OneMount project directory."
        exit 1
    fi
    
    # Check if dev script is executable
    if [ ! -x "$SCRIPT_DIR/dev" ]; then
        echo -e "${YELLOW}Making dev script executable...${NC}"
        chmod +x "$SCRIPT_DIR/dev"
    fi
    
    local target_shell="$1"
    
    if [ -z "$target_shell" ]; then
        # Auto-detect shell
        target_shell=$(detect_shell)
        
        if [ "$target_shell" = "unknown" ]; then
            echo -e "${RED}Error: Could not detect your shell.${NC}"
            echo "Please specify the shell manually:"
            show_usage
            exit 1
        fi
        
        echo -e "${BLUE}Detected shell: $target_shell${NC}"
    fi
    
    # Install completion
    install_completion "$target_shell"
    
    echo ""
    echo -e "${GREEN}🎉 Shell completion installation complete!${NC}"
    echo ""
    echo "You can now use tab completion with the OneMount dev CLI:"
    echo "  ./scripts/dev <TAB>"
    echo "  ./scripts/dev build <TAB>"
    echo "  ./scripts/dev test <TAB>"
}

# Handle command line arguments
case "$1" in
    -h|--help)
        show_usage
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
