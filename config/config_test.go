package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadWithDefault(t *testing.T) {
	os.Args = append(os.Args, "-f", "./config.json")
	flag.Parse()
	dft := []byte(`
	{
		"id": 1000,
		"name": "ava-test",
		"city": "Shanghai"
	}
	`)

	var conf struct {
		ID    int    `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Phone string `json:"phone,omitempty"`
		City  string `json:"city,omitempty"`
	}

	e := LoadWithDefault(&conf, dft)
	assert.NoError(t, e)
	assert.Equal(t, 1000, conf.ID)         // same in default and file
	assert.Equal(t, "ava-test", conf.Name) // empty in file, use default value
	assert.Equal(t, "1234", conf.Phone)    // empty default, use value in file
	assert.Equal(t, "Suzhou", conf.City)   // override by file
}
