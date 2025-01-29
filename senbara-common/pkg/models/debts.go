package models

import "github.com/pojntfx/senbara/senbara-common/pkg/tables"

type (
	CreateDebtParams        = tables.CreateDebtParams
	GetDebtsParams          = tables.GetDebtsParams
	SettleDebtParams        = tables.SettleDebtParams
	GetDebtAndContactParams = tables.GetDebtAndContactParams
	UpdateDebtParams        = tables.UpdateDebtParams
)

type (
	GetDebtsRow          = tables.GetDebtsRow
	GetDebtAndContactRow = tables.GetDebtAndContactRow
)
