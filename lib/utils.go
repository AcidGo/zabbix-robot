package utils

import (
    "errors"
    "emcdoing/json"
    "regexp"
)

func RegexpGenMap(content string, compile *regexp.Regexp) (m map[string]interface{}, err error) {
    m = make(map[string]interface{})

    if compile == nil {
        err = errors.New("regexp compile is nil")
        return 
    }

    match := compile.FindStringSubmatch(content)
    groupNames := compile.SubexpNames()
    if len(match) != len(groupNames) {
        err = errors.New("number of matched items is not equal the group size")
        return 
    }

    for i, name := range groupNames {
        if i != 0 && name != "" {
            m[name] = match[i]
        }
    }

    var data []byte
    data, err = json.Marshal(m)
    if err != nil {
        return 
    }

    if len(data) == 0 && len(content) != 0 {
        err = errors.New("result of regexp is not equal content's length")
        return 
    }

    return m, nil
}

func MapToStruct(m map[string]interface{}, v interface{}) (error) {
    j, err := json.Marshal(m)
    if err != nil {
        return err
    }

    err = json.Unmarshal(j, v)
    if err != nil {
        return err
    }

    return nil
}

func ExtractTags(tagKey)