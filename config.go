package main

import (
	"flag"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

var configInfo *Config

type (
	Config struct {
		Host    string
		Port    int64
		Path    string
		Metrics MetricsConfig
	}

	MetricConfig struct {
		Name        string
		Description string
		Type        string
		Source      string
	}

	MetricsConfig []*MetricConfig
)

func getConfig() *Config {
	return configInfo
}

func parseConfig(configfile string) {
	file, err := ioutil.ReadFile(configfile)

	if err != nil {
		log.Fatalf("parse configfile error, %v\n", err)
	}

	configInfo = &Config{}
	if err = yaml.Unmarshal(file, configInfo); err != nil {
		log.Fatalf("parse configfile error, %v\n", err)
	}
}

func init() {
	configFile := flag.String("c", "config.yml", "the path of config file")
	flag.Parse()

	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	parseConfig(*configFile)
}
