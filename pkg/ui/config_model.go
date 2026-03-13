package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grantios/instagrant/pkg/config"
	"github.com/grantios/instagrant/pkg/util"
)

// ConfigModel handles configuration selection and editing
type ConfigModel struct {
	configs          []string
	selectedIndex    int
	config           *config.Config
	configName       string // Name of the selected config
	editing          bool
	editField        string
	textInput        textinput.Model
	confirmed        bool
	width            int
	height           int
	showPassword     bool   // Toggle for showing password in plain text
	errorMsg         string // Error message to display
	askingForName    bool   // Asking for template name
	newTemplateName  string // Name for new template
	showingDiskWarn  bool   // Showing disk overwrite warning
	showingLastChance bool  // Showing final confirmation
	diskSize         string // Formatted disk size for display
	showingExternalSelector bool // Showing external drive selector
	showingPreserveSelector bool // Showing preserve drive selector
	availableDisks   []string // List of available disk devices
	selectedDiskIndex int    // Selected disk in the list
	inputStep        int     // 0: select disk, 1: input name, 2: input mount, 3: select fs
	selectedFsIndex  int     // Selected filesystem
	currentDrive     config.ExternalDrive // Drive being configured
	currentPreserve  config.PreserveDrive // Drive being configured for preserve
	showingDisclaimer bool   // Showing disclaimer/welcome screen
	disclaimerAgreed  bool   // User has agreed to disclaimer
	showingRequirements bool // Showing requirements screen
}

type ConfigConfirmedMsg struct {
	Config *config.Config
}

func NewConfigModel(configs []string) *ConfigModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return &ConfigModel{
		configs:           configs,
		selectedIndex:     0,
		textInput:         ti,
		width:             80,
		height:            24,
		showingDisclaimer: true, // Show disclaimer first
	}
}

// showDiskWarning displays the disk overwrite confirmation screen
func (m *ConfigModel) validateConfig() error {
	// Check for devices that are in both preserve and external drives
	preserveDevices := make(map[string]bool)
	for _, p := range m.config.Preserve {
		preserveDevices[p.Device] = true
	}
	
	for _, e := range m.config.External {
		if preserveDevices[e.Device] {
			return fmt.Errorf("device %s cannot be in both preserve drives and external drives", e.Device)
		}
	}
	
	return nil
}

func (m *ConfigModel) showDiskWarning() {
	m.showingDiskWarn = true
	
	// Get disk size
	if size, err := util.GetDiskSize(m.config.Disk.Device); err == nil {
		// Format size nicely
		const unit = 1024 * 1024 * 1024 // GB
		if size >= unit {
			gb := float64(size) / float64(unit)
			if gb >= 1024 {
				m.diskSize = fmt.Sprintf("%.1f TB", gb/1024)
			} else {
				m.diskSize = fmt.Sprintf("%.1f GB", gb)
			}
		} else {
			m.diskSize = fmt.Sprintf("%d MB", size/(1024*1024))
		}
	} else {
		m.diskSize = "unknown size"
	}
}

func (m *ConfigModel) Init() tea.Cmd {
	return nil
}

func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle disclaimer screen first
		if m.showingDisclaimer {
			switch msg.String() {
			case "A":
				m.showingDisclaimer = false
				m.disclaimerAgreed = true
				m.showingRequirements = true // Show requirements next
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		
		// Handle requirements screen
		if m.showingRequirements {
			switch msg.String() {
			case "enter", " ":
				m.showingRequirements = false
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		
		if m.askingForName {
			switch msg.String() {
			case "enter":
				m.askingForName = false
				name := m.textInput.Value()
				if err := m.saveTemplateFile(name); err != nil {
					m.errorMsg = fmt.Sprintf("Failed to create template: %v", err)
				} else {
					m.errorMsg = fmt.Sprintf("✓ Template '%s.cue' created in current directory", name)
					// Reload config list to show new template
					newConfigs := config.ListConfigs()
					m.configs = newConfigs
				}
				return m, nil
			case "esc":
				m.askingForName = false
				return m, nil
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		} else if m.showingDiskWarn {
			switch msg.String() {
			case "y":
				// Validate configuration before proceeding
				if err := m.validateConfig(); err != nil {
					m.errorMsg = err.Error()
					m.showingDiskWarn = false
					return m, nil
				}
				// Show second confirmation
				m.showingDiskWarn = false
				m.showingLastChance = true
				return m, nil
			case "n", "esc":
				// User cancelled, go back to config editing
				m.showingDiskWarn = false
				return m, nil
			}
		} else if m.showingLastChance {
			switch msg.String() {
			case "Y":
				// User confirmed final warning, proceed with installation
				m.showingLastChance = false
				m.confirmed = true
				return m, tea.Quit
			case "N", "esc":
				// User cancelled, go back to config editing
				m.showingLastChance = false
				return m, nil
			}
		} else if m.showingPreserveSelector {
			switch m.inputStep {
			case 0: // Select disk
				switch msg.String() {
				case "up", "k":
					if m.selectedDiskIndex > 0 {
						m.selectedDiskIndex--
					}
				case "down", "j":
					if m.selectedDiskIndex < len(m.availableDisks)-1 {
						m.selectedDiskIndex++
					}
				case "enter":
					if len(m.availableDisks) > 0 {
						m.currentPreserve.Device = m.availableDisks[m.selectedDiskIndex]
						m.inputStep = 1
						m.textInput.SetValue("")
						m.textInput.Placeholder = "mount point (e.g., /mnt/data)"
					}
				case "esc":
					m.showingPreserveSelector = false
				}
			case 1: // Input mount point
				switch msg.String() {
				case "enter":
					m.currentPreserve.MountPoint = m.textInput.Value()
					m.config.Preserve = append(m.config.Preserve, m.currentPreserve)
					m.showingPreserveSelector = false
				case "esc":
					m.inputStep = 0
				default:
					m.textInput, cmd = m.textInput.Update(msg)
					return m, cmd
				}
			}
			return m, nil
		} else if m.showingExternalSelector {
			switch m.inputStep {
			case 0: // Select disk
				switch msg.String() {
				case "up", "k":
					if m.selectedDiskIndex > 0 {
						m.selectedDiskIndex--
					}
				case "down", "j":
					if m.selectedDiskIndex < len(m.availableDisks)-1 {
						m.selectedDiskIndex++
					}
				case "enter":
					if len(m.availableDisks) > 0 {
						m.currentDrive.Device = m.availableDisks[m.selectedDiskIndex]
						m.inputStep = 1
						m.textInput.SetValue("")
						m.textInput.Placeholder = "mount name (e.g., media)"
					}
				case "esc":
					m.showingExternalSelector = false
				}
			case 1: // Input name
				switch msg.String() {
				case "enter":
					m.currentDrive.Label = m.textInput.Value()
					m.inputStep = 2
					m.textInput.SetValue("")
					m.textInput.Placeholder = "mount point (e.g., /mnt/media)"
				case "esc":
					m.inputStep = 0
				default:
					m.textInput, cmd = m.textInput.Update(msg)
					return m, cmd
				}
			case 2: // Input mount point
				switch msg.String() {
				case "enter":
					m.currentDrive.MountPoint = m.textInput.Value()
					m.inputStep = 3
					m.selectedFsIndex = 0 // Default to first (xfs)
				case "esc":
					m.inputStep = 1
					m.textInput.SetValue(m.currentDrive.Label)
					m.textInput.Placeholder = "mount name (e.g., media)"
				default:
					m.textInput, cmd = m.textInput.Update(msg)
					return m, cmd
				}
			case 3: // Select filesystem
				filesystems := []string{"xfs", "ext4", "btrfs", "ntfs", "exfat"}
				switch msg.String() {
				case "up", "k":
					if m.selectedFsIndex > 0 {
						m.selectedFsIndex--
					}
				case "down", "j":
					if m.selectedFsIndex < len(filesystems)-1 {
						m.selectedFsIndex++
					}
				case "enter":
					m.currentDrive.Filesystem = filesystems[m.selectedFsIndex]
					m.config.External = append(m.config.External, m.currentDrive)
					m.showingExternalSelector = false
				case "esc":
					m.inputStep = 2
					m.textInput.SetValue(m.currentDrive.MountPoint)
					m.textInput.Placeholder = "mount point (e.g., /mnt/media)"
				}
			}
			return m, nil
		} else if m.editing {
			switch msg.String() {
			case "enter":
				m.applyEdit()
				m.editing = false
				return m, nil
			case "esc":
				m.editing = false
				return m, nil
			case "tab":
				if m.editField == "password" || m.editField == "rootpassword" {
					m.showPassword = !m.showPassword
					if m.showPassword {
						m.textInput.EchoMode = textinput.EchoNormal
					} else {
						m.textInput.EchoMode = textinput.EchoPassword
					}
				}
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.config != nil {
				// Go back to main config list
				m.config = nil
				m.configName = ""
				m.selectedIndex = 0
				m.errorMsg = ""
			}
			return m, nil
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			maxIndex := len(m.configs) + 17 // configs + fields + confirm
			if m.selectedIndex < maxIndex {
				m.selectedIndex++
			}
		case "enter":
			if m.config == nil {
				// Select config
				if m.selectedIndex < len(m.configs) {
					configName := m.configs[m.selectedIndex]
					if configName == "---" {
						return m, nil
					}
					m.loadConfig(configName)
				}
			} else {
				// Handle field selection
				fieldIndex := m.selectedIndex - len(m.configs)
				if fieldIndex >= 0 && fieldIndex <= 17 {
					if fieldIndex == 17 {
						// Confirm button
						if m.configName == "template" {
							m.askingForName = true
							m.textInput.SetValue("")
							m.textInput.Placeholder = "template name"
							return m, nil
						} else if m.configName == "default" {
							if err := m.saveTemplateFile("default"); err != nil {
								m.errorMsg = fmt.Sprintf("Failed to save default config: %v", err)
							} else {
								m.errorMsg = "Saved as default.cue in current directory"
								// Go back to main config list
								m.config = nil
								m.configName = ""
								m.selectedIndex = 0
							}
							return m, nil
						} else {
							// Show disk overwrite warning before confirming
							m.showDiskWarning()
							return m, nil
						}
					} else if fieldIndex == 10 {
						// Legacy boot toggle
						if m.config != nil {
							m.config.Disk.LegacyBoot = !m.config.Disk.LegacyBoot
						}
					} else if fieldIndex == 13 {
						// Preserve room toggle
						if m.config != nil {
							m.config.Disk.PreserveRoom = !m.config.Disk.PreserveRoom
						}
					} else if fieldIndex == 14 {
						// Preserve drives selector
						m.showingPreserveSelector = true
						m.inputStep = 0
						m.selectedDiskIndex = 0
						disks, err := util.GetAvailableDisks()
						if err != nil {
							m.errorMsg = fmt.Sprintf("Failed to list disks: %v", err)
							m.showingPreserveSelector = false
						} else {
							m.availableDisks = disks
						}
					} else if fieldIndex == 15 {
						// External drives selector
						m.showingExternalSelector = true
						m.inputStep = 0
						m.selectedDiskIndex = 0
						disks, err := util.GetAvailableDisks()
						if err != nil {
							m.errorMsg = fmt.Sprintf("Failed to list disks: %v", err)
							m.showingExternalSelector = false
						} else {
							m.availableDisks = disks
						}
					} else if fieldIndex == 16 {
						// Disable skeleton toggle
						if m.config != nil {
							if m.config.Skeleton.Default == "" && m.config.Skeleton.Profile == "" {
								m.config.Skeleton.Default = "default"
								m.config.Skeleton.Profile = m.config.Profile
							} else {
								m.config.Skeleton.Default = ""
								m.config.Skeleton.Profile = ""
							}
						}
					} else {
						// Start editing field
						m.startEdit()
					}
				}
			}
		case "e":
			if m.config != nil && !m.editing {
				fieldIndex := m.selectedIndex - len(m.configs)
				// Only allow 'e' for text-editable fields (0-9 are base fields, 11-12 are packages)
				if fieldIndex >= 0 && fieldIndex < 10 || fieldIndex == 11 || fieldIndex == 12 {
					m.startEdit()
				}
			}
		case "backspace":
			if m.config != nil && !m.editing {
				fieldIndex := m.selectedIndex - len(m.configs)
				if fieldIndex == 14 && len(m.config.Preserve) > 0 {
					// Clear preserve drives
					m.config.Preserve = []config.PreserveDrive{}
				} else if fieldIndex == 15 && len(m.config.External) > 0 {
					// Clear external drives
					m.config.External = []config.ExternalDrive{}
				}
			}
		case "d":
			if m.config == nil && m.selectedIndex < len(m.configs) {
				configName := m.configs[m.selectedIndex]
				if configName != "---" && config.IsLocalConfig(configName) {
					filename := fmt.Sprintf("%s.cue", configName)
					if err := os.Remove(filename); err != nil {
						m.errorMsg = fmt.Sprintf("Failed to delete %s: %v", filename, err)
					} else {
						m.errorMsg = fmt.Sprintf("✓ Deleted %s", filename)
						// Reload config list
						m.configs = config.ListConfigs()
						// Adjust selection if needed
						if m.selectedIndex >= len(m.configs) {
							m.selectedIndex = len(m.configs) - 1
						}
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ConfigConfirmedMsg:
		m.confirmed = true
		return m, tea.Quit
	}

	return m, nil
}

func (m *ConfigModel) loadConfig(name string) {
	// Load config from embedded or external file
	cfg, err := config.Load(fmt.Sprintf("%s.cue", name))
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to load config: %v", err)
		return
	}
	m.config = cfg
	m.configName = name
	m.errorMsg = ""
	m.selectedIndex = len(m.configs) // Move to first field
}
