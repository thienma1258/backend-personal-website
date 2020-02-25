package rest

import (
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Api struct {
}

func RegisterRoutes(router *mux.Router) *mux.Router {
	//router.Methods().
	////router.Methods("POST").Path("/v1/content_json").HandlerFunc(GetTest)
	router = RegisterUserApi(router)
	router = RegisterGalleryApi(router)
	return router
}
