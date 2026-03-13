// Default configuration - base packages and settings
// Minimal base system with essential utilities

profile: "default"
target: "/target"

disk: {
	device:       "/dev/sda"
	legacy_boot:  false
	preserve_room: false
}

system: {
	hostname:   "archlinux"
	timezone:   "America/Chicago"
	locale:     "en_US.UTF-8"
	keymap:     "us"
	kernel:     "linux-lts"
	desktop:    "none"
	gpu_driver: "auto"
	services: [
		"sshd.service", "fstrim.timer", "NetworkManager",
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
	base: [
		   "linux",
		   "linux-lts",
		   "linux-firmware",
		   "base",
		   "base-devel",
		   "sudo",
		   "efibootmgr",
		   "btrfs-progs",
		   "xfsprogs",
		   "networkmanager",
		   "reflector",
		   "curl",
		   "rsync",
		   "git",
		   "zsh",
		   "just",
	   ]
	extra: [
		"tmux",
		"btop",
		"fastfetch",
		"glow",
		"gum",
		"fzf",
		"acpi",
		"restic",
		"rclone",
		"syncthing",
		"openssh",
		"tailscale",
		"git-lfs",
		"wget",
		"flatpak",
		"podman",
		"distrobox",
		"unzip",
		"exfat-utils",
		"ntfs-3g",
		"net-tools",
		"iftop",
		"github-cli",
	]
	aur: []
}

skeleton: {
	default: "./skel/default"
	profile: "./skel/default"
}
