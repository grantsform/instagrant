// Minimal configuration template
// Copy and customize this for your own configs

profile: "custom"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
	// Optional: Customize partition sizes
	// partitions: {
	// 	boot: "9G"
	// 	swap: "33G"
	// 	root: "123G"
	// 	room: "remaining"
	// }
}

system: {
	hostname:   "archlinux"
	timezone:   "America/Chicago"
	locale:     "en_US.UTF-8"
	keymap:     "us"
	kernel:     "linux-lts"
	desktop:    "none"
	gpu_driver: "auto"
	services: []
	// boot_options: "quiet splash" // Additional kernel parameters
}

user: {
	username:      "user"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/home/user"
	shell:         "/usr/bin/zsh"
	// groups: [
	// 	"wheel",
	// 	"audio",
	// 	"video",
	// ]
}

packages: {
	extra: []
	aur:   []
}

skeleton: {
	default: "./skel/default"
}
