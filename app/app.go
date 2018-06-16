package app

import (
	"bitbucket.org/futebolear/user"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/calicaliche/broping-core-service/barroom"
)

func init() {
	r := mux.NewRouter()
	user.RegisterAPI(r.PathPrefix("/users").Subrouter())
	barroom.RegisterAPI(r.PathPrefix("/bars").Subrouter())
	http.Handle("/", r)
}
