package main

import (
	"fmt"
	"github.com/clanko/gadget/cmd"
	"github.com/clanko/gadget/config"
	"net"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

type builder struct {
	config        config.Config
	runningBinary *exec.Cmd
	debugger      *exec.Cmd
	port          int
}

func NewBuilder(conf config.Config) builder {
	return builder{
		config: conf,
	}
}

func (b *builder) build() {
	if b.runningBinary != nil && b.runningBinary.Process != nil {
		cmd.KillPid(b.runningBinary.Process.Pid)
	}

	if b.debugger != nil && b.debugger.Process != nil {
		cmd.KillPid(b.debugger.Process.Pid)
	}

	err := b.buildBinary()
	if err != nil {
		cmd.PrintfDanger("Failed to build binary...")

		return
	}

	b.runBinary()
	b.runDebugger()
}

func (b *builder) buildBinary() error {
	args := []string{"build", "-C=" + b.config.Path, "-o", b.config.Name}

	args = append(args, b.config.BuildArgs...)

	buildCmd := exec.Command("go", args...)

	output, err := buildCmd.CombinedOutput()

	_ = output

	if err != nil {
		cmd.PrintfDanger("Build: " + err.Error())
		cmd.PrintfDanger(string(output))

		return err
	} else {
		cmd.PrintfSuccess("Binary built at " + b.config.Path + "/" + b.config.Name)
	}

	return nil
}

func (b *builder) runBinary() {
	if b.config.Address == "" {
		freePort, err := b.getListenerPort(8080)
		if err != nil {
			panic(cmd.FormatDanger("Failed to configure an app address. You may need to specify one in gadget.toml"))
		}

		b.config.Address = "localhost:" + strconv.Itoa(freePort)
	}

	b.runningBinary = exec.Command(b.config.Path+"/"+b.config.Name, b.config.Address)
	b.runningBinary.Dir = b.config.Path
	b.runningBinary.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdOut, err := b.runningBinary.StdoutPipe()
	if err != nil {
		cmd.PrintfDanger(err.Error())
	}

	stdErr, err := b.runningBinary.StderrPipe()
	if err != nil {
		cmd.PrintfDanger(err.Error())
	}

	err = b.runningBinary.Start()
	if err != nil {
		cmd.PrintfDanger(err.Error())
	} else {
		cmd.PrintfSuccess("Running on http://" + b.config.Address)
	}

	go cmd.PrintReadCloser(stdOut, func(line string) {
		print(cmd.FormatSuccess(line))
	})
	go cmd.PrintReadCloser(stdErr, func(line string) {
		print(cmd.FormatDanger(line))
	})

	// Wait for the initial output
	time.Sleep(1 * time.Second)
}

func (b *builder) runDebugger() {
	if b.port == 0 {
		port, err := b.getListenerPort(b.config.ListenPort)
		if err != nil {
			panic(cmd.FormatDanger(err.Error()))
		}

		b.port = port
	}

	debuggerArgs := []string{
		"attach",
		fmt.Sprintf("--listen=:%v", b.port),
		"--headless=true",
		"--api-version=2",
		strconv.Itoa(b.runningBinary.Process.Pid),
		"--accept-multiclient",
		"--continue",
	}

	// make sure out port is free
	isFree := b.waitForPortFree()
	if isFree != true {
		cmd.PrintfDanger("Failed waiting for port: %v", b.port)
	}

	b.debugger = exec.Command("dlv", debuggerArgs...)
	// something in the env if not set to empty, causing level=warning msg="CGO_CFLAGS already set, Cgo code could be optimized." layer=dlv
	b.debugger.Env = []string{}
	b.debugger.Dir = b.config.Path
	b.debugger.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	debugStdOut, err := b.debugger.StdoutPipe()
	if err != nil {
		cmd.PrintfDanger(err.Error())
	}

	debugStdErr, err := b.debugger.StderrPipe()
	if err != nil {
		cmd.PrintfDanger(err.Error())
	}

	err = b.debugger.Start()
	if err != nil {
		cmd.PrintfDanger(err.Error())
	}

	go cmd.PrintReadCloser(debugStdOut, func(line string) {
		print(cmd.FormatSuccess(line))
	})
	go cmd.PrintReadCloser(debugStdErr, func(line string) {
		print(cmd.FormatDanger(line))
	})

	// Wait for the initial output from running the debugger
	time.Sleep(1 * time.Second)
}

func (b *builder) getListenerPort(preferredPort int) (port int, err error) {
	listenPreferred, err := net.Listen("tcp", ":"+strconv.Itoa(preferredPort))
	if err != nil {
		cmd.PrintfWarning(err.Error() + "\nTrying alternate port")
	} else {
		defer listenPreferred.Close()

		return listenPreferred.Addr().(*net.TCPAddr).Port, err
	}

	listenAlt, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	defer listenAlt.Close()

	return listenAlt.Addr().(*net.TCPAddr).Port, err
}

func (b *builder) waitForPortFree() bool {
	maxWait := time.NewTimer(5 * time.Second)
	increment := time.NewTicker(500 * time.Millisecond)

	defer maxWait.Stop()
	defer increment.Stop()

	for {
		select {
		case <-maxWait.C:
			return false

		case <-increment.C:
			_, err := net.Dial("tcp", ":"+strconv.Itoa(b.port))
			if err != nil {
				// no connection, probably open port
				return true
			}
		}
	}
}
