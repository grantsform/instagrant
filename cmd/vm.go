package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Virtual machine management for testing",
	Long:  `Manage QEMU virtual machines for testing Instagrant installations.`,
}

var vmSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup and boot Arch ISO for installation",
	Long:  `Download Arch Linux ISO, create disk image, and boot into live environment for installation.`,
	RunE:  runVMSetup,
}

var vmBootCmd = &cobra.Command{
	Use:   "boot",
	Short: "Boot from installed disk image",
	Long:  `Boot the installed system from the disk image without the live ISO.`,
	RunE:  runVMBoot,
}

var vmCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check disk image contents",
	Long:  `Inspect the disk image for partitions and basic information.`,
	RunE:  runVMCheck,
}

func init() {
	rootCmd.AddCommand(vmCmd)
	vmCmd.AddCommand(vmSetupCmd)
	vmCmd.AddCommand(vmBootCmd)
	vmCmd.AddCommand(vmCheckCmd)
}

func ensureQEMU() error {
	if _, err := exec.LookPath("qemu-system-x86_64"); err != nil {
		fmt.Println("Installing qemu-full...")
		cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", "qemu-full")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install qemu-full: %w", err)
		}
	}
	return nil
}

func ensureOVMF() error {
	if _, err := os.Stat("/usr/share/edk2-ovmf/x64/OVMF.4m.fd"); os.IsNotExist(err) {
		fmt.Println("Installing OVMF...")
		cmd := exec.Command("sudo", "pacman", "-S", "--noconfirm", "ovmf")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install ovmf: %w", err)
		}
	}
	return nil
}

func ensureArchLinux() error {
	// Check /etc/os-release for Arch Linux
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return fmt.Errorf("cannot detect operating system: %w", err)
	}

	content := string(data)
	if !strings.Contains(content, "Arch Linux") {
		return fmt.Errorf("VM commands are only supported on Arch Linux systems")
	}

	// Verify pacman is available
	if _, err := exec.LookPath("pacman"); err != nil {
		return fmt.Errorf("pacman not found - this appears to not be an Arch Linux system")
	}

	return nil
}

func getTestDir() string {
	return ".test"
}

func ensureTestDir() error {
	return os.MkdirAll(getTestDir(), 0755)
}

func downloadArchISO() error {
	isoPath := filepath.Join(getTestDir(), "archlinux.iso")
	if _, err := os.Stat(isoPath); err == nil {
		fmt.Printf("ISO already exists: %s\n", isoPath)
		return nil
	}

	fmt.Println("Downloading latest Arch Linux ISO...")
	// Get the latest ISO date
	cmd := exec.Command("curl", "-sL", "https://archlinux.org/releng/releases/json/")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get ISO info: %w", err)
	}

	// Extract version from JSON (simple approach)
	lines := strings.Split(string(output), "\n")
	var isoDate string
	for _, line := range lines {
		if strings.Contains(line, `"version"`) {
			parts := strings.Split(line, `"`)
			if len(parts) >= 4 {
				isoDate = strings.TrimSpace(parts[3])
				break
			}
		}
	}

	if isoDate == "" {
		isoDate = "2025.12.01" // fallback
	}

	fmt.Printf("Downloading archlinux-%s-x86_64.iso...\n", isoDate)
	cmd = exec.Command("curl", "-L", "--progress-bar", "-o", isoPath,
		fmt.Sprintf("https://geo.mirror.pkgbuild.com/iso/%s/archlinux-%s-x86_64.iso", isoDate, isoDate))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createDiskImage() error {
	diskPath := filepath.Join(getTestDir(), "disk.qcow2")
	if _, err := os.Stat(diskPath); err == nil {
		fmt.Printf("Disk image already exists: %s\n", diskPath)
		return nil
	}

	fmt.Println("Creating 250G test disk image...")
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", diskPath, "250G")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyInstagrantToTest() error {
	tiosDir := filepath.Join(getTestDir(), "tios")
	if err := os.MkdirAll(tiosDir, 0755); err != nil {
		return fmt.Errorf("failed to create tios directory: %w", err)
	}

	srcPath := "./instagrant"
	dstPath := filepath.Join(tiosDir, "instagrant")

	// Copy the binary
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination binary: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make it executable
	if err := os.Chmod(dstPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	fmt.Printf("Copied instagrant binary to %s\n", tiosDir)
	return nil
}

func runVMSetup(cmd *cobra.Command, args []string) error {
	if err := ensureArchLinux(); err != nil {
		return err
	}

	if err := copyInstagrantToTest(); err != nil {
		return fmt.Errorf("failed to copy instagrant binary: %w", err)
	}

	if err := ensureQEMU(); err != nil {
		return err
	}

	if err := ensureOVMF(); err != nil {
		return err
	}

	if err := downloadArchISO(); err != nil {
		return fmt.Errorf("failed to download ISO: %w", err)
	}

	if err := createDiskImage(); err != nil {
		return fmt.Errorf("failed to create disk image: %w", err)
	}

	fmt.Println("Starting QEMU with Arch Linux ISO...")
	fmt.Println("The .test/tios directory will be available as /dev/vdb (FAT filesystem)")
	fmt.Println("You can mount it with: mount /dev/vdb1 /mnt")
	fmt.Println("Then you can run: /mnt/instagrant")

	qemuCmd := exec.Command("qemu-system-x86_64",
		"-enable-kvm",
		"-m", "4G",
		"-smp", "4",
		"-cpu", "host",
		"-bios", "/usr/share/edk2-ovmf/x64/OVMF.4m.fd",
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", filepath.Join(getTestDir(), "disk.qcow2")),
		"-drive", fmt.Sprintf("file=fat:rw:%s,if=virtio,format=raw", filepath.Join(getTestDir(), "tios")),
		"-cdrom", filepath.Join(getTestDir(), "archlinux.iso"),
		"-boot", "d",
		"-nic", "user,model=virtio-net-pci",
		"-vga", "std", "-device", "VGA,edid=on,xres=840,yres=840",
		"-display", "gtk",
	)

	qemuCmd.Stdout = os.Stdout
	qemuCmd.Stderr = os.Stderr
	qemuCmd.Stdin = os.Stdin

	return qemuCmd.Run()
}

func runVMBoot(cmd *cobra.Command, args []string) error {
	if err := ensureArchLinux(); err != nil {
		return err
	}

	diskPath := filepath.Join(getTestDir(), "disk.qcow2")
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Errorf("disk image not found: %s\nRun 'instagrant vm setup' first to create and install", diskPath)
	}

	// Check disk size
	if stat, err := os.Stat(diskPath); err == nil {
		size := stat.Size()
		if size < 1000000000 { // Less than ~1GB
			fmt.Printf("Warning: Disk image is very small (%d bytes). Installation may not have completed.\n", size)
			fmt.Println("Run 'instagrant vm setup' to install the system first.")
		}
	}

	if err := ensureQEMU(); err != nil {
		return err
	}

	if err := ensureOVMF(); err != nil {
		return err
	}

	fmt.Println("Booting installed system from disk...")
	fmt.Println("If this fails to boot from disk, the installation may not have completed.")
	fmt.Println("Make sure you completed the full installation process with 'instagrant vm setup' first.")
	fmt.Println("The .test/tios directory will be available as /dev/vdb (FAT filesystem)")
	fmt.Println("You can mount it with: mount /dev/vdb /mnt")
	fmt.Println("Then you can run: /mnt/instagrant")
	fmt.Println()

	qemuCmd := exec.Command("qemu-system-x86_64",
		"-enable-kvm",
		"-m", "4G",
		"-smp", "4",
		"-cpu", "host",
		"-bios", "/usr/share/edk2-ovmf/x64/OVMF.4m.fd",
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio,index=0,media=disk", diskPath),
		"-drive", fmt.Sprintf("file=fat:rw:%s,if=virtio,format=raw", filepath.Join(getTestDir(), "tios")),
		"-boot", "strict=on,order=c,menu=on",
		"-net", "none",
		"-vga", "std", "-device", "VGA,edid=on,xres=1600,yres=1200",
		"-display", "gtk",
		"-no-reboot",
	)

	qemuCmd.Stdout = os.Stdout
	qemuCmd.Stderr = os.Stderr
	qemuCmd.Stdin = os.Stdin

	return qemuCmd.Run()
}

func runVMCheck(cmd *cobra.Command, args []string) error {
	diskPath := filepath.Join(getTestDir(), "disk.qcow2")
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Errorf("disk image not found: %s", diskPath)
	}

	fmt.Println("=== Disk Image Info ===")
	qemuImgCmd := exec.Command("qemu-img", "info", diskPath)
	qemuImgCmd.Stdout = os.Stdout
	qemuImgCmd.Stderr = os.Stderr
	if err := qemuImgCmd.Run(); err != nil {
		return fmt.Errorf("failed to get disk info: %w", err)
	}

	fmt.Println()
	fmt.Println("=== Partition Table (if any) ===")

	// Try to read partition table using guestfish
	if _, err := exec.LookPath("guestfish"); err == nil {
		guestfishCmd := exec.Command("guestfish", "-a", diskPath, "-i", "part-list", "/dev/sda")
		output, err := guestfishCmd.CombinedOutput()
		if err != nil {
			fmt.Println("No partitions found or guestfish error:", string(output))
		} else {
			fmt.Print(string(output))
		}
	} else {
		fmt.Println("Install libguestfs for partition inspection: sudo pacman -S libguestfs")
	}

	return nil
}