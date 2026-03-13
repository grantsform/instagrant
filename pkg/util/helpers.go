package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PhaseHelper provides convenient methods for common phase operations
type PhaseHelper struct {
	exec      *Executor
	targetDir string
}

// NewPhaseHelper creates a helper for common phase operations
func NewPhaseHelper(exec *Executor, targetDir string) *PhaseHelper {
	return &PhaseHelper{
		exec:      exec,
		targetDir: targetDir,
	}
}

// Pacman installs packages using pacman
func (h *PhaseHelper) Pacman(packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	cmd := fmt.Sprintf("pacman -S --noconfirm --needed %s", strings.Join(packages, " "))
	return h.exec.Chroot(h.targetDir, cmd)
}

// RunAsUser runs a command as a specific user in chroot
func (h *PhaseHelper) RunAsUser(username, command string) error {
	cmd := fmt.Sprintf("runuser -u %s -- bash -c '%s'", username, command)
	return h.exec.Chroot(h.targetDir, cmd)
}

// WriteFile writes content to a file in the target system
func (h *PhaseHelper) WriteFile(path, content string) error {
	fullPath := h.targetDir + path
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// MkdirAll creates a directory in the target system
func (h *PhaseHelper) MkdirAll(path string, perm os.FileMode) error {
	fullPath := h.targetDir + path
	return os.MkdirAll(fullPath, perm)
}

// Symlink creates a symlink in the target system
func (h *PhaseHelper) Symlink(oldname, newname string) error {
	fullNew := h.targetDir + newname
	if err := os.MkdirAll(filepath.Dir(fullNew), 0755); err != nil {
		return err
	}
	return os.Symlink(oldname, fullNew)
}

// SetPassword sets a user's password using chpasswd
func (h *PhaseHelper) SetPassword(username, password string) error {
	cmd := fmt.Sprintf("echo '%s:%s' | chpasswd", username, password)
	return h.exec.Chroot(h.targetDir, cmd)
}

// EnableService enables a systemd service
func (h *PhaseHelper) EnableService(service string) error {
	return h.exec.Chroot(h.targetDir, fmt.Sprintf("systemctl enable %s", service))
}

// EnableServices enables multiple systemd services
func (h *PhaseHelper) EnableServices(services ...string) error {
	for _, svc := range services {
		if err := h.EnableService(svc); err != nil {
			return err
		}
	}
	return nil
}

// InstallYay installs the yay AUR helper for a user
func (h *PhaseHelper) InstallYay(username, homeDir string) error {
	// Install dependencies
	if err := h.Pacman("git", "base-devel"); err != nil {
		return err
	}
	
	// Clean previous yay directory
	h.exec.Chroot(h.targetDir, fmt.Sprintf("rm -rf %s/yay", homeDir))
	
	// Clone and build yay
	buildCmd := fmt.Sprintf("cd %s && git clone https://aur.archlinux.org/yay.git && cd yay && makepkg -si --noconfirm", homeDir)
	if err := h.RunAsUser(username, buildCmd); err != nil {
		return err
	}
	
	// Cleanup
	return h.exec.Chroot(h.targetDir, fmt.Sprintf("rm -rf %s/yay", homeDir))
}

// InstallAURPackages installs AUR packages using yay
func (h *PhaseHelper) InstallAURPackages(username string, packages ...string) error {
	if len(packages) == 0 {
		return nil
	}
	cmd := fmt.Sprintf("yay -S --noconfirm --needed %s", strings.Join(packages, " "))
	return h.RunAsUser(username, cmd)
}
