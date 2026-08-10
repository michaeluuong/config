// * Configger defines the objects that are stored in ConfigMan.
package config

// Had to rename this file because configger.go is stuck in cache hell!

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/michaeluuong/utilize/jsoning"
	"github.com/michaeluuong/utilize/reflections"
)

// Configger defines the methods required for a basic config.
type Configger interface {
	// SetupConfig is the setup routine for this Configger.
	SetupConfig() error
	// IsSetup returns true if SetupConfig has been called.
	IsSetup() bool
	// PanicOnErr true caller should panic
	PanicOnErr(err error) error
	// loadFromConfigFile loads the parent/embedding object with configurations from a JSON config file.
	loadFromConfigFile() error
	// setWithEnvValues overwrites any configs with values from environment variables.
	setWithEnvValues() error
}

// DefaultConfigger provides a default implementationfor Configger except for the Setup() method.
// This way the implementation is forced to set itself up (i.e. the parent is mandatory).
type DefaultConfigger struct {
	configFile            string `json:"-"` // The full path to the JSON config file
	DefaultConfigFilename string `json:"-"` // The default name of the config file (no path)
	DefaultParentValues   any    `json:"-"` // Default values for the parent
	DirName               string `json:"-"` // The parent directory of the config file
	EnvVar                string `json:"-"` // The env var for the config file
	isSetup               bool   `json:"-"` // true if this DefaultConfigger has been setup
	parent                any    `json:"-"` // The parent (that is composed of DefaultLogger)
	IsPanic               bool   `json:"-"`
}

/*func (d *DefaultConfigger) String() string {
	return ""
}*/

// DefaultSetupConfig is the default routine for setting up Configger and sets IsSetup to true.
// If a config file cannot be found and DefaultParentValues is populated, create the config file with DefaultParentValues.
// Override this for a custom routine.
//   - parentOpt is the parent object. It is required here but optional if being overridden in the parent
//
// Return error
//   - parentOpt is missing
//   - problem loading the parent from the JSON config file
//   - problem loading environment variables
//   - the config file doesn't exist
func (d *DefaultConfigger) DefaultSetupConfig(parent Configger) error {
	if parent == nil {
		return errors.New("parent is required")

	}

	cFile, err := FindConfigFile(d.EnvVar, d.DefaultConfigFilename, d.DirName)
	slog.Info("FindConfigFile()", "EnvVar", d.EnvVar, "DefaultConfigFilename", d.DefaultConfigFilename, "cFile", cFile, "err", err)
	if cFile != "" {
		d.SetDefaults(DefaultConfigger{
			parent:     parent,
			configFile: cFile,
		})

	}

	if err != nil {
		slog.Error("FindConfigFile()|could not find a config file for configger", "err", err)
		if cFile != "" {
			var configValues any
			if d.DefaultParentValues != nil {
				slog.Info("configValues", "d.DefaultParentValues", d.DefaultParentValues)
				configValues = d.DefaultParentValues

			} else {
				slog.Info("No parent values")
				configValues = d.parent

			}

			if err = d.SaveConfigFile(configValues); err != nil {
				slog.Error("SaveConfigFile()|could not save config file", "err", err)
				return err

			}
			slog.Info("CopyStruct", "parent", d.parent)
			reflections.CopyStruct(configValues, d.parent)

			err = nil

		}

		return err

	}

	if err := d.loadFromConfigFile(); err != nil {
		slog.Error("loadFromConfigFile()|could not load config file into configger", "err", err)
		return err

	}

	if err := d.setWithEnvValues(); err != nil {
		slog.Error("setWithEnvValues()|could not load environment variables into configger", "err", err)
		return err

	}

	d.isSetup = true

	return nil

}

// ConfigFile returns the full path to the config file.
func (d *DefaultConfigger) ConfigFile() string {
	return d.configFile

}

// SetConfigFile stores the configuration file.
//   - cFile the full path to the configuration file (only sets if non-empty)
func (d *DefaultConfigger) SetConfigFile(cFile string) {
	if cFile != "" {
		d.configFile = cFile

	}

}

// IsSetup is true if the SetupConfig method has been run successfully.
func (d *DefaultConfigger) IsSetup() bool {
	return d.isSetup

}

// SetIsSetup sets the value of isSetup.
func (d *DefaultConfigger) SetIsSetup(isSetup bool) {
	d.isSetup = isSetup

}

// PanicOnErr soils the sleeper upon the err of my ways is IsPanic is true.
//   - err is the error to display
//
// Return error if IsPanic is false.
func (d *DefaultConfigger) PanicOnErr(err error) error {
	if d.IsPanic {
		panic(err)

	} else {
		return err

	}

}

// SaveConfigFile writes the object to the config file. If the object is nil the parent will be saved.
//   - obj is the object to print; prints parent if nil
func (d *DefaultConfigger) SaveConfigFile(obj any) error {
	var thisObj any = obj
	if obj == nil {
		thisObj = d.parent

	}

	slog.Debug("WriteObjToJSONFile()|writing JSON file", "configFile", d.configFile, "obj", obj)
	return jsoning.NewJsonify().WriteObjToJSONFile(d.configFile, thisObj)

}

// SetDefaults sets up default values.
//   - dcObj is a DefaultConfigger object with values to fill this DefaultConfigger object
func (d *DefaultConfigger) SetDefaults(dcObj DefaultConfigger) {
	if dcObj != (DefaultConfigger{}) {
		d.SetConfigFile(dcObj.ConfigFile())

		if envVarChk := dcObj.EnvVar; envVarChk != "" {
			d.EnvVar = envVarChk

		}

		if defaultConfigFilenameChk := dcObj.DefaultConfigFilename; defaultConfigFilenameChk != "" {
			d.DefaultConfigFilename = defaultConfigFilenameChk

		}

		if defaultParentValuesChk := dcObj.DefaultParentValues; defaultParentValuesChk != nil {
			d.DefaultParentValues = defaultParentValuesChk

		}

		if dirNameChk := dcObj.DirName; dirNameChk != "" {
			d.DirName = dirNameChk

		}

		if parentChk := dcObj.parent; parentChk != nil {
			d.parent = parentChk

		}

	}

}

// loadFromJSONFile unmarshals the JSON object in configFile to the parent.
//   - jsonFilename is the file containing the JSON config object
//
// Return error if
//   - Unable to open file
//   - Unable to read file
//   - Unable to unmarshal JSON object
func (d *DefaultConfigger) loadFromConfigFile() error {
	var parentAny any = d.parent
	return jsoning.NewJsonify().Unmarshal2Struct(d.configFile, &parentAny)

}

// envVars gets enviroment variables that correspond to a json tag in the parent.
//
// Return a map keyed by the JSON tag (that corresponds to the environment variable) and valued by the env value.
func (d *DefaultConfigger) envVars() map[string]string {
	// Get fields and tags from Config
	_, jsonTagToName := reflections.FieldAndTagNames(d.parent, "json")

	envVars := map[string]string{}
	// Get configs that exist in Config from the environment
	for k, v := range jsonTagToName {
		kUpper := strings.ToUpper(k)
		env := os.Getenv(kUpper)
		if env != "" {
			envVars[v] = env

		}

	}

	return envVars

}

// setWithEnvValues overwrites the parent's values (from the config file) with values from environment variables.
//
// Return error if the value of the environment variable is not convertible to the struct's field type.
func (d *DefaultConfigger) setWithEnvValues() error {
	val := reflect.ValueOf(d.parent).Elem()
	envVarsToValues := d.envVars()
	for k, v := range envVarsToValues {
		//slog.Info("envVars", "k", k, "v", v)
		field := val.FieldByName(k)
		err := reflections.SetStructFieldByType(field, v)
		if err != nil {
			return err

		}

	}

	return nil

}
