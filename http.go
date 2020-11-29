package main

import (
    "net/http"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    var bodyString string
    var bodyMap map[string]interface{}
    var rRemote string
    var rHeader map[string][]string

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

    rRemote = r.Header.Get(InnerRemoteAddrField)
    if len(rRemote) == 0 {
        log.Error("not found InnerRemoteAddrField in the header")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "not found InnerRemoteAddrField in the header")
        return 
    }

    regexpPlan := r.Header.Get(RegexpHeaderField)
    log.WithFields(log.Fields{
        "regexpPlan": regexpPlan,
    }).Debug("get the regexp plan from header")

    bodyMap = regexpDeal(regexpPlan)
    if len(bodyMap) == 0 {
        log.Error("arter regexpDeal, the result is empty")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "arter regexpDeal, the result is empty")
        return 
    }

    if yes, err := isIgnore(r.Header, bodyMap); yes {
        log.Debug("catch the ignore selector, ignore it ......")
        return 
    }

    rHeader = repairHeader(r.Header, RelayIgnoreHeaderFields)

    
}