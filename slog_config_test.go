package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/jsoning"
)

// go test -benchtime=1s -bench . -cpuprofile cpu.prof
// go tool pprof cpu.prof
func BenchmarkLogs(b *testing.B) {
	/*for i := 0; i < b.N; i++ {

	}*/

}

func TestSlogLevel_String_main(t *testing.T) {
	var actualSlogLevel SlogLevel

	//--------------------------------------------------------------------------------------------
	// Happy
	actualSlogLevel = SDebug
	assert.Equal(t, SDebug, actualSlogLevel)

	// Default
	defaultSlogLevel := new(SlogLevel)
	assert.Equal(t, "INFO", defaultSlogLevel.String())

	//--------------------------------------------------------------------------------------------
	// String
	var levelString string

	levelString = "DEBUG"
	actualSlogLevel = SDebug
	assert.Equal(t, levelString, actualSlogLevel.String())

	levelString = "INFO"
	actualSlogLevel = SInfo
	assert.Equal(t, levelString, actualSlogLevel.String())

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSlogLevelFromString_main(t *testing.T) {
	var testSlogLevel SlogLevel

	//--------------------------------------------------------------------------------------------
	// Happy
	testSlogLevel.SlogLevelFromString("ERROR")
	assert.Equal(t, SError, testSlogLevel)

	testSlogLevel.SlogLevelFromString("WARN")
	assert.Equal(t, SWarn, testSlogLevel)

	testSlogLevel.SlogLevelFromString("INFO")
	assert.Equal(t, SInfo, testSlogLevel)

	testSlogLevel.SlogLevelFromString("DEBUG")
	assert.Equal(t, SDebug, testSlogLevel)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSlogLevel_Marshal_UnmarshalJSON_main(t *testing.T) {
	var testSlogLevel SlogLevel = SWarn
	jsonify := jsoning.NewJsonify()

	type TestSlogLevel struct {
		TestSlogLevelFile SlogLevel
	}

	//--------------------------------------------------------------------------------------------
	// Happy
	testSlogLevelToFile := &TestSlogLevel{TestSlogLevelFile: testSlogLevel}

	// Marshal
	filename := "test_unmarshal.txt"
	jsonify.WriteObjToJSONFile(filename, testSlogLevelToFile)

	// Unmarshal
	var testSlogLevelFromFile any = &TestSlogLevel{}
	jsonify.Unmarshal2Struct(filename, &testSlogLevelFromFile)
	assert.Equal(t, testSlogLevelToFile, testSlogLevelFromFile)

	// Invalid JSON
	err := testSlogLevel.UnmarshalJSON([]byte("juke"))
	fmt.Printf("err: %v\n", err)

	os.Remove(filename)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSlogLevel_ToLevel_main(t *testing.T) {
	var testSlogLevel SlogLevel
	var actualSlogLevel slog.Level

	//--------------------------------------------------------------------------------------------
	// Happy
	testSlogLevel = SDebug
	actualSlogLevel = testSlogLevel.toLevel()
	assert.Equal(t, slog.LevelDebug, actualSlogLevel)

	testSlogLevel = SError
	actualSlogLevel = testSlogLevel.toLevel()
	assert.Equal(t, slog.LevelError, actualSlogLevel)

	testSlogLevel = SInfo
	actualSlogLevel = testSlogLevel.toLevel()
	assert.Equal(t, slog.LevelInfo, actualSlogLevel)

	testSlogLevel = SWarn
	actualSlogLevel = testSlogLevel.toLevel()
	assert.Equal(t, slog.LevelWarn, actualSlogLevel)

	// Default is INFO
	testSlogLevel = *new(SlogLevel)
	actualSlogLevel = testSlogLevel.toLevel()
	assert.Equal(t, slog.LevelInfo, actualSlogLevel)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestNewSlogConfig_main(t *testing.T) {
	slogConfig := NewSlogConfig()
	slogConfig.Add_source = true
	assert.Implements(t, new(Configger), slogConfig)

	configMan := GetConfigManInstance()
	err := configMan.AddConfig(slogConfig)
	assert.Nil(t, err)

	/*slogConfigConfigger := configMan.ConfigCfgr(&SlogConfig{})
	toSlogConfig := ConfiggerToSlogConfig(slogConfigConfigger)
	assert.IsType(t, new(SlogConfig), toSlogConfig)*/

	slogConfigPtr := configMan.ConfigPtr(&SlogConfig{})
	slogConfigConfigger := (*slogConfigPtr).(Configger)
	toSlogConfig := ConfiggerToSlogConfig(slogConfigConfigger)
	assert.IsType(t, new(SlogConfig), toSlogConfig)

	//--------------------------------------------------------------------------------------------
	// Wrong type
	type NewType struct {
		DefaultConfigger
	}
	var newType any = &NewType{}
	_, isTargetType := newType.(SlogConfig)
	assert.False(t, isTargetType)

	//--------------------------------------------------------------------------------------------
	// dirName
	slogConfigDir := NewSlogConfig("audios")
	slogConfigDir.Add_source = true
	assert.Implements(t, new(Configger), slogConfig)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSlogConfig_SetupConfig_secondary(t *testing.T) {

	//--------------------------------------------------------------------------------------------
	badSlogConfig := &SlogConfig{
		DefaultConfigger: DefaultConfigger{},
	}
	err := badSlogConfig.SetupConfig()
	assert.EqualError(t, err, "envVar or cFilename must be specified")

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSetSlogger_secondary(t *testing.T) {
	workDir, _ := os.Getwd()
	outDir := filepath.Join(workDir, "log")
	outFilename := "slog_config_not_dev.txt"
	outFile := filepath.Join(outDir, outFilename)

	//--------------------------------------------------------------------------------------------
	// Dev is true so set log level to Debug
	slogConfigDev := &SlogConfig{
		Dev:       true,
		Log_level: SInfo,
	}
	slogConfigDev.SetSlogger()

	//--------------------------------------------------------------------------------------------
	// Dev is not set, send logging to file
	slogConfigNotDev := &SlogConfig{
		Dev:      false,
		Out_file: outFile,
	}
	slogConfigNotDev.SetSlogger()
	checkString := "CHECK"
	slog.Info(checkString)
	logFileCheck, err := filing.ReadFileToString(outFile)
	assert.Nil(t, err)
	logFileCheckLines := strings.Split(logFileCheck, "\n")
	logFileCheckLine := logFileCheckLines[len(logFileCheckLines)-2]

	today := time.Now().Format("2006-01-02")
	assert.Regexp(t, "^time="+today+"T.* level=INFO msg="+checkString, logFileCheckLine)

	//--------------------------------------------------------------------------------------------
	// Dev is not set, failed to open file for logging so send to Stderr
	slogConfigNotDev2 := &SlogConfig{
		Out_file: outFile,
	}
	_ = os.Chmod(outFile, 0000)
	slogConfigNotDev2.SetSlogger()

	/* Can't redirect the stderr that the slog handler is using
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	fmt.Printf("err: %v\n", err)
	os.Stderr = w

	captured := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		captured <- buf.String()
	}()
	slog.Info(checkString)

	w.Close()
	os.Stderr = oldStderr
	fmt.Printf("captured: %s\n", <-captured)
	assert.ErrorContains(t, err, ": permission denied")*/

	err = os.Chmod(outFile, 0777)
	os.Remove(outFile)
	os.Remove(outDir)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestAnyToSlogConfig_main(t *testing.T) {
	slogConfig := NewSlogConfig()
	var configgerAny any = slogConfig

	var slogConfigPtr *SlogConfig = NewSlogConfig()
	var configgerAnyPtr any = slogConfigPtr

	//--------------------------------------------------------------------------------------------
	// Happy
	slogConfig = AnyToSlogConfig(configgerAny)
	assert.True(t, slogConfig.Add_source)

	slogConfig = AnyToSlogConfig(configgerAnyPtr)
	assert.True(t, slogConfig.Add_source)

	slogConfig = AnyToSlogConfig(&configgerAnyPtr)
	assert.True(t, slogConfig.Add_source)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestConfiggerToSlogConfig_main(t *testing.T) {
	var configger Configger
	slogConfig := NewSlogConfig()
	slogConfig.Add_source = false
	configger = slogConfig

	var emptyConfigger Configger

	//--------------------------------------------------------------------------------------------
	// Happy
	slogConfig = ConfiggerToSlogConfig(configger)
	assert.False(t, slogConfig.Add_source)

	// Failed type assertion
	slogConfig = ConfiggerToSlogConfig(emptyConfigger)
	assert.True(t, slogConfig.Add_source)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}
