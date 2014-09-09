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

func nodes(c context, v *validator) {
	var n map[interface{}]interface{}
	if err := yaml.Unmarshal(c.content, &n); err == nil {
		fmt.Printf("%s\n%#v\n", c.content, n)
		checkStructure(n, config.CloudConfig{}, c, v)
	}
}

func checkStructure(n map[interface{}]interface{}, c interface{}, ctx context, v *validator) {
	ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(c)

	for k, sn := range n {
		ctx := ctx

		for {
			tokens := strings.SplitN(string(ctx.content), "\n", 2)
			line := tokens[0]
			if len(tokens) > 1 {
				ctx.content = []byte(tokens[1])
			} else {
				ctx.content = []byte{}
			}
			ctx.line++

			if strings.TrimSpace(strings.Split(line, ":")[0]) == fmt.Sprint(k) {
				break
			}
		}

		/*if ct.Kind() != reflect.Struct {
			v.report.Warning(ctx.line, fmt.Sprintf("unrecognized key %q", k))
		}*/
		foundIndex := -1
		for i := 0; i < ct.NumField(); i++ {
			if ct.Field(i).Tag.Get("yaml") == k {
				foundIndex = i
				break
			}
		}
		if foundIndex < 0{
			v.report.Warning(ctx.line, fmt.Sprintf("unrecognized key %q", k))
			continue
		}

		sc := cv.Field(foundIndex).Interface()

		if sn == nil {
			v.report.Warning(ctx.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
			continue
		}

		if sn, ok := sn.(map[interface{}]interface{}); ok {
			checkStructure(sn, sc, ctx, v)
		}

		switch reflect.TypeOf(sc).Kind() {
		case reflect.Struct:
		case reflect.Slice:
			if reflect.TypeOf(sn).Kind() != reflect.Slice {
				fmt.Printf("%T\n", sn)
				v.report.Warning(ctx.line, fmt.Sprintf("incorrect type for %q (want %T)", k,  sc))
				break
			}

			switch reflect.TypeOf(sc).Elem().Kind() {
			case reflect.Struct:
				if sn, ok := sn.([]map[interface{}]interface{}); ok {
					checkStructure(sn[0], sc.([]interface{})[0], ctx, v)
				} else {
					fmt.Printf("alex %T\n", sn)
					v.report.Warning(ctx.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
				}
			default:
				if reflect.TypeOf(sn).Elem().AssignableTo(reflect.TypeOf(sc).Elem()) {
					fmt.Printf("%T\n", sn)
					v.report.Warning(ctx.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
				}
			}
		default:
			if reflect.TypeOf(sn) != reflect.TypeOf(sc) {
					fmt.Printf("%T\n", sn)
				v.report.Warning(ctx.line, fmt.Sprintf("incorrect type for %q (want %T)", k, sc))
			}
		}
	}
}
