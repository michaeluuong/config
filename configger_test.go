package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/jsoning"
)

// go test -benchtime=1s -bench . -cpuprofile cpu.prof
// go tool pprof cpu.prof
func BenchmarkConfigger(b *testing.B) {
	/*for i := 0; i < b.N; i++ {
	}*/

}

// Manual implementation
type TestManualConfiggerImpl struct{ isSetup bool }

func (t *TestManualConfiggerImpl) SetupConfig() error {
	t.isSetup = true

	return nil

}

func (t *TestManualConfiggerImpl) IsSetup() bool { return t.isSetup }

func (t *TestManualConfiggerImpl) loadFromConfigFile() error { return nil }

func (t *TestManualConfiggerImpl) setWithEnvValues() error { return nil }

// --------------------------------------------------------------------------------------------
// Use default implementation.
type TestDefaultConfigger struct {
	Field1 string `json:"Field1"`
	DefaultConfigger
}

func (t *TestDefaultConfigger) SetupConfig() error {
	return t.DefaultConfigger.DefaultSetupConfig(t)

}

// --------------------------------------------------------------------------------------------
type TestBadConfigger struct {
	Field1 bool `json:"Field1"`
	DefaultConfigger
}

func (t *TestBadConfigger) SetupConfig() error {
	return t.DefaultConfigger.DefaultSetupConfig(t)

}

// --------------------------------------------------------------------------------------------
type TestDefaultConfiggerNoParent struct {
	Field1 string
	DefaultConfigger
}

func (t *TestDefaultConfiggerNoParent) SetupConfig() error {
	return t.DefaultConfigger.DefaultSetupConfig(nil)

}

var fullConfigger = &TestDefaultConfigger{
	Field1: "Field 1",
	DefaultConfigger: DefaultConfigger{
		DefaultParentValues: &TestDefaultConfigger{
			Field1: "Default Field1 Value",
		},
	},
}

func TestConfigger_main(t *testing.T) {
	//--------------------------------------------------------------------------------------------
	// Happy
	assert.Implementsf(t, (*Configger)(nil), new(TestDefaultConfigger), "TestDefaultConifgger should have implemented Configger by embedding DefaultConfigger")

	assert.Implementsf(t, (*Configger)(nil), new(TestManualConfiggerImpl), "TestManualConfiggerImpl should have inmplemented Configger with it's implementation methods")

	type testUnimplementedConfigger struct{}
	assert.NotImplements(t, (*Configger)(nil), new(testUnimplementedConfigger))

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestConfigger_DefaultSetupConfig_main(t *testing.T) {
	var err error

	testDefaultConfigger := &TestDefaultConfigger{}
	testCfgFilename := "test_cfg1.json"
	jsoning.NewJsonify().WriteObjToJSONFile(testCfgFilename, fullConfigger)

	//--------------------------------------------------------------------------------------------
	// No config file
	err = testDefaultConfigger.SetupConfig()
	assert.EqualError(t, err, "envVar or cFilename must be specified")

	// From environment variable
	os.Setenv("CONFIG_FILE", testCfgFilename)
	testConfiggerEnv := &TestDefaultConfigger{DefaultConfigger: DefaultConfigger{EnvVar: "CONFIG_FILE"}}
	err = testConfiggerEnv.SetupConfig()
	assert.Equal(t, fullConfigger.Field1, testConfiggerEnv.Field1)

	// DefaultConfigFilename (Since testCfgFilename has no path it should show up in working directory)
	testConfiggerFile := &TestDefaultConfigger{
		DefaultConfigger: DefaultConfigger{
			DefaultConfigFilename: testCfgFilename,
		},
	}
	err = testConfiggerFile.SetupConfig()
	assert.Equal(t, fullConfigger.Field1, testConfiggerFile.Field1)

	// New Config file (working directory)
	testConfiggerNewFile := &TestDefaultConfigger{
		DefaultConfigger: DefaultConfigger{
			DefaultConfigFilename: "new_cfg.json",
			DefaultParentValues:   fullConfigger,
		},
	}
	err = testConfiggerNewFile.SetupConfig()
	assert.Equal(t, fullConfigger.Field1, testConfiggerNewFile.Field1)

	//--------------------------------------------------------------------------------------------
	// Parent is nil
	testDefaultConfiggerNoParent := &TestDefaultConfiggerNoParent{}
	err = testDefaultConfiggerNoParent.SetupConfig()
	assert.EqualError(t, err, "parent is required")

	//--------------------------------------------------------------------------------------------
	newCfgFile := testConfiggerNewFile.DefaultConfigger.ConfigFile()
	os.Unsetenv("CONFIG_FILE")
	os.Remove(newCfgFile)
	os.Remove(testCfgFilename)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestConfigger_DefaultSetupConfig_secondary(t *testing.T) {
	var err error
	wd, _ := os.Getwd()
	testDir2 := filepath.Join(wd, "config2")
	testConfigFile2 := filepath.Join(testDir2, "configger_test_cfg2.json")

	//--------------------------------------------------------------------------------------------
	// Default Parent values (if d.DefaultParentValues != nil)
	os.Setenv("CONFIG_FILE2", testConfigFile2)
	defaultParentConfigger := &TestDefaultConfigger{
		Field1: "Field 1",
		DefaultConfigger: DefaultConfigger{
			EnvVar: "CONFIG_FILE2",
			DefaultParentValues: &TestDefaultConfigger{
				Field1: "Default Field1 Value",
			},
		},
	}
	defaultParentConfigger.SetupConfig()

	var anyResultConfigger any
	anyResultConfigger = &TestDefaultConfigger{}
	err = jsoning.NewJsonify().Unmarshal2Struct(testConfigFile2, &anyResultConfigger)
	assert.Nil(t, err)
	resultConfigger, ok := anyResultConfigger.(*TestDefaultConfigger)

	assert.Truef(t, ok, "should be able to type assert anyResultConfigger to to *TestDefaultConfigger")
	assert.Equal(t, "Default Field1 Value", resultConfigger.Field1)

	os.Remove(testConfigFile2)

	//--------------------------------------------------------------------------------------------
	// No Default Parent Values (if d.DefaultParentValues != nil else ->)
	noParentValuesConfigger := &TestDefaultConfigger{
		Field1: "Field 1",
		DefaultConfigger: DefaultConfigger{
			EnvVar: "CONFIG_FILE2",
		},
	}
	noParentValuesConfigger.SetupConfig()

	err = os.RemoveAll("config2")

	//--------------------------------------------------------------------------------------------
	// Can't write config file (d.SaveConfigFile())
	fmt.Printf("\n\nFUCK2a\n\n")
	testDir2a := filepath.Join(wd, "config2a")
	testConfigFile2a := filepath.Join(testDir2a, "configger_test_cfg2a.json")
	os.Setenv("CONFIG_FILE2A", testConfigFile2a)
	defaultParentConfigger2 := &TestDefaultConfigger{
		Field1: "Field 1",
		DefaultConfigger: DefaultConfigger{
			EnvVar: "CONFIG_FILE2A",
			DefaultParentValues: &TestDefaultConfigger{
				Field1: "Default Field1 Value",
			},
		},
	}
	_ = os.Mkdir("config2a", 0000)
	err = defaultParentConfigger2.SetupConfig()
	filing.Cat(testConfigFile2a)

	_ = os.Chmod("config2a", 0777)
	os.Remove(testConfigFile2a)
	os.Remove("config2a")

	//--------------------------------------------------------------------------------------------
	// Cleanup
	os.Unsetenv("CONFIG_FILE2")

}

func TestConfigger_DefaultSetupConfig_tertiary(t *testing.T) {
	var err error
	workDir, _ := os.Getwd()

	testConfigDir3 := filepath.Join(workDir, "config3")
	testConfigFile3 := filepath.Join(testConfigDir3, "configger_test_cfg3.json")
	os.Mkdir(testConfigDir3, 0755)

	//--------------------------------------------------------------------------------------------
	// Error (d.loadFromConfigFile())
	os.Setenv("CONFIG_FILE3", testConfigFile3)
	os.WriteFile(testConfigFile3, []byte("JSON\n"), 0755)
	loadFromConfigConfigger := &TestDefaultConfigger{
		Field1: "Field 1",
		DefaultConfigger: DefaultConfigger{
			EnvVar: "CONFIG_FILE3",
		},
	}
	filing.Cat(testConfigFile3)
	err = loadFromConfigConfigger.SetupConfig()
	assert.ErrorContains(t, err, "Error unmarshaling JSON: invalid character")

	os.Remove(testConfigFile3)

	//--------------------------------------------------------------------------------------------
	// Error (setWithEnvValues())
	type Field1 struct {
		Field1 bool `json:"Field1"`
	}
	field1 := &Field1{Field1: false}
	jsoning.NewJsonify().WriteObjToJSONFile(testConfigFile3, field1)
	os.Setenv("FIELD1", "fail")
	testBadConfigger := &TestBadConfigger{
		DefaultConfigger: DefaultConfigger{
			EnvVar: "CONFIG_FILE3",
		},
	}
	err = testBadConfigger.SetupConfig()
	assert.ErrorContains(t, err, "strconv.ParseBool: parsing")

	//--------------------------------------------------------------------------------------------
	// Cleanup
	os.Remove(testConfigFile3)
	os.Remove(testConfigDir3)
	os.Unsetenv("CONFIG_FILE")

}

type ex struct {
	Field1 string
}

func TestConfigger_Set_ConfigFile_main(t *testing.T) {
	var actualConfig string
	testDefaultConfigger := &TestDefaultConfigger{DefaultConfigger: DefaultConfigger{EnvVar: "CONFIG_FILE"}}
	wd, _ := os.Getwd()
	testDir := filepath.Join(wd, "config2")
	testConfigFile := filepath.Join(testDir, "set_configfile_test_cfg.json")

	//--------------------------------------------------------------------------------------------
	// Happy
	testDefaultConfigger.SetConfigFile(testConfigFile)
	actualConfig = testDefaultConfigger.ConfigFile()
	assert.Equal(t, testConfigFile, actualConfig)

	testDefaultConfigger.SetConfigFile("")
	actualConfig = testDefaultConfigger.ConfigFile()
	assert.Equal(t, testConfigFile, actualConfig)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestConfigger_SaveConfigFile_main(t *testing.T) {

	//--------------------------------------------------------------------------------------------
	// Happy
	testSaveFileName := "testsave_cfg.json"
	expected := &ex{Field1: "field 1"}
	fullConfigger.SetConfigFile(testSaveFileName)
	fullConfigger.SaveConfigFile(expected)

	// Empty
	fullConfigger.Field1 = "Nil Test"
	fullConfigger.SaveConfigFile(nil)
	assert.Equal(t, fullConfigger.Field1, "Nil Test")

	//--------------------------------------------------------------------------------------------
	// Cleanup
	os.Remove(testSaveFileName)

}
func TestConfigger_IsSetup_SetIsSetup_main(t *testing.T) {

	//--------------------------------------------------------------------------------------------
	// Happy
	fullConfigger.SetIsSetup(false)
	assert.False(t, fullConfigger.IsSetup())

	fullConfigger.SetIsSetup(true)
	isSetup := fullConfigger.IsSetup()
	assert.True(t, isSetup)

	//--------------------------------------------------------------------------------------------
	// Cleanup

}

func TestConfigger_SetDefaults_main(t *testing.T) {
	cFile := "test_cfg.json"

	//--------------------------------------------------------------------------------------------
	// Happy
	fullConfigger.SetDefaults(DefaultConfigger{
		parent:     fullConfigger,
		configFile: cFile,
	})
	assert.Equal(t, fullConfigger, fullConfigger.parent)
	assert.Equal(t, fullConfigger.DefaultConfigger.configFile, fullConfigger.configFile)

	//--------------------------------------------------------------------------------------------
	// EnvVar
	fullConfigger.SetDefaults(DefaultConfigger{
		EnvVar: "CONFIG_FILE",
		parent: fullConfigger,
	})
	assert.Equal(t, fullConfigger.DefaultConfigger.EnvVar, "CONFIG_FILE")

	//--------------------------------------------------------------------------------------------
	// DefaultConfigFilename
	fullConfigger.SetDefaults(DefaultConfigger{
		DefaultConfigFilename: cFile,
		parent:                fullConfigger,
	})
	assert.Equal(t, fullConfigger.DefaultConfigger.DefaultConfigFilename, cFile)

	//--------------------------------------------------------------------------------------------
	// DefaultParentValues
	defaultParentValues := TestDefaultConfigger{Field1: "default parent values"}
	fullConfigger.SetDefaults(DefaultConfigger{
		parent:              fullConfigger,
		DefaultParentValues: defaultParentValues,
	})
	assert.Equal(t, fullConfigger.DefaultConfigger.DefaultParentValues, defaultParentValues)
	Prvp("fullConfigger", fullConfigger)

	//--------------------------------------------------------------------------------------------
	// DirName
	fullConfigger.SetDefaults(DefaultConfigger{
		DirName: "audios",
		parent:  fullConfigger,
	})
	assert.Equal(t, fullConfigger.DefaultConfigger.DirName, "audios")

	//--------------------------------------------------------------------------------------------
	// Cleanup

}
