package core

import (
	"context"
	"time"
)

func (g *gbtEmployeeCore) ConstructGbtEmployee(ctx context.Context, data GbtEmployeeData) IGbtEmployeeCore {
	return &gbtEmployeeCore{
		Data: GbtEmployeeData{
			Name:    data.Name,
			Gender:  data.Gender,
			Address: data.Address,
		}}
}

func (g *gbtEmployeeCore) SetName(ctx context.Context, name string) {
	g.Data.Name = name
	return
}

func (g *gbtEmployeeCore) SetGender(ctx context.Context, gender string) {
	g.Data.Gender = gender
	return
}

func (g *gbtEmployeeCore) SetAddress(ctx context.Context, address string) {
	g.Data.Address = address
	return
}

func (g *gbtEmployeeCore) Save(ctx context.Context) error {
	tx, err := dbGbtMaster.Beginx()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	err = tx.QueryRowContext(ctx, "", nil, g.Data.Name, g.Data.Gender, g.Data.Address)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	rp, err := pkgCachePipeline.NewPipeline(ctx)
	if err != nil {
		return err
	}
	err = rp.HMSet(ctx, "", map[string]string{})
	if err != nil {
		return err
	}
	err = rp.Expire(ctx, "", time.Duration(0))
	if err != nil {
		return err
	}
	err = rp.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
