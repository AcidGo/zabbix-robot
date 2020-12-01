package main

import (
    "encoding/json"
    "log"
    "io/ioutil"
    "net/http"
)


func main() {
    log.Println("mock of zabbix-robot written by AcidGo")
    http.HandleFunc("/zabbix", func (w http.ResponseWriter, r *http.Request) {
        log.Println("Header:", r.Header)
        s, _ := ioutil.ReadAll(r.Body)
        log.Println(string(s))
        var jsonRes interface{}
        err := json.Unmarshal(s, &jsonRes)
        if err != nil {
            log.Println("get error when Unmarshal json")
            log.Println(err)
        } else {
            log.Printf("after json: %v\n", jsonRes)
        }
    })
    http.ListenAndServe(":19952", nil)
}
