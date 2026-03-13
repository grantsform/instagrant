// Steam Deck configuration
// Gaming handheld with Steam and gaming tools

profile: "steamdeck"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:     "STEAMDECK"
	timezone:     "America/Chicago"
	locale:       "en_US.UTF-8"
	keymap:       "us"
	kernel:       "linux-lts"
	desktop:      "hyprland"
	gpu_driver:   "amd"
	boot_options: ""
	services: [
		"steamdeck",
		"bluetooth",
	]
}

user: {
	username:      "STEAMPUNK"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/room/steampunk"
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
		"bluez",
		"bluez-utils",
	]
	extra: [
		// Gaming and Steam
		"steam",
		"proton-ge-custom",
		"gamemode",
		"feral-gamemode",
		"mangohud",
		"vulkan-tools",
		"lutris",
		"wine",
		"wine-gecko",
		"wine-mono",

		// Desktop environment
		"hyprland",
		"waybar",
		"wofi",
		"dunst",
		"swaybg",
		"grim",
		"slurp",
		"wl-clipboard",
	]
	aur: [
		"steamdeck-dkms",
		"jupiter-hw-support",
	]
}

external_drives: [
	{
		device:     "/dev/sdb"
		mount_point: "/drv/games"
		label:      "games"
		filesystem: "xfs"
	}
]

skeleton: {
	default:  "steamdeck"
	profile:  "steamdeck"
}