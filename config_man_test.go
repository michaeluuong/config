package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/michaeluuong/utilize/filing"
)

func BenchmarkConfig_man(b *testing.B) {
	/*for i := 0; i < b.N; i++ {

	}*/

}

var configMan = GetConfigManInstance()

type TestConfig struct {
	Field1           string
	DefaultConfigger // Default implementation of Configger
}

func (t *TestConfig) IsSetup() bool {
	return t.isSetup
}

func (t *TestConfig) SetupConfig() error {
	return t.DefaultConfigger.DefaultSetupConfig(t)

}

var testConfig = &TestConfig{
	Field1: "field1",
	DefaultConfigger: DefaultConfigger{
		DefaultConfigFilename: "config_man_test_cfg.json",
	},
}

var testEmptyConfig = &TestConfig{
	Field1:           "field1",
	DefaultConfigger: DefaultConfigger{},
}

func TestAddConfig_Config_Singleton_main(t *testing.T) {
	var err error

	//--------------------------------------------------------------------------------------------
	// Happy
	err = configMan.AddConfig(testConfig)
	assert.Nil(t, err)
	myActualTestConfig := configMan.Config(&TestConfig{})
	assert.Equal(t, testConfig, myActualTestConfig)
	assert.Implements(t, (*Configger)(nil), testConfig)

	myActualTestConfig, configMan = nil, nil
	configMan = GetConfigManInstance()
	assertedActualTestConfig, ok := configMan.Config(new(TestConfig)).(*TestConfig)
	assert.True(t, ok, "type assertion to TestConfig failed")
	assert.Implements(t, (*Configger)(nil), assertedActualTestConfig)
	assert.Equal(t, testConfig, assertedActualTestConfig)

	//--------------------------------------------------------------------------------------------
	// Test Empty
	err = configMan.AddConfig(testEmptyConfig)
	assert.EqualError(t, err, "envVar or cFilename must be specified")

	//--------------------------------------------------------------------------------------------
	myActTestConfig, ok := myActualTestConfig.(TestConfig)
	os.Remove(myActTestConfig.ConfigFile())

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestAssertConfigger_main(t *testing.T) {
	var anyTestConfigAsserted *TestConfig
	var ok bool

	err := configMan.AddConfig(testConfig)
	assert.Nil(t, err)

	//--------------------------------------------------------------------------------------------
	// Happy
	anyTestConfig := configMan.ConfigPtr(&TestConfig{})
	assert.Implements(t, (*Configger)(nil), *anyTestConfig)
	fmt.Printf("anyTestConfig: %#v\n", *anyTestConfig)

	anyTestConfigAsserted, ok = AssertConfigger[TestConfig](anyTestConfig)
	assert.True(t, ok)
	assert.Implements(t, (*Configger)(nil), anyTestConfigAsserted)
	assert.Equal(t, testConfig, anyTestConfigAsserted)

	//--------------------------------------------------------------------------------------------
	// Not ok
	type NoConfig struct{}
	anyNoConfig := any(&NoConfig{})
	anyTestConfigAsserted, ok = AssertConfigger[TestConfig](&anyNoConfig)
	assert.False(t, ok)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestSet_IsLoggerSet_main(t *testing.T) {
	var actualValue bool

	//--------------------------------------------------------------------------------------------
	// Check if actually set to true
	configMan.SetIsLoggerSet(true)
	expectedSetValue := configMan.isLoggerSet
	assert.True(t, expectedSetValue)

	actualValue = configMan.IsLoggerSet()
	assert.True(t, actualValue)

	// Check if actually set to false
	configMan.SetIsLoggerSet(false)
	actualValue = configMan.IsLoggerSet()
	assert.False(t, actualValue)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestResetConfigManInstance_main(t *testing.T) {
	configMan = ResetConfigManInstance()

	//--------------------------------------------------------------------------------------------
	// Happy
	myActualTestConfig, ok := configMan.Config(&TestConfig{}).(*TestConfig)
	assert.False(t, ok, "type assertion to TestConfig failed")
	assert.Nil(t, myActualTestConfig)
	assert.NotEqual(t, testConfig, myActualTestConfig)

	Prvp("myActualTestConfig", myActualTestConfig)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestFindConfigFile_main(t *testing.T) {
	var actualConfigFile, cFile, cFilename, cDir string
	var err error
	parentDir := "audios"

	//--------------------------------------------------------------------------------------------
	// Environment
	cFile = "/Users/irenepiechota-wong/DEV/GO/audios/audiostag/config/audiostag_cfg.json"
	os.Setenv("CONFIG_FILE", cFile)
	actualConfigFile, err = FindConfigFile("CONFIG_FILE", "")
	assert.Nil(t, err)
	assert.Equal(t, cFile, actualConfigFile)

	//--------------------------------------------------------------------------------------------
	// UserConfigDir
	userConfigDir, _ := os.UserConfigDir()
	cFilename = "audiostag_cfg.json"
	cFile = filepath.Join(userConfigDir, parentDir, "config", cFilename)
	cDir = filepath.Dir(cFile)
	if !filing.Exists(cDir) {
		os.MkdirAll(cDir, os.ModePerm)

	}
	os.WriteFile(cFile, []byte("{}"), os.ModePerm)

	actualConfigFile, err = FindConfigFile("", cFilename, parentDir)
	assert.Equal(t, cFile, actualConfigFile)

	os.Remove(cFile)
	os.Remove(cDir)

	// Empty
	actualConfigFile, err = FindConfigFile("", "")
	assert.EqualError(t, err, "envVar or cFilename must be specified")

	//--------------------------------------------------------------------------------------------
	// DNE
	actualConfigFile, err = FindConfigFile("", "dne.txt")
	assert.EqualError(t, err, "could not find a config file")

	//--------------------------------------------------------------------------------------------
	// UserHomeDir
	userHomeDir, _ := os.UserHomeDir()
	cFilename = filepath.Join("audios", "audiostag_cfg.json")
	cFile = filepath.Join(userHomeDir, cFilename)
	cDir = filepath.Dir(cFile)
	if !filing.Exists(cDir) {
		os.MkdirAll(cDir, os.ModePerm)

	}
	os.WriteFile(cFile, []byte("{}"), os.ModePerm)

	actualConfigFile, err = FindConfigFile("", cFilename)
	Prvp("actualConfigFile", actualConfigFile)
	assert.Equal(t, cFile, actualConfigFile)
	os.Remove(cFile)
	os.Remove(cDir)

	//--------------------------------------------------------------------------------------------
	// Working directory
	wd, _ := os.Getwd()
	cFilename = filepath.Join("audios", "audiostag_cfg.json")
	cFile = filepath.Join(wd, cFilename)
	cDir = filepath.Dir(cFile)
	if !filing.Exists(cDir) {
		os.MkdirAll(cDir, os.ModePerm)

	}
	os.WriteFile(cFile, []byte("{}"), os.ModePerm)

	actualConfigFile, err = FindConfigFile("", cFilename)
	assert.Equal(t, cFile, actualConfigFile)
	os.Remove(cFile)
	os.Remove(cDir)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}
