# Instagrant

A modern, API-driven Arch Linux installer written in Go, inspired by archinstall but with a focus on configuration-driven installation and modularity.

## Features

- **Interactive TUI**: Easy-to-use terminal interface for configuring your installation
- **Config-Driven**: Define your entire installation in CUE configuration files
- **Btrfs + Snapper**: Built-in support for Btrfs subvolumes and automatic snapshots
- **Multiple Profiles**: Pre-configured profiles for workstation, server, media center, etc.

## Quick Start

### Prerequisites

- Booted into Arch Linux live environment
- Internet connection
- Target disk with sufficient space

### Download

```bash
# Download the latest binary
curl -Lo instagrant https://github.com/grantios/instagrant/releases/download/latest/instagrant && chmod +x instagrant

# Or install directly to /usr/local/bin
sudo install -m 755 <(curl -Lo- https://github.com/grantios/instagrant/releases/download/latest/instagrant) /usr/local/bin/instagrant
```

### Run

```bash
# Run instagrant (interactive TUI)
sudo ./instagrant
```

The interactive TUI will guide you through selecting a configuration profile and customizing your installation.

## Configuration

Instagrant uses CUE configuration files to define every aspect of the installation.

### Configuration Structure

```cue
profile: "workstation"      // Profile name
target:  "/target"          // Installation target directory

disk: {
    device:       "/dev/sda"   // Target disk
    legacy_boot:  false        // Use UEFI (true for BIOS)
    preserve_room: false       // Keep existing room partition
}

system: {
    hostname:   "archlinux"
    timezone:   "America/Chicago"
    locale:     "en_US.UTF-8"
    keymap:     "us"
    kernel:     "linux-lts"    // linux, linux-lts, linux-zen, linux-hardened
    desktop:    "plasma"       // none, plasma, hyprland, gnome
    gpu_driver: "auto"         // auto, nvidia, amd, intel, modesetting
    services: [
        "NetworkManager",
        "sshd.service",
    ]
}

user: {
    username:      "user"
    password:      "change!"
    root_password: "change!"
    home_dir:      "/home/user"
    shell:         "/usr/bin/zsh"
}

packages: {
    extra: [                   // Additional packages from official repos
        "firefox",
        "vim",
    ]
    aur: [                     // AUR packages (installed via yay)
        "visual-studio-code-bin",
    ]
}

skeleton: {
    default: "./skel/default"      // Default skeleton files
    profile: "./skel/workstation"  // Profile-specific skeleton (optional)
}
```

### Available Profiles

- **default** - Base system with essential utilities
- **workstation** - Full desktop with KDE Plasma (includes KDE applications)
- **ministation** - Minimal KDE Plasma desktop (same as workstation but without KDE applications)
- **homeserver** - Home server setup
- **mediacenter** - Media center with auto-login
- **steamdeck** - Steam Deck-style gaming setup
- **smartclock** - Minimal smart display
- **devestation** - Development workstation
- **nextclouder** - Nextcloud server
- **template** - Blank template for custom configs

## Installation Phases

Instagrant executes installation in discrete, resumable phases:

### Setup Phases (Pre-Chroot)
1. **partition** - Partition and format disks
2. **mount** - Mount partitions with Btrfs subvolumes
3. **pacstrap** - Install base system
4. **fstab** - Generate fstab

### Chroot Phases
5. **timezone** - Configure timezone and locale
6. **hostname** - Set hostname and hosts file
7. **user** - Create user account
8. **packages** - Install desktop and GPU drivers
9. **aur** - Install AUR packages via yay
10. **bootloader** - Install and configure systemd-boot
11. **services** - Enable system services
12. **skeleton** - Apply skeleton files
13. **snapper** - Configure Btrfs snapshots

### Building from Source

```bash
git clone https://github.com/grantios/instagrant
cd instagrant
go build -o instagrant
```

### Testing with Virtual Machines

Instagrant includes built-in VM commands for testing installations:

```bash
# Setup and boot Arch ISO for installation
sudo ./instagrant vm setup

# Boot from installed disk image
sudo ./instagrant vm boot

# Check disk image contents
./instagrant vm check
```

The VM commands automatically download the latest Arch Linux ISO, create a test disk image, and manage QEMU dependencies.

## Skeleton Files

Skeleton files are copied to the target system during installation. Organize them in:

- `skel/default/@sys/` - System files (copied to `/`)
- `skel/default/@user/` - User files (copied to `$HOME`)
- `skel/default/@root/` - Root files (copied to `/root`)

Profile-specific skeletons override default skeletons.