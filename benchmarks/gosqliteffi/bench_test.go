package gosqliteffi_test

import (
	"database/sql"
	"testing"

	"github.com/Aswanidev-vs/gosqlite-ffi/sqliteffi"
	"github.com/Aswanidev-vs/gosqlite-ffi/benchmarks/internal/suite"
)

func init() { sql.Register("gosqliteffi-bench", sqliteffi.Driver{}) }

func open(path string) (*sql.DB, error) { return sql.Open("gosqliteffi-bench", path) }

func BenchmarkDriver(b *testing.B) { suite.Benchmarks(b, open) }
func TestConcurrentStress(t *testing.T) { suite.Stress(t, open) }
