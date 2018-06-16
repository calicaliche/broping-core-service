package misc

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/appengine/datastore"
	"net/http"
	"errors"
)

const fmtInternalErrMsg = `{"Error": %q,"StatusCode":%d}`

func WriteResponse(w http.ResponseWriter, resp interface{}, code int, err error) {
	response := struct {
		Content    interface{}
		Error      string
		StatusCode int
	}{
		Content:    resp,
		StatusCode: code,
	}
	if err != nil {
		response.Error = err.Error()
	}
	js, err := json.Marshal(&response)
	if err != nil {
		write(w, http.StatusInternalServerError, buildInternalErrorMsg(err.Error()))
		return
	}
	write(w, code, js)
}

func Decode(r *http.Request, structure interface{}) error {
	if r.Body == nil {
		return errors.New("Request doesn't contained body")
	}
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(&structure)
}

func write(w http.ResponseWriter, code int, json []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(json)
}

func buildInternalErrorMsg(msg string) []byte {
	return []byte(fmt.Sprintf(fmtInternalErrMsg, msg, http.StatusInternalServerError))
}

func GetPathVariable(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

func ReplyKey(w http.ResponseWriter, key *datastore.Key) {
	response := struct {
		Key string
	}{
		Key: key.Encode(),
	}
	WriteResponse(w, response, http.StatusOK, nil)
}
