package http

import (
    "fmt"
    "encoding/json"
    "net/http"
    "reflect"
    "strings"

    "github.com/AcidGo/zabbix-robot/state"
    "github.com/AcidGo/zabbix-robot/utils"
)

func sendThroughString(remote string, header map[string][]string, s string) (string, error) {
    body, length, _ := utils.StringToBody(s)
    client := &http.Client{}
    req, err := http.NewRequest("POST", remote, nil)
    if err != nil {
        return "", err
    }

    req.Header = header
    req.Body = body
    req.ContentLength = length
    rsp, err := client.Do(req)
    if err != nil {
        return "", err
    }

    res, _, _ := utils.BodyToString(rsp.Body)
    return res, nil
}

func sendThroughMap(remote string, header map[string][]string, m map[string]interface{}) (string, int, error) {
    data, err := json.Marshal(m)
    if err != nil {
        return "", nil, err
    }

    body, length, err := utils.BytesToBody(data)
    if err != nil {
        return "", err
    }

    client := &http.Client{}
    req, err := http.NewRequest("POST", remote, nil)
    if err != nil {
        return "", err
    }

    req.Header = header
    req.Body = body
    req.ContentLength = length
    rsp, err := client.Do(req)
    if err != nil {
        return "", err
    }

    res, _, _ := utils.BodyToString(rsp.Body)
    return res, nil
}

func SendThrough(remote string, header map[string][]string, data interface{}, s *state.SvrStater) (string, error) {
    var err error
    var res string
    switch data.(type) {
    case string:
        res, err = sendThroughString(remote, header, data.(string))
    case map[string]interface{}:
        res, err = sendThroughMap(remote, header, data.(map[string]interface{}))
    default:
        err = fmt.Errorf("not support the send through type: %T", data)
    }

    return res, err
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

    res, err := sendThroughMap(remote, header, data)
    if err != nil {
        return res, err
    }

    return res, nil
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