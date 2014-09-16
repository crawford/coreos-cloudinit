package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/coreos/coreos-cloudinit/system"
)

func TestEnvironmentApply(t *testing.T) {
	os.Setenv("COREOS_PUBLIC_IPV4", "1.2.3.4")
	os.Setenv("COREOS_PRIVATE_IPV4", "5.6.7.8")
	os.Setenv("COREOS_PUBLIC_IPV6", "1234::")
	os.Setenv("COREOS_PRIVATE_IPV6", "5678::")
	for _, tt := range []struct {
		subs  map[string]string
		input string
		out   string
	}{
		{
			// Substituting both values directly should always take precedence
			// over environment variables
			map[string]string{
				"$public_ipv4":  "192.0.2.3",
				"$private_ipv4": "192.0.2.203",
				"$public_ipv6":  "fe00:1234::",
				"$private_ipv6": "fe00:5678::",
			},
			`[Service]
ExecStart=/usr/bin/echo "$public_ipv4 $public_ipv6"
ExecStop=/usr/bin/echo $private_ipv4 $private_ipv6
ExecStop=/usr/bin/echo $unknown`,
			`[Service]
ExecStart=/usr/bin/echo "192.0.2.3 fe00:1234::"
ExecStop=/usr/bin/echo 192.0.2.203 fe00:5678::
ExecStop=/usr/bin/echo $unknown`,
		},
		{
			// Substituting one value directly while falling back with the other
			map[string]string{"$private_ipv4": "127.0.0.1"},
			"$private_ipv4\n$public_ipv4",
			"127.0.0.1\n1.2.3.4",
		},
		{
			// Falling back to environment variables for both values
			map[string]string{"foo": "bar"},
			"$private_ipv4\n$public_ipv4",
			"5.6.7.8\n1.2.3.4",
		},
		{
			// No substitutions
			nil,
			"$private_ipv4\nfoobar",
			"5.6.7.8\nfoobar",
		},
		{
			// Escaping substitutions
			map[string]string{"$private_ipv4": "127.0.0.1"},
			`\$private_ipv4
$private_ipv4
addr: \$private_ipv4
\\$private_ipv4`,
			`$private_ipv4
127.0.0.1
addr: $private_ipv4
\$private_ipv4`,
		},
		{
			// No substitutions with escaping
			nil,
			"\\$test\n$test",
			"\\$test\n$test",
		},
	} {

		env := NewEnvironment("./", "./", "./", "", "", tt.subs)
		got := env.Apply(tt.input)
		if got != tt.out {
			t.Fatalf("Environment incorrectly applied.\ngot:\n%s\nwant:\n%s", got, tt.out)
		}
	}
}

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

	env := NewEnvironment("./", "./", "./", "", "", subs)
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

func TestEnvironmentFileNil(t *testing.T) {
	subs := map[string]string{
		"$public_ipv4":  "",
		"$private_ipv4": "",
		"$public_ipv6":  "",
		"$private_ipv6": "",
	}

	env := NewEnvironment("./", "./", "./", "", "", subs)
	ef := env.DefaultEnvironmentFile()
	if ef != nil {
		t.Fatalf("Environment file not nil: %v", ef)
	}
}
