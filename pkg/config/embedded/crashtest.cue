profile: "crashtest"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "crashtest"
	timezone:   "UTC"
	locale:     "en_US.UTF-8"
	keymap:     "colemak"
	kernel:     "linux"
	desktop:    "none"
	gpu_driver: "auto"
	services: []
}

user: {
	username:      "test"
	password:      "test"
	root_password: "test"
	home_dir:      "/room/test"
	shell:         "/usr/bin/bash"
}

packages: {
	extra: []
	aur:   ["chibi-scheme"]
}

skeleton: {
	default: "./skel/default"
}