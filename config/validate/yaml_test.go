package validate

import (
	"reflect"
	"testing"
)

func TestSyntax(t *testing.T) {
	for _, tt := range []struct {
		c string
		e []Entry
	}{
		{},
		{
			c: "	",
			e: []Entry{{entryError, "found character that cannot start any token", 1}},
		},
		{
			c: "a:\na",
			e: []Entry{{entryError, "could not find expected ':'", 2}},
		},
	} {
		v := validator{report: &Report{}}
		syntax(context{content: []byte(tt.c)}, &v)

		if e := v.report.Entries(); !reflect.DeepEqual(tt.e, e) {
			t.Fatalf("bad report (%s): want %#v, got %#v", tt.c, tt.e, e)
		}
	}
}

func TestNodes(t *testing.T) {
	for _, tt := range []struct {
		c string
		e []Entry
	}{
		{},
		{
			c: "test:",
			e: []Entry{{entryWarning, "unrecognized key \"test\"", 1}},
		},
		{
			c: "hostname:",
			e: []Entry{{entryWarning, "incorrect type for \"hostname\" (want string)", 1}},
		},
		{
			c: "hostname:\n  - bad",
			e: []Entry{{entryWarning, "incorrect type for \"hostname\" (want string)", 1}},
		},
		{
			c: "hostname: 4",
		},
		{
			c: "coreos:\n  etcd:\n    discover:",
			e: []Entry{{entryWarning, "unrecognized key \"discover\"", 3}},
		},
		{
			c: "coreos:\n  etcd:\n    discovery: good",
		},
		{
			c: "ssh_authorized_keys:\n  bad",
			e: []Entry{{entryWarning, "incorrect type for \"ssh_authorized_keys\" (want []string)", 1}},
		},
		{
			c: "ssh_authorized_keys:\n  - good\n  - 2",
		},
		{
			c: "ssh_authorized_keys:\n  - good",
		},
		{
			c: "users:\n  bad",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want []struct)", 1}},
		},
		{
			c: "users:\n  - bad",
			e: []Entry{{entryWarning, "incorrect type for \"users\" (want struct)", 1}},
		},
		{
			c: "users:\n  - name: good",
		},
		{
			c: "coreos:\n  units:\n    - bad",
			e: []Entry{{entryWarning, "incorrect type for \"units\" (want struct)", 2}},
		},
		{
			c: "coreos:\n  units:\n    - name:\n      - bad",
			e: []Entry{{entryWarning, "incorrect type for \"name\" (want string)", 3}},
		},
		{
			c: "coreos:\n  units:\n    - enable: bad",
			e: []Entry{{entryWarning, "incorrect type for \"enable\" (want bool)", 3}},
		},
		{
			c: "coreos:\n  units:\n    - name: test.service\n    - enable: true",
		},
	} {
		v := validator{report: &Report{}}
		nodes(context{content: []byte(tt.c)}, &v)

		if e := v.report.Entries(); !reflect.DeepEqual(tt.e, e) {
			t.Fatalf("bad report (%s): want %#v, got %#v", tt.c, tt.e, e)
		}
	}
}
