package sqliteffi

import (
	"testing"

	"go.uber.org/goleak"
)

// VerifyTestMain fails the package test run if tests leave goroutines behind.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
