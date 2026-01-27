// Package signals provides Gurgeh-specific signal emission and persistence.
package signals

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/mistakeknot/autarch/pkg/signals"
)

const maxDescriptionLen = 4096

const createTableSQL = `
CREATE TABLE IF NOT EXISTS signals (
	id TEXT PRIMARY KEY,
	spec_id TEXT NOT NULL,
	type TEXT NOT NULL,
	affected_field TEXT NOT NULL,
	severity TEXT NOT NULL,
	source TEXT NOT NULL DEFAULT 'gurgeh',
	title TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL,
	created_at TEXT NOT NULL,
	dismissed_at TEXT,
	UNIQUE(spec_id, type, affected_field)
);
CREATE INDEX IF NOT EXISTS idx_signals_spec_active ON signals(spec_id) WHERE dismissed_at IS NULL;
`

// Store manages durable signal persistence in .gurgeh/signals/signals.db (SQLite WAL mode).
type Store struct {
	db *sql.DB
}

// NewStore opens or creates the signal database at projectRoot/.gurgeh/signals/signals.db.
func NewStore(projectRoot string) (*Store, error) {
	dir := filepath.Join(projectRoot, ".gurgeh", "signals")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create signals dir: %w", err)
	}

	dbPath := filepath.Join(dir, "signals.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open signals db: %w", err)
	}

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("init signals schema: %w", err)
	}

	// Set file permissions to 0600 (owner read/write only).
	if err := os.Chmod(dbPath, 0600); err != nil {
		// Non-fatal — log but continue.
	}

	return &Store{db: db}, nil
}

// Emit persists a signal. Uses INSERT OR IGNORE for deduplication via the
// UNIQUE constraint on (spec_id, type, affected_field).
func (s *Store) Emit(sig signals.Signal) error {
	desc := sig.Detail
	if len(desc) > maxDescriptionLen {
		desc = desc[:maxDescriptionLen]
	}

	_, err := s.db.Exec(`INSERT OR IGNORE INTO signals
		(id, spec_id, type, affected_field, severity, source, title, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sig.ID, sig.SpecID, string(sig.Type), sig.AffectedField,
		string(sig.Severity), sig.Source, sig.Title, desc,
		sig.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("emit signal: %w", err)
	}
	return nil
}

// Active returns all non-dismissed signals for a spec.
func (s *Store) Active(specID string) ([]signals.Signal, error) {
	rows, err := s.db.Query(`SELECT id, spec_id, type, affected_field, severity, source, title, description, created_at
		FROM signals WHERE spec_id = ? AND dismissed_at IS NULL ORDER BY created_at`, specID)
	if err != nil {
		return nil, fmt.Errorf("query active signals: %w", err)
	}
	defer rows.Close()

	return scanSignals(rows)
}

// ActiveAll returns all non-dismissed signals across all specs.
func (s *Store) ActiveAll() ([]signals.Signal, error) {
	rows, err := s.db.Query(`SELECT id, spec_id, type, affected_field, severity, source, title, description, created_at
		FROM signals WHERE dismissed_at IS NULL ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("query all active signals: %w", err)
	}
	defer rows.Close()

	return scanSignals(rows)
}

// Dismiss marks a signal as dismissed by setting dismissed_at.
func (s *Store) Dismiss(id string) error {
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(`UPDATE signals SET dismissed_at = ? WHERE id = ? AND dismissed_at IS NULL`, now, id)
	if err != nil {
		return fmt.Errorf("dismiss signal: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("signal %s not found or already dismissed", id)
	}
	return nil
}

// Count returns the number of active (non-dismissed) signals for a spec.
func (s *Store) Count(specID string) int {
	var count int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM signals WHERE spec_id = ? AND dismissed_at IS NULL`, specID).Scan(&count)
	return count
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func scanSignals(rows *sql.Rows) ([]signals.Signal, error) {
	var result []signals.Signal
	for rows.Next() {
		var (
			sig       signals.Signal
			sigType   string
			severity  string
			createdAt string
		)
		if err := rows.Scan(&sig.ID, &sig.SpecID, &sigType, &sig.AffectedField,
			&severity, &sig.Source, &sig.Title, &sig.Detail, &createdAt); err != nil {
			return nil, fmt.Errorf("scan signal: %w", err)
		}
		sig.Type = signals.SignalType(sigType)
		sig.Severity = signals.Severity(severity)
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			sig.CreatedAt = t
		}
		result = append(result, sig)
	}
	return result, rows.Err()
}

// EmitAll persists multiple signals, ignoring duplicates.
func (s *Store) EmitAll(sigs []signals.Signal) error {
	for _, sig := range sigs {
		if err := s.Emit(sig); err != nil {
			// Log but continue — partial failure is acceptable for signals.
			if !strings.Contains(err.Error(), "UNIQUE constraint") {
				return err
			}
		}
	}
	return nil
}
