package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

var filePath string

func init() {
	flag.StringVar(&filePath, "f", "./config", `config file path, default to "./config"`)
}

// LoadConfig loads a json config file located by commandline flag '-config'
func LoadConfig(v interface{}) error {
	buf, e := ioutil.ReadFile(filePath)
	if e != nil {
		return errors.Wrap(e, "read config file failed, did you set the right path?")
	}
	if e := json.Unmarshal(buf, v); e != nil {
		return errors.Wrap(e, "unmarshal config file failed, please check the file path and content")
	}
	return nil
}

func LoadWithDefault(v interface{}, cfgDefault []byte) error {
	vFile := reflect.New(reflect.TypeOf(v).Elem()).Interface()
	if e := LoadConfig(vFile); e != nil {
		return e
	}

	if e := json.Unmarshal(cfgDefault, v); e != nil {
		return errors.Wrap(e, "unmarshal default config failed")
	}

	return mergo.MergeWithOverwrite(v, vFile)
}
