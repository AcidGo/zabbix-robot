package main

import (
    "bytes"
    "io"
    "io/ioutil"
    "os"

    log "github.com/sirupsen/logrus"
)


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

func bodyToString(body io.ReadCloser) (string, int64, io.ReadCloser) {
    contents, err := ioutil.ReadAll(body)
    if err != nil {
        log.WithFields(log.Fields{
            "error": err,
            }).Error("in bodyToString, get error when read all the body")
    }

    length := int64(len(contents))

    newReadCloser := ioutil.NopCloser(bytes.NewReader(contents))
    s := string(contents)
    return s, length, newReadCloser
}

func isIgnore(header map[string][]string, content map[string]interface{}) (bool, error) {
    var err error

    for key, val := range header {
        if v, ok := IgnoreSelector[key]; ok {
            for _, subVal := range val {
                if v == subVal {
                    return true, nil
                }
            }
        }
    }

    for key, val := range content {
        if v, ok := IgnoreSelector[key]; ok {
            var tmpS string
            switch val.(type) {
            case int:
                tmpS = strconv.Itoa(val)
            case string:
                tmpS = val
            default:
                log.Warnf("cannot case the type of %v", val)
                continue
            }
            if v == tmpS {
                return true, nil
            }
        }
    }

    return false, nil
}

func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}

func regexpDeal(regexpPlan string, bodyStr string) map[string]interface{} {
    log.Debug("into regexpDeal ......")
    var err error
    result := make(map[string]interface{})
    var data []byte

    if regexpPlan == MsgRegexpOkHeader && RegexpOkCompile != nil {
        match := RegexpOkCompile.FindStringSubmatch(bodyStr)
        groupNames := RegexpOkCompile.SubexpNames()
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
    } else if regexpPlan == MsgRegexpProblemHeader && RegexpProblemCompile != nil {
        match := RegexpProblemCompile.FindStringSubmatch(bodyStr)
        groupNames := RegexpProblemCompile.SubexpNames()
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
            "regexpPlan": regexpPlan,
            }).Error("not match the regexpPlan")
    }

    if err == nil {
        log.Debug("start use regexp for tag conv")
        for tagKey, tagCompile := range RegexpTagconvCompiles {
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
        }
    }

    data, errJson := json.Marshal(result)
    if errJson != nil {
        log.WithFields(log.Fields{
            "error": errJson,
            }).Error("try to Marshal the result to json is failed")
    }

    if err != nil || len(data) == 0 {
        log.Error("found an error, change to original content to body")
        return result
    }

    log.Debug("finish deal with request msg with regexp successfully")
    return result
}

func repairHeader(header map[string][]string, ignoreSelector []string) map[string][]string {
    newHeader := make(map[string][]string)
    for key, val := range header {
        if _, hasFind := Find(ignoreSelector, key); hasFind {
            continue
        }
        newHeader[key] = val
    }

    return newHeader
}