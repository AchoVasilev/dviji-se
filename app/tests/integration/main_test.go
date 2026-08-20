package integration

import (
	"os"
	"server/tests/integration/testdb"
	"testing"
)

func TestMain(m *testing.M) {
	// The public auth routes are only registered when registration is enabled,
	// and that now defaults to off. Config is a lazily loaded singleton, so it
	// has to be set before any test reads it.
	os.Setenv("ALLOW_REGISTRATION", "true")

	code := m.Run()
	testdb.Terminate()
	os.Exit(code)
}
