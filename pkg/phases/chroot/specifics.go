package chroot

import (
	"fmt"
	"path/filepath"

	"github.com/grantios/instagrant/pkg/phase"
	"github.com/grantios/instagrant/pkg/util"
)

// SpecificsPhase handles desktop-specific configurations
type SpecificsPhase struct {
	phase.BasePhase
}

func NewSpecificsPhase() *SpecificsPhase {
	return &SpecificsPhase{
		BasePhase: phase.NewBasePhase(
			"specifics",
			"Apply desktop-specific configurations",
			true, // isChroot
		),
	}
}

func (s *SpecificsPhase) Execute(ctx *phase.Context) error {
	exec := ctx.Exec

	// Handle Plasma desktop configuration
	if ctx.Config.System.Desktop == "plasma" {
		if err := s.configurePlasma(ctx, exec); err != nil {
			return fmt.Errorf("failed to configure Plasma: %w", err)
		}
	}

	ctx.Logger.Success("Desktop-specific configurations applied")
	return nil
}

func (s *SpecificsPhase) configurePlasma(ctx *phase.Context, exec *util.Executor) error {
	ctx.Logger.Stage("Configuring Plasma desktop environment")

	userHome := ctx.Config.User.HomeDir

	// Create directory structure
	grantiosDir := filepath.Join(userHome, ".grantios")
	kdeDir := filepath.Join(grantiosDir, "kde-plasma")
	pluginsDir := filepath.Join(kdeDir, "plugins")
	themingDir := filepath.Join(kdeDir, "theming")

	dirs := []string{grantiosDir, kdeDir, pluginsDir, themingDir}
	for _, dir := range dirs {
		if err := exec.Chroot(ctx.TargetDir, fmt.Sprintf("mkdir -p %s", dir)); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Clone theme repositories
	ctx.Logger.Info("Cloning theme repositories...")
	repos := []string{
		"https://github.com/grantios/gravity",
		"https://github.com/grantios/grantite-kde-theme",
		"https://github.com/grantios/grantite-gtk-theme",
		"https://github.com/grantios/grantite-cursors",
		"https://github.com/grantios/papirus-icon-theme",
		"https://github.com/grantios/papirus-folders",
		"https://github.com/grantios/ampersans-icon-theme",
	}

	for _, repo := range repos {
		repoName := filepath.Base(repo)

		cloneCmd := fmt.Sprintf("su - %s -c 'cd %s && git clone %s %s'", ctx.Config.User.Username, themingDir, repo, repoName)
		if err := exec.Chroot(ctx.TargetDir, cloneCmd); err != nil {
			ctx.Logger.Warn("Failed to clone %s: %v", repo, err)
			continue
		}

		ctx.Logger.Info("Successfully cloned %s", repoName)
	}

	// Copy wallpaper and logo files to user home
	ctx.Logger.Info("Setting up wallpapers and logos...")
	gravityDir := filepath.Join(themingDir, "gravity")
	
	// Copy .wall.webp and .logo.png to user home
	copyFiles := []string{".wall.webp", ".logo.png"}
	for _, file := range copyFiles {
		srcFile := filepath.Join(gravityDir, file)
		if err := exec.Chroot(ctx.TargetDir, fmt.Sprintf("cp %s %s/", srcFile, userHome)); err != nil {
			ctx.Logger.Warn("Failed to copy %s: %v", file, err)
		}
	}

	ctx.Logger.Success("Plasma desktop configured successfully")
	return nil
}