package http

import (
	"github.com/julienschmidt/httprouter"

	"github.com/golang-base-template/util/middleware"
)

func AssignRoutes(router *httprouter.Router) {
	router.GET("/get-data/:source", middleware.ChainReq(GetData, middleware.InitContext, middleware.SetHeader))
	router.POST("/create-gbt-employee", middleware.ChainReq(CreateGbtEmployee, middleware.InitContext, middleware.SetHeader, middleware.CSRF))
}
