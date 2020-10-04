package actionCutoff

import (
    "zabbix-robot/action"
    "zabbix-robot/utils"

    "errors"
    "net/http"
    "regexp"

    log "github.com/sirupsen/logrus"
)

type ActionCutoff struct {
    *Action
    msgRegexpOKHeader string
    regexpOkCompile *regexp.Regexp
    msgRegexpProblemHeader string
    regexpProblemCompile *regexp.Regexp
    regexpTagconvCompiles map[string]*regexp.Regexp
}

func (a *ActionCutoff) PreDo(r *http.Request) (header map[string][]string, data map[string]interface{}, tag map[string]string, err error) {
    s, _, _ := utils.bodyToString(r.Body)
    log.WithFields(log.Fields{
        "method": r.Method,
        "body": s,
    }).Debugf("in router %s, get a new requests", a.uri)

    if r.Method != http.MethodPost {
        log.Error("the request method is not expected")
        err = errors.New("please use POST method")
        return
    }

    bodyStatus := r.Header.Get("Status")
    data, tag = a.regexpDeal(bodyStatus, s)

    return r.Header, data, tag, err
}

func (a *Action) PreCheck(header map[string][]string, data map[string]interface{}, tag map[string]string) (err error) {
    
}

func (a *ActionCutoff) SetOkCompile(s string, r *regexp.Regexp) {
    a.msgRegexpOKHeader = s
    a.regexpOkCompile = r
}

func (a *ActionCutoff) SetProblemCompile(s string, r *regexp.Regexp) {
    a.msgRegexpProblemHeader = s
    a.regexpProblemCompile = r
}

func (a *ActionCutoff) SetTagCompile(mr map[string]*regexp.Regexp) {
    a.regexpTagconvCompiles = mr
}

func (a *ActionCutoff) regexpDeal(bodyStatus string, contents string) (msgMap map[string]interface{}, tags map[string]string) {
    log.Debug("start regexpDeal")
    s := string(contents)
    var data []byte
    result := make(map[string]interface{})

    if bodyStatus == a.msgRegexpOKHeader && a.regexpOkCompile != nil {
        match := a.regexpOkCompile.FindStringSubmatch(s)
        groupNames := a.regexpOkCompile.SubexpNames()
        if len(match) != len(groupNames) {
            log.WithFields(log.Fields{
                "LenOfMatch": len(match),
                "LenOfGroupnames": len(groupNames),
                }).Error("The LenOfMatch and LenOfGroupnames are not equel")
            err = errors.New("get error in match")
        } else {
            for i, name := range groupNames {
                if i != 0 && name != "" {
                    result[name] = match[i]
                }
            }
        }
    } else if bodyStatus == a.msgRegexpProblemHeader && a.regexpProblemCompile != nil {
        match := a.regexpProblemCompile.FindStringSubmatch(s)
        groupNames := a.regexpProblemCompile.SubexpNames()
        if len(match) != len(groupNames) {
            log.WithFields(log.Fields{
                "LenOfMatch": len(match),
                "LenOfGroupnames": len(groupNames),
                }).Error("The LenOfMatch and LenOfGroupnames are not equel")
            err = errors.New("get error in match")
        } else {
            for i, name := range groupNames {
                if i != 0 && name != "" {
                    result[name] = match[i]
                }
            }
        }
    } else {
        log.WithFields(log.Fields{
            "bodyStatus": bodyStatus,
            }).Error("not match the bodyStatus")
    }

    if err == nil {
        log.Debug("start use regexp for tag conv")
        for tagKey, tagCompile := range a.regexpTagconvCompiles {
            tagConvRes := make(map[string]string)
            for resKey, resValue := range result {
                if resKey == tagKey {
                    match := tagCompile.FindAllStringSubmatch(resValue.(string), -1)
                    for _, matchArr := range match {
                        if len(matchArr) > 2 {
                            tagConvRes[strings.TrimSpace(matchArr[1])] = strings.TrimSpace(matchArr[2])
                        }
                    }
                }
            }
            result[tagKey] = tagConvRes
            tags = tagConvRes
        }
    }

    data, errJson := json.Marshal(result)
    if errJson != nil {
        log.WithFields(log.Fields{
            "error": errJson,
            }).Error("try to Marshal the result to json is failed")
    }

    err = json.Unmarshal(data, &msgMap)

    if err != nil || len(data) == 0 {
        log.Error("found an error, change to original content to body")
        return contents
    }

    log.Debug("finish deal with request msg with regexp successfully")
    return msgMap, tags
}