# template Configuration Skeleton

This directory contains configuration files that will be copied during installation.

## Directory Structure:
- **@sys/** - Files copied to system root (/)
- **@root/** - Files copied to root user's home (/root/)
- **@user/** - Files copied to configured user's home

## Usage:
Place configuration files in the appropriate directories above.
They will be automatically copied during the installation process.

Example:
- @sys/etc/systemd/system/my-service.service
- @user/.config/my-app/config.ini
- @root/.zshrc
