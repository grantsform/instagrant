// Home Server configuration
// Minimal server setup with no desktop environment

profile: "homeserver"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "archserver"
	timezone:   "UTC"
	locale:     "en_US.UTF-8"
	keymap:     "us"
	kernel:     "linux-lts"
	desktop:    "none"
	gpu_driver: "modesetting"
	services: [
		"docker.service",
		"nginx.service",
		"fail2ban.service",
		"ufw.service",
	]
}

user: {
	username:      "admin"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/home/admin"
	shell:         "/usr/bin/bash"
}

packages: {
	extra: [
		"samba",
		"docker",
	]
	aur: []
}

skeleton: {
	default: "./skel/default"
	profile: "./skel/homeserver"
}
