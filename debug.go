package rpc

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"sort"
)

const debugText = `<html>
	<body>
	<title>Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th>
		{{range .Method}}
			<tr>
			<td align=left font=fixed>{{.Name}}(*http.Request, {{.Type.RequestDataType}},) ({{.Type.ResponseDataType}} error)</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugMethod struct {
	Type             *methodType
	Name             string
	RequestDataType  interface{}
	ResponseDataType interface{}
}

type methodArray []debugMethod

type debugService struct {
	Service *service
	Name    string
	Method  methodArray
}

type serviceArray []debugService

func (s serviceArray) Len() int           { return len(s) }
func (s serviceArray) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s serviceArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (m methodArray) Len() int           { return len(m) }
func (m methodArray) Less(i, j int) bool { return m[i].Name < m[j].Name }
func (m methodArray) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

// Runs at /debug/rpc
func (server *Server) DebugHandlerFuncJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("content-type","application/json")
	err := json.NewEncoder(w).Encode(server.getServices())
	if err != nil {
		log.Printf("err %s",err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Runs at /debug/rpc
func (server *Server) DebugHandlerFunc(w http.ResponseWriter, req *http.Request) {
	err := server.writeDebug(w)
	if err != nil {
		fmt.Fprintln(w, "rpc: error executing template:", err.Error())
	}
}

func (server *Server) writeDebug(w io.Writer) error {
	// Build a sorted version of the data.

	return debug.Execute(w, server.getServices())
}
func (server *Server) getServices() *serviceArray {
	var services serviceArray
	for snamei, svci := range server.services {
		svc := svci
		ds := debugService{svc, snamei, make(methodArray, 0, len(svc.methods))}
		for mname, method := range svc.methods {
			m:=debugMethod{
				Type:             method,
				Name:             mname,
				RequestDataType:  reflect.New(method.RequestDataType.Elem()).Interface(),
				ResponseDataType: reflect.New(method.ResponseDataType.Elem()).Interface(),
			}
			log.Printf("DATA TYPE IN: %q",m)
			ds.Method = append(ds.Method,m )
		}
		sort.Sort(ds.Method)
		services = append(services, ds)
	}
	sort.Sort(services)
	return &services
}
