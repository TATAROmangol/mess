package tokenissuer

import (
	"flag"
	"os"
	"testing"
)

var CFG *Config

func TestMain(m *testing.M) {
	var configPath = flag.String("config-path", "", "path to config")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}

	CFG = cfg
	os.Exit(m.Run())
}
