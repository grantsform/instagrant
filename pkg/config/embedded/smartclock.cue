// Smart Clock configuration
// Hyprland + Godot development environment with rotated display

profile: "smartclock"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:     "LEGRUNE"
	timezone:     "America/Chicago"
	locale:       "en_US.UTF-8"
	keymap:       "colemak"
	kernel:       "linux-lts"
	desktop:      "hyprland"
	gpu_driver:   "auto"
	boot_options: "fbcon=rotate:2"
	services: [
		"hyprland",
		"pipewire",
		"wireplumber",
	]
}

user: {
	username:      "CLOCKED"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/room/eight"
	shell:         "/usr/bin/zsh"
}

packages: {
	base: [
		"linux-lts",
		"linux-lts-headers",
		"base",
		"base-devel",
		"git",
		"networkmanager",
	]
	extra: [
		// Development
		"godot",
		"neovim",
		"python",
		"nodejs",
		"npm",

		// Desktop environment
		"hyprland",
		"waybar",
		"wofi",
		"dunst",
		"swaybg",
		"grim",
		"slurp",
		"wl-clipboard",

		// Applications
		"dolphin",
		"chromium",
		"wireplumber",
		"brightnessctl",

	]
	aur: [
		"tty-clock",
		"hyprpaper",
	]
}

skeleton: {
	default:  "smartclock"
	profile:  "smartclock"
}