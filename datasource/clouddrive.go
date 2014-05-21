package datasource

import(
	"io/ioutil"
	"path"
)

const (
	Path = "latest"
	File = "user-data"
)

type cloudDrive struct {
	path string
}

func NewCloudDrive(path string) *cloudDrive {
	return &cloudDrive{path}
}

func (self *cloudDrive) Fetch() ([]byte, error) {
	return ioutil.ReadFile(path.Join(self.path, Path, File))
}

func (self *cloudDrive) Type() string {
	return "cloud-drive"
}
