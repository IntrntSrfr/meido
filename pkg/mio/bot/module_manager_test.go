package bot

import (
	"testing"

	"github.com/intrntsrfr/meido/pkg/mio/test"
)

func TestModuleManager(t *testing.T) {
	mngr := NewModuleManager(test.NewTestLogger())

	if _, err := mngr.FindModule("test"); err == nil {
		t.Errorf("ModuleManager.FindModule() not returning error when not finding a module")
	}
	mod := NewTestModule(nil, "test", test.NewTestLogger())
	mngr.RegisterModule(mod)
	if _, err := mngr.FindModule("test"); err != nil {
		t.Errorf("ModuleManager.FindModule() returning error when finding a module")
	}

	if _, err := mngr.FindCommand("test"); err == nil {
		t.Errorf("ModuleManager.FindCommand() not returning error when not finding a command")
	}
	mod.RegisterCommands(&ModuleCommand{Name: "test"})
	if _, err := mngr.FindCommand("test"); err != nil {
		t.Errorf("ModuleManager.FindCommand() returning error when finding a command")
	}

	if _, err := mngr.FindPassive("test"); err == nil {
		t.Errorf("ModuleManager.FindPassive() not returning error when not finding a passive")
	}
	mod.RegisterPassives(&ModulePassive{Name: "test"})
	if _, err := mngr.FindPassive("test"); err != nil {
		t.Errorf("ModuleManager.FindPassive() returning error when finding a passive")
	}
}

func TestModuleManager_FailedHook(t *testing.T) {
	mngr := NewModuleManager(test.NewTestLogger())
	mod := NewTestModule(nil, "test", test.NewTestLogger())
	mod.hookShouldFail = true

	mngr.RegisterModule(mod)
	if len(mngr.Modules) != 0 {
		t.Errorf("len(ModuleManager.Modules) should be 0 after failed hook")
	}
}
