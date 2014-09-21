package system

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

type UpdateConfig struct {
	config.UpdateConfig
}

// File generates an `/etc/coreos/update.conf` file (if any update
// configuration options are set in cloud-config) by either rewriting the
// existing file on disk, or starting from `/usr/share/coreos/update.conf`
func (uc UpdateConfig) File(root string) (*config.File, error) {
	// Generate the list of possible substitutions to be performed based on the options that are configured
	subs := map[string]string{}
	uct := reflect.TypeOf(uc.UpdateConfig)
	ucv := reflect.ValueOf(uc.UpdateConfig)
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

	if len(subs) == 0 {
		return nil, nil
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

	return &config.File{
		Path:               path.Join("etc", "coreos", "update.conf"),
		RawFilePermissions: "0644",
		Content:            out,
	}, nil
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
