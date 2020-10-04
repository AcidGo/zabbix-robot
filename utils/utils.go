package utils

import (
    "io"
    "io/ioutil"
    "reflect"
)

func bodyToString(body io.ReadCloser) (string, io.ReadCloser, int64) {
    contents, err := ioutil.ReadAll(body)
    if err != nil {
        log.WithFields(log.Fields{
            "error": err,
            }).Error("in bodyToString, get error when readall the body")
    }
    length := int64(len(contents))
    newReadCloser := ioutil.NopCloser(bytes.NewReader(contents))
    s := string(contents)
    return s, newReadCloser, length
}

func relayPass(remote string, header map[string][]string, body io.ReadCloser) error {
    s, body, length := bodyToString(body)
    log.WithFields(log.Fields{
        "body": s,
        }).Info("into relayPass, get the body string format")

    client := &http.Client{}
    req, err := http.NewRequest("POST", remote, nil)
    if err != nil {
        return err
    }

    req.Header = header
    req.Body = body
    req.ContentLength = length
    rsp, err := client.Do(req)
    if err != nil {
        log.Error("relayPass get error:", err)
        return err
    }
    log.WithFields(log.Fields{
        "respon": rsp,
        }).Debug("respon from remote http server")
    s, _, _ = bodyToString(rsp.Body)
    log.WithFields(log.Fields{
        "responBody": s,
        }).Debug("string response body from remote http server")

    return nil
}

func relayDelay(remote string, header map[string][]string, data map[string]interface{}, dataDelay map[string]interface{}) error {
    log.Debug("before all cook, the length of data is: ", len(data))
    for k, v := range dataDelay {
        if IsFunc(v) {
            log.WithFields(log.Fields{
                "key": k,
                }).Debug("in relayDelay, found a func value")
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
        log.Debug("after cook, the type of delay data is: ", data[k])
    }
    log.Debug("after all cook, the length of data is: ", len(data))
    bodyJson, _ := json.Marshal(data)
    body := ioutil.NopCloser(bytes.NewReader(bodyJson))
    err := relayPass(remote, header, body)

    return err
}

func IsFunc(v interface{}) bool {
   return reflect.TypeOf(v).Kind() == reflect.Func
}

func FileExists(path string) bool {
    _, err := os.Stat(path)
    if err != nil {
        if os.IsExist(err) {
            return true
        }
        return false
    }
    return true
}

