package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/jamesread/greyvar/dat-editor/pkg/editorapi"
	log "github.com/sirupsen/logrus"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	datDir := flag.String("dat", "", "path to dat directory (default: ../server/dat or GREYVAR_DAT_DIR)")
	resDir := flag.String("res", "", "path to res directory (default: ../res or GREYVAR_RES)")
	flag.Parse()

	if *datDir != "" {
		os.Setenv("GREYVAR_DAT_DIR", *datDir)
	}

	if *resDir != "" {
		os.Setenv("GREYVAR_RES", *resDir)
	}

	resRoot := editorapi.ResolveResDir(*resDir)
	log.Infof("dat-editor listening on %s (dat=%s, res=%s)", *addr, editorapi.DatDirLabel(), resRoot)

	handler := editorapi.NewServer(resRoot)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
