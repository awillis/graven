package main

import (
	"github.com/urfave/cli"
	"os"
	"fmt"
	"github.com/cbegin/graven/commands"
	"github.com/cbegin/graven/domain"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Author = "Clinton Begin"
	app.Copyright = "Clinton Begin"
	app.Description = "A build automation tool for Go."
	app.Name = "graven"
	app.Usage = "A build automation tool for Go."

	app.Commands = []cli.Command{
		commands.BuildCommand,
	}

	p, err := domain.FindProject()
	if err != nil {
		fmt.Println("Could not find project.yaml in current or parent path.")
		return
	}

	app.Metadata = map[string]interface{}{"project":p}
	//fmt.Printf("%+v %v\n", p, err)

	// new -- initializes new directory and project.yaml
	// clean -- deletes target dir
	// test -- runs tests with flags, coverage etc.
	// build -- compiles all platforms with flags
	// package -- clean, test, build, package archives
	// deploy -- deploy one artifact to one repository
	// release [major|minor|patch] package, deploy each archive
	// docker?


	app.Run(os.Args)
}