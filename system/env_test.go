package system

import (
	"testing"
	"os"
	"io/ioutil"
	"path"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/system"
)

func TestEnvironmentFile(t *testing.T) {
	subs := map[string]string{
		"$public_ipv4":  "1.2.3.4",
		"$private_ipv4": "5.6.7.8",
		"$public_ipv6":  "1234::",
		"$private_ipv6": "5678::",
	}
	expect := "COREOS_PRIVATE_IPV4=5.6.7.8\nCOREOS_PRIVATE_IPV6=5678::\nCOREOS_PUBLIC_IPV4=1.2.3.4\nCOREOS_PUBLIC_IPV6=1234::\n"

	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	env := config.NewEnvironment("./", "./", "./", "", "", subs)
	ef := env.DefaultEnvironmentFile()
	err = system.WriteEnvFile(ef, dir)
	if err != nil {
		t.Fatalf("WriteEnvFile failed: %v", err)
	}

	fullPath := path.Join(dir, "etc", "environment")
	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != expect {
		t.Fatalf("File has incorrect contents: %q", contents)
	}
}
