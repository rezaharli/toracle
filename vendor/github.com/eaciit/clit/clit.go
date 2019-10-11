package clit

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eaciit/appconfig"
	"github.com/eaciit/toolkit"
)

var (
	EnableConfig = true
	EnableLog    = true
	Log          *toolkit.LogEngine

	preFn   func() error
	closeFn func() error
	flags   = map[string]*string{}

	configs     map[string]*appconfig.Config
	configpaths = map[string]string{}

	exeDir string

	isParsed = false
)

func init() {
	SetFlag("config", "", "Location of the config file")
}

func SetFlag(name, value, usage string) *string {
	f := flag.String(name, value, usage)
	flags[name] = f
	return f
}

func SetPreFn(fn func() error) {
	preFn = fn
}

func SetCloseFn(fn func() error) {
	closeFn = fn
}

func Flag(name string) string {
	f := flags[name]
	return *f
}

func Parse() {
	if !isParsed {
		flag.Parse()
		isParsed = true
	}
}

func LoadConfigFromFlag(flagName, configName, defaultPath string) {
	if flagName == "" {
		flagName = "config"
	}

	if configName == "" {
		configName = "default"
	}

	Parse()
	if EnableConfig {
		config := Flag(flagName)
		if config == "" {
			AddConfig(configName, defaultPath)
		} else {
			AddConfig(configName, config)
		}
	}
}

func Commit() error {
	var err error
	Parse()

	if EnableConfig {
		for n, f := range configpaths {
			if err = ReadConfig(n, f); err != nil {
				return fmt.Errorf("error reading config file %s. %s", f, err.Error())
			}
		}
	}

	if EnableLog && Log == nil {
		if Log, err = toolkit.NewLog(true, false, "", "", ""); err != nil {
			return fmt.Errorf("error preparing log. %s", err.Error())
		}
	}

	if preFn != nil {
		preFn()
	}

	return nil
}

func ExeDir() string {
	if exeDir == "" {
		exeDir, _ = os.Executable()
		exeDir = filepath.Dir(exeDir)
	}
	return exeDir
}

func AddConfig(name, path string) {
	if name == "" {
		name = "default"
	}
	configpaths[name] = path
}

func ReadConfig(name, path string) error {
	if name == "" {
		name = "default"
	}

	if name == "default" && path == "" {
		path := Flag("config")
		if path == "" {
			path = filepath.Join(ExeDir(), "app.json")
		}
	} else if path == "" {
		return errors.New("path can't be empty")
	}

	initConfigs()
	config := new(appconfig.Config)
	if err := config.SetConfigFile(path); err != nil {
		return err
	}
	fmt.Printf("Read config: %s\n", path)

	configs[name] = config
	configpaths[name] = path
	return nil
}

func Config(name, key string, def interface{}) interface{} {
	initConfigs()
	if name == "" {
		name = "default"
	}
	config, found := configs[name]
	if !found {
		return def
	}
	v := config.GetDefault(key, def)
	return v
}

func SetConfig(name, key string, value interface{}) {
	initConfigs()
	if name == "" {
		name = "default"
	}
	config, found := configs[name]
	if !found {
		return
	}
	config.Set(key, value)
}

func WriteConfig(name string) error {
	initConfigs()
	if name == "" {
		name = "default"
	}
	config, found := configs[name]
	if !found {
		return fmt.Errorf("can not write config. config %s is not yet initialized", name)
	}
	return config.Write()
}

func initConfigs() {
	if configs == nil {
		configs = map[string]*appconfig.Config{}
	}
}

func Value(name, configgroup string, def string) string {
	flagvalue := Flag(name)
	if flagvalue != "" {
		return flagvalue
	}

	cfgvalue := Config(configgroup, name, def)
	return toolkit.ToString(cfgvalue)
}

func Close() {
	if closeFn != nil {
		closeFn()
	}
}
