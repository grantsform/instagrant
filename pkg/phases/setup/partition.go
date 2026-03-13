package setup

import (
	"fmt"
	"time"

	"github.com/grantios/instagrant/pkg/phase"
	"github.com/grantios/instagrant/pkg/util"
)

// PartitionPhase handles disk partitioning and formatting
type PartitionPhase struct {
	phase.BasePhase
}

func NewPartitionPhase() *PartitionPhase {
	return &PartitionPhase{
		BasePhase: phase.NewBasePhase(
			"partition",
			"Partition and format disks",
			false,
		),
	}
}

func (p *PartitionPhase) Validate(ctx *phase.Context) error {
	if !util.IsBlockDevice(ctx.Config.Disk.Device) {
		return fmt.Errorf("disk %s does not exist or is not a block device", ctx.Config.Disk.Device)
	}
	return nil
}

func (p *PartitionPhase) Execute(ctx *phase.Context) error {
	disk := ctx.Config.Disk.Device
	exec := ctx.Exec
	
	// Check if we're preserving the room partition
	preserveRoom := ctx.Config.Disk.PreserveRoom
	if preserveRoom {
		ctx.Logger.Info("Preserve room enabled, will skip room partition formatting")
	}
	
	ctx.Logger.Stage("Wiping disk %s", disk)
	
	// Deactivate swap, wipe and zero disk
	exec.Run("swapoff", "-a")
	if err := exec.Run("wipefs", "-af", disk); err != nil {
		return fmt.Errorf("failed to wipe disk: %w", err)
	}
	if err := exec.Run("dd", "if=/dev/zero", "of="+disk, "bs=1M", "count=100", "status=none"); err != nil {
		return fmt.Errorf("failed to zero disk: %w", err)
	}
	
	ctx.Logger.Stage("Creating partitions")
	
	if ctx.Config.Disk.LegacyBoot {
		return p.createMBRPartitions(ctx, disk)
	}
	return p.createGPTPartitions(ctx, disk, preserveRoom)
}

func (p *PartitionPhase) createGPTPartitions(ctx *phase.Context, disk string, preserveRoom bool) error {
	exec := ctx.Exec
	partPrefix := util.GetPartitionScheme(disk)
	parts := []string{partPrefix + "1", partPrefix + "2", partPrefix + "3", partPrefix + "4"}
	
	// Create GPT partition table with all partitions at once
	sgdiskCmd := fmt.Sprintf("sgdisk --zap-all "+
		"-n 1:0:+9G -t 1:ef00 -c 1:das "+
		"-n 2:0:+33G -t 2:8200 -c '2:Linux swap' "+
		"-n 3:0:+123G -t 3:8300 -c '3:Linux root' "+
		"-n 4:0:0 -t 4:8300 -c '4:Linux room' %s", disk)
	
	if err := exec.RunSh(sgdiskCmd); err != nil {
		return fmt.Errorf("failed to partition: %w", err)
	}
	
	exec.Run("partprobe", disk)
	time.Sleep(2 * time.Second)
	
	ctx.Logger.Stage("Formatting partitions")
	
	// Format partitions (skip room if preserveRoom is enabled)
	formatCmds := map[string][]string{
		"boot": {"mkfs.vfat", "-F", "32", "-n", "das", parts[0]},
		"swap": {"mkswap", "-L", "swap", parts[1]},
		"root": {"mkfs.btrfs", "-f", "-L", "root", parts[2]},
	}
	
	// Only format room partition if not preserving
	if !preserveRoom {
		formatCmds["room"] = []string{"mkfs.xfs", "-f", "-L", "room", parts[3]}
	} else {
		ctx.Logger.Info("Skipping room partition formatting (preserve mode)")
	}
	
	exec.Run("wipefs", "-af", parts[2]) // Clean root before formatting
	
	for name, cmd := range formatCmds {
		if err := exec.Run(cmd[0], cmd[1:]...); err != nil {
			// Special message for btrfs device busy errors
			if name == "root" && cmd[0] == "mkfs.btrfs" {
				ctx.Logger.Error("")
				ctx.Logger.Error("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				ctx.Logger.Error("This happens from time to time... The easiest thing to do here")
				ctx.Logger.Error("is to just reboot computer and try again. Like do a fresh-boot")
				ctx.Logger.Error("of the ArchLinux install medium. See ya in a sec!")
				ctx.Logger.Error("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				ctx.Logger.Error("")
			}
			return fmt.Errorf("failed to format partition: %w", err)
		}
	}
	
	// Handle external drives
	for _, ext := range ctx.Config.External {
		ctx.Logger.Stage("Formatting external drive %s", ext.Device)
		extPart := util.GetPartitionScheme(ext.Device) + "1"
		
		// Partition and format in one go
		if err := exec.RunSh(fmt.Sprintf("sgdisk --zap-all -n 1:0:0 -t 1:8300 %s && partprobe %s", ext.Device, ext.Device)); err != nil {
			return fmt.Errorf("failed to partition external drive: %w", err)
		}
		
		time.Sleep(1 * time.Second)
		
		// Format with appropriate filesystem
		mkfsMap := map[string]string{
			"xfs":   "mkfs.xfs -f -L %s %s",
			"btrfs": "mkfs.btrfs -f -L %s %s",
			"ext4":  "mkfs.ext4 -F -L %s %s",
			"ntfs":  "mkfs.ntfs -f -L %s %s",
			"exfat": "mkfs.exfat -L %s %s",
		}
		
		mkfsCmd, ok := mkfsMap[ext.Filesystem]
		if !ok {
			return fmt.Errorf("unsupported filesystem: %s", ext.Filesystem)
		}
		
		if err := exec.RunSh(fmt.Sprintf(mkfsCmd, ext.Label, extPart)); err != nil {
			return fmt.Errorf("failed to format external drive: %w", err)
		}
	}
	
	return nil
}

func (p *PartitionPhase) createMBRPartitions(ctx *phase.Context, disk string) error {
	return fmt.Errorf("legacy boot not yet implemented")
}
