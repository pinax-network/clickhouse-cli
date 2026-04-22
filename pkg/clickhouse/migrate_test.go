package clickhouse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFiles(t *testing.T, names []string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func TestParseMigrationDirectory_OrderedAndIgnoresNonMatching(t *testing.T) {
	dir := writeFiles(t, []string{
		"2_second.sql",
		"1_first.sql",
		"3_third.sql",
		"README.md",
		"not_a_migration.sql",
	})

	got, err := parseMigrationDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 migrations, got %d: %+v", len(got), got)
	}
	for i, mf := range got {
		if mf.Seq != i+1 {
			t.Errorf("index %d: expected seq %d, got %d", i, i+1, mf.Seq)
		}
	}
}

func TestParseMigrationDirectory_EmptyReturnsNoError(t *testing.T) {
	dir := t.TempDir()
	got, err := parseMigrationDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %+v", got)
	}
}

func TestParseMigrationDirectory_Duplicate(t *testing.T) {
	dir := writeFiles(t, []string{"1_a.sql", "1_b.sql"})
	_, err := parseMigrationDirectory(dir)
	if err == nil {
		t.Fatal("expected error for duplicate sequence")
	}
	if !strings.Contains(err.Error(), "duplicate sequence number") {
		t.Errorf("expected duplicate error, got: %v", err)
	}
}

func TestParseMigrationDirectory_Gap(t *testing.T) {
	dir := writeFiles(t, []string{"1_a.sql", "3_c.sql"})
	_, err := parseMigrationDirectory(dir)
	if err == nil {
		t.Fatal("expected error for gap in sequence")
	}
	if !strings.Contains(err.Error(), "missing sequence number") {
		t.Errorf("expected missing sequence error, got: %v", err)
	}
}

func TestParseMigrationDirectory_DoesNotStartAtOne(t *testing.T) {
	dir := writeFiles(t, []string{"2_a.sql", "3_b.sql"})
	_, err := parseMigrationDirectory(dir)
	if err == nil {
		t.Fatal("expected error when sequence does not start at 1")
	}
}

func TestNewMigration_RejectsInvalidIdentifiers(t *testing.T) {
	cases := []string{
		"onlyone",
		"a.b.c",
		"bad-db.table",
		"db.bad-table",
		"db.table;DROP",
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			if _, err := NewMigration(nil, "/tmp", tc, false, false); err == nil {
				t.Fatalf("expected error for %q", tc)
			}
		})
	}
}

func TestNewMigration_AcceptsValidIdentifiers(t *testing.T) {
	m, err := NewMigration(nil, "/tmp", "my_db.schema_migrations", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.database != "my_db" || m.table != "schema_migrations" {
		t.Errorf("unexpected parse: %q.%q", m.database, m.table)
	}
	if !m.clusterMode {
		t.Error("clusterMode should be propagated")
	}
}
