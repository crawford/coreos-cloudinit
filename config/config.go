package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"

	"github.com/coreos/coreos-cloudinit/third_party/launchpad.net/goyaml"
)

// CloudConfigFile represents a CoreOS specific configuration option that can generate
// an associated system.File to be written to disk
type CloudConfigFile interface {
	// File should either return (*system.File, error), or (nil, nil) if nothing
	// needs to be done for this configuration option.
	File(root string) (*system.File, error)
}

// CloudConfigUnit represents a CoreOS specific configuration option that can generate
// associated system.Units to be created/enabled appropriately
type CloudConfigUnit interface {
	Units(root string) ([]system.Unit, error)
}

// CloudConfig encapsulates the entire cloud-config configuration file and maps directly to YAML
type CloudConfig struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Coreos            struct {
		Etcd   EtcdEnvironment  `yaml:"etcd"`
		Fleet  FleetEnvironment `yaml:"fleet"`
		OEM    OEMRelease       `yaml:"oem"`
		Update UpdateConfig     `yaml:"update"`
		Units  []system.Unit    `yaml:"units"`
	} `yaml:"coreos"`
	WriteFiles        []system.File `yaml:"write_files"`
	Hostname          string        `yaml:"hostname"`
	Users             []system.User `yaml:"users"`
	ManageEtcHosts    EtcHosts      `yaml:"manage_etc_hosts"`
	NetworkConfigPath string        `yaml:"-"`
	NetworkConfig     string        `yaml:"-"`
}

func normalizeKeys(config string) string {
	yamlKey := regexp.MustCompile(`^.*?:`)
	return yamlKey.ReplaceAllStringFunc(config, func(match string) string {
		return strings.Replace(match, "-", "_", -1)
	})
}

// NewCloudConfig instantiates a new CloudConfig from the given contents (a
// string of YAML), returning any error encountered.
func NewCloudConfig(contents string) (*CloudConfig, error) {
	var cfg CloudConfig
	err := goyaml.Unmarshal([]byte(normalizeKeys(contents)), &cfg)
	if err != nil {
		return &cfg, err
	}
	return &cfg, nil
}

func (cc CloudConfig) String() string {
	bytes, err := goyaml.Marshal(cc)
	if err != nil {
		return ""
	}

	stringified := string(bytes)
	stringified = fmt.Sprintf("#cloud-config\n%s", stringified)

	return stringified
}
