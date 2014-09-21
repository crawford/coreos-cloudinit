package system

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/coreos/coreos-cloudinit/config"
)

const DefaultIpv4Address = "127.0.0.1"

type EtcHosts struct {
	config.EtcHosts
}

func (eh EtcHosts) generateEtcHosts() (out string, err error) {
	if eh.EtcHosts != "localhost" {
		return "", errors.New("Invalid option to manage_etc_hosts")
	}

	// use the operating system hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s\n", DefaultIpv4Address, hostname), nil

}

func (eh EtcHosts) File(root string) (*config.File, error) {
	if eh.EtcHosts == "" {
		return nil, nil
	}

	etcHosts, err := eh.generateEtcHosts()
	if err != nil {
		return nil, err
	}

	return &config.File{
		Path:               path.Join("etc", "hosts"),
		RawFilePermissions: "0644",
		Content:            etcHosts,
	}, nil
}
