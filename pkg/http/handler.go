package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	basicctrl "github.com/golang-base-template/internal/basic_module/controller"
	gbtctrl "github.com/golang-base-template/internal/gbt_employee/controller"
	"github.com/golang-base-template/util/response"
)

var (
	basicCtrl       basicctrl.BasicController
	gbtEmployeeCtrl gbtctrl.IGbtEmployeeController
)

func Init() {
	if basicCtrl == nil {
		basicCtrl = basicctrl.NewBasicController()
	}
	if gbtEmployeeCtrl == nil {
		gbtEmployeeCtrl = gbtctrl.NewGbtEmployeeCtrl()
	}
}

func GetData(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()
	res := response.New(r.Header.Get("Origin"), "true")
	source := p.ByName("source")
	result, err := basicCtrl.GetData(ctx)
	result += "." + source
	if err != nil {
		res.WriteError(w, http.StatusInternalServerError, []string{"error when get data"}, err.Error())
		return
	}

	res.WriteResponse(w, result)
	return
}
