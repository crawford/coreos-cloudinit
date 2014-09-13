package validate

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"

	"github.com/coreos/coreos-cloudinit/third_party/gopkg.in/yaml.v1"
)

var (
	YamlRules []rule = []rule{
		syntax,
		nodes,
	}
	yamlLineError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
	yamlError     = regexp.MustCompile(`^YAML error: (?P<msg>.*)$`)
)

func syntax(c context, v *validator) {
	if err := yaml.Unmarshal(c.content, &struct{}{}); err != nil {
		matches := yamlLineError.FindStringSubmatch(err.Error())
		if len(matches) > 0 {
			line, err := strconv.Atoi(matches[1])
			if err != nil {
				panic(err)
			}
			msg := matches[2]
			v.report.Error(c.line+line, msg)
			return
		}

		matches = yamlError.FindStringSubmatch(err.Error())
		if len(matches) > 0 {
			msg := matches[1]
			v.report.Error(c.line+1, msg)
			return
		}

		panic("couldn't parse yaml error")
	}
}

type node map[interface{}]interface{}

func nodes(c context, v *validator) {
	var n node
	if err := yaml.Unmarshal(c.content, &n); err == nil {
		checkNode(n, toNode(config.CloudConfig{}, ""), c, v)
	}
}

func toNode(s interface{}, prefix string) node {
	prefix += " "
	sv := reflect.ValueOf(s)

	if sv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%T is not a struct (%s)", s, sv.Kind()))
	}

	n := make(node)
	for i := 0; i < sv.Type().NumField(); i++ {
		ft := sv.Type().Field(i)
		fv := sv.Field(i)
		k := ft.Tag.Get("yaml")

		if k == "-" || k == "" {
			continue
		}

		switch fv.Kind() {
		case reflect.Struct:
			//fmt.Printf("%s%s: struct\n", prefix, k)
			n[k] = toNode(fv.Interface(), prefix)
		case reflect.Slice:
			et := ft.Type.Elem()

			switch et.Kind() {
			case reflect.Struct:
				//fmt.Printf("%s%s: []struct\n", prefix, k)
				n[k] = []node{toNode(reflect.New(et).Elem().Interface(), prefix)}
			default:
				//fmt.Printf("%s%s: []%s\n", prefix, k, et.Kind())
				//n[k] = reflect.SliceOf(et)
				n[k] = fv.Interface()
				fmt.Printf("%s %T %#v\n", reflect.TypeOf(n[k]).Kind(), n[k], n[k])
			}
		default:
			//fmt.Printf("%s%s: %s\n", prefix, k, fv.Kind())
			n[k] = fv.Interface()
		}
	}
	return n
}

func checkNode(n, g node, c context, v *validator) {
	fmt.Printf("NODE: %#v\n", n)
	fmt.Printf("GOOD: %#v\n", g)
	for k, sn := range n {
		c := c

		for {
			tokens := strings.SplitN(string(c.content), "\n", 2)
			line := tokens[0]
			if len(tokens) > 1 {
				c.content = []byte(tokens[1])
			} else {
				c.content = []byte{}
			}
			c.line++

			if strings.TrimSpace(strings.Split(line, ":")[0]) == fmt.Sprint(k) {
				break
			}
		}

		fmt.Printf("KEY: %s\n", k)
		// Is the key found?
		sc, ok := g[k]
		if !ok {
			v.report.Warning(c.line, fmt.Sprintf("unrecognized key %q", k))
			continue
		}
		if sc == nil {
			panic(fmt.Sprintf("reference node %q is nil", k))
		}

		// If its a struct, we have to go deeper...
		fmt.Printf(" got %q (%T), want %q (%T)\n", sn, sn, sc, sc)
		//nsn, nk := sn.(node)
		nsc, ck := sc.(node)
		nsn, nk := sn.(map[interface{}]interface{})
		//nsc, ck := sc.(map[interface{}]interface{})
		fmt.Printf("%v %v\n", nk, ck)
		//if sn, ok := sn.(map[interface{}]interface{}); ok {
		if nk && ck {
			fmt.Printf(" good: got %#v (%T), want %s (%T)\n", nsn, nsn, nsc, nsc)
			checkNode(nsn, nsc, c, v)
			continue
		}

		// The []string for ssh_authorized_keys is some sort of struct type
		// (validate.node{"ssh_authorized_keys":(*reflect.rtype)}) instead of a slice
		// of strings.
		// ["good"] ([]interface {}) = []interface {}{"good"}

		// Is it the right type?
		fmt.Printf(" bad: got %#v (%T), want %q (%T)\n", sn, sn, sc, sc)
		sct := reflect.TypeOf(sc)
		snt := reflect.TypeOf(sn)
		if sn == nil {
			v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
			continue
		}

		fmt.Printf(" %T %T\n", sn, sc)
		if snt.Kind() == reflect.Slice && sct.Kind() == reflect.Slice {
			fmt.Printf(" SLICE: %v %v\n", snt.Elem(), sct.Elem())
		} else {
			fmt.Printf(" %v %v\n", snt, sct)
			if !snt.ConvertibleTo(sct) {
				v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
			}
		}
	}
}
