package integration

import (
	"os"
	"server/tests/integration/testdb"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.Terminate()
	os.Exit(code)
}
