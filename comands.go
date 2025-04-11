package main

import (
	"bufio"
	"github.com/clanko/gadget/cmd"
	"github.com/clanko/gadget/config"
	"github.com/clanko/scaffold"
	"os"
)

type command interface {
	execute(input *bufio.Scanner, args []string)
}

type watchCommand struct {
	gsh *gadgetShell
}

func (command watchCommand) execute(input *bufio.Scanner, args []string) {
	if command.gsh.watcher != nil && command.gsh.watcher.isWatching {
		cmd.PrintfWarning("Watcher already watching")

		return
	}

	watcher := newWatcher(command.gsh.config)

	command.gsh.watcher = &watcher

	cmd.PrintfInfo("Watching files")

	runWatcher(command.gsh.watcher, command.gsh.builder)
}

type unwatchCommand struct {
	gsh *gadgetShell
}

func (command unwatchCommand) execute(input *bufio.Scanner, args []string) {
	if command.gsh.watcher != nil && command.gsh.watcher.isWatching {
		command.gsh.watcher.endWatch()

		command.gsh.watcher = nil
	}
}

type devCommand struct {
	gsh *gadgetShell
}

func (command devCommand) execute(input *bufio.Scanner, args []string) {
	if command.gsh.watcher != nil && command.gsh.watcher.isWatching {
		command.gsh.watcher.endWatch()

		command.gsh.watcher = nil
	}

	command.gsh.builder.runBuildDebug()

	watcher := newWatcher(command.gsh.config)

	command.gsh.watcher = &watcher

	cmd.PrintfInfo("Watching files")

	runWatcher(command.gsh.watcher, command.gsh.builder)
}

type debugCommand struct {
	gsh *gadgetShell
}

func (command debugCommand) execute(input *bufio.Scanner, args []string) {
	command.gsh.builder.runBuildDebug()
}

type runCommand struct {
	gsh *gadgetShell
}

func (command runCommand) execute(input *bufio.Scanner, args []string) {
	command.gsh.builder.stopRunningProcesses()

	err := command.gsh.builder.buildBinary()
	if err != nil {
		cmd.PrintfDanger("%v", err)
	}

	command.gsh.builder.runBinary()
}

type buildCommand struct {
	gsh *gadgetShell
}

func (command buildCommand) execute(input *bufio.Scanner, args []string) {
	command.gsh.builder.stopRunningProcesses()

	err := command.gsh.builder.buildBinary()
	if err != nil {
		cmd.PrintfDanger("%v", err)
	}
}

type makeCommand struct {
}

func (make makeCommand) execute(input *bufio.Scanner, args []string) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	templatesPath := home + "/" + gadgetCliConfigDir + "/make-templates"

	err = os.MkdirAll(templatesPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	path := args[0]

	// make at destination
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if path == "gadget-config" {
		gadgetConfigFile := wd + "/gadget.toml"
		// make sure there's not already a gadget.toml
		_, err = os.Stat(gadgetConfigFile)
		if err == nil {
			cmd.PrintfInfo("It looks like there's already a gadget.toml here. Remove or rename it before generating another")

			return
		}

		// generate a gadget.toml in the current directory
		newFile, err := os.Create(gadgetConfigFile)
		defer func(newFile *os.File) {
			err := newFile.Close()
			if err != nil {

			}
		}(newFile)

		if err != nil {
			cmd.PrintfDanger("%v", err)

			return
		}

		_, err = newFile.WriteString(config.GetSampleConfigFileContent())
		if err != nil {
			cmd.PrintfDanger("%v", err)
		}

		return
	}

	newTemplatePath := templatesPath + "/" + path

	// check if template path exists.
	_, err = os.Stat(newTemplatePath)
	if err != nil {
		cmd.FormatDanger("Failed to locate template directory: %v", path)

		return
	}

	scaf, err := scaffold.Init(newTemplatePath)
	if err != nil {
		cmd.FormatDanger("Failed to initialize scaffold. Does %v/scaffold.toml exist?", newTemplatePath)
	}

	// get all tokens. Those that don't have a ValueToken set, prompt for value
	tokens := scaf.GetTokens()
	for _, token := range tokens {
		if token.ValueToken == "" {
			// prompt for token value
			for token.Value == "" {
				print(cmd.FormatInfo("Enter value for token %v: ", token.Name))

				input.Scan()
				if input.Text() != "" {
					scaf.RegisterTokenValue(token.Name, input.Text())
					token.Value = input.Text()
				}
			}
		}
	}

	scaf.OnMake(func(created string) {
		cmd.PrintfSuccess("Created: %v", created)
	})

	err = scaf.Make(wd)
	if err != nil {
		panic(err)
	}
}
