package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListenAddrPrefersPortEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	cfg := Defaults()
	assert.Equal(t, ":9090", ListenAddr(cfg))
}

func TestListenAddrUsesConfigListen(t *testing.T) {
	t.Setenv("PORT", "")
	cfg := Defaults()
	cfg.Listen = ":8081"
	assert.Equal(t, ":8081", ListenAddr(cfg))
}
