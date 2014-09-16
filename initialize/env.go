package initialize

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type Environment struct {
	root          string
	configRoot    string
	workspace     string
	netconfType   string
	sshKeyName    string
	substitutions map[string]string
}

// TODO(jonboulle): this is getting unwieldy, should be able to simplify the interface somehow
func NewEnvironment(root, configRoot, workspace, netconfType, sshKeyName string, substitutions map[string]string) *Environment {
	if substitutions == nil {
		substitutions = make(map[string]string)
	}
	// If certain values are not in the supplied substitution, fall back to retrieving them from the environment
	for k, v := range map[string]string{
		"$public_ipv4":  os.Getenv("COREOS_PUBLIC_IPV4"),
		"$private_ipv4": os.Getenv("COREOS_PRIVATE_IPV4"),
		"$public_ipv6":  os.Getenv("COREOS_PUBLIC_IPV6"),
		"$private_ipv6": os.Getenv("COREOS_PRIVATE_IPV6"),
	} {
		if _, ok := substitutions[k]; !ok {
			substitutions[k] = v
		}
	}
	return &Environment{root, configRoot, workspace, netconfType, sshKeyName, substitutions}
}

func (e *Environment) Workspace() string {
	return path.Join(e.root, e.workspace)
}

func (e *Environment) Root() string {
	return e.root
}

func (e *Environment) ConfigRoot() string {
	return e.configRoot
}

func (e *Environment) NetconfType() string {
	return e.netconfType
}

func (e *Environment) SSHKeyName() string {
	return e.sshKeyName
}

func (e *Environment) SetSSHKeyName(name string) {
	e.sshKeyName = name
}

// Apply goes through the map of substitutions and replaces all instances of
// the keys with their respective values. It supports escaping substitutions
// with a leading '\'.
func (e *Environment) Apply(data string) string {
	for key, val := range e.substitutions {
		matchKey := strings.Replace(key, `$`, `\$`, -1)
		replKey := strings.Replace(key, `$`, `$$`, -1)

		// "key" -> "val"
		data = regexp.MustCompile(`([^\\]|^)`+matchKey).ReplaceAllString(data, `${1}`+val)
		// "\key" -> "key"
		data = regexp.MustCompile(`\\`+matchKey).ReplaceAllString(data, replKey)
	}
	return data
}

func (e *Environment) DefaultEnvironmentFile() *system.EnvFile {
	ef := system.EnvFile{
		File: &system.File{
			Path: "/etc/environment",
		},
		Vars: map[string]string{},
	}
	if ip, ok := e.substitutions["$public_ipv4"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PUBLIC_IPV4"] = ip
	}
	if ip, ok := e.substitutions["$private_ipv4"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PRIVATE_IPV4"] = ip
	}
	if ip, ok := e.substitutions["$public_ipv6"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PUBLIC_IPV6"] = ip
	}
	if ip, ok := e.substitutions["$private_ipv6"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PRIVATE_IPV6"] = ip
	}
	if len(ef.Vars) == 0 {
		return nil
	} else {
		return &ef
	}
}

func environmentString(e interface{}) string {
	et := reflect.TypeOf(e)
	ev := reflect.ValueOf(e)

	out := "[Service]\n"
	for i := 0; i < et.NumField(); i++ {
		val := ev.Field(i).String()
		if val != "" {
			key := et.Field(i).Tag.Get("env")
			out += fmt.Sprintf("Environment=\"%s=%s\"\n", key, val)
		}
	}
	return out
}

func environmentLen(e interface{}) int {
	ev := reflect.ValueOf(e)

	count := 0
	for i := 0; i < ev.NumField(); i++ {
		if ev.Field(i).String() != "" {
			count++
		}
	}
	return count
}
