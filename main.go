package main

func init() {
    var err error

    err = initLog()
    if err != nil {
        log.Fatal("get err when init log: ", err)
    }

    err = initFlag()
    if err != nil {
        log.Fatal("get err when init flag: ", err)
    }

    err = initConfigParse()
    if err != nil {
        log.Fatal("get err when init parse config: ", err)
    }
    log.Debug("finished to load config from ", programConfigPath)

    err = initRegexp()
    if err != nil {
        log.Fatal("get err when init regexp: ", err)
    }

    err = initLimitUnitGroup()
    if err != nil {
        log.Fatal("get err when init limitUnitGroup: ", err)
    }

    err = initReport()
    if err != nil {
        log.Fatal("get err when init reportor: ", err)
    }
}

func purge() {
    if reportDB != nil {
        reportDB.Close()
    }
}

func main() {
    log.Printf("%s written by %s, the version is %s", programName, programAuthor, programVersion)
    log.Printf("the %s is starting to run ......", programName)
    http.HandleFunc(httpURL, mainHandler)
    http.ListenAndServer(httpListenPort, nil)
    purge()
}