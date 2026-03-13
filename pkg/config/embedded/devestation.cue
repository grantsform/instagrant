// Devestation configuration
// Development workstation with multiple external drives

profile: "devestation"
target: "/target"

disk: {
	device:      "/dev/sdb"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "DEVESTATION"
	timezone:   "America/Chicago"
	locale:     "en_US.UTF-8"
	keymap:     "colemak"
	kernel:     "linux"
	desktop:    "hyprland"
	gpu_driver: "nvidia-open"
	services: [
		"hyprland",
		"docker",
		"libvirtd",
	]
}

user: {
	username:      "STEIN"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/room/stein"
	shell:         "/usr/bin/zsh"
}

packages: {
	base: [
		"linux",
		"linux-headers",
		"base",
		"base-devel",
		"git",
		"networkmanager",
	]
	extra: [
		// Desktop environment
		"hyprland",
		"kitty",
		"waybar",
		"wofi",
		"dunst",
		"swaybg",
		"grim",
		"slurp",
		"wl-clipboard",

		// Applications
		"chromium",
		"alacritty",
	]
	aur: [
		"visual-studio-code-bin",
	]
}

external_drives: []

skeleton: {
	default:  "devestation"
	profile:  "devestation"
}