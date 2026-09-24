package main

import (
	"fmt"
	"os"

	"github.com/jamesread/greyvar/datlib/lint"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Infof("datlint")

	datDir, err := lint.DiscoverDatDir()
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	log.Infof("datdir: %v", datDir)

	resDir := lint.ResolveResDir(datDir)
	if resDir == "" {
		log.Warnf("res dir not found (set GREYVAR_RES_DIR); texture existence checks will warn")
	} else {
		log.Infof("resdir: %v", resDir)
	}

	checks := lint.LintAll(lint.Options{DatDir: datDir, ResDir: resDir})
	errorCount, warnCount := 0, 0
	for _, check := range checks {
		errs, warns := printCheck(check)
		errorCount += errs
		warnCount += warns
	}

	if errorCount == 0 && warnCount == 0 {
		log.Infof("Finished. All good.")
		return
	}

	if warnCount > 0 {
		log.Warnf("Warnings: %v", warnCount)
	}
	if errorCount == 0 {
		log.Infof("Finished with warnings only.")
		return
	}

	log.Errorf("Errors!: %v", errorCount)
	os.Exit(1)
}

func printCheck(check lint.Check) (errors, warnings int) {
	prefix := fmt.Sprintf("%-8s %s", check.Type, check.Filename)
	if len(check.Issues) == 0 {
		log.Infof("%s: OK", prefix)
		return 0, 0
	}

	log.Infof("%s:", prefix)
	for _, issue := range check.Issues {
		switch issue.Severity {
		case lint.SeverityWarning:
			log.Warnf("  %s", issue.Issue)
			warnings++
		default:
			log.Errorf("  %s", issue.Issue)
			errors++
		}
	}
	return errors, warnings
}
