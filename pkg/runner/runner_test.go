package runner

import (
	"os"
	"testing"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// setup
	tempDir := os.TempDir()
	os.Setenv("RUNNER_DATADIR", tempDir)

	t.Run("runner should run test based on execution data", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("hello I'm test content")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusPassed)
	})

}
