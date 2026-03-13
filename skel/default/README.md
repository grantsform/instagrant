# Default Configuration Skeleton

This directory contains default skeleton files that are copied to **ALL** installations, regardless of the specific configuration chosen.

## Purpose

Files in this skeleton provide base configuration that applies to all system types:
- workstation
- homeserver  
- mediacenter
- smartclock
- Any custom configurations

## Directory Structure

- **@sys/** - Files copied to system root (/) - applied first
- **@root/** - Files copied to root user's home (/root/) - applied first  
- **@user/** - Files copied to configured user's home - applied first

## Override Behavior

Specific configuration skeletons (like `workstation/`) are copied **after** the default skeleton, allowing them to override any default files.

## Current Contents

- **@sys/tios/post.sh** - Post-installation setup script, available at `/tios/post.sh` on installed systems

## Usage

Add common configuration files here that should be available on all installations. Specific configurations can override these files by placing files with the same relative paths in their own skeleton directories.
