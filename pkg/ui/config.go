package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/grantios/instagrant/pkg/config"
)

// StartConfigUI starts the configuration selection UI
func StartConfigUI(configs []string) (*config.Config, error) {
	clearTerminal()

	model := NewConfigModel(configs)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	configModel, ok := finalModel.(*ConfigModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	if configModel.confirmed && configModel.config != nil {
		return configModel.config, nil
	}

	return nil, fmt.Errorf("configuration cancelled")
}
