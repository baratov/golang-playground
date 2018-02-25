package server

import (
	"encoding/json"
	"net/http"
)

// JSend format, the simplest I found
// https://labs.omniti.com/labs/jsend

const (
	fieldStatus  = "status"
	fieldData    = "data"
	fieldMessage = "message"

	statusSuccess = "success"
	statusFail    = "fail"
	statusError   = "error"
)

type Response struct {
	fields map[string]interface{}
	rw     http.ResponseWriter
}

func withWriter(w http.ResponseWriter) *Response {
	return &Response{
		fields: make(map[string]interface{}),
		rw:     w,
	}
}

func (r *Response) Data(data interface{}) *Response {
	return r.
		Field(fieldData, data).
		Field(fieldStatus, statusSuccess)
}

func (r *Response) Error(err error) *Response {
	if err != nil {
		return r.
			Field(fieldMessage, err.Error()).
			Field(fieldStatus, statusFail)
	} else {
		return r
	}
}

func (r *Response) Field(name string, value interface{}) *Response {
	r.fields[name] = value
	return r
}

func (r *Response) WriteResponse() {
	j, err := json.Marshal(r.fields)
	if err != nil {
		panic(err)
	}

	r.rw.Header().Set("Content-Type", "application/json")
	r.rw.WriteHeader(http.StatusOK)
	r.rw.Write(j)
}
