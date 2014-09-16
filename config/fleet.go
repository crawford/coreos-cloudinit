package config

import (
	"github.com/coreos/coreos-cloudinit/system"
)

type FleetEnvironment struct {
	AgentTTL                string `yaml:"agent_ttl"                 env:"FLEET_AGENT_TTL"`
	EngineReconcileInterval string `yaml:"engine_reconcile_interval" env:"FLEET_ENGINE_RECONCILE_INTERVAL"`
	EtcdCAFile              string `yaml:"etcd_cafile"               env:"FLEET_ETCD_CAFILE"`
	EtcdCertFile            string `yaml:"etcd_certfile"             env:"FLEET_ETCD_CERTFILE"`
	EtcdKeyFile             string `yaml:"etcd_keyfile"              env:"FLEET_ETCD_KEYFILE"`
	EtcdRequestTimeout      string `yaml:"etcd_request_timeout"      env:"FLEET_ETCD_REQUEST_TIMEOUT"`
	EtcdServers             string `yaml:"etcd_servers"              env:"FLEET_ETCD_SERVERS"`
	Metadata                string `yaml:"metadata"                  env:"FLEET_METADATA"`
	PublicIP                string `yaml:"public_ip"                 env:"FLEET_PUBLIC_IP"`
	Verbosity               string `yaml:"verbosity"                 env:"FLEET_VERBOSITY"`
}

func (fe FleetEnvironment) String() string {
	return "[Service]\n" + environmentString(fe)
}

// Units generates a Unit file drop-in for fleet, if any fleet options were
// configured in cloud-config
func (fe FleetEnvironment) Units(root string) ([]system.Unit, error) {
	if environmentLen(fe) == 0 {
		return nil, nil
	}
	fleet := system.Unit{
		Name:    "fleet.service",
		Runtime: true,
		DropIn:  true,
		Content: fe.String(),
	}
	return []system.Unit{fleet}, nil
}
