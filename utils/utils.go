package utils

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "io/ioutil"
    "os"
    "reflect"
    "regexp"
    "strings"
)

type DelayFunc func(remote string, header map[string][]string, data map[string]interface{}, delayData map[string]interface{}) error

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

func IsFunc(v interface{}) bool {
   return reflect.TypeOf(v).Kind() == reflect.Func
}

func BytesToBody(data []byte) (io.ReadCloser, int64, error) {
    body := ioutil.NopCloser(bytes.NewReader(data))
    return body, int64(len(data)), nil
}

func BodyToString(body io.ReadCloser) (string, int64, error) {
    t, err := ioutil.ReadAll(body)
    if err != nil {
        return "", 0, err
    }

    length := int64(len(t))
    s := string(t)
    return s, length, nil
}

func StringToBody(content string) (io.ReadCloser, int64, error) {
    data := []byte(content)
    return BytesToBody(data)
}

func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}

func RegexpDealContent(content string, compile *regexp.Regexp) (map[string]interface{}, error) {
    var err error
    var data []byte
    result := make(map[string]interface{})

    if compile == nil {
        return result, errors.New("the regexp's compile is nil")
    }

    match := compile.FindStringSubmatch(content)
    groupNames := compile.SubexpNames()
    if len(match) != len(groupNames) {
        return result, errors.New("the number of matched item is not equal the group size")
    }

    for i, name := range groupNames {
        if i != 0 && name != "" {
            result[name] = match[i]
        }
    }

    data, err = json.Marshal(result)
    if err != nil {
        return result, err
    }

    if len(data) == 0 && len(content) != 0 {
        return result, errors.New("after conv, length of data is not equal raw content")
    }

    return result, nil
}

func RegexpDealTag(data map[string]interface{}, compiles map[string]*regexp.Regexp) (map[string]interface{}, error) {
    for key, compile := range compiles {
        result := make(map[string]string)
        for dataK, dataV := range data {
            if dataK == key {
                match := compile.FindAllStringSubmatch(dataV.(string), -1)
                for _, arr := range match {
                    if len(arr) > 2 {
                        result[strings.TrimSpace(arr[1])] = strings.TrimSpace(arr[2])
                    }
                }
                data[key] = result
            }
        }
    }

    return data, nil
}

func MapGet(m map[string]interface{}, k string, v string) interface{} {
    if x, found := m[k]; found {
        return x
    } else {
        return v
    }
}