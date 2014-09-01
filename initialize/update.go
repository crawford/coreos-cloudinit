package initialize

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

const (
	locksmithUnit    = "locksmithd.service"
	updateEngineUnit = "update-engine.service"
)

// updateOption represents a configurable update option, which, if set, will be
// written into update.conf, replacing any existing value for the option
type updateOption struct {
	key    string   // key used to configure this option in cloud-config
	valid  []string // valid values for the option
	prefix string   // prefix for the option in the update.conf file
	value  string   // used to store the new value in update.conf (including prefix)
	seen   bool     // whether the option has been seen in any existing update.conf
}

// isValid checks whether a supplied value is valid for this option
func isValid(valid, val string) bool {
	if len(valid) == 0 {
		return true
	}
	for _, v := range strings.Split(valid, ",") {
		if val == v {
			return true
		}
	}
	return false
}

type UpdateConfig struct {
	RebootStrategy string `yaml:"reboot-strategy" env:"REBOOT_STRATEGY" valid:"best-effort,etcd-lock,reboot,off"`
	Group          string `yaml:"group"           env:"GROUP"`
	Server         string `yaml:"server"          env:"SERVER"`
}

// File generates an `/etc/coreos/update.conf` file (if any update
// configuration options are set in cloud-config) by either rewriting the
// existing file on disk, or starting from `/usr/share/coreos/update.conf`
func (uc UpdateConfig) File(root string) (*system.File, error) {
	if environmentLen(uc) == 0 {
		return nil, nil
	}

	// Generate the list of possible substitutions to be performed based on the options that are configured
	subs := map[string]string{}
	uct := reflect.TypeOf(uc)
	ucv := reflect.ValueOf(uc)
	for i := 0; i < uct.NumField(); i++ {
		ft := uct.Field(i)
		fv := ucv.Field(i)

		val := fv.String()
		if val == "" {
			continue
		}
		valid := ft.Tag.Get("valid")
		if !isValid(valid, val) {
			return nil, errors.New(fmt.Sprintf("invalid value %q for option %q (valid options: %q)", val, ft.Name, valid))
		}
		env := ft.Tag.Get("env")
		subs[env] = fmt.Sprintf("%s=%s", env, val)
	}

	etcUpdate := path.Join(root, "etc", "coreos", "update.conf")
	usrUpdate := path.Join(root, "usr", "share", "coreos", "update.conf")

	conf, err := os.Open(etcUpdate)
	if os.IsNotExist(err) {
		conf, err = os.Open(usrUpdate)
	}
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conf)
	var out string

	for scanner.Scan() {
		line := scanner.Text()
		for env, value := range subs {
			if strings.HasPrefix(line, env) {
				line = value
				delete(subs, env)
				break
			}
		}
		out += line
		out += "\n"
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	for _, value := range subs {
		out += value
		out += "\n"
	}

	return &system.File{
		Path:               path.Join("etc", "coreos", "update.conf"),
		RawFilePermissions: "0644",
		Content:            out,
	}, nil
}

// Units generates units for the cloud-init initializer to act on:
// - a locksmith system.Unit, if "reboot-strategy" was set in cloud-config
// - an update_engine system.Unit, if "group" or "server" was set in cloud-config
func (uc UpdateConfig) Units(root string) ([]system.Unit, error) {
	var units []system.Unit
	if uc.RebootStrategy != "" {
		ls := &system.Unit{
			Name:    locksmithUnit,
			Command: "restart",
			Mask:    false,
			Runtime: true,
		}

		if uc.RebootStrategy == "off" {
			ls.Command = "stop"
			ls.Mask = true
		}
		units = append(units, *ls)
	}

	if uc.Group != "" || uc.Server != "" {
		ue := system.Unit{
			Name:    updateEngineUnit,
			Command: "restart",
		}
		units = append(units, ue)
	}

	return units, nil
}
