package controller

import (
	"context"

	"github.com/golang-base-template/internal/gbt_employee/core"
)

func (ge *gbtEmployee) CreateGBTEmployee(ctx context.Context, data GbtEmployeeData) (result GbtEmployeeData, err error) {
	gbtEmployee := gbtEmployeeCore.ConstructGbtEmployee(ctx, core.GbtEmployeeData{
		Name:    data.Name,
		Gender:  data.Gender,
		Address: data.Address,
	})
	err = gbtEmployee.Save(ctx)
	if err != nil {
		return GbtEmployeeData{}, err
	}
	return data, nil
}
