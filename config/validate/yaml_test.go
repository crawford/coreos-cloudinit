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
			e: []Entry{{1, "found character that cannot start any token", errorEntry}},
		},
		{
			c: "a:\na",
			e: []Entry{{2, "could not find expected ':'", errorEntry}},
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
			e: []Entry{{1, "unrecognized key \"test\"", warningEntry}},
		},
		{
			c: "hostname:",
			e: []Entry{{1, "incorrect type for \"hostname\" (want string)", warningEntry}},
		},
		{
			c: "coreos:\n  etcd:\n    discover:",
			e: []Entry{{3, "unrecognized key \"discover\"", warningEntry}},
		},
		{
			c: "coreos:\n  etcd:\n    discovery: good",
		},
		{
			c: "ssh_authorized_keys:\n  bad",
			e: []Entry{{1, "incorrect type for \"ssh_authorized_keys\" (want []string)", warningEntry}},
		},
		{
			c: "ssh_authorized_keys:\n  - good",
		},
		{
			c: "users:\n  - bad",
			e: []Entry{{1, "incorrect type for \"users\" (want []system.User)", warningEntry}},
		},
		{
			c: "users:\n  - name: good",
		},
	} {
		v := validator{report: &Report{}}
		nodes(context{content: []byte(tt.c)}, &v)

		if e := v.report.Entries(); !reflect.DeepEqual(tt.e, e) {
			t.Fatalf("bad report (%s): want %#v, got %#v", tt.c, tt.e, e)
		}
	}
}
