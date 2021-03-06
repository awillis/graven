package commands

import (
	"os/exec"
	"os"
	"fmt"
	"path"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/urfave/cli"
	"github.com/cbegin/graven/domain"
)

var BuildCommand = cli.Command{
	Name:        "build",
	Usage:       "Builds the current project",
	Action: build,
}

func build(c *cli.Context) error {
	if err := clean(c); err != nil {
		return err
	}

	var merr error
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	project, err := domain.FindProject()
	if err != nil {
		return err
	}

	for _, artifact := range project.Artifacts {
		a := artifact
		wg.Add(len(artifact.Targets))
		for _, target := range artifact.Targets {
			t := target
			go func() {
				defer wg.Done()
				err := buildTarget(project, &a, &t)
				if err != nil {
					mutex.Lock()
					merr = multierror.Append(merr, err)
					mutex.Unlock()
				}

			}()
		}
	}
	wg.Wait()

	return merr
}

func buildTarget(project *domain.Project, artifact *domain.Artifact, target *domain.Target) error {
	classifiedPath := project.TargetPath(artifact.Classifier)
	if _, err := os.Stat(classifiedPath); os.IsNotExist(err) {
		os.Mkdir(classifiedPath, 0755)
	}

	for _, resource := range artifact.Resources {
		resourcePath := project.ProjectPath(resource)
		baseProjectPath := project.ProjectPath()
		if len(resourcePath[len(baseProjectPath):]) < 1 {
			return fmt.Errorf("Resource path cannot be the entire project folder: %s", resourcePath)
		}
		err := domain.CopyDir(resourcePath, classifiedPath)
		if err != nil {
			return err
		}
	}

	return runBuildCommand(classifiedPath, project, target)
}

func runBuildCommand(classifiedPath string, project *domain.Project, target *domain.Target) error {
	var c *exec.Cmd
	if target.Flags == "" {
		c = exec.Command("go", "build", "-o", path.Join(classifiedPath, target.Executable), target.Package)
	} else {
		c = exec.Command("go", "build", "-o", path.Join(classifiedPath, target.Executable), target.Flags, target.Package)
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Dir = project.ProjectPath()

	environment := []string{}
	for k, v := range target.Environment {
		environment = append(environment, fmt.Sprintf("%s=%s", k, v))
	}
	gopath, _ := os.LookupEnv("GOPATH")
	environment = append(environment, fmt.Sprintf("%s=%s", "GOPATH", gopath))
	c.Env = environment

	err := c.Run()
	if err != nil {
		return fmt.Errorf("Build error. %v", err)
	}

	if !c.ProcessState.Success() {
		return fmt.Errorf("Build command exited in an error state. %v", c)
	}
	return err
}