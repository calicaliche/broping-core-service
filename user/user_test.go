package user

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	inst       aetest.Instance
	routerTest = func() *mux.Router {
		r := mux.NewRouter()
		RegisterAPI(r.PathPrefix("/users").Subrouter())
		return r
	}()
)

func TestMain(m *testing.M) {
	var err error
	inst, err = aetest.NewInstance(nil)
	if err != nil {
		os.Exit(-1)
	}
	e := m.Run()
	inst.Close()
	os.Exit(e)
}

func TestHappyCase(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	u := User{"pate", "reyPate", "pate@gmail.com", true}
	if _, err := Put(c, &u); err != nil {
		t.Fatal(err)
	}
	if _, err := Delete(c, &u); err != nil {
		t.Fatal(err)
	}
	u = User{}
	if err := Get(c, &u, "pate"); err != nil {
		t.Fatal(err)
	}
	if u.Active == true {
		t.Errorf("User Active: %t want: %t", u.Active, false)
	}
}

func TestPostHandler(t *testing.T) {
	testCases := []struct {
		method   string
		code     int
		content  *User
		expected *User
	}{
		{
			method:   "POST",
			code:     http.StatusOK,
			content:  &User{Username: "pate", Password: "reyPate", Email: "pate@gmail.com"},
			expected: &User{"pate", "reyPate", "pate@gmail.com", true},
		},
		{
			method:   "POST",
			code:     http.StatusBadRequest,
			content:  &User{Username: "pate", Password: "reyPate", Email: "pate@gmail.com"},
			expected: &User{"pate", "reyPate", "pate@gmail.com", true},
		},
		{
			method:   "POST",
			code:     http.StatusOK,
			content:  &User{Username: "alonso", Password: "solis10", Email: "asolis10@gmail.com"},
			expected: &User{"alonso", "solis10", "asolis10@gmail.com", true},
		},
		{
			method:   "POST",
			code:     http.StatusBadRequest,
			content:  nil,
			expected: nil,
		},
	}

	for _, tt := range testCases {
		body, err := json.Marshal(tt.content)
		if err != nil {
			t.Fatal(err)
		}

		req, err := inst.NewRequest(tt.method, "/users/", bytes.NewBuffer(body))
		if err != nil {
			t.Errorf("inst.NewRequest failed: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		routerTest.ServeHTTP(resp, req)

		// Validate that code is the expected
		if resp.Code != tt.code {
			t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, tt.code, resp.Body.String())
		}

		// Check against dataStore.
		if tt.expected != nil {
			c := appengine.NewContext(req)

			var user User
			if err := Get(c, &user, tt.content.Username); err != nil {
				t.Fatal(err)
			}
			// Validate expected data
			if user.Username != tt.expected.Username {
				t.Errorf("Got username %q; want %q", user.Username, tt.expected.Username)
			}
			if user.Active != tt.expected.Active {
				t.Errorf("User Active: %t Want: %t", user.Active, tt.expected.Active)
			}
		}
	}
}

func TestGetHandler(t *testing.T) {
	// Prepare dataStore for tests
	u := &User{"zidane", "zizou", "zizou@gmail.com", true}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	req, err := inst.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("inst.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	routerTest.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, http.StatusOK, resp.Body.String())
	}

	testCases := []struct {
		method   string
		code     int
		name     string
		expected *User
	}{
		{
			method:   "GET",
			code:     http.StatusOK,
			name:     "zidane",
			expected: &User{"zidane", "zizou", "zizou@gmail.com", true},
		},
		{
			method: "GET",
			code:   http.StatusNotFound,
			name:   "figo",
		},
	}

	for _, tt := range testCases {
		req, err := inst.NewRequest(tt.method, "/users/"+tt.name, nil)
		if err != nil {
			t.Errorf("inst.NewRequest failed: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		routerTest.ServeHTTP(resp, req)
		// Validate that code is the expected
		if resp.Code != tt.code {
			t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, tt.code, resp.Body.String())
		}
		// Validate expected data if http Code == http.StatusOk. Check against dataStore.
		if tt.code == http.StatusOK {
			c := appengine.NewContext(req)

			var user User
			if err := Get(c, &user, tt.name); err != nil {
				t.Fatal(err)
			}

			if user.Username != tt.expected.Username {
				t.Errorf("Got username %q; want %q", user.Username, tt.expected.Username)
			}
			if user.Active != tt.expected.Active {
				t.Errorf("User Active: %t Want: %t", user.Active, tt.expected.Active)
			}
		}
	}
}

func TestPutHandler(t *testing.T) {
	// Prepare dataStore for tests
	u := &User{"delpiero", "juve10", "delpiero@gmail.com", true}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	req, err := inst.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("inst.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	routerTest.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, http.StatusOK, resp.Body.String())
	}

	testCases := []struct {
		method   string
		code     int
		name     string
		content  *User
		expected *User
	}{
		{
			method:   "PUT",
			code:     http.StatusOK,
			name:     "delpiero",
			content:  &User{"delpiero", "juve10delpiero", "delpiero@gmail.com", true},
			expected: &User{"delpiero", "juve10delpiero", "delpiero@gmail.com", true},
		},
		{
			method:  "PUT",
			code:    http.StatusNotFound,
			name:    "davids",
			content: &User{"delpiero", "juve10delpiero", "delpiero@gmail.com", true},
		},
		{
			method:  "PUT",
			code:    http.StatusBadRequest,
			name:    "delpiero",
			content: nil,
		},
	}

	for _, tt := range testCases {
		body, err := json.Marshal(tt.content)
		if err != nil {
			t.Fatal(err)
		}
		req, err := inst.NewRequest(tt.method, "/users/"+tt.name, bytes.NewBuffer(body))
		if err != nil {
			t.Errorf("inst.NewRequest failed: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		routerTest.ServeHTTP(resp, req)
		// Validate that code is the expected
		if resp.Code != tt.code {
			t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, tt.code, resp.Body.String())
		}
		// Validate expected data if http Code == http.StatusOk. Check against dataStore.
		if tt.code == http.StatusOK {
			c := appengine.NewContext(req)

			var user User
			if err := Get(c, &user, tt.name); err != nil {
				t.Fatal(err)
			}

			if user.Username != tt.expected.Username {
				t.Errorf("Got username %q; want %q", user.Username, tt.expected.Username)
			}
			if user.Password != tt.expected.Password {
				t.Errorf("Passowrd: %t Want: %t", user.Active, tt.expected.Active)
			}
		}
	}
}

func TestDeleteHandler(t *testing.T) {
	// Prepare dataStore for tests
	u := &User{"maradona", "diego10", "d10s@gmail.com", true}
	body, err := json.Marshal(&u)
	if err != nil {
		t.Fatal(err)
	}
	req, err := inst.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("inst.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	routerTest.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, http.StatusOK, resp.Body.String())
	}

	testCases := []struct {
		method   string
		code     int
		name     string
		content  *User
		expected *User
	}{
		{
			method:   "DELETE",
			code:     http.StatusOK,
			name:     "maradona",
			expected: &User{"maradona", "diego10", "d10s@gmail.com", false},
		},
		{
			method: "DELETE",
			code:   http.StatusNotFound,
			name:   "batistuta",
		},
	}

	for _, tt := range testCases {
		req, err := inst.NewRequest(tt.method, "/users/"+tt.name, bytes.NewBuffer(body))
		if err != nil {
			t.Errorf("inst.NewRequest failed: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		routerTest.ServeHTTP(resp, req)
		// Validate that code is the expected
		if resp.Code != tt.code {
			t.Errorf("Got response code %d; want %d; body:\n%s", resp.Code, tt.code, nil)
		}
		// Validate expected data if http Code == http.StatusOk. Check against dataStore.
		if tt.code == http.StatusOK {
			c := appengine.NewContext(req)

			var user User
			if err := Get(c, &user, tt.name); err != nil {
				t.Fatal(err)
			}

			if user.Username != tt.expected.Username {
				t.Errorf("Got username %q; want %q", user.Username, tt.expected.Username)
			}
			if user.Active != tt.expected.Active {
				t.Errorf("Passowrd: %t Want: %t", user.Active, tt.expected.Active)
			}
		}
	}
}
