package cmd

import (
	"bufio"
	"github.com/clanko/scaffold"
	"os"
)

type command interface {
	execute(input *bufio.Scanner, args []string)
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

	newTemplatePath := templatesPath + "/" + path

	// check if template path exists.
	_, err = os.Stat(newTemplatePath)
	if err != nil {
		FormatDanger("Failed to locate template directory: %v", path)

		return
	}

	scaf, err := scaffold.Init(newTemplatePath)
	if err != nil {
		FormatDanger("Failed to initialize scaffold. Does %v/scaffold.toml exist?", newTemplatePath)
	}

	// get all tokens. Those that don't have a ValueToken set, prompt for value
	tokens := scaf.GetTokens()
	for _, token := range tokens {
		if token.ValueToken == "" {
			// prompt for token value
			for token.Value == "" {
				print(FormatInfo("Enter value for token %v: ", token.Name))

				input.Scan()
				if input.Text() != "" {
					scaf.RegisterTokenValue(token.Name, input.Text())
					token.Value = input.Text()
				}
			}
		}
	}

	// make at destination
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	scaf.OnMake(func(created string) {
		PrintfSuccess("Created: %v", created)
	})

	err = scaf.Make(wd)
	if err != nil {
		panic(err)
	}
}
