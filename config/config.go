package config

import (
	"github.com/clanko/gadget/cmd"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type Config struct {
	Name          string   `toml:"app_name"`
	Path          string   `toml:"app_path"`
	Address       string   `toml:"app_address"`
	BuildArgs     []string `toml:"build_args"`
	ListenPort    int      `toml:"listen_port"`
	ListenHost    string   `toml:"listen_host"`
	ExcludeDirs   []string `toml:"exclude_dirs"`
	ExcludeFiles  []string `toml:"exclude_files"`
	ExcludeExts   []string `toml:"exclude_exts"`
	ExcludePrefix []string `toml:"exclude_prefix"`
	IncludeDirs   []string `toml:"include_dirs"`
	IncludeFiles  []string `toml:"include_files"`
}

func GetConfig(configPath string) Config {
	config := getDefaultConfig()

	content, err := os.ReadFile(configPath)
	if err != nil {
		cmd.PrintfWarning("No gadget.toml found. Using default config")

		return config
	}

	err = toml.Unmarshal(content, &config)
	if err != nil {
		panic(cmd.FormatDanger("Failed to parse config file\n%v", err))
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(cmd.FormatDanger(err.Error()))
	}

	// loop through and modify any relative paths to absolute
	for i, file := range config.ExcludeFiles {
		if string(file[0]) != "/" {
			config.ExcludeFiles[i] = wd + "/" + file
		}
	}

	for i, file := range config.ExcludeDirs {
		if string(file[0]) != "/" {
			config.ExcludeDirs[i] = wd + "/" + file
		}
	}

	for i, file := range config.IncludeFiles {
		if string(file[0]) != "/" {
			config.IncludeFiles[i] = wd + "/" + file
		}
	}

	for i, file := range config.IncludeDirs {
		if string(file[0]) != "/" {
			config.IncludeDirs[i] = wd + "/" + file
		}
	}

	config.ExcludeFiles = append(config.ExcludeFiles, config.Path+"/"+config.Name)

	return config
}

func getDefaultConfig() Config {
	wd, err := os.Getwd()
	if err != nil {
		panic(cmd.FormatDanger(err.Error()))
	}

	return Config{
		Name:       "gadget_binary",
		Path:       wd,
		BuildArgs:  []string{`-gcflags=all=-N -l`},
		ListenPort: 3811,
		ListenHost: "127.0.0.1",
	}
}

func GetSampleConfigFileContent() string {
	return `
# Skip the command line flags and use a gadget.toml configuration file.
# Either pass the config location as a command or just execute gadget from the same directory.
# Flags specified in the command will override values set in the configuration file.

# The name of the binary.
# app_name = "gadget_binary"

# The address to run the app on. If none is specified, Gadget will try to find an available port on localhost to use.
# app_address = "localhost:8090"

# The path to the project to build, if no value is set, Gadget will use the current directory.
# app_path = "/path/to/project"

# Arguments to pass to go when building the binary.
# build_args = ["-gcflags=all=-N -l"]

# The port to connect debugger. If not set, Gadget will find an available port.
# listen_port = 3811

# The host to connect debugger. If not set, will default to 127.0.0.1
# listen_host = 127.0.0.1

# Directories that shouldn't trigger rebuild.
# exclude_dirs = []

# Files that shouldn't trigger rebuild. The generated binary is automatically excluded, no need to include it.
exclude_files = ["README.md"]

# Files with these extensions won't trigger rebuild.
# exclude_exts = []

# Files beginning with a sequence of chars matching these won't trigger rebuild.
# exclude_prefix = []

# Directories that should prompt rebuild. Useful for specifying custom paths to other modules in use by a project.
# include_dirs = []

# Files that should prompt rebuild.
# include_files = []

`
}
