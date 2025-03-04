package models

import "github.com/pojntfx/senbara/senbara-common/internal/tables"

type (
	CreateActivityParams        = tables.CreateActivityParams
	GetActivitiesParams         = tables.GetActivitiesParams
	DeleteActivityParams        = tables.DeleteActivityParams
	GetActivityAndContactParams = tables.GetActivityAndContactParams
	UpdateActivityParams        = tables.UpdateActivityParams
)

type (
	CreateActivityRow        = tables.CreateActivityRow
	UpdateActivityRow        = tables.UpdateActivityRow
	GetActivitiesRow         = tables.GetActivitiesRow
	GetActivityAndContactRow = tables.GetActivityAndContactRow
)
