package system

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"

	"github.com/coreos/coreos-cloudinit/network"
)

const (
	runtimeNetworkPath = "/run/systemd/network"
)

func DownNetworkInterfaces(interfaces []network.InterfaceGenerator) error {
	sysInterfaceMap := make(map[string]net.Interface)
	if systemInterfaces, err := net.Interfaces(); err == nil {
		for _, iface := range systemInterfaces {
			sysInterfaceMap[iface.Name] = iface
		}
	} else {
		return err
	}

	for _, iface := range interfaces {
		if systemInterface, ok := sysInterfaceMap[iface.Name()]; ok {
			err := exec.Command("ip", "link", "set", systemInterface.Name, "down").Run()
			if err != nil {
				fmt.Printf("Error while downing interface '%s' (%s). Continuing...\n", systemInterface.Name, err)
			}
		}
	}

	return nil
}

func Probe8012q() error {
	return exec.Command("modprobe", "8021q").Run()
}

func RestartNetworkd() error {
	_, err := RunUnitCommand("restart", "systemd-networkd.service")
	return err
}

func WriteNetworkdConfigs(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		filename := path.Join(runtimeNetworkPath, fmt.Sprintf("%s.netdev", iface.Name()))
		if err := writeConfig(filename, iface.GenerateNetdevConfig()); err != nil {
			return err
		}
		filename = path.Join(runtimeNetworkPath, fmt.Sprintf("%s.link", iface.Name()))
		if err := writeConfig(filename, iface.GenerateLinkConfig()); err != nil {
			return err
		}
		filename = path.Join(runtimeNetworkPath, fmt.Sprintf("%s.network", iface.Name()))
		if err := writeConfig(filename, iface.GenerateNetworkConfig()); err != nil {
			return err
		}
	}
	return nil
}

func writeConfig(filename string, config string) error {
	if config == "" {
		return nil
	}

	if file, err := os.Create(filename); err == nil {
		io.WriteString(file, config)
		file.Close()
		return nil
	} else {
		return err
	}
}
