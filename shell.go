package main

import (
	"bufio"
	"github.com/clanko/gadget/cmd"
	"github.com/clanko/gadget/config"
	"io"
	"os"
	"os/exec"
	"strings"
)

var gadgetCliConfigDir = ".clanko-gadget-cli"

type gadgetShell struct {
	builder *builder
	watcher *watcher
	config  config.Config
}

func newGadgetShell(b *builder, w *watcher, config config.Config) gadgetShell {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	templatesPath := home + "/" + gadgetCliConfigDir + "/make-templates"

	err = os.MkdirAll(templatesPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	gsh := gadgetShell{}
	gsh.watcher = w
	gsh.builder = b
	gsh.config = config

	return gsh
}

func (gsh gadgetShell) getCommands() map[string]command {
	registeredCommands := make(map[string]command)

	registeredCommands["make"] = makeCommand{}
	registeredCommands["build"] = buildCommand{&gsh}
	registeredCommands["run"] = runCommand{&gsh}
	registeredCommands["debug"] = debugCommand{&gsh}
	registeredCommands["dev"] = devCommand{&gsh}
	registeredCommands["watch"] = watchCommand{&gsh}
	registeredCommands["unwatch"] = unwatchCommand{&gsh}

	return registeredCommands
}

func (gsh gadgetShell) getCommand(key string) command {
	return gsh.getCommands()[key]
}

func (gsh gadgetShell) hasCommand(key string) bool {
	for registered := range gsh.getCommands() {
		if registered == key {
			return true
		}
	}

	return false
}

func (gsh gadgetShell) run() {
	input := bufio.NewScanner(os.Stdin)

	for {
		printPrompt()

		input.Scan()

		bin, args := gsh.splitCommand(input.Text())
		// first check if it's a registered command for gadget
		if gsh.hasCommand(bin) != false {
			gsh.getCommand(bin).execute(input, args)
		} else {
			// If we don't have a command check if it's a valid shell command
			command := exec.Command(bin, args...)

			stdout, err := command.Output()
			if err != nil {
				// error executing command. print help message
				cmd.PrintfWarning("Unknown command: " + bin)
				cmd.PrintfWarning("Enter \"help\" for a list of commands")
			} else {
				println(string(stdout))
			}
		}
	}
}

func (gsh gadgetShell) splitCommand(input string) (string, []string) {
	split := strings.Split(input, " ")
	var bin string
	args := make([]string, 0)
	for i := range split {
		if i == 0 {
			bin = split[0]
		} else {
			args = append(args, split[i])
		}
	}

	return bin, args
}

func printPrompt() {
	print(cmd.FormatSuccess("gadget->: "))
}

func printReadCloser(readCloser io.ReadCloser, printFunc func(string)) {
	reader := bufio.NewReader(readCloser)
	line, err := reader.ReadString('\n')
	for err == nil {
		if line != "" {
			printFunc(line)

			return
		}

		line, err = reader.ReadString('\n')
	}
}
