package http

import (
    "fmt"
    "encoding/json"
    "net/http"
    "reflect"
    "strings"

    "github.com/AcidGo/zabbix-robot/utils"
    log "github.com/sirupsen/logrus"
)

func sendThroughString(remote string, header map[string][]string, s string) (string, int, error) {
    log.Debugf("call sendThroughString, the content is: %s", s)
    body, length, _ := utils.StringToBody(s)
    client := &http.Client{}
    req, err := http.NewRequest("POST", remote, nil)
    if err != nil {
        return "", 0, err
    }

    req.Header = header
    req.Body = body
    req.ContentLength = length
    rsp, err := client.Do(req)
    if err != nil {
        return "", 0, err
    }

    res, _, _ := utils.BodyToString(rsp.Body)
    return res, rsp.StatusCode, nil
}

func sendThroughMap(remote string, header map[string][]string, m map[string]interface{}) (string, int, error) {
    data, err := json.Marshal(m)
    if err != nil {
        return "", 0, err
    }

    body, length, err := utils.BytesToBody(data)
    if err != nil {
        return "", 0, err
    }

    client := &http.Client{}
    req, err := http.NewRequest("POST", remote, nil)
    if err != nil {
        return "", 0, err
    }

    req.Header = header
    req.Body = body
    req.ContentLength = length
    rsp, err := client.Do(req)
    if err != nil {
        return "", 0, err
    }

    res, _, _ := utils.BodyToString(rsp.Body)
    return res, rsp.StatusCode, nil
}

func SendThrough(remote string, header map[string][]string, data interface{}) (string, int, error) {
    var err error
    var res string
    var statusCode int
    switch data.(type) {
    case string:
        res, statusCode, err = sendThroughString(remote, header, data.(string))
    case map[string]interface{}:
        res, statusCode, err = sendThroughMap(remote, header, data.(map[string]interface{}))
    default:
        err = fmt.Errorf("not support the send through type: %T", data)
    }

    return res, statusCode, err
}

func SendDelayMap(remote string, header map[string][]string, data map[string]interface{}, dataDelay map[string]interface{}) (string, error) {
    var err error
    for k, v := range dataDelay {
        if utils.IsFunc(v) {
            f := reflect.ValueOf(v)
            vv := f.Call([]reflect.Value{})
            ss := make([]string, len(vv))
            for _, i := range vv {
                if len(strings.Trim(i.String(), " ")) == 0 {
                    continue
                }
                ss = append(ss, strings.Trim(i.String(), " "))
            }
            data[k] = strings.Join(ss, "")
        } else {
            data[k] = v
        }
    }

    res, _, err := sendThroughMap(remote, header, data)
    return res, err
}

func RepairHeader(header http.Header, ignoreFileds []string) (http.Header, error) {
    var err error
    newHeader := make(map[string][]string)

    for k, v := range header {
        if _, ok := utils.Find(ignoreFileds, k); ok {
            continue
        }
        newHeader[k] = v
    }

    return newHeader, err
}