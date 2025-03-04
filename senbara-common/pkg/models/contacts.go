package models

import "github.com/pojntfx/senbara/senbara-common/internal/tables"

type (
	CreateContactParams         = tables.CreateContactParams
	GetContactParams            = tables.GetContactParams
	DeleteContactParams         = tables.DeleteContactParams
	DeleteDebtsForContactParams = tables.DeleteDebtsForContactParams
	UpdateContactParams         = tables.UpdateContactParams
)

type (
	Contact = tables.Contact
)
