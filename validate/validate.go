package validate

import (
	"strings"
	"fmt"
)

type validator struct {
	report Reporter
	tests []test
}

func Validate(config []byte) (Reporter, error) {
	v := &validator{&Report{}, []test{{ruleContext{config, 0}, baseRule}}}

	for len(v.tests) > 0 {
		t := v.tests[0]
		v.tests = v.tests[1:]

		if err := func(t test, v *validator) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%s", r)
				}
			}()
			t.rule(t.context, v)
			return nil
		}(t, v); err != nil {
			return v.report, err
		}
	}

	return v.report, nil
}


func baseRule(c ruleContext, v *validator) {
	header := strings.SplitN(string(c.content), "\n", 2)[0]
	if header == "#cloud-config" {

	} else if strings.HasPrefix("#!", header) {

	} else {
		v.report.Error(c.currentLine + 1, "must be \"#cloud-config\" or \"#!\"")
	}
}


