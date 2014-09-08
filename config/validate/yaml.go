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
	goyamlLineError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
	goyamlError     = regexp.MustCompile(`^YAML error: (?P<msg>.*)$`)
)

func syntax(c context, v *validator) {
	if err := goyaml.Unmarshal(c.content, &struct{}{}); err != nil {
		matches := goyamlLineError.FindStringSubmatch(err.Error())
		if len(matches) > 0 {
			line, err := strconv.Atoi(matches[1])
			if err != nil {
				panic(err)
			}
			msg := matches[2]
			v.report.Error(c.line+line, msg)
			return
		}

		matches = goyamlError.FindStringSubmatch(err.Error())
		if len(matches) > 0 {
			msg := matches[1]
			v.report.Error(c.line+1, msg)
			return
		}

		panic("couldn't parse goyaml error")
	}
}

func nodes(c context, v *validator) {
	var n node
	if err := goyaml.Unmarshal(c.content, &n); err == nil {
		fmt.Printf("%#v\n", toNode(config.CloudConfig{}))
		checkNode(n, toNode(config.CloudConfig{}), c, v)
	}
}

func toNode(s interface{}) node {
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)

	if sv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%T is not a struct", s))
	}

	n := make(node)
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)
		fv := sv.Field(i)
		k := ft.Tag.Get("yaml")

		if k == "-" {
			continue
		}

		fmt.Printf("%s: %s\n", k, fv.Kind())
		switch fv.Kind() {
		case reflect.Struct:
			n[k] = toNode(fv.Interface())
		case reflect.Slice:
			fmt.Printf("SLICE %s %T\n", ft.Type, fv.Interface())
			switch ft.Type.Elem().Kind() {
			case reflect.Struct:
			fmt.Println("[]STRUCT")
				n[k] = toNode(fv.Interface())
			default:
			fmt.Println("[]?")
				n[k] = fv.Interface()
			}
		default:
			n[k] = fv.Interface()
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
			fmt.Printf("got %T, want %T\n", v, sc)
			if sn, ok := v.(map[interface{}]interface{}); ok {
				checkNode(node(sn), sc.(node), cfg, val)
			} else {
				fmt.Printf("%#v\n", v)
				if reflect.TypeOf(v).Kind() == reflect.Slice && reflect.TypeOf(sc).Kind() == reflect.Slice {
					fmt.Printf("%v %v\n", reflect.TypeOf(v).Elem(), reflect.TypeOf(sc).Elem())
				} else {
					if !reflect.TypeOf(v).ConvertibleTo(reflect.TypeOf(sc)) {
						val.report.Warning(cfg.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
					}
				}
			}
		} else {
			val.report.Warning(cfg.line, fmt.Sprintf("unrecognized key %q", k))
		}
	}
}
