package chroot

import (
	"fmt"
	"strings"

	"github.com/grantios/instagrant/pkg/phase"
)

// UserPhase creates the user account
type UserPhase struct {
	phase.BasePhase
}

func NewUserPhase() *UserPhase {
	return &UserPhase{
		BasePhase: phase.NewBasePhase(
			"user",
			"Create user account",
			true, // isChroot
		),
	}
}

func (u *UserPhase) Execute(ctx *phase.Context) error {
	cfg := ctx.Config.User
	h := ctx.Helper
	exec := ctx.Exec
	
	ctx.Logger.Stage("Creating user %s", cfg.Username)
	
	// Create user with all options in one command
	groups := strings.Join(cfg.Groups, ",")
	cmd := fmt.Sprintf("useradd -m -d %s -G %s -s %s %s", cfg.HomeDir, groups, cfg.Shell, cfg.Username)
	if err := exec.Chroot(ctx.TargetDir, cmd); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	// Set passwords
	ctx.Logger.Stage("Setting passwords")
	if err := h.SetPassword(cfg.Username, cfg.Password); err != nil {
		return err
	}
	if err := h.SetPassword("root", cfg.RootPassword); err != nil {
		return err
	}
	
	// Configure sudo
	ctx.Logger.Stage("Configuring sudo")
	if err := exec.Chroot(ctx.TargetDir, "sed -i 's/^# %wheel ALL=(ALL:ALL) ALL/%wheel ALL=(ALL:ALL) NOPASSWD: ALL/' /etc/sudoers"); err != nil {
		return fmt.Errorf("failed to configure sudo: %w", err)
	}
	
	// Setup /room directory
	ctx.Logger.Stage("Setting up /room directory")
	exec.Chroot(ctx.TargetDir, fmt.Sprintf("chown -R %s:%s /room && chmod 755 /room", cfg.Username, cfg.Username))
	
	// Setup XDG directories (simplified)
	ctx.Logger.Stage("Setting up XDG directories")
	xdgSetup := `
mkdir -p ~/.local/{cache,config,share,share/Trash,state,mount}
rm -rf ~/.config ~/.cache ~/.trash ~/.state ~/.mount
ln -sf ~/.local/config ~/.config
ln -sf ~/.local/cache ~/.cache
ln -sf ~/.local/share/Trash ~/.trash
ln -sf ~/.local/state ~/.state
ln -sf ~/.local/mount ~/.mount
ln -sf /mnt ~/.local/mount/mnt
ln -sf /run/media ~/.local/mount/auto
`
	
	// For root
	exec.Chroot(ctx.TargetDir, xdgSetup)
	
	// For user
	h.RunAsUser(cfg.Username, xdgSetup)
	
	// Install oh-my-zsh
	ctx.Logger.Stage("Installing oh-my-zsh")
	
	// Root oh-my-zsh installation
	rootOmzInstall := `sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended && mv ~/.oh-my-zsh ~/.ohmy && sed -i 's|export ZSH="$HOME/.oh-my-zsh"|export ZSH="$HOME/.ohmy"|' ~/.zshrc && sed -i 's/ZSH_THEME="robbyrussell"/ZSH_THEME="avit"/' ~/.zshrc`
	exec.Chroot(ctx.TargetDir, rootOmzInstall)
	exec.Chroot(ctx.TargetDir, "chsh -s /usr/bin/zsh")
	
	// User oh-my-zsh installation
	userOmzInstall := `sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended && mv ~/.oh-my-zsh ~/.ohmy && sed -i "s|export ZSH=\"\$HOME/.oh-my-zsh\"|export ZSH=\"\$HOME/.ohmy\"|" ~/.zshrc && sed -i "s/ZSH_THEME=\"robbyrussell\"/ZSH_THEME=\"candy\"/" ~/.zshrc`
	h.RunAsUser(cfg.Username, userOmzInstall)
	exec.Chroot(ctx.TargetDir, fmt.Sprintf("chsh -s /usr/bin/zsh %s", cfg.Username))
	
	// Add snapper rollback alias to override default 'snapper rollback' with snapper-rollback
	snapperAlias := `
# Override snapper rollback with snapper-rollback for proper btrfs rollback
snapper() {
    if [[ "$1" == "rollback" ]]; then
        shift
        sudo snapper-rollback "$@"
    else
        command snapper "$@"
    fi
}
`
	exec.Chroot(ctx.TargetDir, fmt.Sprintf("echo '%s' >> /root/.zshrc", snapperAlias))
	exec.Chroot(ctx.TargetDir, fmt.Sprintf("echo '%s' >> %s/.zshrc", snapperAlias, cfg.HomeDir))
	
	ctx.Logger.Success("User account created")
	ctx.Logger.Warn("IMPORTANT: Change passwords after first login!")
	
	return nil
}
