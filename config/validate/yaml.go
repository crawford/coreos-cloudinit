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
	yamlKey       = regexp.MustCompile(`^ *-? ?(?P<key>.*):`)
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
				//	fmt.Printf("%s %T %#v\n", reflect.TypeOf(n[k]).Kind(), n[k], n[k])
			}
		default:
			//fmt.Printf("%s%s: %s\n", prefix, k, fv.Kind())
			n[k] = fv.Interface()
		}
	}
	return n
}

func checkNode(n, g node, c context, v *validator) {
	//	fmt.Printf("NODE: %#v\n", n)
	//	fmt.Printf("GOOD: %#v\n", g)
	//fields:
	for k, sn := range n {
		c := c

		fmt.Printf("KEY: %s\n", k)
		fmt.Printf("CONTENT: %q\n", c.content)
		{
			for {
				tokens := strings.SplitN(string(c.content), "\n", 2)
				line := tokens[0]
				fmt.Printf(" %s\n", line)
				if line == "" {
					panic(fmt.Sprintf("key %q not found in content", k))
				}

				if len(tokens) > 1 {
					c.content = []byte(tokens[1])
				} else {
					c.content = []byte{}
				}
				c.line++

				matches := yamlKey.FindStringSubmatch(line)
				if len(matches) > 0 && matches[1] == fmt.Sprint(k) {
					break
				}
			}
		}

		// Is the key found?
		sg, ok := g[k]
		if !ok {
			v.report.Warning(c.line, fmt.Sprintf("unrecognized key %q", k))
			continue
		}
		if sg == nil {
			panic(fmt.Sprintf("reference node %q is nil", k))
		}

		if sg, ok := sg.([]node); ok {
			fmt.Printf(" WANT []NODE:\n")
			if sn, ok := sn.([]interface{}); ok {
				fmt.Printf("  GOT []INTERFACE{}: %q (%T)\n", sn, sn)
				ssg := sg[0]
				for _, ssn := range sn {
					fmt.Printf("   GOT INTERFACE: %q (%T)\n", ssn, ssn)
					fmt.Printf("   WANT: %q (%T)\n", ssg, ssg)
					if ssn, ok := ssn.(map[interface{}]interface{}); ok {
						fmt.Printf("   CHECKING: %q   ?=   %q\n", ssn, ssg)
						checkNode(ssn, ssg, c, v)
					} else {
						v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sg))
						continue
					}
				}
				continue
			} else {
				v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sg))
				continue
			}
		}

		if sg, ok := sg.(node); ok {
			if sn, ok := sn.(map[interface{}]interface{}); ok {
				checkNode(sn, sg, c, v)
				continue
			} else {
				v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sg))
				continue
			}
		}

		if !isSameType(reflect.ValueOf(sn), reflect.ValueOf(sg)) {
			v.report.Warning(c.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sg))
		}
	}
}

func isSameType(n, g reflect.Value) bool {
	if !n.IsValid() {
		return false
	}
	for n.Kind() == reflect.Interface {
		n = n.Elem()
	}
	fmt.Printf("Types: %s (%s), %s (%s)\n", n.Type(), n.Kind(), g.Type(), g.Kind())

	//fmt.Printf(" bad: got %#v (%T), want %q (%T)\n", sn, sn, sg, sg)
	switch g.Kind() {
	case reflect.String:
		fmt.Println(" STRING")
		fmt.Printf("  %#v (%s)\n", n.Interface(), n.Kind())
		return n.Kind() == reflect.String

	case reflect.Slice:
		fmt.Println(" SLICE")
		if n.Kind() != reflect.Slice {
			return false
		}

		elem := func(v reflect.Value) reflect.Value {
			if v.Len() > 0 {
				return v.Index(0)
			}
			return reflect.Indirect(reflect.New(v.Type().Elem()))
		}
		fmt.Printf("  %s (%s) -> %s (%s)\n", n.Type(), n.Kind(), elem(n).Type(), elem(n).Kind())
		fmt.Printf("  %s (%s) -> %s (%s)\n", g.Type(), g.Kind(), elem(g).Type(), elem(g).Kind())

		sg := reflect.Indirect(reflect.New(g.Type().Elem()))
		for i := 0; i < n.Len(); i++ {
			if !isSameType(n.Index(i), sg) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("unhandled kind %s", g.Kind()))
	}
}
