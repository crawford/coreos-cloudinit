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
		{
			c: "",
		},
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
		{
			c: `test: a`,
			e: []Entry{{1, "unrecognized key \"test\"", warningEntry}},
		},
		{
			c: `coreos:
  etcd:
    discover:`,
			e: []Entry{{3, "unrecognized key \"discover\"", warningEntry}},
		},
	} {
		v := validator{report: &Report{}}
		nodes(context{content: []byte(tt.c)}, &v)

		if e := v.report.Entries(); !reflect.DeepEqual(tt.e, e) {
			t.Fatalf("bad report (%s): want %#v, got %#v", tt.c, tt.e, e)
		}
	}
}
