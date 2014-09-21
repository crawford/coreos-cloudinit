package config

import (
	"testing"
	"io/ioutil"
	"os"
	"path"
	"fmt"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestCloudConfigManageEtcHosts(t *testing.T) {
	contents := `
manage_etc_hosts: localhost
`
	cfg, err := config.NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	manageEtcHosts := cfg.ManageEtcHosts

	if manageEtcHosts != "localhost" {
		t.Errorf("ManageEtcHosts value is %q, expected 'localhost'", manageEtcHosts)
	}
}

func TestManageEtcHostsInvalidValue(t *testing.T) {
	eh := EtcHosts("invalid")
	if f, err := eh.File(""); err == nil || f != nil {
		t.Fatalf("EtcHosts File succeeded with invalid value!")
	}
}
