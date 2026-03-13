// Ministation configuration
// Minimal KDE Plasma workstation (based on workstation)

profile: "ministation"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "MINI"
	timezone:   "America/Chicago"
	locale:     "en_US.UTF-8"
	keymap:     "colemak"
	kernel:     "linux-lts"
	desktop:    "plasma-minimal"
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
		// Same extras as workstation (KDE full apps + tools)
		"yakuake",
		"network-manager-applet",
		"plasma-nm",
		"proton-vpn-gtk-app",
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
		"papirus-icon-theme",
		"noto-fonts",
		"noto-fonts-emoji",
		"easyeffects",
		"pipewire",
		"pipewire-alsa",
		"pipewire-pulse",
		"chromium",
		"firefox",
		"thunderbird",
		"signal-desktop",
		"element-desktop",
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
