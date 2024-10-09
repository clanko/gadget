package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

var TestConfig = Config{
	Name:          "appNameTest",
	Address:       "localhost:8080",
	Path:          "/path/test",
	BuildArgs:     []string{"-gcflags=all=-N -l", "testarg"},
	ListenPort:    3811,
	ExcludeDirs:   []string{"_testdata/exclude_dir", "_testdata/exclude_dir/included_dir/excluded_dir"},
	ExcludeFiles:  []string{"/path/test/appNameTest", "_testdata/watched_dir/excluded_file.go"},
	ExcludeExts:   []string{"_testdata/watched_dir/extension_test.txt"},
	ExcludePrefix: []string{"_testdata/watched_dir/prefixed"},
	IncludeDirs:   []string{"_testdata/exclude_dir/included_dir"},
	IncludeFiles:  []string{"_testdata/exclude_dir/included_dir/excluded_dir/included_file.go"},
}

type addableStrings []string

func (e *addableStrings) Add(value string) {
	*e = append(*e, value)
}

func compareStringSlices(expectedSlice, givenSlice []string) (notFound string, unexpected string) {
	notFoundArgs := ""
	for _, expectedArg := range expectedSlice {
		found := false
		for _, givenArg := range givenSlice {
			if expectedArg == givenArg {
				found = true
			}
		}

		if found == false {
			notFoundArgs = notFoundArgs + expectedArg + "\n"
		}
	}

	unexpectedArgs := ""
	for _, givenArg := range givenSlice {
		found := false
		for _, expectedArg := range expectedSlice {
			if expectedArg == givenArg {
				found = true
			}
		}

		if found == false {
			unexpectedArgs = unexpectedArgs + givenArg + "\n"
		}
	}

	return notFoundArgs, unexpectedArgs
}

func TestGetConfig(t *testing.T) {
	// GetConfig with a non-existent config file returns the default Config
	conf := GetConfig("")

	defaultConf := getDefaultConfig()

	if reflect.DeepEqual(conf, defaultConf) == false {
		t.Errorf("GetConfig without config file didn't match the default Config")
	}

	// GetConfig with an existing config file, returns Config loaded with its values
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	loadedConf := GetConfig(wd + "/../_testdata/config.toml")

	// Test loaded config has expected values.
	expectedValues := addableStrings{}
	unexpectedValues := addableStrings{}

	if TestConfig.Name != loadedConf.Name {
		expectedValues.Add("Name " + TestConfig.Name)
		unexpectedValues.Add("Name " + loadedConf.Name)
	}

	if TestConfig.Address != loadedConf.Address {
		expectedValues.Add("Address " + TestConfig.Address)
		unexpectedValues.Add("Address " + loadedConf.Address)
	}

	if TestConfig.Path != loadedConf.Path {
		expectedValues.Add("Path " + TestConfig.Path)
		unexpectedValues.Add("Path " + loadedConf.Path)
	}

	if reflect.DeepEqual(TestConfig.BuildArgs, loadedConf.BuildArgs) {
		notFoundArgs, unexpectedArgs := compareStringSlices(TestConfig.BuildArgs, loadedConf.BuildArgs)

		if notFoundArgs != "" {
			expectedValues.Add("BuildArgs not found: " + notFoundArgs)
		}

		if unexpectedArgs != "" {
			unexpectedValues.Add("BuildArgs unexpected " + unexpectedArgs)
		}
	}

	if TestConfig.ListenPort != loadedConf.ListenPort {
		expectedValues.Add("ListenPort " + strconv.Itoa(TestConfig.ListenPort))
		unexpectedValues.Add("ListenPort " + strconv.Itoa(loadedConf.ListenPort))
	}

	if reflect.DeepEqual(TestConfig.ExcludeDirs, loadedConf.ExcludeDirs) == false {
		SetSliceCompareValues(
			"ExcludeDirs",
			TestConfig.ExcludeDirs,
			loadedConf.ExcludeDirs,
			&expectedValues,
			&unexpectedValues,
		)
	}

	if reflect.DeepEqual(TestConfig.ExcludeFiles, loadedConf.ExcludeFiles) == false {
		SetSliceCompareValues(
			"ExcludeFiles",
			TestConfig.ExcludeFiles,
			loadedConf.ExcludeFiles,
			&expectedValues,
			&unexpectedValues,
		)
	}

	if reflect.DeepEqual(TestConfig.ExcludeExts, loadedConf.ExcludeExts) == false {
		SetSliceCompareValues(
			"ExcludeExts",
			TestConfig.ExcludeExts,
			loadedConf.ExcludeExts,
			&expectedValues,
			&unexpectedValues,
		)
	}

	if reflect.DeepEqual(TestConfig.ExcludePrefix, loadedConf.ExcludePrefix) == false {
		SetSliceCompareValues(
			"ExcludePrefix",
			TestConfig.ExcludePrefix,
			loadedConf.ExcludePrefix,
			&expectedValues,
			&unexpectedValues,
		)
	}

	if reflect.DeepEqual(TestConfig.IncludeDirs, loadedConf.IncludeDirs) == false {
		SetSliceCompareValues(
			"IncludeDirs",
			TestConfig.IncludeDirs,
			loadedConf.IncludeDirs,
			&expectedValues,
			&unexpectedValues,
		)
	}

	if reflect.DeepEqual(TestConfig.IncludeFiles, loadedConf.IncludeFiles) == false {
		SetSliceCompareValues(
			"IncludeFiles",
			TestConfig.IncludeFiles,
			loadedConf.IncludeFiles,
			&expectedValues,
			&unexpectedValues,
		)
	}

	if len(expectedValues) > 0 || len(unexpectedValues) > 0 {
		message := ""
		if len(expectedValues) > 0 {
			message = "Expected:\n" + strings.Join(expectedValues, "\n")
		}

		if len(unexpectedValues) > 0 {
			message = message + "Got:\n" + strings.Join(unexpectedValues, "\n")
		}

		t.Errorf("Loaded config did not have the expected values\n%v", message)
	}
}

func SetSliceCompareValues(
	property string,
	expectedSlice []string,
	givenSlice []string,
	expectedValues *addableStrings,
	unexpectedValues *addableStrings,
) {
	notFoundArgs, unexpectedArgs := compareStringSlices(expectedSlice, givenSlice)

	if notFoundArgs != "" {
		expectedValues.Add(property + " not found: " + notFoundArgs)
	}

	if unexpectedArgs != "" {
		unexpectedValues.Add(property + " unexpected " + unexpectedArgs)
	}
}
