package rpc

import (
	"bytes"
	"net/http"
	"testing"
)

type CalculatorRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type CalculatorResponse struct {
	Result int `json:"result"`
}

type CalculatorService Service

func (c *CalculatorService) Add(request *http.Request, data *CalculatorRequest) (response *CalculatorResponse, err error) {

	return &CalculatorResponse{
		Result: data.A+data.B,
	}, nil
}

func TestShouldRegisterCalculator(t *testing.T) {
	err:=DefaultServer.Register(new(CalculatorService),"pipi")
	if err != nil {
		t.Errorf("test should have registered methods instead %s",err.Error())
	}
}

func TestShouldGetDebugOutput(t *testing.T)  {
	err:=DefaultServer.Register(new(CalculatorService),"pipi")
	if err != nil {
		t.Errorf("test should have registered methods instead %s",err.Error())
	}
	var debugInfo bytes.Buffer
	DefaultServer.writeDebug(&debugInfo)

	t.Logf("Info: %s",debugInfo.String())
}