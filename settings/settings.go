package settings

import (
	"fmt"
	"io/ioutil"

	"github.com/eljuanchosf/gocafier/logging"

	"gopkg.in/yaml.v2"
)

var Values Config

//Config represents the config structure for the package
type Config struct {
	Email struct {
		Body    string      `yaml:"body"`
		Cc      interface{} `yaml:"cc"`
		From    string      `yaml:"from"`
		Subject string      `yaml:"subject"`
		To      string      `yaml:"to"`
	} `yaml:"email"`
	Packages []string `yaml:"packages"`
	SMTP     struct {
		Port   int    `yaml:"port"`
		Server string `yaml:"server"`
	} `yaml:"smtp"`
}

//LoadConfig reads the specified config file
func LoadConfig(filename string) {
	if filename == "." {
		filename = "config.yml"
	}

	logging.LogStd(fmt.Sprintf("Loading config from %s", filename), true)

	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &Values)
	if err != nil {
		panic(err)
	}
}
