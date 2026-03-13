// Media Center configuration
// Optimized for Kodi/media playback with Hyprland

profile: "mediacenter"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "mediacenter"
	timezone:   "America/Chicago"
	locale:     "en_US.UTF-8"
	keymap:     "us"
	kernel:     "linux-lts"
	desktop:    "hyprland"
	gpu_driver: "auto"
	services: [
		"jellyfin.service",
		"plexmediaserver.service",
	]
}

user: {
	username:      "media"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/home/media"
	shell:         "/usr/bin/zsh"
}

packages: {
	extra: [
		"kodi",
		"retroarch",
		"libretro",
		"rtorrent",
		"blueman",
		"jellyfin-web",
		"jellyfin-server",
		"chromium",
		"pavucontrol",
		"easyeffects",
		"dolphin",
		"wireplumber",
		"brightnessctl",
		"mpv",
		"yt-dlp",
		"smbclient",
		"nfs-utils",
	]
	aur: [
		"steamlink",
	]
}

skeleton: {
	default: "./skel/default"
	profile: "./skel/mediacenter"
}

// Example: Large media drive
external_drives: [{
	device:      "/dev/sdb"
	mount_point: "/media/storage"
	label:       "media"
	filesystem:  "xfs"
}]
