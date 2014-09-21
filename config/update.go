package config

const (
	locksmithUnit    = "locksmithd.service"
	updateEngineUnit = "update-engine.service"
)

type UpdateConfig struct {
	RebootStrategy string `yaml:"reboot-strategy" env:"REBOOT_STRATEGY" valid:"best-effort,etcd-lock,reboot,off"`
	Group          string `yaml:"group"           env:"GROUP"`
	Server         string `yaml:"server"          env:"SERVER"`
}

// Units generates units for the cloud-init initializer to act on:
// - a locksmith Unit, if "reboot-strategy" was set in cloud-config
// - an update_engine Unit, if "group" or "server" was set in cloud-config
func (uc UpdateConfig) Units(root string) ([]Unit, error) {
	var units []Unit
	if uc.RebootStrategy != "" {
		ls := &Unit{
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
		ue := Unit{
			Name:    updateEngineUnit,
			Command: "restart",
		}
		units = append(units, ue)
	}

	return units, nil
}
