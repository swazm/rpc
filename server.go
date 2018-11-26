package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

type Request struct {
	Id      int             `json:"id" validate:"nonzero"`
	Method  string          `json:"method" validate:"nonzero"`
	Service string          `json:"service" validate:"nonzero"`
	Data    json.RawMessage `json:"data" validate:"nonzero"`
}

type Response struct {
	Id    int         `json:"id"`
	Data  interface{} `json:"data"`
	Error *Error      `json:"error"`
}

type Server struct {
	services map[string]*service // map[string]*service
}

func (server *Server) Register(rcvr interface{}, name string) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()

	s.methods = suitableMethods(s.typ)
	if len(s.methods) == 0 {
		return errors.New("rpc.Register: type " + sname + " has no exported methods of suitable type")

	}

	if server.services[sname] != nil {
		return errors.New("rpc: service already defined: " + sname)
	} else {
		server.services[sname] = s
	}
	return nil
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{
		services:make(map[string]*service),
	}
}

var DefaultServer = NewServer()

func (server *Server) HandleFunc(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var request Request
	err := decoder.Decode(&request)

	response := new(Response)

	status := http.StatusOK
	if err != nil {
		response.Error = NewError(http.StatusBadRequest, err.Error())
		status = http.StatusBadRequest
		log.Printf("error while deserializing request %s", err.Error())
	} else {
		//Check for the existance of the service
		if server.services[request.Service] == nil {
			response.Error = NewError(http.StatusBadRequest, fmt.Sprintf("invalid service '%s'", request.Service))
			status = http.StatusBadRequest
		} else if server.services[request.Service].methods[request.Method] == nil {
			response.Error = NewError(http.StatusBadRequest, fmt.Sprintf("invalid method '%s' on service '%s'", request.Method, request.Service))
			status = http.StatusBadRequest
		} else {
			m := server.services[request.Service].methods[request.Method]

			var requestData reflect.Value
			// Decode the argument value.
			requestDataTypeIsValue := false // if true, need to indirect before calling.
			if m.RequestDataType.Kind() == reflect.Ptr {
				requestData = reflect.New(m.RequestDataType.Elem())
			} else {
				requestData = reflect.New(m.RequestDataType)
				requestDataTypeIsValue = true
			}
			// argv guaranteed to be a pointer now.

			err := json.Unmarshal([]byte(request.Data), requestData.Interface());
			if err != nil {
				response.Error = NewError(http.StatusInternalServerError, fmt.Sprintf("error while calling '%s' on service :'%s'", request.Method, request.Service,err.Error()))
				status = http.StatusInternalServerError
			}
			if requestDataTypeIsValue {
				requestData = requestData.Elem()
			}
			//call the method
			returnValues := m.method.Func.Call([]reflect.Value{server.services[request.Service].rcvr,reflect.ValueOf(req), requestData})
			log.Printf("Return values %+v",returnValues[0])
			if returnValues[1].Interface() != nil {
				status = http.StatusBadRequest
				response.Error = NewError(http.StatusInternalServerError, fmt.Sprintf("error while calling '%s' on service '%s' : %s", request.Method, request.Service, returnValues[1].Interface()))
			}

			response.Data = returnValues[0].Interface()

		}

	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		response.Error = NewError(http.StatusInternalServerError, err.Error())
		status = http.StatusInternalServerError
		log.Printf("error while serializing request %s", err.Error())
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(responseBytes)
}
