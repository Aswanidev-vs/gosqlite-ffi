package ncruces_test

import (
	"database/sql"
	"testing"

	"github.com/Aswanidev-vs/gosqlite-ffi/benchmarks/internal/suite"
	_ "github.com/ncruces/go-sqlite3/driver"
)

func open(path string) (*sql.DB, error) { return sql.Open("sqlite3", path) }

func BenchmarkDriver(b *testing.B) { suite.Benchmarks(b, open) }
func TestConcurrentStress(t *testing.T) { suite.Stress(t, open) }
