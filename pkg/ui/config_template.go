package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (m *ConfigModel) saveTemplateFile(name string) error {
	filename := fmt.Sprintf("%s.cue", name)

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	filePath := filepath.Join(cwd, filename)

	// Marshal config to CUE format
	cueContent := m.marshalToCUE(name)

	// Write to current directory
	if err := os.WriteFile(filePath, []byte(cueContent), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

func (m *ConfigModel) marshalToCUE(profileName string) string {
	c := m.config
	var b strings.Builder

	b.WriteString("package config\n\n")

	// Include the schema definition matching embedded configs
	b.WriteString("#Config: {\n")
	b.WriteString("\tprofile: string\n\n")
	b.WriteString("\tdisk?: {\n")
	b.WriteString("\t\tdevice?:      string\n")
	b.WriteString("\t\tlegacy_boot?: bool\n")
	b.WriteString("\t\tpartitions?: {...}\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tsystem?: {\n")
	b.WriteString("\t\thostname?:     string\n")
	b.WriteString("\t\ttimezone?:     string\n")
	b.WriteString("\t\tlocale?:       string\n")
	b.WriteString("\t\tkeymap?:       string\n")
	b.WriteString("\t\tkernel?:       string\n")
	b.WriteString("\t\tdesktop?:      string\n")
	b.WriteString("\t\tgpu_driver?:   string\n")
	b.WriteString("\t\tservices?:     [...string]\n")
	b.WriteString("\t\tboot_options?: string\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tuser?: {\n")
	b.WriteString("\t\tusername?:      string\n")
	b.WriteString("\t\tpassword?:      string\n")
	b.WriteString("\t\troot_password?: string\n")
	b.WriteString("\t\thome_dir?:      string\n")
	b.WriteString("\t\tshell?:         string\n")
	b.WriteString("\t\tgroups?:        [...string]\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tpackages?: {\n")
	b.WriteString("\t\tbase?:  [...string]\n")
	b.WriteString("\t\textra?: [...string]\n")
	b.WriteString("\t\taur?:   [...string]\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tskeleton?: {\n")
	b.WriteString("\t\tdefault?: string\n")
	b.WriteString("\t\tprofile?: string\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\texternal_drives?: [...{\n")
	b.WriteString("\t\tdevice:      string\n")
	b.WriteString("\t\tmount_point: string\n")
	b.WriteString("\t\tlabel:       string\n")
	b.WriteString("\t\tfilesystem:  string\n")
	b.WriteString("\t}]\n\n")
	b.WriteString("\tpreserve_drives?: [...{\n")
	b.WriteString("\t\tdevice:      string\n")
	b.WriteString("\t\tmount_point: string\n")
	b.WriteString("\t}]\n")
	b.WriteString("}\n\n")

	b.WriteString("config: #Config & {\n")
	b.WriteString(fmt.Sprintf("\tprofile: %q\n\n", profileName))

	// Disk section
	b.WriteString("\tdisk: {\n")
	b.WriteString(fmt.Sprintf("\t\tdevice: %q\n", c.Disk.Device))
	b.WriteString(fmt.Sprintf("\t\tlegacy_boot: %t\n", c.Disk.LegacyBoot))
	b.WriteString("\t}\n\n")

	// System section
	b.WriteString("\tsystem: {\n")
	b.WriteString(fmt.Sprintf("\t\thostname: %q\n", c.System.Hostname))
	b.WriteString(fmt.Sprintf("\t\ttimezone: %q\n", c.System.Timezone))
	b.WriteString(fmt.Sprintf("\t\tlocale:   %q\n", c.System.Locale))
	b.WriteString("\t}\n\n")

	// User section
	b.WriteString("\tuser: {\n")
	b.WriteString(fmt.Sprintf("\t\tusername: %q\n", c.User.Username))
	b.WriteString(fmt.Sprintf("\t\tpassword: %q\n", c.User.Password))
	b.WriteString(fmt.Sprintf("\t\troot_password: %q\n", c.User.RootPassword))
	b.WriteString(fmt.Sprintf("\t\thome_dir: %q\n", c.User.HomeDir))
	b.WriteString("\t}\n\n")

	// Packages
	hasPackages := len(c.Packages.Base) > 0 || len(c.Packages.Extra) > 0 || len(c.Packages.AUR) > 0
	if hasPackages {
		b.WriteString("\tpackages: {\n")

		if len(c.Packages.Base) > 0 {
			b.WriteString("\t\tbase: [\n")
			for _, pkg := range c.Packages.Base {
				b.WriteString(fmt.Sprintf("\t\t\t%q,\n", pkg))
			}
			b.WriteString("\t\t]\n")
		}

		if len(c.Packages.Extra) > 0 {
			b.WriteString("\t\textra: [\n")
			for _, pkg := range c.Packages.Extra {
				b.WriteString(fmt.Sprintf("\t\t\t%q,\n", pkg))
			}
			b.WriteString("\t\t]\n")
		}

		if len(c.Packages.AUR) > 0 {
			b.WriteString("\t\taur: [\n")
			for _, pkg := range c.Packages.AUR {
				b.WriteString(fmt.Sprintf("\t\t\t%q,\n", pkg))
			}
			b.WriteString("\t\t]\n")
		}

		b.WriteString("\t}\n\n")
	}

	// Preserve drives
	if len(c.Preserve) > 0 {
		b.WriteString("\tpreserve: [\n")
		for _, d := range c.Preserve {
			b.WriteString("\t\t{\n")
			b.WriteString(fmt.Sprintf("\t\t\tdevice: %q\n", d.Device))
			b.WriteString(fmt.Sprintf("\t\t\tmount_point: %q\n", d.MountPoint))
			b.WriteString("\t\t},\n")
		}
		b.WriteString("\t]\n\n")
	}

	// External drives
	if len(c.External) > 0 {
		b.WriteString("\texternal: [\n")
		for _, d := range c.External {
			b.WriteString("\t\t{\n")
			b.WriteString(fmt.Sprintf("\t\t\tdevice: %q\n", d.Device))
			b.WriteString(fmt.Sprintf("\t\t\tmount_point: %q\n", d.MountPoint))
			b.WriteString(fmt.Sprintf("\t\t\tlabel: %q\n", d.Label))
			b.WriteString(fmt.Sprintf("\t\t\tfilesystem: %q\n", d.Filesystem))
			b.WriteString("\t\t},\n")
		}
		b.WriteString("\t]\n")
	}

	b.WriteString("}\n")

	return b.String()
}