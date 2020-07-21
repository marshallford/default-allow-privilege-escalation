package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigError(t *testing.T) {
	invalidYaml := []byte("foobar")
	dir, err := ioutil.TempDir("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.Setenv("CONFIGPATH", dir)
	defer os.Unsetenv("CONFIGPATH")

	tmpConfig := filepath.Join(dir, "config.yaml")
	if err := ioutil.WriteFile(tmpConfig, invalidYaml, 0666); err != nil {
		t.Fatal(err)
	}

	_, err = New()
	expectedError := "While parsing config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `foobar` into map[string]interface {}"
	if expectedError != err.Error() {
		t.Errorf("expected error %s, got %s", expectedError, err.Error())
	}
}

func TestNewConfig(t *testing.T) {
	os.Setenv("SERVER_PORT", "1")
	defer os.Unsetenv("SERVER_PORT")
	config, _ := New()
	port := config.GetInt("server.port")
	if port != 1 {
		t.Errorf("expected port %d, got %d", 1, port)
	}
}
