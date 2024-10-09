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
		// no config. load default config
		return config
	}

	err = toml.Unmarshal(content, &config)
	if err != nil {
		panic(cmd.FormatDanger("Failed to parse config file\n%v", err))
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
	}
}
