// * SlogConfig is a Configger object for log/slog.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/michaeluuong/utilize/stringy"
)

const slogConfigFilename = "slog_cfg.json"

// SlogLevel is an enum for slog.Level debug levels.
type SlogLevel slog.Level

const (
	SDebug SlogLevel = SlogLevel(slog.LevelDebug)
	SError SlogLevel = SlogLevel(slog.LevelError)
	SInfo  SlogLevel = SlogLevel(slog.LevelInfo) // Default
	SWarn  SlogLevel = SlogLevel(slog.LevelWarn)
)

// String returns SlogLevel in uppercase to match environment variables.
func (s SlogLevel) String() string {
	switch s {
	case SDebug:
		return "DEBUG"
	case SError:
		return "ERROR"
	case SWarn:
		return "WARN"
	}

	return "INFO"

}

// SlogLevelFromString gets the SlogLevel from a string.
//   - slogLevel is the slog.Level as a string (i.e. Debug|Error|Info|Warn)
func (s *SlogLevel) SlogLevelFromString(slogLevel string) {
	slogLevelU := strings.ToUpper(slogLevel)
	switch slogLevelU {
	case SDebug.String():
		*s = SDebug

	case SError.String():
		*s = SError

	case SInfo.String():
		*s = SInfo

	case SWarn.String():
		*s = SWarn

	}

}

// MarshalJSON custom marshaler for SlogLevel.
//
// Return JSON data as a []byte or an error if unable to marshal the object data.
func (s *SlogLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())

}

// UnmarshallJSON custom unmarshaler for SlogLevel.
//   - data is JSON object from file
//
// Return an error if unable to unmarshal the JSON object.
func (s *SlogLevel) UnmarshalJSON(data []byte) error {
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err

	}

	s.SlogLevelFromString(strVal)

	return nil

}

// toLevel returns SlogLevel as slog.Level.
func (s SlogLevel) toLevel() slog.Level {
	return slog.Level(s)

}

// SlogConfig is a Configger object that represents all available configurations for log/slog.
// Composed of DefaultConfigger which implements Configger.
//
//	SlogConfig was exported but was meant to be created with NewSlogConfig().
//	If you have to create it yourself make sure to include DefaultConfigger.
type SlogConfig struct {
	Add_source bool      `json:"add_source"` // true adds filename/method name to the log line
	Dev        bool      `json:"dev"`        // true changes certain behaviors for development (e.g. Log_level is Debug).
	Log_func   bool      `json:"log_func"`   // true allows ReplacAttr to perform replacements (e.g. Add_source)
	Log_json   bool      `json:"log_json"`   // true prints log lines in json format
	Log_level  SlogLevel `json:"log_level"`  // LogLevel.Debug|Error|Info|Warn
	Out_file   string    `json:"out_file"`   // The file where log statements go (default is stderr)
	//	ConfigFilename   string    `json:"-"`          // The full path to the json configuration file
	DefaultConfigger // Default implementation of Configger
}

// toLevel converts SlogLevel to slog.Level.
// Return slogLevel as slog.Level.
func (s *SlogConfig) toLevel() slog.Level {
	//	return slog.Level(s.Log_level)
	return s.Log_level.toLevel()

}

// Setup overrides Configger/DefaultConfigger and sets up the object for logging.
//   - parentOpt is only to meet the implementation requirements of Configger.
func (s *SlogConfig) SetupConfig() error {
	if err := s.DefaultConfigger.DefaultSetupConfig(s); err != nil {
		return err

	}

	s.SetSlogger()

	return nil

}

// CustomSlogHandler wraps slog.Handler to customize logging behavior.
type CustomSlogHandler struct {
	slog.Handler
}

// Handle is a custom wrapper that also prints slog Info and Error message to stdout.
func (c *CustomSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= slog.LevelError || r.Level == slog.LevelInfo {
		extracts := make(map[string]string)
		r.Attrs(func(attr slog.Attr) bool {
			extracts[attr.Key] = attr.Value.String()
			return true

		})

		if r.Level == slog.LevelInfo {
			jsonData, _ := json.Marshal(extracts)
			fmt.Fprintf(os.Stderr, "%s, %s\n", stringy.CaseString(r.Message, stringy.SentenceCase), jsonData)

		} else if r.Level == slog.LevelError {
			fmt.Fprintf(os.Stderr, "ERROR: %s, %#v\n", r.Message, extracts)

		}

	}

	return c.Handler.Handle(ctx, r)

}

// SetLogger configures and sets an slog logger.
func (s *SlogConfig) SetSlogger() {
	slogLevel := s.toLevel()
	var logWriter io.Writer

	if s.Dev {
		if slogLevel == SInfo.toLevel() {
			slogLevel = SDebug.toLevel()

		}

		logWriter = os.Stderr

	}

	if s.Out_file != "" {
		rotator := &lumberjack.Logger{
			Filename:   s.Out_file,
			MaxSize:    50,
			MaxAge:     30,
			MaxBackups: 30,
		}
		slog.Info("defining rotator", "rotator", rotator)
		logWriter = rotator

	}

	logOpts := &slog.HandlerOptions{
		AddSource: s.Add_source,
		Level:     slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Check if this attribute is the source file/line info
			if a.Key == slog.SourceKey && s.Log_func {
				source, ok := a.Value.Any().(*slog.Source)
				// Extract only the file name from the full path
				if ok {
					return slog.String(
						a.Key, filepath.Base(source.File)+"/"+filepath.Base(source.Function)+"()",
					)

				}

			}

			return a
		},
	}

	var logger *slog.Logger
	if s.Log_json {
		jsonHandler := slog.NewJSONHandler(logWriter, logOpts)

		var customErrorHandler *CustomSlogHandler
		if logWriter != os.Stderr {
			customErrorHandler = &CustomSlogHandler{Handler: jsonHandler}
			logger = slog.New(customErrorHandler)

		} else {
			logger = slog.New(jsonHandler)

		}

	} else {
		textHandler := slog.NewTextHandler(logWriter, logOpts)
		logger = slog.New(textHandler)

	}
	slog.SetDefault(logger)
	slog.Debug("SlogConfig", "slc", s)

}

// NewSlogConfig returns an SlogConfig object configured with values from configFile.
func NewSlogConfig(dirName ...string) *SlogConfig {
	var directoryName string
	if len(dirName) == 0 {
		directoryName = os.Getenv("PROGNAME")
		if directoryName == "" {
			u, err := user.Current()
			if err == nil {
				directoryName = u.Username

			}

		}

	} else {
		directoryName = dirName[0]

	}

	newSlogConfig := &SlogConfig{
		DefaultConfigger: DefaultConfigger{
			DefaultConfigFilename: slogConfigFilename,
			DirName:               directoryName,
			DefaultParentValues: &SlogConfig{
				Add_source: true,
				Dev:        true,
				Log_func:   true,
				Log_json:   true,
				Log_level:  SDebug,
			},
		},
	}
	newSlogConfig.SetupConfig()

	return newSlogConfig

}

func NewSlogConfigMan() {
	GetConfigManInstance().AddConfig(NewSlogConfig())

}

// AnyToSlogConfig takes an object of type Configger and returns a *SlogConfig.
//   - configger is an object of type Configger
//
// Return configger as type *SlogConfig or an empty struct of type *SlogConfig if configger is not of type SlogConfig.
func AnyToSlogConfig(configger any) *SlogConfig {
	toConfigger, ok := configger.(Configger)

	s, ok := toConfigger.(*SlogConfig)
	if !ok {
		return NewSlogConfig()

	}

	return s

}

func ConfiggerToSlogConfig(configger Configger) *SlogConfig {
	s, ok := configger.(*SlogConfig)
	if !ok {
		return NewSlogConfig()

	}

	return s

}
