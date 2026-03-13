#!/usr/bin/env bash
set -euo pipefail

# Post-installation script for Arch Linux
# Run this after first boot to set up additional features

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Retry pacman commands up to 3 times
retry_pacman() {
    local attempts=3
    local cmd="$*"
    for ((i=1; i<=attempts; i++)); do
        log_info "Attempt $i/$attempts: $cmd"
        if $cmd; then
            return 0
        else
            log_error "Attempt $i failed: $cmd"
            if [[ $i -lt $attempts ]]; then
                log_warn "Retrying in 5 seconds..."
                sleep 5
            fi
        fi
    done
    log_error "All attempts failed: $cmd"
    return 1
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
    log_error "Don't run this script as root. Run as your regular user."
    exit 1
fi

log_info "Starting post-installation setup..."

# Update system
log_info "Updating system..."
retry_pacman sudo pacman -Syu --noconfirm

# Install useful packages
log_info "Installing additional packages..."
retry_pacman sudo pacman -S --noconfirm \
    firefox \
    code \
    discord \
    steam \
    lutris \
    wine \
    gamemode \
    mangohud \
    flatpak \
    timeshift \
    ufw \
    fail2ban \
    rkhunter \
    clamav

# Enable services
log_info "Enabling services..."
sudo systemctl enable ufw
sudo systemctl enable fail2ban
sudo ufw enable

# Setup flatpak
log_info "Setting up Flatpak..."
flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo

# Install common flatpaks
flatpak install -y flathub com.spotify.Client
flatpak install -y flathub org.libreoffice.LibreOffice
flatpak install -y flathub org.videolan.VLC

# Setup user directories
log_info "Setting up user directories..."
sudo pacman -S --noconfirm xdg-user-dirs
xdg-user-dirs-update

# Install oh-my-zsh
if [[ ! -d "$HOME/.oh-my-zsh" ]]; then
    log_info "Installing Oh My Zsh..."
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
fi

# Install yay packages (if yay is available)
if command -v yay &> /dev/null; then
    log_info "Installing AUR packages..."
    yay -S --noconfirm \
        spotify \
        visual-studio-code-bin \
        proton-ge-custom-bin \
        mangohud \
        goverlay
fi

# Setup git
log_info "Configuring Git..."
read -p "Enter your Git name: " git_name
read -p "Enter your Git email: " git_email
git config --global user.name "$git_name"
git config --global user.email "$git_email"

# Setup SSH key
if [[ ! -f "$HOME/.ssh/id_ed25519" ]]; then
    log_info "Generating SSH key..."
    ssh-keygen -t ed25519 -C "$git_email" -f "$HOME/.ssh/id_ed25519" -N ""
    log_info "Add this SSH key to your GitHub/GitLab account:"
    cat "$HOME/.ssh/id_ed25519.pub"
fi

# Setup firewall rules
log_info "Configuring firewall..."
sudo ufw allow ssh
sudo ufw allow http
sudo ufw allow https

# Setup timeshift for system snapshots
log_info "Setting up Timeshift..."
sudo timeshift --create --comments "Post-install snapshot"

log_success "Post-installation setup complete!"
echo ""
echo "Next steps:"
echo "1. Reboot to apply all changes"
echo "2. Login and enjoy your Arch Linux system!"
echo "3. Consider running 'sudo rkhunter --check' for security audit"
echo "4. Update clamav database: sudo freshclam"