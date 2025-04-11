# Gadget CLI

## Application Development Utility
* Live reload, with the ability to include/exclude paths. Facilitates development where local packages may be updated, prompting app rebuild.
* Attaches Delve to the running process for debugging.
* Setup templates for code generation using the scaffold package.

## Building
* go build -o gadget github.com/clanko/gadget

## Installation
* sudo mv gadget /usr/bin/gadget

## Usage
* Run `gadget dev` to build and debug your application, and watch for file changes that signal gadget to reload. Gadget then enters the gadget shell awaiting commands.
* Running `gadget` will simply enter the gadget shell.
* The best way to run gadget, is with a gadget.toml configuration. You can generate one by entering gadget shell and running: `make gadget-config` 
* Run `gadget -v 1 dev` for verbose output.

## Interactive Shell Commands
- build
  - Builds application binary
  - - Stops previously running binary and debugger
- run
  - Builds and runs binary
  - - Stops previously running binary and debugger
- debug
  - Builds and runs binary and debugger
  - - Stops previously running binary and debugger
- dev
  - Builds and runs binary and debugger, and watches files
  - - Stops previously running binary and debugger
- watch
  - Starts file watcher
- unwatch
  - Stops file watcher
- make gadget-config
  - Generates a gadget.toml configuration file in the current directory.
- make {template-name}
  - Generate files based on a Scaffold template

## Template Generation
* Gadget uses Scaffold under the hood, to generate templated files. These templates should be placed in folders in ~/.clanko-gadget-cli/make-templates/
* Read more about Scaffold here: https://github.com/clanko/scaffold
* After setting up your template, from gadget shell enter `make your-template`

## Debugging
### GoLand
* Go to Edit Configurations and create a Go Remote Configuration.
* Make sure host and port match the values shown by the API server when running gadget dev.