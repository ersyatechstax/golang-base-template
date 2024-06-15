package core

import (
	"context"
	"log"

	"github.com/golang-base-template/util/cache"
	"github.com/golang-base-template/util/database"
)

var (
	dbGbtMaster      database.Database
	dbGbtSlave       database.Database
	pkgCachePipeline cache.ICachePipeline
)

type (
	IGbtEmployeeCore interface {
		ConstructGbtEmployee(ctx context.Context, data GbtEmployeeData) IGbtEmployeeCore
		SetName(ctx context.Context, name string)
		SetGender(ctx context.Context, gender string)
		SetAddress(ctx context.Context, name string)
		Save(ctx context.Context) (err error)
	}
	gbtEmployeeCore struct {
		Data GbtEmployeeData
	}

	GbtEmployeeData struct {
		EmployeeID int64  `json:"employee_id"`
		Name       string `json:"name"`
		Gender     string `json:"gender"`
		Address    string `json:"address"`
	}
)

func NewGbtEmployeeCore() IGbtEmployeeCore {
	if dbGbtMaster == nil {
		db, err := database.GetDB("gbt", database.MasterReplication)
		if err != nil {
			log.Println("error when init database master gbt: ", err.Error())
		}
		dbGbtMaster = db
	}

	if dbGbtSlave == nil {
		db, err := database.GetDB("gbt", database.SlaveReplication)
		if err != nil {
			log.Println("error when init database slave gbt: ", err.Error())
		}
		dbGbtSlave = db
	}

	if pkgCachePipeline == nil {
		pkgCachePipeline = cache.NewPkgPipeline()
	}

	return &gbtEmployeeCore{}
}
