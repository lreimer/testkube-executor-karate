package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	junit "github.com/joshdk/go-junit"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/kubeshop/testkube/pkg/executor"
	"github.com/kubeshop/testkube/pkg/executor/output"
	"github.com/kubeshop/testkube/pkg/executor/runner"
	"github.com/kubeshop/testkube/pkg/executor/secret"
)

type Params struct {
	Datadir string // RUNNER_DATADIR
}

func NewRunner() *KarateRunner {
	return &KarateRunner{
		params: Params{
			Datadir: os.Getenv("RUNNER_DATADIR"),
		},
	}
}

type KarateRunner struct {
	params Params
}

const FEATURE_TYPE = "feature"
const PROJECT_TYPE = "project"

func (r *KarateRunner) Run(execution testkube.Execution) (result testkube.ExecutionResult, err error) {
	// check that the datadir exists
	_, err = os.Stat(r.params.Datadir)
	if errors.Is(err, os.ErrNotExist) {
		return result, err
	}

	// prepare the arguments, always use JUnit XML report
	args := []string{"-f", "junit:xml"}
	args = append(args, execution.Args...)

	var directory string
	karateType := strings.Split(execution.TestType, "/")[1]
	if karateType == FEATURE_TYPE && execution.Content.IsFile() {
		directory = r.params.Datadir
		_ = os.Rename(filepath.Join(directory, "test-content"), filepath.Join(directory, "test-content.feature"))
		args = append(args, "test-content.feature")
	} else if karateType == PROJECT_TYPE && execution.Content.IsDir() {
		directory = filepath.Join(r.params.Datadir, "repo")
		// feature file needs to be part of args
	} else {
		return result.Err(fmt.Errorf("unsupported content for test type %s", execution.TestType)), nil
	}

	envManager := secret.NewEnvManagerWithVars(execution.Variables)
	envManager.GetVars(execution.Variables)
	// simply set the ENVs to use during execution
	for _, env := range execution.Variables {
		os.Setenv(env.Name, env.Value)
	}

	// convert executor env variables to runner env variables
	for key, value := range execution.Envs {
		os.Setenv(key, value)
	}

	output.PrintEvent("Running", directory, "karate", args)
	output, err := executor.Run(directory, "karate", envManager, args...)
	output = envManager.Obfuscate(output)

	if err == nil {
		result.Status = testkube.ExecutionStatusPassed
	} else {
		result.Status = testkube.ExecutionStatusFailed
		result.ErrorMessage = err.Error()
		if strings.Contains(result.ErrorMessage, "exit status 1") {
			result.ErrorMessage = "there are test failures"
		} else {
			// ZAP was unable to run at all, wrong args?
			return result, nil
		}
	}

	result.Output = string(output)
	result.OutputType = "text/plain"

	junitReportPath := filepath.Join(directory, "target", "karate-reports")
	err = filepath.Walk(junitReportPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".xml" {
			suites, _ := junit.IngestFile(path)
			for _, suite := range suites {
				for _, test := range suite.Tests {
					result.Steps = append(
						result.Steps,
						testkube.ExecutionStepResult{
							Name:     test.Name,
							Duration: test.Duration.String(),
							Status:   testStepStatus(test.Status),
						})
				}
			}
		}

		return nil
	})

	return result, err
}

// GetType returns runner type
func (r *KarateRunner) GetType() runner.Type {
	return runner.TypeMain
}

func testStepStatus(in junit.Status) (out string) {
	switch string(in) {
	case "passed":
		return string(testkube.PASSED_ExecutionStatus)
	case "skipped":
		// we could ignore this otherwise
		return string(testkube.PASSED_ExecutionStatus)
	default:
		return string(testkube.FAILED_ExecutionStatus)
	}
}
