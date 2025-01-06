package mio_test

import (
	"testing"

	"github.com/intrntsrfr/meido/pkg/mio"
)

func TestModuleManager(t *testing.T) {
	mngr := mio.NewModuleManager(mio.NewDiscardLogger())

	if _, err := mngr.FindModule("test"); err == nil {
		t.Errorf("ModuleManager.FindModule() not returning error when not finding a module")
	}
	mod := NewTestModule(nil, "test", mio.NewDiscardLogger())
	mngr.RegisterModule(mod)
	if _, err := mngr.FindModule("test"); err != nil {
		t.Errorf("ModuleManager.FindModule() returning error when finding a module")
	}

	if _, err := mngr.FindCommand("test"); err == nil {
		t.Errorf("ModuleManager.FindCommand() not returning error when not finding a command")
	}
	mod.RegisterCommands(&mio.ModuleCommand{Name: "test"})
	if _, err := mngr.FindCommand("test"); err != nil {
		t.Errorf("ModuleManager.FindCommand() returning error when finding a command")
	}

	if _, err := mngr.FindPassive("test"); err == nil {
		t.Errorf("ModuleManager.FindPassive() not returning error when not finding a passive")
	}
	mod.RegisterPassives(&mio.ModulePassive{Name: "test"})
	if _, err := mngr.FindPassive("test"); err != nil {
		t.Errorf("ModuleManager.FindPassive() returning error when finding a passive")
	}
}

func TestModuleManager_FailedHook(t *testing.T) {
	mngr := mio.NewModuleManager(mio.NewDiscardLogger())
	mod := NewTestModule(nil, "test", mio.NewDiscardLogger())
	mod.hookShouldFail = true

	mngr.RegisterModule(mod)
	if len(mngr.Modules) != 0 {
		t.Errorf("len(ModuleManager.Modules) should be 0 after failed hook")
	}
}
