package config

import (
	"os"
	"strconv"
	"errors"
)

type File struct {
	Encoding           string `yaml:"-"`
	Content            string `yaml:"content"`
	Owner              string `yaml:"owner"`
	Path               string `yaml:"path"`
	RawFilePermissions string `yaml:"permissions"`
}

func (f *File) Permissions() (os.FileMode, error) {
	if f.RawFilePermissions == "" {
		return os.FileMode(0644), nil
	}

	// Parse string representation of file mode as octal
	perm, err := strconv.ParseInt(f.RawFilePermissions, 8, 32)
	if err != nil {
		return 0, errors.New("Unable to parse file permissions as octal integer")
	}
	return os.FileMode(perm), nil
}


