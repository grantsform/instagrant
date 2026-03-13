package setup

import (
	"fmt"
	"os"

	"github.com/grantios/instagrant/pkg/phase"
	"github.com/grantios/instagrant/pkg/util"
)

// MountPhase handles mounting partitions
type MountPhase struct {
	phase.BasePhase
}

func NewMountPhase() *MountPhase {
	return &MountPhase{
		BasePhase: phase.NewBasePhase("mount", "Mount partitions", false),
	}
}

func (m *MountPhase) Execute(ctx *phase.Context) error {
	exec := ctx.Exec
	targetDir := ctx.TargetDir
	
	partPrefix := util.GetPartitionScheme(ctx.Config.Disk.Device)
	parts := map[string]string{
		"root": partPrefix + "3",
		"boot": partPrefix + "1",
		"swap": partPrefix + "2",
		"room": partPrefix + "4",
	}
	
	ctx.Logger.Stage("Creating btrfs subvolumes")
	
	// Ensure target directory exists with proper permissions
	ctx.Logger.Info("Ensuring target directory %s exists...", targetDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}
	
	// Mount root temporarily to create subvolumes
	os.MkdirAll(targetDir+"/root-tmp", 0755)
	exec.Run("mount", parts["root"], ctx.TargetDir+"/root-tmp")
	defer exec.Run("umount", ctx.TargetDir+"/root-tmp")
	
	// Create all subvolumes at once
	for _, sv := range []string{"@", "@home", "@var", "@snapshots"} {
		if err := exec.Run("btrfs", "subvolume", "create", ctx.TargetDir+"/root-tmp/"+sv); err != nil {
			return fmt.Errorf("failed to create subvolume %s: %w", sv, err)
		}
	}
	
	exec.Run("umount", ctx.TargetDir+"/root-tmp")
	
	ctx.Logger.Stage("Mounting subvolumes")
	
	// Mount root and create all mount points
	os.MkdirAll(targetDir, 0755)
	opts := "noatime,compress=zstd:1,space_cache=v2,discard=async"
	
	// Mount root subvolume first
	if err := exec.Run("mount", "-o", opts+",subvol=@", parts["root"], targetDir); err != nil {
		return fmt.Errorf("failed to mount root: %w", err)
	}
	
	// Create .btrfsroot mount point in the mounted filesystem
	os.MkdirAll(targetDir+"/.btrfsroot", 0755)
	
	// Mount the entire Btrfs filesystem to /.btrfsroot
	if err := exec.Run("mount", "-o", opts+",subvol=/", parts["root"], targetDir+"/.btrfsroot"); err != nil {
		return fmt.Errorf("failed to mount btrfs root: %w", err)
	}
	
	// Create all mount points
	for _, dir := range []string{"/boot", "/home", "/var", "/.snapshots", "/room"} {
		os.MkdirAll(targetDir+dir, 0755)
	}
	
	// Mount all subvolumes and partitions
	mounts := [][]string{
		{"-o", opts + ",subvol=@home", parts["root"], targetDir + "/home"},
		{"-o", opts + ",subvol=@var", parts["root"], targetDir + "/var"},
		{"-o", opts + ",subvol=@snapshots", parts["root"], targetDir + "/.snapshots"},
		{parts["boot"], targetDir + "/boot"},
		{parts["room"], targetDir + "/room"},
	}
	
	for _, m := range mounts {
		if err := exec.Run("mount", m...); err != nil {
			return fmt.Errorf("failed to mount: %w", err)
		}
	}
	
	// Activate swap
	exec.Run("swapon", parts["swap"])
	
	// Mount external and preserved drives
	for _, ext := range ctx.Config.External {
		extPart := util.GetPartitionScheme(ext.Device) + "1"
		mountPoint := targetDir + ext.MountPoint
		os.MkdirAll(mountPoint, 0755)
		ctx.Logger.Stage("Mounting external drive %s", ext.Device)
		exec.Run("mount", extPart, mountPoint)
	}
	
	for _, pres := range ctx.Config.Preserve {
		mountPoint := targetDir + pres.MountPoint
		os.MkdirAll(mountPoint, 0755)
		ctx.Logger.Stage("Mounting preserved drive %s", pres.Device)
		exec.Run("mount", pres.Device, mountPoint)
	}
	
	ctx.Logger.Success("All partitions mounted")
	return nil
}
