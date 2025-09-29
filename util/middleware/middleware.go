package middleware

import (
	"context"
	"net/http"
	"strings"

	utilContext "github.com/golang-base-template/util/context"
	"github.com/golang-base-template/util/csrf"
	"github.com/golang-base-template/util/response"
	"github.com/julienschmidt/httprouter"
)

type (
	Chain func(httprouter.Handle) httprouter.Handle
)

// ChainReq is a middleware for.....
func ChainReq(endHandler httprouter.Handle, chains ...Chain) httprouter.Handle {
	if len(chains) == 0 {
		return endHandler
	}

	return chains[0](ChainReq(endHandler, chains[1:]...))
}

var InitContext = func(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		r = r.WithContext(context.Background())
		next(w, r, p)
	}
}

// SetHeader is for add response header for common JSON api
var SetHeader = func(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTION")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie, Source-Type, Origin, Content-Filename")
		next(w, r, p)
	}
}

var CSRF = func(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := r.Context()
		// check if it comes from apps then skip csrf checking
		xDevice, _ := utilContext.GetDevice(ctx)
		if strings.HasPrefix(xDevice, "ios") || strings.HasPrefix(xDevice, "android") {
			next(w, r, p)
			return
		}

		csrfCheck := csrf.New()
		if csrfCheck.Check(r.Header.Get("origin"), r.Header.Get("Referer")) == false {
			resp := response.New(r.Header.Get("origin"), "true")
			resp.WriteError(w, http.StatusForbidden, "Unauthenticated", "Unauthenticated")
			return
		}
		next(w, r, p)
	}
}
