package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigOk(t *testing.T) {
	simpleConfig := `---
repo_url: www.example.com
tasks:
  test:
    description: Hello World
    category: test
    env:
      IS_TEST: "true"
    steps:
      - cmd: ls
        args: -la
`
	reader := strings.NewReader(simpleConfig)
	config, err := Load(reader)
	assert.NoError(t, err, "Unexpected error while loading config from file")
	assert.Equal(t, "www.example.com", config.RepoUrl)
}

func TestLoadConfigFail_InvalidYamlSchema(t *testing.T) {
	reader := strings.NewReader("foo bar")
	config, err := Load(reader)
	assert.ErrorContains(t, err, "cannot unmarshal")
	assert.Empty(t, config)
}
