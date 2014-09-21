package config

import (
	"sort"
	"strings"
	"testing"
)

const (
	base = `SERVER=https://example.com
GROUP=thegroupc`
	configured = base + `
REBOOT_STRATEGY=awesome
`
	expected = base + `
REBOOT_STRATEGY=etcd-lock
`
)

func TestEmptyUpdateConfig(t *testing.T) {
	uc := &UpdateConfig{}
	f, err := uc.File("")
	if err != nil {
		t.Error("unexpected error getting file from empty UpdateConfig")
	}
	if f != nil {
		t.Errorf("getting file from empty UpdateConfig should have returned nil, got %v", f)
	}
	uu, err := uc.Units("")
	if err != nil {
		t.Error("unexpected error getting unit from empty UpdateConfig")
	}
	if len(uu) != 0 {
		t.Errorf("getting unit from empty UpdateConfig should have returned zero units, got %d", len(uu))
	}
}

func TestInvalidUpdateOptions(t *testing.T) {
	if !isValid("one,two", "one") {
		t.Error("update option did not accept valid option \"one\"")
	}
	if isValid("one,two", "three") {
		t.Error("update option accepted invalid option \"three\"")
	}
	for _, s := range []string{"one", "asdf", "foobarbaz"} {
		if !isValid("", s) {
			t.Errorf("update option with no \"valid\" field did not accept %q", s)
		}
	}

	uc := &UpdateConfig{RebootStrategy: "wizzlewazzle"}
	f, err := uc.File("")
	if err == nil {
		t.Errorf("File did not give an error on invalid UpdateOption")
	}
	if f != nil {
		t.Errorf("File did not return a nil file on invalid UpdateOption")
	}
}

func TestServerGroupOptions(t *testing.T) {
	u := &UpdateConfig{Group: "master", Server: "http://foo.com"}

	want := `
GROUP=master
SERVER=http://foo.com`

	f, err := u.File(dir)
	if err != nil {
		t.Errorf("unexpected error getting file from UpdateConfig: %v", err)
	} else if f == nil {
		t.Error("unexpectedly got empty file from UpdateConfig")
	} else {
		out := strings.Split(f.Content, "\n")
		sort.Strings(out)
		got := strings.Join(out, "\n")
		if got != want {
			t.Errorf("File has incorrect contents, got %v, want %v", got, want)
		}
	}

	uu, err := u.Units(dir)
	if err != nil {
		t.Errorf("unexpected error getting units from UpdateConfig: %v", err)
	} else if len(uu) != 1 {
		t.Errorf("unexpected number of files returned from UpdateConfig: want 1, got %d", len(uu))
	} else {
		unit := uu[0]
		if unit.Name != "update-engine.service" {
			t.Errorf("bad name for generated unit: want update-engine.service, got %s", unit.Name)
		}
		if unit.Command != "restart" {
			t.Errorf("bad command for generated unit: want restart, got %s", unit.Command)
		}
	}
}

func TestRebootStrategies(t *testing.T) {
	strategies := []struct {
		name     string
		line     string
		uMask    bool
		uCommand string
	}{
		{"best-effort", "REBOOT_STRATEGY=best-effort", false, "restart"},
		{"etcd-lock", "REBOOT_STRATEGY=etcd-lock", false, "restart"},
		{"reboot", "REBOOT_STRATEGY=reboot", false, "restart"},
		{"off", "REBOOT_STRATEGY=off", true, "stop"},
	}
	for _, s := range strategies {
		uc := &UpdateConfig{RebootStrategy: s.name}
		f, err := uc.File(dir)
		if err != nil {
			t.Errorf("update failed to generate file for reboot-strategy=%v: %v", s.name, err)
		} else if f == nil {
			t.Errorf("generated empty file for reboot-strategy=%v", s.name)
		} else {
			seen := false
			for _, line := range strings.Split(f.Content, "\n") {
				if line == s.line {
					seen = true
					break
				}
			}
			if !seen {
				t.Errorf("couldn't find expected line %v for reboot-strategy=%v", s.line)
			}
		}
		uu, err := uc.Units(dir)
		if err != nil {
			t.Errorf("failed to generate unit for reboot-strategy=%v!", s.name)
		} else if len(uu) != 1 {
			t.Errorf("unexpected number of units for reboot-strategy=%v: %d", s.name, len(uu))
		} else {
			u := uu[0]
			if u.Name != locksmithUnit {
				t.Errorf("unit generated for reboot strategy=%v had bad name: %v", s.name, u.Name)
			}
			if u.Mask != s.uMask {
				t.Errorf("unit generated for reboot strategy=%v had bad mask: %t", s.name, u.Mask)
			}
			if u.Command != s.uCommand {
				t.Errorf("unit generated for reboot strategy=%v had bad command: %v", s.name, u.Command)
			}
		}
	}
}
