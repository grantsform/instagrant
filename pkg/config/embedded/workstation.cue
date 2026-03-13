// Workstation configuration
// Full desktop workstation with KDE Plasma

profile: "workstation"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "GATED"
	timezone:   "America/Chicago"
	locale:     "en_US.UTF-8"
	keymap:     "colemak"
	kernel:     "linux-lts"
	desktop:    "plasma"
	gpu_driver: "auto"
	services: [
		"sddm",
		"syncthing@${USERNAME}.service",
	]
}

user: {
	username:      "stein"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/room/stein"
	shell:         "/usr/bin/zsh"
}

packages: {
	extra: [
		// Workstation-specific
		"yakuake",
		"network-manager-applet",
		"plasma-nm",
		"proton-vpn-gtk-app",
		
		// Multimedia and creative tools
		"vlc",
		"streamlink",
		"yt-dlp",
		"mpv",
		"mpd",
		"steam",
		"godot",
		"blender",
		"krita",
		"kdenlive",
		"obs-studio",
		"obsidian",
		"libreoffice",
		"lmms",
		"audacity",
		"musescore",
		
		// Fonts and themes
		"papirus-icon-theme",
		"noto-fonts",
		"noto-fonts-emoji",
		
		// Audio
		"easyeffects",
		"pipewire",
		"pipewire-alsa",
		"pipewire-pulse",
		
		// Browsers and communication
		"chromium",
		"firefox",
		"thunderbird",
		"signal-desktop",
		"element-desktop",
		
		// Development tools
		"raylib",
		"sdl2",
		"sdl2_net",
		"sdl2_image",
		"sdl2_mixer",
		"sdl2_ttf",
		"qemu-full",
		"llvm",
		"lldb",
		"clang",
		"cmake",
		"meson",
		"ninja",
		"sbcl",
		"roswell",
		"racket",
		"fennel",
	]
	aur: [
		"papirus-folders-git",
		"bazaar",
		"visual-studio-code-bin",
		"chez-scheme",
		"chibi-scheme",
		"onlyoffice-bin",
		"planify",
		"notesnook-bin",
		"electronmail-bin",
		"steamlink",
		"vesktop-bin",
		"chatterino2-bin",
		"blockbench-bin",
		"sidequest-bin",
	]
}

skeleton: {
	default: "./skel/default"
	profile: "./skel/workstation"
}
