package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	os.Setenv("SERVER_PORT", "1")
	defer os.Unsetenv("SERVER_PORT")
	config := New()
	port := config.GetInt("server.port")
	if port != 1 {
		t.Errorf("expected port %d, got %d", 1, port)
	}
}
