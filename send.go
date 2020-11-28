package main

import (
    "errors"
)


func sendPreCheck(header map[string][]string, content map[string]interface{}) error {
    if _, ok := header[httpHeaderFieldRemote]; !ok {
        return errors.New("not found the remote field %s in the request header", httpHeaderFieldRemote)
    }

    if len(content) == 0 {
        return errors.New("the content from request is empty")
    }

    return nil
}