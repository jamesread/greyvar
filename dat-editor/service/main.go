package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/jamesread/greyvar/dat-editor/service/buildinfo"
	"github.com/jamesread/greyvar/dat-editor/service/internal/api"
	"github.com/jamesread/greyvar/dat-editor/service/internal/config"
	"github.com/jamesread/greyvar/dat-editor/service/internal/paths"
)

func main() {
	configDir := flag.String("configdir", "", "directory containing config.yaml")
	flag.Parse()
	if *configDir != "" {
		config.SetConfigDir(*configDir)
	}

	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	cfg := config.LoadConfig()
	if path := config.GetConfigPath(); path != "" {
		logrus.Infof("Config loaded from %s", path)
	}

	paths.ApplyDatDir(cfg.DatDir)
	resRoot := paths.ResolveResDir(cfg.ResDir)
	logrus.Infof("Greyvar dat-editor %s listening on %s (dat=%s, res=%s)",
		buildinfo.Version, config.ListenAddr(cfg), paths.DatDirLabel(), resRoot)

	svc := api.New(cfg, resRoot)
	addr := config.ListenAddr(cfg)
	return serveHTTP(addr, api.BuildMux(svc))
}

func serveHTTP(addr string, handler http.Handler) error {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		Protocols:         protocols,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}
