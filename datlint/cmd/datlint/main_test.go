package main

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/lint"
)

func TestPrintCheckOK(t *testing.T) {
	errs, warns := printCheck(lint.Check{Type: "entdef", Filename: "demo.yml"})
	if errs != 0 || warns != 0 {
		t.Fatalf("counts error=%d warn=%d", errs, warns)
	}
}

func TestPrintCheckIssues(t *testing.T) {
	check := lint.Check{Type: "entdef", Filename: "demo.yml"}
	check.AddWarning("soft problem")
	check.AddError("hard problem")
	errs, warns := printCheck(check)
	if errs != 1 || warns != 1 {
		t.Fatalf("counts error=%d warn=%d", errs, warns)
	}
}
