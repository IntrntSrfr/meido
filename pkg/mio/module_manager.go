package mio

import (
	"strings"

	"go.uber.org/zap"
)

type ModuleManager struct {
	Modules map[string]Module
	log     *zap.Logger
}

func NewModuleManager(log *zap.Logger) *ModuleManager {
	return &ModuleManager{
		Modules: make(map[string]Module),
		log:     log.Named("ModuleManager"),
	}
}

func (m *ModuleManager) RegisterModule(mod Module) {
	m.log.Info("adding module", zap.String("name", mod.Name()))
	err := mod.Hook()
	if err != nil {
		m.log.Error("could not register module", zap.Error(err))
		return
	}
	m.Modules[mod.Name()] = mod
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
