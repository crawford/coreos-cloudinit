package validate

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"

	"github.com/coreos/coreos-cloudinit/third_party/launchpad.net/goyaml"
)

type node map[interface{}]interface{}

var (
	YamlRules []rule = []rule{
		syntax,
		nodes,
	}
	goyamlError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
)

func syntax(c context, v *validator) {
	if err := goyaml.Unmarshal(c.content, &struct{}{}); err != nil {
		matches := goyamlError.FindStringSubmatch(err.Error())
		line, err := strconv.Atoi(matches[1])
		if err != nil {
			panic(err)
		}
		msg := matches[2]
		v.report.Error(c.line+line+1, msg)
	}
}

func nodes(c context, v *validator) {
	var n node
	if err := goyaml.Unmarshal(c.content, &n); err == nil {
		checkNode(n, toNode(config.CloudConfig{}), c, v)
	}
}

func toNode(s interface{}) node {
	n := make(node)
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)

	if sv.Kind() != reflect.Struct {
		return n
	}

	for i := 0; i < st.NumField(); i++ {
		k := st.Field(i).Tag.Get("yaml")
		if k != "-" {
			n[k] = toNode(sv.Field(i).Interface())
		}
	}
	return n
}

func checkNode(n, c node, cfg context, val *validator) {
	for k, v := range n {
		cfg := cfg

		for {
			tokens := strings.SplitN(string(cfg.content), "\n", 2)
			line := tokens[0]
			if len(tokens) > 1 {
				cfg.content = []byte(tokens[1])
			} else {
				cfg.content = []byte{}
			}
			cfg.line++

			if strings.TrimSpace(strings.Split(line, ":")[0]) == fmt.Sprint(k) {
				break
			}
		}

		if sc, ok := c[k]; ok {
			if sn, ok := v.(map[interface{}]interface{}); ok {
				checkNode(node(sn), sc.(node), cfg, val)
			}
		} else {
			val.report.Warning(cfg.line, fmt.Sprintf("unrecognized key %q", k))
		}
	}
}
