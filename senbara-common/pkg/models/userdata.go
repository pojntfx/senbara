package models

import (
	"database/sql"
	"time"
)

type (
	ExportedEntityIdentifier = struct {
		EntityName string `json:"entityName"`
	}
)

type (
	ExportedJournalEntry = struct {
		ExportedEntityIdentifier

		ID        int32     `json:"id"`
		Title     string    `json:"title"`
		Date      time.Time `json:"date"`
		Body      string    `json:"body"`
		Rating    int32     `json:"rating"`
		Namespace string    `json:"namespace"`
	}

	ExportedContact = struct {
		ExportedEntityIdentifier

		ID        int32        `json:"id"`
		FirstName string       `json:"firstName"`
		LastName  string       `json:"lastName"`
		Nickname  string       `json:"nickname"`
		Email     string       `json:"email"`
		Pronouns  string       `json:"pronouns"`
		Namespace string       `json:"namespace"`
		Birthday  sql.NullTime `json:"birthday"`
		Address   string       `json:"address"`
		Notes     string       `json:"notes"`
	}

	ExportedDebt = struct {
		ExportedEntityIdentifier

		ID          int32         `json:"id"`
		Amount      float64       `json:"amount"`
		Currency    string        `json:"currency"`
		Description string        `json:"description"`
		ContactID   sql.NullInt32 `json:"contactId"`
	}

	ExportedActivity = struct {
		ExportedEntityIdentifier

		ID          int32         `json:"id"`
		Name        string        `json:"name"`
		Date        time.Time     `json:"date"`
		Description string        `json:"description"`
		ContactID   sql.NullInt32 `json:"contactId"`
	}
)
