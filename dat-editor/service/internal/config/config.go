package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	ConfigVersion int    `yaml:"configVersion"`
	Listen        string `yaml:"listen"`
	SiteTitle     string `yaml:"siteTitle"`
	DatDir        string `yaml:"datDir"`
	ResDir        string `yaml:"resDir"`
}

var configDirOverride string

func SetConfigDir(dir string) {
	configDirOverride = dir
}

func GetConfigPath() string {
	if configDirOverride != "" {
		return existingPath(filepath.Join(configDirOverride, "config.yaml"))
	}
	return firstExistingPath([]string{
		"./config.yaml",
		"./config/config.yaml",
		os.Getenv("DEDITOR_CONFIG_FILE"),
	})
}

func firstExistingPath(candidates []string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if found := existingPath(p); found != "" {
			return found
		}
	}
	return ""
}

func existingPath(p string) string {
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

func Defaults() *Config {
	return &Config{
		ConfigVersion: 1,
		Listen:        ":8080",
		SiteTitle:     "Greyvar Dat Editor",
	}
}

func LoadConfig() *Config {
	cfg := Defaults()
	path := GetConfigPath()
	if path != "" {
		loadConfigFile(cfg, path)
	}
	applyEnvOverrides(cfg)
	applyFallbacks(cfg)
	return cfg
}

func loadConfigFile(cfg *Config, path string) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return
	}
	_ = k.Unmarshal("", cfg)
}

func applyFallbacks(cfg *Config) {
	if cfg.Listen == "" {
		cfg.Listen = ":8080"
	}
	if cfg.SiteTitle == "" {
		cfg.SiteTitle = "Greyvar Dat Editor"
	}
}

func applyEnvOverrides(cfg *Config) {
	setIfEnv(&cfg.DatDir, "GREYVAR_DAT_DIR")
	setIfEnv(&cfg.ResDir, "GREYVAR_RES")
	setIfEnv(&cfg.SiteTitle, "SITE_TITLE")
}

func setIfEnv(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func ListenAddr(cfg *Config) string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return normalizePort(port)
	}
	if cfg != nil && cfg.Listen != "" {
		return cfg.Listen
	}
	return ":8080"
}

func normalizePort(port string) string {
	if strings.Contains(port, ":") {
		return port
	}
	return ":" + port
}
