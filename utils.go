package main

import (
    "io"
    "os"
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