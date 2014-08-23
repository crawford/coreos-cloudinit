package initialize

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/coreos/coreos-cloudinit/system"
)

type EtcdEnvironment struct {
	Addr                string `yaml:"addr"                  env:"ETCD_ADDR"`
	BindAddr            string `yaml:"bind_addr"             env:"ETCD_BIND_ADDR"`
	CAFile              string `yaml:"ca_file"               env:"ETCD_CA_FILE"`
	CertFile            string `yaml:"cert_file"             env:"ETCD_CERT_FILE"`
	ClusterActiveSize   string `yaml:"cluster-active_size"   env:"ETCD_CLUSTER_ACTIVE_SIZE"`
	ClusterRemoveDelay  string `yaml:"cluster-remove_delay"  env:"ETCD_CLUSTER_REMOVE_DELAY"`
	ClusterSyncInterval string `yaml:"cluster-sync_interval" env:"ETCD_CLUSTER_SYNC_INTERVAL"`
	Cors                string `yaml:"cors"                  env:"ETCD_CORS"`
	CPUProfileFile      string `yaml:"cpu_profile_file"      env:"ETCD_CPU_PROFILE_FILE"`
	DataDir             string `yaml:"data_dir"              env:"ETCD_DATA_DIR"`
	Discovery           string `yaml:"discovery"             env:"ETCD_DISCOVERY"`
	HTTPReadTimeout     string `yaml:"http_read_timeout"     env:"ETCD_HTTP_READ_TIMEOUT"`
	HTTPWriteTimeout    string `yaml:"http_write_timeout"    env:"ETCD_HTTP_WRITE_TIMEOUT"`
	KeyFile             string `yaml:"key_file"              env:"ETCD_KEY_FILE"`
	MaxClusterSize      string `yaml:"max_cluster_size"      env:"ETCD_MAX_CLUSTER_SIZE"`
	MaxResultBuffer     string `yaml:"max_result_buffer"     env:"ETCD_MAX_RESULT_BUFFER"`
	MaxRetryAttempts    string `yaml:"max_retry_attempts"    env:"ETCD_MAX_RETRY_ATTEMPTS"`
	Name                string `yaml:"name"                  env:"ETCD_NAME"`
	PeerAddr            string `yaml:"peer-addr"             env:"ETCD_PEER_ADDR"`
	PeerBindAddr        string `yaml:"peer-bind_addr"        env:"ETCD_PEER_BIND_ADDR"`
	PeerCAFile          string `yaml:"peer-ca_file"          env:"ETCD_PEER_CA_FILE"`
	PeerCertFile        string `yaml:"peer-cert_file"        env:"ETCD_PEER_CERT_FILE"`
	PeerKeyFile         string `yaml:"peer-key_file"         env:"ETCD_PEER_KEY_FILE"`
	Peers               string `yaml:"peers"                 env:"ETCD_PEERS"`
	PeersFile           string `yaml:"peers_file"            env:"ETCD_PEERS_FILE"`
	Snapshot            string `yaml:"snapshot"              env:"ETCD_SNAPSHOT"`
	Verbose             string `yaml:"verbose"               env:"ETCD_VERBOSE"`
	VeryVerbose         string `yaml:"very_verbose"          env:"ETCD_VERY_VERBOSE"`
}

func (ee EtcdEnvironment) String() string {
	eet := reflect.TypeOf(ee)
	eev := reflect.ValueOf(ee)

	out := "[Service]\n"
	for i := 0; i < eet.NumField(); i++ {
		val := eev.Field(i).String()
		if val != "" {
			key := eet.Field(i).Tag.Get("env")
			out += fmt.Sprintf("Environment=\"%s=%s\"\n", key, val)
		}
	}
	return out
}

// Units creates a Unit file drop-in for etcd, using any configured options.
func (ee EtcdEnvironment) Units(root string) ([]system.Unit, error) {
	if ee.len() == 0 {
		return nil, nil
	}

	if ee.Name == "" {
		if machineID := system.MachineID(root); machineID != "" {
			ee.Name = machineID
		} else if hostname, err := system.Hostname(); err == nil {
			ee.Name = hostname
		} else {
			return nil, errors.New("Unable to determine default etcd name")
		}
	}

	etcd := system.Unit{
		Name:    "etcd.service",
		Runtime: true,
		DropIn:  true,
		Content: ee.String(),
	}
	return []system.Unit{etcd}, nil
}

func (ee EtcdEnvironment) len() int {
	eev := reflect.ValueOf(ee)

	count := 0
	for i := 0; i < eev.NumField(); i++ {
		if eev.Field(i).String() != "" {
			count++
		}
	}
	return count
}
