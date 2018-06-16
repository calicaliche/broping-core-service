package barroom

import (
	"google.golang.org/appengine"
	"github.com/gorilla/mux"
	"net/http"
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"bitbucket.org/futebolear/misc"
	"errors"
)

type Bar struct {
	Id string
	Name string
	Location appengine.GeoPoint
	Active bool
}

func RegisterAPI(r *mux.Router){
	r.Path("/").Methods("POST").HandlerFunc(postHandler)
	r.Path("/{id}").Methods("GET").HandlerFunc(getHandler)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var b Bar
	if err := CreateBarFromPathVariable(r, c, &b); err != nil {
		misc.WriteResponse(w, "", http.StatusNotFound, err)
		return
	}
	misc.WriteResponse(w, b, http.StatusOK, nil)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var b Bar
	if err := misc.Decode(r, &b); err != nil {
		misc.WriteResponse(w, "", http.StatusBadRequest, err)
		return

	}
	// TODO this has to be a validate method in a next PR
	if b.Id == "" {
		misc.WriteResponse(w, "", http.StatusBadRequest, errors.New("ID is mandatory"))
		return
	}
	// Validate if user exists
	if err := Get(c, &b, b.Id); err == nil {
		misc.WriteResponse(w, "", http.StatusBadRequest, errors.New("ID already exists"))
		return
	}
	// Activate user before saving
	b.Active = true
	putAndReply(w, c, &b)
}

func CreateBarFromPathVariable(r *http.Request, c context.Context, bar *Bar) error {
	id := misc.GetPathVariable(r, "id")
	// Retrieve user from data store
	if err := Get(c, bar, id); err != nil {
		return errors.New("ID " + id + " does not exist")
	}
	return nil
}



func putAndReply(w http.ResponseWriter, c context.Context, bar *Bar) {
	key, err := Put(c, bar)
	if err != nil {
		misc.WriteResponse(w, "", http.StatusInternalServerError, err)
		return
	}
	misc.ReplyKey(w, key)
}

func deleteAndReply(w http.ResponseWriter, c context.Context, bar *Bar) {
	key, err := Delete(c, bar)
	if err != nil {
		misc.WriteResponse(w, "", http.StatusInternalServerError, err)
		return
	}
	misc.ReplyKey(w, key)
}

func Get(c context.Context, bar *Bar, barId string) error {
	return datastore.Get(c, generateKey(c, barId), bar)
}

func Delete(c context.Context, bar *Bar) (*datastore.Key, error) {
	if err := Get(c, &Bar{}, bar.Id); err != nil {
		return nil, err
	}
	// Soft delete bar
	bar.Active = false
	return Put(c, bar)
}

func Put(c context.Context, bar *Bar) (*datastore.Key, error) {
	key := generateKey(c, bar.Id)
	return datastore.Put(c, key, bar)
}

func generateKey(c context.Context, barId string) *datastore.Key {
	return datastore.NewKey(c, "Bar", barId, 0, nil)
}
