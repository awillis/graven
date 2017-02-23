package commands

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"os"
	"strings"
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli"
	"github.com/cbegin/graven/domain"
	"gopkg.in/yaml.v2"
	"path"
)

var InitCommand = cli.Command{
	Name: "init",
	Usage:       "Initializes a project directory",
	Action: initialize,
}

type ClassifierTemplate struct {
	Classifier   string
	Archive      string
	Extension    string
	OS           string
	Architecture string
}

type PackagePath struct {
	Package string
	Path string
}

var (
	darwinTemplate = ClassifierTemplate{
		Classifier: "darwin",
		Archive: "tgz",
		Extension: "",
		OS: "darwin",
		Architecture: "amd64",
	}
	linuxTemplate = ClassifierTemplate{
		Classifier: "linux",
		Archive: "tar.gz",
		Extension: "",
		OS: "linux",
		Architecture: "amd64",
	}
	winTemplate = ClassifierTemplate{
		Classifier: "win",
		Archive: "zip",
		Extension: ".exe",
		OS: "windows",
		Architecture: "amd64",
	}
	templates = []ClassifierTemplate{
		darwinTemplate,
		linuxTemplate,
		winTemplate,
	}
)

func initialize(c *cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectPath := path.Join(wd, "projectx.yaml")

	packages := &[]PackagePath{}

	if err = filepath.Walk(wd, getInitializeWalkerFunc(wd, packages)); err != nil {
		return err
	}

	artifacts := []domain.Artifact{}

	// TODO: determine default name (better than "app"

	for _, template := range templates {
		targets := []domain.Target{}

		for i, p := range *packages {
			if p.Package == "main" {
				pkg := fmt.Sprintf(".%v", p.Path)
				targets = append(targets, domain.Target{
					Executable: fmt.Sprintf("%v%v%v", "app", i, template.Extension),
					Package: pkg,
					Flags: "",
					Environment:map[string]string{
						"GOOS":template.OS,
						"GOARCH":template.Architecture,
					},
				})
			}
		}

		artifacts = append(artifacts, domain.Artifact{
			Classifier:template.Classifier,
			Resources: []string{},
			Archive:template.Archive,
			Targets:targets,
		})
	}

	newProject := &domain.Project{}
	newProject.Name = "github.com/org/myProject"
	newProject.Version = "0.0.1"
	newProject.Artifacts = artifacts

	bytes, err := yaml.Marshal(newProject)
	if err != nil {
		return err
	}
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return fmt.Errorf("%v already exists. No changes made.", projectPath)
	}

	if err := ioutil.WriteFile(projectPath, bytes, 0655); err != nil {
		return err
	}

	return err
}

func getInitializeWalkerFunc(basePath string, packages *[]PackagePath) filepath.WalkFunc {
	fs := token.NewFileSet()
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			subDir := path[len(basePath):]
			subDirParts := strings.Split(subDir, string(filepath.Separator))
			matches, _ := filepath.Glob(filepath.Join(path, "*.go"));
			if len(matches) > 0 && !contains(subDirParts, map[string]struct{}{
				"vendor":struct{}{},
				"target":struct{}{},
				".git":struct{}{}}) {
				ast, err := parser.ParseDir(fs, path, nil, parser.PackageClauseOnly)
				if err != nil {
					fmt.Println(err)
					return err
				}
				for _, v := range ast {
					shortPath := path[len(basePath):]
					*packages = append(*packages, PackagePath{
						Package: v.Name,
						Path: shortPath,
					})
				}
			}
		}
		return nil
	}
}
