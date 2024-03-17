package bot

import (
	"strings"

	"go.uber.org/zap"
)

type ModuleManager struct {
	Modules map[string]Module
	logger  *zap.Logger
}

func NewModuleManager(logger *zap.Logger) *ModuleManager {
	logger = logger.Named("ModuleManager")
	return &ModuleManager{
		Modules: make(map[string]Module),
		logger:  logger,
	}
}

func (m *ModuleManager) RegisterModule(mod Module) {
	err := mod.Hook()
	if err != nil {
		m.logger.Error("Failed to register module", zap.String("module", mod.Name()), zap.Error(err))
		return
	}
	m.Modules[mod.Name()] = mod
	m.logger.Info("Registered module", zap.String("name", mod.Name()))
}

func (m *ModuleManager) FindModule(name string) (Module, error) {
	for _, m := range m.Modules {
		if strings.EqualFold(m.Name(), name) {
			return m, nil
		}
	}
	return nil, ErrModuleNotFound
}

func (m *ModuleManager) FindCommand(name string) (*ModuleCommand, error) {
	for _, m := range m.Modules {
		if cmd, err := m.FindCommand(name); err == nil {
			return cmd, nil
		}
	}
	return nil, ErrCommandNotFound
}

func (m *ModuleManager) FindPassive(name string) (*ModulePassive, error) {
	for _, m := range m.Modules {
		if pas, err := m.FindPassive(name); err == nil {
			return pas, nil
		}
	}
	return nil, ErrPassiveNotFound
}

func (m *ModuleManager) FindApplicationCommand(name string) (*ModuleApplicationCommand, error) {
	for _, m := range m.Modules {
		if cmd, err := m.FindApplicationCommand(name); err == nil {
			return cmd, nil
		}
	}
	return nil, ErrApplicationCommandNotFound
}
