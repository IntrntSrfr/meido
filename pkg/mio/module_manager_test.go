package mio

import (
	"testing"
)

func TestModuleManager(t *testing.T) {
	mngr := NewModuleManager(testLogger())

	if _, err := mngr.FindModule("test"); err == nil {
		t.Errorf("ModuleManager.FindModule() not returning error when not finding a module")
	}
	mod := newTestModule(nil, "test", testLogger())
	mngr.RegisterModule(mod)
	if _, err := mngr.FindModule("test"); err != nil {
		t.Errorf("ModuleManager.FindModule() returning error when finding a module")
	}

	if _, err := mngr.FindCommand("test"); err == nil {
		t.Errorf("ModuleManager.FindCommand() not returning error when not finding a command")
	}
	mod.RegisterCommand(&ModuleCommand{Name: "test"})
	if _, err := mngr.FindCommand("test"); err != nil {
		t.Errorf("ModuleManager.FindCommand() returning error when finding a command")
	}

	if _, err := mngr.FindPassive("test"); err == nil {
		t.Errorf("ModuleManager.FindPassive() not returning error when not finding a passive")
	}
	mod.RegisterPassive(&ModulePassive{Name: "test"})
	if _, err := mngr.FindPassive("test"); err != nil {
		t.Errorf("ModuleManager.FindPassive() returning error when finding a passive")
	}
}

func TestModuleManager_FailedHook(t *testing.T) {
	mngr := NewModuleManager(testLogger())
	mod := newTestModule(nil, "test", testLogger())
	mod.hookShouldFail = true

	mngr.RegisterModule(mod)
	if len(mngr.Modules) != 0 {
		t.Errorf("len(ModuleManager.Modules) should be 0 after failed hook")
	}
}
