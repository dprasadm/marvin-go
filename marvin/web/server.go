package server

import (
    "http"
    "fmt"
    "os"
    "marvin/cloud/request"
)

const marvinEndPoint = "99games.mobi:9900"
//const marvinEndPoint = "127.0.0.1:9900"

func registerRequestHandlers() {
    http.Handle("/marvin/coins/", http.HandlerFunc(genericHttpPostRequestHandler))
    http.Handle("/marvin/registration/", http.HandlerFunc(genericHttpPostRequestHandler))
    http.Handle("/marvin/unregister/", http.HandlerFunc(genericHttpPostRequestHandler))
    http.Handle("/marvin/messages/", http.HandlerFunc(genericHttpPostRequestHandler))
    http.Handle("/marvin/scores/", http.HandlerFunc(genericHttpPostRequestHandler))
    http.Handle("/marvin/achievements/", http.HandlerFunc(genericHttpPostRequestHandler))
    http.Handle("/marvin/leaderboard/", http.HandlerFunc(genericHttpPostRequestHandler))
    
    http.Handle("/", http.HandlerFunc(genericInvalidRequestHandler))
}

func Run() {
    registerRequestHandlers()
    //err := http.ListenAndServeTLS(marvinEndPoint, "cert.pem", "key.pem", nil)
    err := http.ListenAndServe(marvinEndPoint, nil)
    if err != nil {
        fmt.Printf("ListenAndServe: %v\n", err)
        os.Exit(1)
    }
}

type ExecutionPipe struct {
    resWriter http.ResponseWriter
    request   *http.Request
}

func (xp* ExecutionPipe) Execute() {
    httpReqReder := &cloud.HTTPRequestReader{}
    json := httpReqReder.Read(xp.request)
    httpReqValidator := &cloud.HTTPRequestValidator{}
    request := httpReqValidator.Validate(xp.request, json)
    
    response := request.Perform()
    
    (cloud.NewHTTPResponseWriter(xp.resWriter)).Write(response)
}

func genericHttpPostRequestHandler(w http.ResponseWriter, req *http.Request) {
    if req.Method == "POST" {
        xp := &ExecutionPipe{w, req}
        xp.Execute()
    }
    // req.Body.Close() done in request.HTTPRequestReader.Read(*http.Request) 
}

func genericInvalidRequestHandler(w http.ResponseWriter, req *http.Request) {
    m := []byte("{\"result\": false, \"answer\": \"Unknwon Request. Rejected\"}")
    w.Write(m)
}


