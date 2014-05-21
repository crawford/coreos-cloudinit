package datasource

import (
	"io/ioutil"
	"os"
	"path"
)

const (
	latestPath       = "openstack/latest"
	userDataFilename = "user-data"
)

type cloudDrive struct {
	path string
}

func NewCloudDrive(path string) *cloudDrive {
	return &cloudDrive{path}
}

func (self *cloudDrive) Fetch() ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(self.path, latestPath, userDataFilename))
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}

func (self *cloudDrive) Type() string {
	return "cloud-drive"
}
