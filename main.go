package main

import (
	"flag"
	"github.com/clanko/gadget/cmd"
	"github.com/clanko/gadget/config"
	"os"
	"os/signal"
	"runtime"
)

const (
	GADGET_VERSION = "0.1.0"
)

var (
	help       bool
	appPath    string
	binaryName string
	listenPort int
	verbose    = 0
)

// Gadget CLI
func main() {
	goVersion := runtime.Version()
	cmd.PrintfSuccess("Gadget v: %v", GADGET_VERSION)
	cmd.PrintfSuccess("Go v: %v", goVersion)

	setFlags()

	if help == true {
		flag.Usage()
		return
	}

	conf := getConfigWithFlags()

	builder := NewBuilder(conf)

	// in case of panic
	defer func() {
		cmd.KillPid(builder.debugger.Process.Pid)
	}()

	// Clean up before exiting
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, os.Interrupt)
	go func() {
		<-quitChannel
		if builder.debugger != nil {
			cmd.KillPid(builder.debugger.Process.Pid)
		}

		if builder.runningBinary != nil {
			cmd.KillPid(builder.runningBinary.Process.Pid)
		}

		os.Exit(0)
	}()

	if flag.Arg(0) == "dev" {
		cmd.PrintfInfo("Building...")
		builder.build()

		cmd.PrintfInfo("Watching files")

		watcher := NewWatcher(conf)

		watcher.onEvent = func() {
			// recompile
			println("\nRebuilding")

			builder.build()

			cmd.PrintPrompt()
		}

		go func() {
			watcher.watch()
		}()
	}

	gsh := cmd.NewGadgetShell()

	go gsh.Run()

	cmd.PrintfInfo("^C to exit")

	<-make(chan struct{})
}

func setFlags() {
	const helpUsage = "Print this help message."

	flag.BoolVar(&help, "help", false, helpUsage)
	flag.BoolVar(&help, "h", false, helpUsage+" short-hand")

	flag.IntVar(&verbose, "v", 0, "-v 1")

	flag.StringVar(&appPath, "path", "", "-path /path/to/project")

	flag.StringVar(&binaryName, "binary", "", "-binary name")

	flag.IntVar(&listenPort, "listen", 0, "-listen 2345")

	flag.Parse()
}

func getConfigWithFlags() config.Config {
	conf := config.GetConfig("gadget.toml")

	if appPath != "" {
		conf.Path = appPath
	}

	if binaryName != "" {
		conf.Name = binaryName
	}

	if listenPort != 0 {
		conf.ListenPort = listenPort
	}

	return conf
}
