package main

import (
	"os/user"
	"path/filepath"

	"io/ioutil"

	"github.com/v-braun/go-must"
	yaml "gopkg.in/yaml.v2"
)

type Conf struct {
	Stacks []struct {
		Title string
		Path  string
		Start bool
	}
}

func configLocation() string {
	usr, _ := user.Current()
	home := usr.HomeDir
	cfgPath := filepath.Join(home, ".compose-compose.yml")
	return cfgPath
}

func writeConf(conf *Conf, path string) {
	content, err := yaml.Marshal(conf)
	must.NoError(err, "could not generate yml")

	err = ioutil.WriteFile(path, []byte(content), 0644)
	must.NoError(err, "could save config file")
}

func createDefaultConf(path string) {
	cfg := new(Conf)
	cfg.Stacks = []struct {
		Title string
		Path  string
		Start bool
	}{
		{"my awesome stack", "/path/to/the/stack/folder", false},
	}

	writeConf(cfg, path)
}

func LoadOrCreateConf() *Conf {
	path := configLocation()

	if !pathExists(path) {
		createDefaultConf(path)
	}

	data, err := ioutil.ReadFile(path)
	must.NoError(err, "could not read config file")

	conf := new(Conf)
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		panic(err)
	}
	must.NoError(err, "could not parse yml config")

	return conf
}
