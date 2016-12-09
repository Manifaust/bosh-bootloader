package terraform

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var tempDir func(dir, prefix string) (string, error) = ioutil.TempDir
var writeFile func(file string, data []byte, perm os.FileMode) error = ioutil.WriteFile
var readFile func(filename string) ([]byte, error) = ioutil.ReadFile

type Executor struct {
	cmd terraformCmd
}

type terraformCmd interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

func NewExecutor(cmd terraformCmd) Executor {
	return Executor{cmd: cmd}
}

func (e Executor) Apply(credentials, envID, projectID, zone, region, template, prevTFState string) (string, error) {
	tempDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	credentialsPath := filepath.Join(tempDir, "credentials.json")
	err = writeFile(credentialsPath, []byte(credentials), os.ModePerm)
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(tempDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", err
	}

	if prevTFState != "" {
		err = writeFile(filepath.Join(tempDir, "terraform.tfstate"), []byte(prevTFState), os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	args := []string{"apply"}
	args = append(args, makeVar("project_id", projectID)...)
	args = append(args, makeVar("env_id", envID)...)
	args = append(args, makeVar("region", region)...)
	args = append(args, makeVar("zone", zone)...)
	args = append(args, makeVar("credentials", credentialsPath)...)
	err = e.cmd.Run(os.Stdout, tempDir, args)
	if err != nil {
		return "", err
	}

	tfState, err := readFile(filepath.Join(tempDir, "terraform.tfstate"))
	if err != nil {
		return "", err
	}

	return string(tfState), nil
}

func (e Executor) Destroy(credentials, envID, projectID, zone, region, template, tfState string) error {
	tempDir, err := tempDir("", "")
	if err != nil {
		return err
	}

	credentialsPath := filepath.Join(tempDir, "credentials.json")
	err = writeFile(credentialsPath, []byte(credentials), os.ModePerm)
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(tempDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(tempDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return err
	}

	args := []string{"destroy", "-force"}
	args = append(args, makeVar("project_id", projectID)...)
	args = append(args, makeVar("env_id", envID)...)
	args = append(args, makeVar("region", region)...)
	args = append(args, makeVar("zone", zone)...)
	args = append(args, makeVar("credentials", credentialsPath)...)
	err = e.cmd.Run(os.Stdout, tempDir, args)
	if err != nil {
		return err
	}

	return nil
}

func makeVar(name string, value string) []string {
	return []string{"-var", fmt.Sprintf("%s=%s", name, value)}
}
