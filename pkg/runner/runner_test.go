package runner

import (
	"fmt"
	"github.com/kubeshop/testkube/pkg/executor/secret"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/stretchr/testify/assert"
)

type executorCallArgs struct {
	dir       string
	command   string
	arguments []string
}

func TestRun(t *testing.T) {
	// setup
	setupKarateJar(t)

	tempDir := os.TempDir()
	os.Setenv("RUNNER_DATADIR", tempDir)

	t.Run("basic Karate test feature", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.TestType = "karate/feature"
		writeTestContent(t, tempDir, "../../examples/karate-success.feature")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, testkube.ExecutionStatusPassed, result.Status)
		assert.Len(t, result.Steps, 2)
	})

	t.Run("basic Karate failure feature", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.TestType = "karate/feature"
		writeTestContent(t, tempDir, "../../examples/karate-failure.feature")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, testkube.ExecutionStatusFailed, result.Status)
		assert.Len(t, result.Steps, 1)
	})

	t.Run("project Karate test without repo path", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		repository := testkube.NewGitRepository("http://not-used/", "main")
		execution.Content = &testkube.TestContent{Type_: "git-dir", Repository: repository}
		execution.TestType = "karate/project"
		execution.Args = []string{"."} // tell karate to look for features in start dir
		writeTestContentProject(t, execution.Content.Repository, tempDir, "../../examples/project")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusPassed)
		assert.Len(t, result.Steps, 2)
	})

	t.Run("project Karate test with repo path", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		repository := testkube.NewGitRepository("http://not-used/", "main").WithPath("my-dir")
		execution.Content = &testkube.TestContent{Type_: "git-dir", Repository: repository}
		execution.TestType = "karate/project"
		execution.Args = []string{"."} // tell karate to look for features in start dir
		writeTestContentProject(t, execution.Content.Repository, tempDir, "../../examples/project")

		// when
		result, err := runner.Run(*execution)
		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusPassed)
		assert.Len(t, result.Steps, 2)
	})
}

func TestExecutorRunCall(t *testing.T) {
	// setup
	tempDir := os.TempDir()
	os.Setenv("RUNNER_DATADIR", tempDir)
	karateJarPath = "/home/karate/karate.jar"
	repoContent := &testkube.TestContent{Type_: "git-dir", Repository: testkube.NewGitRepository("http://not-used", "main")}
	repoWithPathContent := &testkube.TestContent{Type_: "git-dir", Repository: testkube.NewGitRepository("http://not-used", "main").WithPath("my-path")}

	tests := []struct {
		name          string
		execution     *testkube.Execution
		expectedArgs  []string
		expectedPath  string
		expectedError string
	}{
		{
			name:         "feature uses default args",
			execution:    newExecution("karate/feature", testkube.NewStringTestContent("")),
			expectedArgs: []string{"-jar", "/home/karate/karate.jar", "-f", "junit:xml", "test-content.feature"},
		},
		{
			name:         "project uses default args",
			execution:    newExecution("karate/project", repoContent, "."),
			expectedArgs: []string{"-jar", "/home/karate/karate.jar", "-f", "junit:xml", "."},
		},
		{
			name:         "project start execution in repo path when specified",
			execution:    newExecution("karate/project", repoWithPathContent, "features"),
			expectedArgs: []string{"-jar", "/home/karate/karate.jar", "-f", "junit:xml", "features"},
			expectedPath: repoWithPathContent.Repository.Path,
		},
		{
			name:         "standalone uses only specified args",
			execution:    newExecution("karate/standalone", repoContent, "-Dsomeurl=https://google.com", "-cp", "/home/karate/karate.jar", "com.intuit.karate.Main", "-f", "junit:xml", "my-path"),
			expectedArgs: []string{"-Dsomeurl=https://google.com", "-cp", "/home/karate/karate.jar", "com.intuit.karate.Main", "-f", "junit:xml", "my-path"},
		},
		{
			name:          "standalone throws error on missing args",
			execution:     newExecution("karate/standalone", repoContent),
			expectedError: "args are required for test type karate/standalone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			var actualCall executorCallArgs
			executorRun = func(dir string, command string, envMngr secret.Manager, arguments ...string) (out []byte, err error) {
				actualCall = executorCallArgs{
					dir:       dir,
					command:   command,
					arguments: arguments,
				}
				return []byte{}, fmt.Errorf("only interested in executor call")
			}

			// when
			result, err := NewRunner().Run(*tt.execution)

			// then
			assert.NoError(t, err)
			if len(tt.expectedError) > 0 {
				assert.Equal(t, tt.expectedError, result.ErrorMessage)
				return
			}
			assert.Equal(t, testkube.ExecutionStatusFailed, result.Status)
			assert.NotNil(t, actualCall)

			if len(tt.expectedPath) > 0 {
				assert.True(t, strings.HasSuffix(actualCall.dir, tt.expectedPath))
			}
			assert.Equal(t, "java", actualCall.command)
			assert.Equal(t, tt.expectedArgs, actualCall.arguments)
		})
	}
}

func newExecution(testType string, content *testkube.TestContent, arguments ...string) *testkube.Execution {
	execution := testkube.NewQueuedExecution()
	execution.TestType = testType
	execution.Content = content
	execution.Args = arguments
	return execution
}

func setupKarateJar(t *testing.T) {
	localJar, err := filepath.Abs("../../karate.jar")
	if err != nil {
		assert.FailNow(t, "can't locate karate.jar, please run `make install-karate`")
	}
	karateJarPath = localJar
}

func writeTestContent(t *testing.T, dir string, file string) {
	featureFile, err := os.ReadFile(file)
	if err != nil {
		assert.FailNow(t, "Unable to read Karate feature file")
	}

	err = os.WriteFile(filepath.Join(dir, "test-content"), featureFile, 0644)
	if err != nil {
		assert.FailNow(t, "Unable to write Karate feature file as test content")
	}
}

func writeTestContentProject(t *testing.T, repo *testkube.Repository, targetDir string, sourceDir string) {

	repoDir := filepath.Join(targetDir, "repo")
	if len(repo.Path) > 0 {
		repoDir = filepath.Join(repoDir, repo.Path)
	}

	err := os.RemoveAll(repoDir)
	if err != nil {
		assert.FailNow(t, "Unable to clear repoDir")
	}

	_, err = os.Stat(repoDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(repoDir, 0777)
		if err != nil {
			assert.FailNow(t, "Unable to create repoDir")
		}
	}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		assert.FailNow(t, "Unable to read sourceDir")
	}
	for _, f := range files {
		if !f.IsDir() {
			file := filepath.Join(sourceDir, f.Name())
			featureFile, err := os.ReadFile(file)
			if err != nil {
				assert.FailNow(t, "Unable to read Karate feature file")
			}

			err = os.WriteFile(filepath.Join(repoDir, f.Name()), featureFile, 0644)
			if err != nil {
				assert.FailNow(t, "Unable to write Karate feature file as test content")
			}
		}
	}
}
