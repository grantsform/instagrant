package chroot

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/grantios/instagrant/pkg/phase"
)

// SkeletonPhase copies skeleton files
type SkeletonPhase struct {
	phase.BasePhase
}

func NewSkeletonPhase() *SkeletonPhase {
	return &SkeletonPhase{
		BasePhase: phase.NewBasePhase(
			"skeleton",
			"Apply skeleton files",
			false, // Runs on host to copy files
		),
	}
}

func (s *SkeletonPhase) Execute(ctx *phase.Context) error {
	skelCfg := ctx.Config.Skeleton
	skeletonApplied := false
	
	// Copy default skeleton
	if skelCfg.Default != "" {
		ctx.Logger.Stage("Applying default skeleton: %s", skelCfg.Default)
		if err := s.copySkeleton(ctx, skelCfg.Default); err != nil {
			return fmt.Errorf("failed to copy default skeleton: %w", err)
		}
		// Check if skeleton was actually applied (not just warned about missing)
		if _, err := os.Stat(skelCfg.Default); err == nil {
			skeletonApplied = true
		}
	}
	
	// Copy profile skeleton
	if skelCfg.Profile != "" {
		ctx.Logger.Stage("Applying profile skeleton: %s", skelCfg.Profile)
		if err := s.copySkeleton(ctx, skelCfg.Profile); err != nil {
			return fmt.Errorf("failed to copy profile skeleton: %w", err)
		}
		// Check if skeleton was actually applied
		if _, err := os.Stat(skelCfg.Profile); err == nil {
			skeletonApplied = true
		}
	}
	
	if skeletonApplied {
		ctx.Logger.Success("Skeleton files applied")
	} else {
		ctx.Logger.Info("No skeleton files applied (optional skeleton directory not found)")
	}
	return nil
}

func (s *SkeletonPhase) copySkeleton(ctx *phase.Context, skelPath string) error {
	if _, err := os.Stat(skelPath); os.IsNotExist(err) {
		ctx.Logger.Warn("Skeleton path does not exist: %s (continuing without skeleton files)", skelPath)
		return nil
	}
	
	// Process each skeleton subdirectory
	// @sys - system files to /
	sysPath := filepath.Join(skelPath, "@sys")
	if _, err := os.Stat(sysPath); err == nil {
		ctx.Logger.Info("Copying system skeleton files from %s to /", sysPath)
		if err := s.copySkeletonDir(ctx, sysPath, ctx.TargetDir); err != nil {
			return fmt.Errorf("failed to copy @sys skeleton: %w", err)
		}
	}
	
	// @root - root home directory files to /root
	rootPath := filepath.Join(skelPath, "@root")
	if _, err := os.Stat(rootPath); err == nil {
		ctx.Logger.Info("Copying root skeleton files from %s to /root", rootPath)
		rootTarget := filepath.Join(ctx.TargetDir, "root")
		if err := os.MkdirAll(rootTarget, 0755); err != nil {
			return fmt.Errorf("failed to create /root: %w", err)
		}
		if err := s.copySkeletonDir(ctx, rootPath, rootTarget); err != nil {
			return fmt.Errorf("failed to copy @root skeleton: %w", err)
		}
	}
	
	// @user - user home directory files
	userPath := filepath.Join(skelPath, "@user")
	if _, err := os.Stat(userPath); err == nil {
		ctx.Logger.Info("Copying user skeleton files from %s to %s", userPath, ctx.Config.User.HomeDir)
		userTarget := filepath.Join(ctx.TargetDir, ctx.Config.User.HomeDir)
		if err := os.MkdirAll(userTarget, 0755); err != nil {
			return fmt.Errorf("failed to create user home: %w", err)
		}
		if err := s.copySkeletonDir(ctx, userPath, userTarget); err != nil {
			return fmt.Errorf("failed to copy @user skeleton: %w", err)
		}
		// Set ownership
		chownCmd := fmt.Sprintf("chown -R %s:%s %s", ctx.Config.User.Username, ctx.Config.User.Username, ctx.Config.User.HomeDir)
		if err := ctx.Exec.Chroot(ctx.TargetDir, chownCmd); err != nil {
			ctx.Logger.Warn("Failed to set ownership on user skeleton files: %v", err)
		}
	}
	
	return nil
}

func (s *SkeletonPhase) copySkeletonDir(ctx *phase.Context, srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Get relative path from source directory
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		
		// Determine target path
		targetPath := filepath.Join(dstDir, relPath)
		
		// Create directory or copy file
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		
		return s.copyFile(path, targetPath, info.Mode())
	})
}

func (s *SkeletonPhase) copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	return err
}
