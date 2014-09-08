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
	var n struct{}
	if err := goyaml.Unmarshal(c.content, &n); err == nil {
		fmt.Printf("%#v\n", n)
		checkStructure(n, config.CloudConfig{}, c, v)
	}
}

func checkStructure(n, c interface{}, ctx context, val *validator) {
	//ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(c)
	nt := reflect.TypeOf(n)
	nv := reflect.ValueOf(n)

	for i := 0; i < nt.NumField(); i++ {
		k := nt.Field(i).Name
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

		if nt.Kind() != reflect.Struct {
			val.report.Warning(ctx.line, fmt.Sprintf("unrecognized key %q", k))
		}
		if _, ok := nt.FieldByName(k); ok {
				sn := nv.FieldByName(k).Interface()
				sc := cv.FieldByName(k).Interface()
				checkStructure(sn, sc, ctx, val)
		} else {
			val.report.Warning(ctx.line, fmt.Sprintf("unrecognized key %q", k))
		}
	}
}
