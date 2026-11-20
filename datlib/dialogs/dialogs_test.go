package dialogs

import (
	"strings"
	"testing"
)

func TestBillsDialogIsValid(t *testing.T) {
	d, err := ReadFile("../../server/dat/dialogs/bill.yml")
	if err != nil {
		t.Fatal(err)
	}
	if d.Speaker != "Bill" || d.Nodes["seeds"].Give != "hay_seeds:5" {
		t.Fatalf("Bill: %+v", d)
	}
}

func TestValidateCatchesBrokenDialogs(t *testing.T) {
	d := &Dialog{Start: "nope", Nodes: map[string]*Node{
		"a": {Text: "hi", Options: []Option{{Label: "go", Goto: "missing"}, {Goto: "a"}}},
		"b": {Text: "dead end"},
	}}
	got := strings.Join(d.Validate(), "\n")
	for _, want := range []string{`start node "nope"`, `unknown node "missing"`, "without a label", `node "b" needs at least one option`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}
