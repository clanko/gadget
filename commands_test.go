package main

import (
	"bufio"
	"os"
	"testing"
)

func TestMakeGadetConfigCommand(t *testing.T) {
	conf := getConfigWithFlags()

	builder := newBuilder(conf)
	watcher := newWatcher(conf)

	gsh := newGadgetShell(&builder, &watcher)

	input := bufio.NewScanner(os.Stdin)

	command := gsh.getCommand("make")

	command.execute(input, []string{"gadget-config"})
}
