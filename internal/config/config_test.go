package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -v -cover ./pkg/config/...

// go test -v -cover -run TestLoadConfig_Valid ./pkg/config

func TestLoadConfig_Valid(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	tempConfigFile, err := os.Create("test-config.yaml")
	require.NoError(err)
	defer os.Remove(tempConfigFile.Name())

	configData := `
directory: "/tmp"
check_frequency: 10
reporting_api: "http://localhost/api"
socket_path: "/tmp/socket"
mongo_uri: "mongodb://user:password@localhost:27017"
`
	_, err = tempConfigFile.Write([]byte(configData))
	require.NoError(err)
	tempConfigFile.Close()

	config, err := LoadConfig("test-config", "./")
	assert.NoError(err)
	assert.NotNil(config)
	assert.Equal("/tmp", config.Directory)
	assert.Equal(10, config.CheckFrequency)
	assert.Equal("http://localhost/api", config.ReportingAPI)
	assert.Equal("/tmp/socket", config.SocketPath)
	assert.Equal("9000", config.HTTPPort)

}

// go test -v -cover -run TestLoadConfig_InValidConfig ./pkg/config

func TestLoadConfig_InValidConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	tempConfigFile, err := os.Create("test-config.yaml")
	require.NoError(err)
	defer os.Remove(tempConfigFile.Name())

	configData := `
directory: "/tmp"
check_frequency: 0
reporting_api: "http://localhost/api"
`

	_, err = tempConfigFile.Write([]byte(configData))
	require.NoError(err)
	tempConfigFile.Close()

	_, err = LoadConfig("test-config", "./")
	assert.Error(err)
	assert.Contains(err.Error(), "error validating config")
}

// go test -v -cover -run TestLoadConfig_InValidDirectory ./pkg/config

func TestLoadConfig_InValidDirectory(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	tempConfigFile, err := os.Create("test-config.yaml")
	require.NoError(err)
	defer os.Remove(tempConfigFile.Name())

	configData := `
directory: ""
check_frequency: 10
reporting_api: "http://localhost/api"
socket_path: "/tmp/socket"
mongo_uri: "mongodb://user:password@localhost:27017"
`

	_, err = tempConfigFile.Write([]byte(configData))
	require.NoError(err)
	tempConfigFile.Close()

	_, err = LoadConfig("test-config", "./")
	assert.Error(err)
	assert.Contains(err.Error(), "invalid directory format")
}
