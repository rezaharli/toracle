package appconfig

import (
	"encoding/json"
	"fmt"

	"github.com/eaciit/toolkit"
	//"errors"
	//"fmt"

	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/eaciit/toolkit"
)

type Config struct {
	filename string
	isLoaded bool
	configs  map[string]interface{}
}

func (c *Config) Load() error {
	var err error = nil
	filename := c.configFileName()
	if isConfigFileExist(filename) == false {
		err := ioutil.WriteFile(filename, []byte("{}"), 0644)
		if err != nil {
			return err
		}
		c.configs = map[string]interface{}{}
	} else {
		fileName := c.configFileName()
		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
		cfgItems := map[string]interface{}{}
		if err = json.Unmarshal(data, &cfgItems); err != nil {
			return err
		}
		c.configs = cfgItems
	}
	c.isLoaded = true
	return err
}

func (c *Config) Write() error {
	var jsonBytes []byte
	if jsonStr, err := json.MarshalIndent(c.configs, "", "\t"); err != nil {
		return err
	} else {
		jsonBytes = []byte(jsonStr)
	}
	fileName := c.configFileName()
	if err := ioutil.WriteFile(fileName, jsonBytes, 0644); err != nil {
		return err
	}
	return nil
}

func (c *Config) WriteObject(obj interface{}) error {
	maps := map[string]interface{}{}
	if err := toolkit.Serde(obj, &maps, ""); err != nil {
		return fmt.Errorf("unable to serialized object. %s", err.Error())
	}
	c.configs = maps
	return c.Write()
}

func (c *Config) Serde(o interface{}) error {
	if !c.isLoaded {
		return fmt.Errorf("config is not yet laoded")
	}
	return toolkit.Serde(c.configs, o, "")
}

func (c *Config) SetConfigFile(pathtofile string) error {
	if pathtofile == "" {
		pathtofile = c.configFileName()
	}
	c.filename = pathtofile
	return c.Load()
}

func (c *Config) configFileName() string {
	if c.filename == "" {
		c.filename = filepath.Join(PathDefault(false), "config.json")
	}
	return c.filename
}

func isConfigFileExist(loc string) bool {
	fn := loc
	_, err := os.Stat(fn)
	if err != nil {
		return os.IsNotExist(err) == false
	}
	return true
}

func (c *Config) HasKey(id string) bool {
	_, exist := c.configs[id]
	return exist
}

func (c *Config) Get(id string) interface{} {
	if !c.isLoaded {
		c.Load()
	}
	ret, exist := c.configs[id]
	if exist == false {
		ret = ""
	}
	return ret
}

func (c *Config) GetDefault(id string, def interface{}) interface{} {
	if !c.isLoaded {
		c.Load()
	}
	ret, exist := c.configs[id]
	if exist == false {
		ret = def
	}
	return ret
}

func (c *Config) Set(id string, value interface{}) error {
	if c.configs == nil {
		c.configs = make(map[string]interface{})
	}
	c.configs[id] = value
	return nil
}
