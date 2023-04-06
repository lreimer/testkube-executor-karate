package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// setup
	tempDir := os.TempDir()
	os.Setenv("RUNNER_DATADIR", tempDir)

	t.Run("basic Karate test feature", func(t *testing.T) {
		// given
		setupKarateJar(t)

		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.TestType = "karate/feature"
		writeTestContent(t, tempDir, "../../examples/karate-success.feature")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusPassed)
		assert.Len(t, result.Steps, 2)
	})

	t.Run("basic Karate failure feature", func(t *testing.T) {
		// given
		setupKarateJar(t)

		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.TestType = "karate/feature"
		writeTestContent(t, tempDir, "../../examples/karate-failure.feature")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusFailed)
		assert.Len(t, result.Steps, 1)
	})

	t.Run("project Karate test without repo path", func(t *testing.T) {
		// given
		setupKarateJar(t)

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
		setupKarateJar(t)

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

func setupKarateJar(t *testing.T) {
	localJar, err := filepath.Abs("../../karate.jar")
	if err != nil {
		assert.FailNow(t, "can't locate karate.jar, please run `make install-karate`")
	}
	KarateJarPath = localJar
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
