package http

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"

	gbtctrl "github.com/golang-base-template/internal/gbt_employee/controller"
	"github.com/golang-base-template/util/response"
)

type (
	GbtEmployeeInput struct {
		Name    string `json:"name"`
		Gender  string `json:"gender"`
		Address string `json:"address"`
	}
)

func CreateGbtEmployee(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := r.Context()

	res := response.New(r.Header.Get("Origin"), "true")

	var input GbtEmployeeInput

	// Decode JSON input
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		res.WriteError(w, http.StatusUnprocessableEntity, []string{response.ErrDecodeRequestBody}, err.Error())
		return
	}

	gbtEmployee, err := gbtEmployeeCtrl.CreateGBTEmployee(ctx, gbtctrl.GbtEmployeeData{
		Name:    input.Name,
		Gender:  input.Gender,
		Address: input.Address,
	})
	if err != nil {
		res.WriteError(w, http.StatusInternalServerError, []string{"error when create gbt employee data"}, err.Error())
		return
	}

	res.WriteResponse(w, gbtEmployee)
	return
}
