package tui

import (
	"database/sql"

	"github.com/mistakeknot/autarch/internal/coldwine/project"
	"github.com/mistakeknot/autarch/internal/coldwine/storage"
)

func LoadReviewQueue(db *sql.DB) ([]string, error) {
	return storage.ListReviewQueue(db)
}

func LoadReviewQueueFromProject() ([]string, error) {
	root, err := project.FindRoot(".")
	if err != nil {
		return nil, err
	}
	db, err := storage.OpenShared(project.StateDBPath(root))
	if err != nil {
		return nil, err
	}
	return storage.ListReviewQueue(db)
}
