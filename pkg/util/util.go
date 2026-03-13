package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandExists checks if a command is available in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsBlockDevice checks if a path is a block device
func IsBlockDevice(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	// Check if it's a device file
	mode := info.Mode()
	return (mode & os.ModeDevice) != 0
}

// GetDiskSize returns the size of a disk in bytes
func GetDiskSize(device string) (uint64, error) {
	cmd := exec.Command("blockdev", "--getsize64", device)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	var size uint64
	_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &size)
	return size, err
}

// GetPartitionScheme determines if a disk uses nvme/mmcblk naming
func GetPartitionScheme(device string) string {
	if strings.Contains(device, "nvme") || strings.Contains(device, "mmcblk") {
		return device + "p"
	}
	return device
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string, perm os.FileMode) error {
	if FileExists(path) {
		return nil
	}
	return os.MkdirAll(path, perm)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	cmd := exec.Command("cp", "-a", src, dst)
	return cmd.Run()
}

// ChownRecursive recursively changes ownership of a directory
func ChownRecursive(path, owner string) error {
	cmd := exec.Command("chown", "-R", owner, path)
	return cmd.Run()
}

// GetAvailableDisks returns a list of available disk devices
func GetAvailableDisks() ([]string, error) {
	cmd := exec.Command("lsblk", "-d", "-n", "-o", "NAME")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var disks []string
	for _, line := range lines {
		disk := strings.TrimSpace(line)
		if disk != "" {
			disks = append(disks, "/dev/"+disk)
		}
	}
	return disks, nil
}
