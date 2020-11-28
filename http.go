package main

import (
    "net/http"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
    var bodyString string
    bodyString, _, _ = bodyToString(r.Body)
    log.WithFields(log.Fields{
        "method": r.Method,
        "body": bodyString,
    }).Debug("get a new accessing request")

    if r.Method != http.MethodPost {
        log.WithFields(log.Fields{
            "method": r.Method,
        }).Error("the method of request is not expected")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "please POST method")
        return 
    }

    
}