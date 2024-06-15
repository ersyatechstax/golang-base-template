package controller

import (
	"context"
	"github.com/golang-base-template/external"
	"github.com/golang-base-template/internal/gbt_employee/core"
)

var (
	gbtEmployeeCore   core.IGbtEmployeeCore
	pkgExternal       external.IPkgExternal
	exampleExtService external.IExampleExtService
)

type (
	IGbtEmployeeController interface {
		CreateGBTEmployee(ctx context.Context, data GbtEmployeeData) (GbtEmployeeData, error)
	}
	gbtEmployee struct{}

	GbtEmployeeData struct {
		EmployeeID int64  `json:"employee_id"`
		Name       string `json:"name"`
		Gender     string `json:"gender"`
		Address    string `json:"address"`
	}
)

func NewGbtEmployeeCtrl() IGbtEmployeeController {
	if gbtEmployeeCore == nil {
		gbtEmployeeCore = core.NewGbtEmployeeCore()
	}

	if pkgExternal == nil {
		pkgExternal = external.NewPkgExternal()
	}

	if exampleExtService == nil {
		exampleExtService = pkgExternal.NewExampleExtService()
	}

	return &gbtEmployee{}
}
