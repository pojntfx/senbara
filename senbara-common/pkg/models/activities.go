package models

import "github.com/pojntfx/senbara/senbara-common/pkg/tables"

type (
	CreateActivityParams        = tables.CreateActivityParams
	GetActivitiesParams         = tables.GetActivitiesParams
	DeleteActivityParams        = tables.DeleteActivityParams
	GetActivityAndContactParams = tables.GetActivityAndContactParams
	UpdateActivityParams        = tables.UpdateActivityParams
)

type (
	GetActivitiesRow         = tables.GetActivitiesRow
	GetActivityAndContactRow = tables.GetActivityAndContactRow
)
