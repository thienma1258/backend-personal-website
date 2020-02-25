package rest

import (
	"dongpham/src/services"
	"github.com/gorilla/mux"
	"net/http"
)

type User struct {
	UserServices *services.UserServices
}

func (userApi *User) GetAllUser(w http.ResponseWriter, r *http.Request) {
	return
}

func RegisterUserApi(router *mux.Router) *mux.Router {
	User := User{}
	router.Methods("GET").Path("v0/Users").HandlerFunc(User.GetAllUser)
	return router
}
