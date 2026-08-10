// Package config caches configuration objects per Configger implementation.
//
//	Configurations can be loaded from
//	 1. environment variables (override and match the json tag for the config value, except in uppercase)
//	 2. config files
//
// * ConfigMan caches Configger objects.
package config

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/reflections"
)

// ConfigMan caches Configger objects per type.
type ConfigMan struct {
	// typeToConfig stores each configure object by its type.
	typeToConfig map[reflect.Type]Configger
	// isLoggerSet should be set to true if logging has been configured and cached.
	isLoggerSet bool
}

// AddConfig runs the configs Setup() method and stores it by its type.
//   - configger is the concrete type to add to cache
//
// Return error from configger.SetupConfig()
// DefaultSetupConfig()
//   - parent is nil
//   - can't find config file
//   - have the name of a config file but can't save default values to it
//   - can't load config file
//   - can't load environment variables
func (c *ConfigMan) AddConfig(configger Configger) error {
	if !configger.IsSetup() {
		if err := configger.SetupConfig(); err != nil {
			return configger.PanicOnErr(err)

		}

	}

	t := reflections.AnyType(configger)
	c.typeToConfig[t] = configger

	return nil

}

// Config gets a config by type from cache.
//   - configger is the concrete type to get
//
// Return the config as an object of any.
func (c *ConfigMan) Config(configger Configger) any {
	t := reflections.AnyType(configger)
	return c.typeToConfig[t]

}

// ConfigPtr gets a config by type from cache.
//   - configger is the concrete type to get
//
// Return the config and an object of *any.
func (c *ConfigMan) ConfigPtr(configger Configger) *any {
	t := reflections.AnyType(configger)
	a := any(c.typeToConfig[t])
	return &a

}

// ConfigCfgr gets a config by type from cache.
//   - configger is the concrete type to get
//
// Return the config as
/*func (c *ConfigMan) ConfigCfgr(configger Configger) Configger {
	t := reflections.AnyType(configger)
	return c.typeToConfig[t]

}*/

// SetIsLoggerSet flags that logging has been setup.
//   - isSet should be set to true if logging has been configured and set
func (c *ConfigMan) SetIsLoggerSet(isSet bool) {
	c.isLoggerSet = isSet

}

// IsLoggerSet is a convenience method that returns true if the logging mechanism has been setup.
//
// Return true if loggings has been setup otherwise false.
func (c *ConfigMan) IsLoggerSet() bool {
	return c.isLoggerSet

}

// Singleton
var (
	configManInstance *ConfigMan
	configManMut      sync.Mutex
)

// GetConfigManInstance creates and/or returns the single instance of a configuration manager.
func GetConfigManInstance() *ConfigMan {
	configManMut.Lock()
	defer configManMut.Unlock()

	if configManInstance == nil {
		configManInstance = &ConfigMan{typeToConfig: map[reflect.Type]Configger{}}

	}

	return configManInstance

}

// resetConfigManInstance resets the configuration manager and returns the new instance.
//
// Return the only instance of the ConfigMan object.
func ResetConfigManInstance() *ConfigMan {
	configManMut.Lock()
	configManInstance = nil
	slog.Info("reset ConfigMan instance", "configManInstance", configManInstance)
	configManMut.Unlock()

	return GetConfigManInstance()

}

// AssertConfigger assert an object of type *any to T.
//   - anyConfgger is the object to assert to T
//
// Return
//   - anyConfigger type asserted to T
//   - true if anyConfigger was asserted to type *T else false
func AssertConfigger[T any](anyConfigger *any) (*T, bool) {
	v, ok := (*anyConfigger).(*T)
	return v, ok

}

// FindConfigFile tries to find an existing config file in order of env variable, user config, user home or working directory.
//   - envVar is the environment variable to look for first
//   - cFilename is the name of the configuration file (basename)
//   - parentDirs are the directories immediately enclosing the cFile
//
// Return the first config file that exists or else the file from os.UserConfigDir()
//  1. enviroment variable
//  2. os.UserConfigDir() + parentDirs + config + cFilename
//  3. os.UserHomeDir() + parentDirs + config + cFilename
//  4. os.Getwd() + parentDirs + config + cFilename
func FindConfigFile(envVar string, cFilename string, parentDirs ...string) (configFile string, err error) {
	if envVar == "" && cFilename == "" {
		return "", errors.New("envVar or cFilename must be specified")

	}

	const pSep = filing.PathSep

	parentDir := ""
	if len(parentDirs) > 0 && parentDirs[0] != "" {
		parentDir = strings.Join(parentDirs, pSep) + pSep

	}

	envVal := os.Getenv(envVar)
	userConfigDir, _ := os.UserConfigDir()
	userConfigFile := userConfigDir + pSep + parentDir + "config" + pSep + cFilename

	if envVal != "" {
		if filing.Exists(envVal) {
			slog.Info("found config file", "envVar", envVar, "envVal", envVal)
			configFile = envVal

		} else {
			slog.Error("envVal != \"\" &&!filing.Exists()|could not find config file", "envVar", envVar, "envVal", envVal)

		}

	} else if cFilename != "" {
		slog.Debug("cFilename!=\"\"|envVar was not provided or the file does not exist", "envVar", envVar)

		if filing.Exists(userConfigFile) {
			slog.Info("found config file from user config directory", "userConfigFile", userConfigFile)
			configFile = userConfigFile

		} else {
			slog.Debug("cFilename != \"\" && !filing.Exists()|config file was not found in user config directory", "userConfigDir", userConfigDir)
			userHomeDir, _ := os.UserHomeDir()
			userHomeFile := userHomeDir + pSep + cFilename
			if filing.Exists(userHomeFile) {
				slog.Info("found config file from user home directory", "userHomeFile", userHomeFile)
				configFile = userHomeFile

			} else {
				slog.Debug("cFilename != \"\" && !filing.Exists()|config file was not found in user home directory", "userHomeDir", userHomeDir)
				workDir, _ := os.Getwd()
				workFile := workDir + pSep + cFilename
				if filing.Exists(workFile) {
					slog.Info("found config file from work directory", "workFile", workFile)
					configFile = workFile

				}

			}

		}

	}

	if configFile == "" {
		if envVal != "" {
			return envVal, errors.New("could not find a config file")

		} else {
			return userConfigFile, errors.New("could not find a config file")

		}

	}

	return configFile, nil

}
