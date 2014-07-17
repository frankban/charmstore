package router

import (
	"encoding/json"
	"net/http"

	"github.com/juju/charmstore/params"
)

// HandleErrors returns a Handler that calls the given function.
// If the function reports an error, it sets the HTTP response
// code and sends the error as a JSON reply by calling
// WriteError.
func HandleErrors(handle func(http.ResponseWriter, *http.Request) error) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) {
		if err := handle(w, req); err != nil {
			WriteError(w, err)
		}
	}
	return http.HandlerFunc(f)
}

// HandleJSON returns a Handler that calls the given function.
// The result is formatted as JSON.
func HandleJSON(handle func(http.ResponseWriter, *http.Request) (interface{}, error)) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) error {
		val, err := handle(w, req)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, val)
	}
	return HandleErrors(f)
}

func WriteError(w http.ResponseWriter, err error) {
	errResp := &params.Error{
		Message: err.Error(),
	}
	if err, ok := err.(params.ErrorCoder); ok {
		errResp.Code = err.ErrorCode()
	}
	// TODO log writeJSON error if it happens?
	WriteJSON(w, http.StatusInternalServerError, errResp)
}

func WriteJSON(w http.ResponseWriter, code int, val interface{}) error {
	// TODO consider marshalling directly to w using json.NewEncoder.
	// pro: this will not require a full buffer allocation.
	// con: if there's an error after the first write, it will be lost.
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
	return nil
}
