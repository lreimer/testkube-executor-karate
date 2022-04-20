package runner

import (
	"io/ioutil"
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
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.TestType = "karate/feature"
		writeTestContent(t, tempDir, "../../examples/karate-test.feature")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusPassed)
	})
}

func writeTestContent(t *testing.T, dir string, file string) {
	featureFile, err := ioutil.ReadFile(file)
	if err != nil {
		assert.FailNow(t, "Unable to read Karate feature file")
	}

	err = ioutil.WriteFile(filepath.Join(dir, "test-content"), featureFile, 0644)
	if err != nil {
		assert.FailNow(t, "Unable to write Karate feature file as test content")
	}
}
