// Nextcloud server configuration
// Self-hosted cloud storage with security hardening

profile: "nextclouder"
target: "/target"

disk: {
	device:      "/dev/sda"
	legacy_boot: false
	preserve_room: false
}

system: {
	hostname:   "nextcloud"
	timezone:   "UTC"
	locale:     "en_US.UTF-8"
	keymap:     "us"
	kernel:     "linux-lts"
	desktop:    "none"
	gpu_driver: "modesetting"
	services: [
		"nginx.service",
		"php-fpm.service",
		"mariadb.service",
		"redis.service",
		"fail2ban.service",
		"ufw.service",
	]
}

user: {
	username:      "nextcloud"
	password:      "change!"
	root_password: "change!"
	home_dir:      "/home/nextcloud"
	shell:         "/usr/bin/bash"
}

packages: {
	extra: [
		// Web server
		"nginx",
		"nginx-mainline",
		
		// PHP and extensions
		"php",
		"php-fpm",
		"php-gd",
		"php-intl",
		"php-apcu",
		"php-imagick",
		
		// Database
		"mariadb",
		
		// Cache
		"redis",
		
		// Security
		"fail2ban",
		"ufw",
		"certbot",
		"certbot-nginx",
		
		// Utilities
		"unzip",
		"wget",
		"htop",
		"tmux",
	]
	aur: [
		"nextcloud",
	]
}

skeleton: {
	default: "./skel/default"
	profile: "./skel/server"
}

// Optional: Separate data drive for Nextcloud storage
// external_drives: [{
// 	device:      "/dev/sdb"
// 	mount_point: "/var/nextcloud/data"
// 	label:       "nextcloud-data"
// 	filesystem:  "xfs"
// }]
