package system

import (
	"path"

	"github.com/coreos/coreos-cloudinit/config"
)

type OEMRelease struct {
	config.OEMRelease
}

func (oem OEMRelease) File(root string) (*config.File, error) {
	if oem.ID == "" {
		return nil, nil
	}

	return &config.File{
		Path:               path.Join("etc", "oem-release"),
		RawFilePermissions: "0644",
		Content:            oem.String(),
	}, nil
}
