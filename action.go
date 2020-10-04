package action

import (
    "zabbix-robot/limitUnit"
    "net/http"
)

type ILimitUnitGroup interface {
    Increase(map[string][]string, map[string]interface{}, map[string]string) (limitUnit.LUStateCode, *limitUnit.ILimitUnit)
}

type Action struct {
    uri         string
}

func NewAction(uri string) *Action {
    return &Action{uri: uri}
}

func (a *Action) Handler(w http.ResponseWriter, r *http.Request) {
    var newHeader map[string][]string
    var newData map[string]interface{}
    var msgTag map[string]string
    var err error

    newHeader, newData, msgTag, err = a.PreDo(r)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, err.Error())
        return
    }

    err = a.PreCheck(newHeader, newData, msgTag)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, err.Error())
        return
    }

    err = a.ExecDo(newHeader, newData, msgTag)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, err.Error())
        return
    }

    return
}

func (a *Action) PreDo(r *http.Request) (header map[string][]string, data map[string]interface{}, tag map[string]string, err error) {
    return 
}

func (a *Action) PreCheck(header map[string][]string, data map[string]interface{}, tag map[string]string) (err error) {
    return
}

func (a *Action) ExecDo(header map[string][]string, data map[string]interface{}, tag map[string]string) (err error) {
    return
}