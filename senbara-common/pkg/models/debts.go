package models

import "github.com/pojntfx/senbara/senbara-common/internal/tables"

type (
	CreateDebtParams        = tables.CreateDebtParams
	GetDebtsParams          = tables.GetDebtsParams
	SettleDebtParams        = tables.SettleDebtParams
	GetDebtAndContactParams = tables.GetDebtAndContactParams
	UpdateDebtParams        = tables.UpdateDebtParams
)

type (
	CreateDebtRow        = tables.CreateDebtRow
	UpdateDebtRow        = tables.UpdateDebtRow
	GetDebtsRow          = tables.GetDebtsRow
	GetDebtAndContactRow = tables.GetDebtAndContactRow
)
