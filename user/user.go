package user

import (
	"bitbucket.org/futebolear/misc"
	"errors"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
	"github.com/gorilla/mux"
)

type User struct {
	Username string
	Password string
	Email    string
	Active   bool
}

func RegisterAPI(r *mux.Router){
	r.Path("/").Methods("POST").HandlerFunc(postHandler)
	r.Path("/{username}").Methods("GET").HandlerFunc(getHandler)
	r.Path("/{username}").Methods("PUT").HandlerFunc(putHandler)
	r.Path("/{username}").Methods("DELETE").HandlerFunc(deleteHandler)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var u User
	if err := CreateUserFromPathVariable(r, c, &u); err != nil {
		misc.WriteResponse(w, "", http.StatusNotFound, err)
		return
	}
	misc.WriteResponse(w, u, http.StatusOK, nil)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var u User
	if err := CreateUserFromPathVariable(r, c, &u); err != nil {
		misc.WriteResponse(w, "", http.StatusNotFound, err)
		return
	}
	deleteAndReply(w, c, &u)
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if err := CreateUserFromPathVariable(r, c, &User{}); err != nil {
		misc.WriteResponse(w, "", http.StatusNotFound, err)
		return
	}
	var u User
	if err := misc.Decode(r, &u); err != nil {
		misc.WriteResponse(w, "", http.StatusBadRequest, err)
		return
	}
	usernameToMatch := misc.GetPathVariable(r, "username")
	if u.Username != usernameToMatch {
		log.Debugf(c, "User.puthandler - Received username %s and path variable %s", u.Username, usernameToMatch)
		misc.WriteResponse(w, "", http.StatusBadRequest, errors.New("Invalid Request"))
		return
	} else {
		putAndReply(w, c, &u)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var u User
	if err := misc.Decode(r, &u); err != nil {
		misc.WriteResponse(w, "", http.StatusBadRequest, err)
		return

	}
	// TODO this has to be a validate method in a next PR
	if u.Username == "" {
		misc.WriteResponse(w, "", http.StatusBadRequest, errors.New("Username is mandatory"))
		return
	}
	// Validate if user exists
	if err := Get(c, &u, u.Username); err == nil {
		misc.WriteResponse(w, "", http.StatusBadRequest, errors.New("Username already exists"))
		return
	}
	// Activate user before saving
	u.Active = true
	putAndReply(w, c, &u)
}

func CreateUserFromPathVariable(r *http.Request, c context.Context, user *User) error {
	username := misc.GetPathVariable(r, "username")
	// Retrieve user from data store
	if err := Get(c, user, username); err != nil {
		return errors.New("Username " + username + " does not exist")
	}
	return nil
}

func putAndReply(w http.ResponseWriter, c context.Context, user *User) {
	key, err := Put(c, user)
	if err != nil {
		misc.WriteResponse(w, "", http.StatusInternalServerError, err)
		return
	}
	misc.ReplyKey(w, key)
}

func deleteAndReply(w http.ResponseWriter, c context.Context, user *User) {
	key, err := Delete(c, user)
	if err != nil {
		misc.WriteResponse(w, "", http.StatusInternalServerError, err)
		return
	}
	misc.ReplyKey(w, key)
}

func Get(c context.Context, user *User, username string) error {
	return datastore.Get(c, generateKey(c, username), user)
}

func Delete(c context.Context, user *User) (*datastore.Key, error) {
	if err := Get(c, &User{}, user.Username); err != nil {
		return nil, err
	}
	// Soft delete user
	user.Active = false
	return Put(c, user)
}

func Put(c context.Context, user *User) (*datastore.Key, error) {
	key := generateKey(c, user.Username)
	return datastore.Put(c, key, user)
}

func generateKey(c context.Context, username string) *datastore.Key {
	return datastore.NewKey(c, "User", username, 0, nil)
}
