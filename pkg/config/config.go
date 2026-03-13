package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

//go:embed embedded/*.cue
var embeddedConfigs embed.FS

// Config represents the complete installation configuration
type Config struct {
	Profile  string          `json:"profile"`
	Target   string          `json:"target"`
	Disk     DiskConfig      `json:"disk"`
	System   SystemConfig    `json:"system"`
	User     UserConfig      `json:"user"`
	Packages PackageConfig   `json:"packages"`
	Skeleton SkeletonConfig  `json:"skeleton"`
	External []ExternalDrive `json:"external_drives,omitempty"`
	Preserve []PreserveDrive `json:"preserve_drives,omitempty"`
}

// DiskConfig holds disk partitioning configuration
type DiskConfig struct {
	Device      string         `json:"device"`
	LegacyBoot  bool           `json:"legacy_boot,omitempty"`
	PreserveRoom bool          `json:"preserve_room,omitempty"`
	Partitions  PartitionSetup `json:"partitions,omitempty"`
}

// PartitionSetup defines partition sizes and layouts
type PartitionSetup struct {
	Boot string `json:"boot,omitempty"` // e.g., "9G"
	Swap string `json:"swap,omitempty"` // e.g., "33G"
	Root string `json:"root,omitempty"` // e.g., "123G"
	Room string `json:"room,omitempty"` // "remaining" or size
}

// SystemConfig holds system-level configuration
type SystemConfig struct {
	Hostname    string   `json:"hostname"`
	Timezone    string   `json:"timezone"`
	Locale      string   `json:"locale"`
	Keymap      string   `json:"keymap"`
	Kernel      string   `json:"kernel"`
	Desktop     string   `json:"desktop"`
	GPUDriver   string   `json:"gpu_driver,omitempty"`
	Services    []string `json:"services"`
	BootOptions string   `json:"boot_options,omitempty"`
}

// UserConfig holds user account configuration
type UserConfig struct {
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	HomeDir      string   `json:"home_dir"`
	Shell        string   `json:"shell"`
	Groups       []string `json:"groups,omitempty"`
	RootPassword string   `json:"root_password"`
}

// PackageConfig holds package installation configuration
type PackageConfig struct {
	Base  []string `json:"base"`
	Extra []string `json:"extra"`
	AUR   []string `json:"aur"`
}

// SkeletonConfig defines skeleton file sources
type SkeletonConfig struct {
	Default string `json:"default"`
	Profile string `json:"profile,omitempty"`
}

// ExternalDrive represents an external drive to be formatted and mounted
type ExternalDrive struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	Label      string `json:"label"`
	Filesystem string `json:"filesystem"` // xfs, btrfs, ext4
}

// PreserveDrive represents an existing drive to be mounted but not formatted
type PreserveDrive struct {
	Device     string `json:"device"` // Can be /dev/sdX, LABEL=xxx, or UUID=xxx
	MountPoint string `json:"mount_point"`
}

// Load reads and parses a CUE configuration file
// First checks for external config in current directory, then falls back to embedded configs
func Load(path string) (*Config, error) {
	filename := filepath.Base(path)
	isDefault := filename == "default.cue"
	
	// Load the requested config
	cfg, err := loadSingleConfig(path)
	if err != nil {
		return nil, err
	}
	
	// If not loading default, merge with default config as base layer
	if !isDefault {
		defaultCfg, err := loadSingleConfig("default.cue")
		if err == nil {
			// Merge: default provides base values, cfg overrides
			cfg = mergeConfigs(defaultCfg, cfg)
		}
		// If default doesn't exist, just use the config as-is
	}
	
	// Apply Go-level defaults (for values not constrained by CUE)
	cfg.applyDefaults()

	return cfg, nil
}

// loadSingleConfig loads a single config file without merging
func loadSingleConfig(path string) (*Config, error) {
	// Create CUE context
	ctx := cuecontext.New()

	var cueContent []byte

	// Check if external config exists in current directory
	externalPath := filepath.Base(path) // Just the filename
	if _, checkErr := os.Stat(externalPath); checkErr == nil {
		// Load from external file in current directory
		content, readErr := os.ReadFile(externalPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read external config: %w", readErr)
		}
		cueContent = content
	} else if _, checkErr := os.Stat(path); checkErr == nil {
		// Load from provided path
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read config: %w", readErr)
		}
		cueContent = content
	} else {
		// Fall back to embedded config
		embeddedPath := "embedded/" + filepath.Base(path)
		content, readErr := embeddedConfigs.ReadFile(embeddedPath)
		if readErr != nil {
			return nil, fmt.Errorf("config not found (tried external and embedded): %s", path)
		}
		cueContent = content
	}
	// Compile the CUE content directly with filename for better error messages
	filename := filepath.Base(path)
	value := ctx.CompileBytes(cueContent, cue.Filename(filename))
	if value.Err() != nil {
		return nil, fmt.Errorf("failed to compile CUE (%s): %w", filename, value.Err())
	}

	// Try to find the config field first
	configValue := value.LookupPath(cue.ParsePath("config"))
	
	// If config field doesn't exist, the CUE fields are already at top level
	// (this happens when CompileBytes unwraps the config: #Config & {...} structure)
	if !configValue.Exists() {
		configValue = value
	}
	
	if configValue.Err() != nil {
		return nil, fmt.Errorf("config field error: %w", configValue.Err())
	}

	// Validate CUE constraints
	if err := configValue.Validate(cue.Concrete(true)); err != nil {
		return nil, fmt.Errorf("CUE validation failed: %w", err)
	}

	// Decode into Config struct
	var cfg Config
	if err := configValue.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}

// mergeConfigs merges two configs with overlay logic
// base provides defaults, overlay overrides non-empty values
// packages.extra and packages.aur are appended (not replaced)
func mergeConfigs(base, overlay *Config) *Config {
	result := *overlay
	
	// Profile comes from overlay
	if result.Profile == "" {
		result.Profile = base.Profile
	}
	
	// Disk settings
	if result.Disk.Device == "" {
		result.Disk.Device = base.Disk.Device
	}
	if !overlay.Disk.LegacyBoot && base.Disk.LegacyBoot {
		result.Disk.LegacyBoot = base.Disk.LegacyBoot
	}
	
	// System settings
	if result.System.Hostname == "" {
		result.System.Hostname = base.System.Hostname
	}
	if result.System.Timezone == "" {
		result.System.Timezone = base.System.Timezone
	}
	if result.System.Locale == "" {
		result.System.Locale = base.System.Locale
	}
	if result.System.Keymap == "" {
		result.System.Keymap = base.System.Keymap
	}
	if result.System.Kernel == "" {
		result.System.Kernel = base.System.Kernel
	}
	if result.System.Desktop == "" {
		result.System.Desktop = base.System.Desktop
	}
	if result.System.GPUDriver == "" {
		result.System.GPUDriver = base.System.GPUDriver
	}
	if len(result.System.Services) == 0 {
		result.System.Services = base.System.Services
	}
	if result.System.BootOptions == "" {
		result.System.BootOptions = base.System.BootOptions
	}
	
	// User settings
	if result.User.Username == "" {
		result.User.Username = base.User.Username
	}
	if result.User.Password == "" {
		result.User.Password = base.User.Password
	}
	if result.User.RootPassword == "" {
		result.User.RootPassword = base.User.RootPassword
	}
	if result.User.HomeDir == "" {
		result.User.HomeDir = base.User.HomeDir
	}
	if result.User.Shell == "" {
		result.User.Shell = base.User.Shell
	}
	if len(result.User.Groups) == 0 {
		result.User.Groups = base.User.Groups
	}
	
	// Packages - base stays as-is, extra and aur get APPENDED from default
	if len(result.Packages.Base) == 0 {
		result.Packages.Base = base.Packages.Base
	}
	// Append default extra packages to overlay extra packages
	result.Packages.Extra = append(base.Packages.Extra, overlay.Packages.Extra...)
	// Append default aur packages to overlay aur packages
	result.Packages.AUR = append(base.Packages.AUR, overlay.Packages.AUR...)
	
	// Skeleton
	if result.Skeleton.Default == "" {
		result.Skeleton.Default = base.Skeleton.Default
	}
	if result.Skeleton.Profile == "" {
		result.Skeleton.Profile = base.Skeleton.Profile
	}
	
	// External and Preserve drives - overlay takes precedence (no merging)
	if len(result.External) == 0 {
		result.External = base.External
	}
	if len(result.Preserve) == 0 {
		result.Preserve = base.Preserve
	}
	
	return &result
}

// applyDefaults sets default values for unspecified fields
func (c *Config) applyDefaults() {
	if c.System.Timezone == "" {
		c.System.Timezone = "America/Chicago"
	}
	if c.System.Locale == "" {
		c.System.Locale = "en_US.UTF-8"
	}
	if c.System.Keymap == "" {
		c.System.Keymap = "us"
	}
	if c.System.Kernel == "" {
		c.System.Kernel = "linux-lts"
	}
	if c.System.Desktop == "" {
		c.System.Desktop = "none"
	}
	if c.System.GPUDriver == "" {
		c.System.GPUDriver = "auto"
	}
	if c.User.Shell == "" {
		c.User.Shell = "/usr/bin/zsh"
	}
	if len(c.User.Groups) == 0 {
		c.User.Groups = []string{"wheel", "audio", "video", "optical", "storage", "power"}
	}
	if c.Disk.Partitions.Boot == "" {
		c.Disk.Partitions.Boot = "9G"
	}
	if c.Disk.Partitions.Swap == "" {
		c.Disk.Partitions.Swap = "33G"
	}
	if c.Disk.Partitions.Root == "" {
		c.Disk.Partitions.Root = "123G"
	}
	if c.Disk.Partitions.Room == "" {
		c.Disk.Partitions.Room = "remaining"
	}

	// Default base packages
	if len(c.Packages.Base) == 0 {
		// Base packages are now defined in embedded/default.cue
		// This fallback is kept for safety but should not be used
		c.Packages.Base = []string{
			"linux",
			"linux-lts",
			"linux-firmware",
			"base",
		}
	}

	// Default services
	if len(c.System.Services) == 0 {
		c.System.Services = []string{
			"NetworkManager",
			"fstrim.timer",
			"reflector.timer",
		}
	}
}



// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Disk.Device == "" {
		return fmt.Errorf("disk device is required")
	}
	if c.System.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if c.User.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.User.Password == "" {
		return fmt.Errorf("user password is required")
	}
	if c.User.RootPassword == "" {
		return fmt.Errorf("root password is required")
	}

	// Validate disk exists
	if _, err := os.Stat(c.Disk.Device); os.IsNotExist(err) {
		return fmt.Errorf("disk device %s does not exist", c.Disk.Device)
	}

	// Validate desktop choice
	validDesktops := map[string]bool{
		"none": true, "plasma": true, "hyprland": true, "gnome": true,
	}
	if !validDesktops[c.System.Desktop] {
		return fmt.Errorf("invalid desktop: %s (valid: none, plasma, hyprland, gnome)", c.System.Desktop)
	}

	// Validate kernel
	validKernels := map[string]bool{
		"linux": true, "linux-lts": true, "linux-zen": true, "linux-hardened": true,
	}
	if !validKernels[c.System.Kernel] {
		return fmt.Errorf("invalid kernel: %s", c.System.Kernel)
	}

	return nil
}

// ListConfigs returns available configuration names from embedded and current directory
// Returns in order: template, default, separator, then other configs
func ListConfigs() []string {
	embeddedConfigsList := make(map[string]bool)
	localConfigs := make(map[string]bool)
	
	// List embedded configs
	entries, err := embeddedConfigs.ReadDir("embedded")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cue") {
				name := strings.TrimSuffix(entry.Name(), ".cue")
				embeddedConfigsList[name] = true
			}
		}
	}
	
	// List configs in current directory
	entries, err = os.ReadDir(".")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cue") {
				name := strings.TrimSuffix(entry.Name(), ".cue")
				// If it's also embedded, it stays in embedded (don't duplicate)
				if !embeddedConfigsList[name] {
					localConfigs[name] = true
				}
			}
		}
	}
	
	// Build result with specific order
	result := []string{}
	
	// Add template first if it exists in embedded
	if embeddedConfigsList["template"] {
		result = append(result, "template")
		delete(embeddedConfigsList, "template")
	}
	
	// Add default second if it exists in embedded
	if embeddedConfigsList["default"] {
		result = append(result, "default")
		delete(embeddedConfigsList, "default")
	}
	
	// Add separator after template/default
	if len(result) > 0 && embeddedConfigsList["crashtest"] {
		result = append(result, "---")
	}
	
	// Add crashtest after separator
	if embeddedConfigsList["crashtest"] {
		result = append(result, "crashtest")
		delete(embeddedConfigsList, "crashtest")
	}
	
	// Add separator after crashtest
	if len(result) > 0 && (len(embeddedConfigsList) > 0 || len(localConfigs) > 0) {
		result = append(result, "---")
	}
	
	// Add remaining embedded configs in specific order
	orderedConfigs := []string{"workstation", "devestation", "homeserver", "nextclouder", "mediacenter", "smartclock", "steamdeck"}
	for _, name := range orderedConfigs {
		if embeddedConfigsList[name] {
			result = append(result, name)
			delete(embeddedConfigsList, name)
		}
	}
	// Add any remaining embedded configs not in the ordered list (shouldn't happen)
	for name := range embeddedConfigsList {
		result = append(result, name)
	}
	
	// Add separator before local configs
	if len(localConfigs) > 0 {
		result = append(result, "---")
	}
	
	// Add local configs
	for name := range localConfigs {
		result = append(result, name)
	}
	
	return result
}

// IsLocalConfig returns true if the config exists as a local file and is not embedded
func IsLocalConfig(name string) bool {
	// Check if it's embedded first
	_, err := embeddedConfigs.ReadFile(fmt.Sprintf("embedded/%s.cue", name))
	if err == nil {
		return false // It's embedded, not local
	}
	
	// Check if it exists as a local file
	_, err = os.Stat(fmt.Sprintf("%s.cue", name))
	return err == nil
}

