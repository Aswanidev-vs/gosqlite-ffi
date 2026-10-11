package modernc_test

import (
	"database/sql"
	"testing"

	"github.com/Aswanidev-vs/gosqlite-ffi/benchmarks/internal/suite"
	_ "modernc.org/sqlite"
)

func open(path string) (*sql.DB, error) { return sql.Open("sqlite", path) }

func BenchmarkDriver(b *testing.B) { suite.Benchmarks(b, open) }
func TestConcurrentStress(t *testing.T) { suite.Stress(t, open) }
