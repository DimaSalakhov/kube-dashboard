package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Monitor        monitorConfig `json:"monitor"`
	KubeconfigPath string        `json:"kubeconfigPath"`
}

type monitorConfig struct {
	Contexts []string `json:"contexts"`
}

func MustParseConfig() config {
	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func parseConfig() (config, error) {
	var (
		cfg        config
		configFile string
	)

	flag.StringVar(&configFile, "config", "./config.json", "config file")
	flag.Parse()

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, errors.Wrap(err, "failed to read config file")
	}

	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return cfg, errors.Wrap(err, "failed to unmarshal config")
	}

	if cfg.KubeconfigPath == "" {
		dir, _ := os.UserHomeDir()
		cfg.KubeconfigPath = filepath.Join(dir, ".kube", "config")
	}

	return cfg, nil
}
