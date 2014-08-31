package validate

import (
	"regexp"
	"strconv"

	"github.com/coreos/coreos-cloudinit/third_party/launchpad.net/goyaml"
)

type node map[interface{}]interface{}

var (
	YamlRules []rule = []rule{
		syntax,
	}
	goyamlError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
	validNodes  = node{
		"hostname": node{},
		"coreos": node{
			"etcd": node{
				"addr":                  node{},
				"bind_addr":             node{},
				"ca_file":               node{},
				"cert_file":             node{},
				"cors":                  node{},
				"cpu_profile_file":      node{},
				"data_dir":              node{},
				"discovery":             node{},
				"http_read_timeout":     node{},
				"http_write_timeout":    node{},
				"key_file":              node{},
				"peers":                 node{},
				"peers_file":            node{},
				"max_cluster_size":      node{},
				"max_result_buffer":     node{},
				"max_retry_attempts":    node{},
				"name":                  node{},
				"snapshot":              node{},
				"verbose":               node{},
				"very_verbose":          node{},
				"peer-addr":             node{},
				"peer-bind_addr":        node{},
				"peer-ca_file":          node{},
				"peer-cert_file":        node{},
				"peer-key_file":         node{},
				"cluster-active_size":   node{},
				"cluster-remove_delay":  node{},
				"cluster-sync_interval": node{},
			},
			"fleet": node{
				"verbosity":                 node{},
				"etcd_servers":              node{},
				"etcd_request_timeout":      node{},
				"etcd_cafile":               node{},
				"etcd_keyfile":              node{},
				"etcd_certfile":             node{},
				"public_ip":                 node{},
				"metadata":                  node{},
				"agent_ttl":                 node{},
				"engine_reconcile_interval": node{},
			},
			"update": node{
				"reboot-strategy": node{},
				"server":          node{},
				"group":           node{},
			},
			"units": node{
				"name":    node{},
				"runtime": node{},
				"enable":  node{},
				"content": node{},
				"command": node{},
				"mask":    node{},
			},
		},
		"ssh_authorized_keys": node{},
		"users": node{
			"name":                     node{},
			"gecos":                    node{},
			"passwd":                   node{},
			"homedir":                  node{},
			"no-create-home":           node{},
			"primary-group":            node{},
			"groups":                   node{},
			"no-user-group":            node{},
			"ssh-authorized-keys":      node{},
			"coreos-ssh-import-github": node{},
			"coreos-ssh-import-url":    node{},
			"system":                   node{},
			"no-log-init":              node{},
		},
		"write_files": node{
			"path":        node{},
			"content":     node{},
			"permissions": node{},
			"owner":       node{},
		},
		"manage_etc_hosts": node{},
	}
)

func syntax(context ruleContext, validator *validator) {
	if err := goyaml.Unmarshal(context.content, &struct{}{}); err != nil {
		matches := goyamlError.FindStringSubmatch(err.Error())
		line, err := strconv.Atoi(matches[1])
		if err != nil {
			panic(err)
		}
		msg := matches[2]
		validator.report.Error(context.currentLine+line, msg)
	}
}

/*func nodes(context ruleContext, validator *validator) error {
	var config node
	if err := goyaml.Unmarshal(context.content, &config); err != nil {
		return nil
	}
	checkNode(config, validNodes, validator.report, string(context.content), 0)
	return nil
}

func checkNode(n, c node, r Reporter, config string, lineNum int) {
	for k, v := range n {
		lineNum := lineNum
		config := config

		for {
			tokens := strings.SplitN(config, "\n", 2)
			line := tokens[0]
			if len(tokens) > 1 {
				config = tokens[1]
			} else {
				config = ""
			}
			lineNum++

			if strings.TrimSpace(strings.Split(line, ":")[0]) == fmt.Sprint(k) {
				break
			}
		}

		if sc, ok := c[k]; ok {
			if sn, ok := v.(map[interface{}]interface{}); ok {
				checkNode(node(sn), sc.(node), r, config, lineNum)
			}
		} else {
			r.Warning(lineNum, fmt.Sprintf("unrecognized key %q", k))
		}
	}
}*/
