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
    rHeader = repairHeader(r.Header, RelayIgnoreHeaderFields)
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

    bodyMap = regexpDeal(regexpPlan, bodyString)
    if len(bodyMap) == 0 {
        log.Error("arter regexpDeal, the result is empty")
        log.Warn("through relay pass the message for remote")
        err = relayPassString(rRemote, rHeader, bodyString)
        if err != nil {
            log.Error("when call relayPassString, get error: ", err)
        }
        return 
    }

    if yes, err := isIgnore(r.Header, bodyMap); yes {
        log.Debug("catch the ignore selector, ignore it ......")
        return 
    }

    limitUnit, err := limitGroup.MatchOne(r.Header, bodyMap)
    if err != nil {
        log.Error(err)
        err = relayPassMap(rRemote, rHeader, bodyMap)
        if err != nil {
            log.Error("when call relayPassMap, get error: ", err)
        }
        return 
    }

    unitStatus, err := limitUnit.Increase()
    if err != nil {
        log.Error("when call Increase() of limitUnit, get error: ", err)
        log.Warn("through relay pass the message for remote")
        err = relayPassMap(rRemote, rHeader, bodyMap)
        if err != nil {
            log.Error("when call relayPassMap, get error: ", err)
        }
        return 
    }

    switch unitStatus {
    case LimitFree:
        log.Debugf("limitUnit %s is free", limitUnit.name)
        err = relayPassMap(rRemote, rHeader, bodyMap)
        if err != nil {
            log.Error("when call relayPassMap, get error: ", err)
        }
        return 
    case LimitWork:
        log.Debug("limitUnit %s had been limited, check for cutoff")
        log.Debug("still in inhibition")
        err = reportDBInsert(ReportDB, ReportTableName, bodyMap)
        if err != nil {
            log.Error("when call reportDBInsert, get error: ", err)
        }
        return 
    case LimitBorn:
        log.Debug("start to create an inhibit")
        dataCurrent := map[string]interface{} {
            "EventType": map[string]string{
                "EventType": RelayInhibitionEventType,
            },
            "EventID": RelayInhibitionEventID,
            "Severity": rSeverity,
            "Status": RelayInhibitionStatus,
            "HostName": RelayInhibitionHostName,
            "HostIP": RelayInhibitionHostIP,
            "EventTime": time.Now().Format("2006-01-02 15:04:05"),
            "EventItem": RelayInhibitionEventItem,
            "Channel": RelayInhibitionChannel,
        }
        dataDelay := map[string]interface{} {
            "Details": func() string {
                log.Debugf("the limitUnit %s inhibit's count is %d", limitUnit.name, limitUnit.InhibitUnit.Count)
                return fmt.Sprintf("[%s]规则触发抑制，此次抑制了[%d]条", limitUnit.name, limitUnit.InhibitUnit.Count)
            },
        }
        _ = limitUnit.inhibitCreate(
            relayDelayMap,
            rRemote,
            rHeader,
            dataCurrent,
            dataDelay,
        )
        log.Debug("finish born an inhibit")
        reportDBInsert(ReportDB, ReportTableName, bodyMap)
        return 
    default:
        log.Errorf("unknow limitUnit status code %d from limitUnit %s", unitStatus, limitUnit.name)
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "unknow limitUnit status code")
        return 
    }
}