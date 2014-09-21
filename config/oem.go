package config

import (
	"fmt"
	"strings"
)

type OEMRelease struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	VersionID    string `yaml:"version-id"`
	HomeURL      string `yaml:"home-url"`
	BugReportURL string `yaml:"bug-report-url"`
}

func (oem OEMRelease) String() string {
	fields := []string{
		fmt.Sprintf("ID=%s", oem.ID),
		fmt.Sprintf("VERSION_ID=%s", oem.VersionID),
		fmt.Sprintf("NAME=%q", oem.Name),
		fmt.Sprintf("HOME_URL=%q", oem.HomeURL),
		fmt.Sprintf("BUG_REPORT_URL=%q", oem.BugReportURL),
	}

	return strings.Join(fields, "\n") + "\n"
}
