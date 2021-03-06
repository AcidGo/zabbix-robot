package main

import (
    "bufio"
    "database/sql"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "net/http"
    "os"
    "regexp"
    "time"

    "github.com/AcidGo/zabbix-robot/config"
    achttp "github.com/AcidGo/zabbix-robot/http"
    "github.com/AcidGo/zabbix-robot/ignore"
    "github.com/AcidGo/zabbix-robot/limit"
    "github.com/AcidGo/zabbix-robot/state"
    "github.com/AcidGo/zabbix-robot/utils"

    "gopkg.in/ini.v1"
    "github.com/lestrrat/go-file-rotatelogs"
    lfshook "github.com/rifflock/lfshook"
    log "github.com/sirupsen/logrus"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
)

var (
    // app info
    AppName                             string
    AppAuthor                           string
    AppVersion                          string
    AppGitCommitHash                    string
    AppBuildTime                        string
    AppGoVersion                        string
)


func init() {
    var err error

    err = initLog()
    if err != nil {
        log.Fatal("get err when init log: ", err)
    }
    log.Debug("finished to init logger")

    err = initFlag()
    if err != nil {
        log.Fatal("get err when init flag: ", err)
    }
    log.Debug("finished to init flag parse")

    err = initConfigParse()
    if err != nil {
        log.Fatal("get err when init parse config: ", err)
    }
    log.Debug("finished to load config from ", config.AppConfigPath)

    err = refreshLog()
    if err != nil {
        log.Fatal("get err when re-fresh log config: ", err)
    }
    log.Debug("finished to re-fresh logger by config")

    err = initIgnore()
    if err != nil {
        log.Fatal("get err when init ignore: ", err)
    }
    log.Debug("finished to init ignore unit by config")

    err = initFormat()
    if err != nil {
        log.Fatal("get err when init regexp: ", err)
    }
    log.Debug("finished to init formater by config")

    err = initTagconv()
    if err != nil {
        log.Fatal("get err when init tag conv: ", err)
    }
    log.Debug("finished to init tag conv by config")

    err = initLimitGroup()
    if err != nil {
        log.Fatal("get err when init limit group: ", err)
    }
    log.Debug("finished to init limit group by config")

    err = initReport()
    if err != nil {
        log.Error("get err when init reportor: ", err)
    } else {
        log.Debug("finished to init report channel by config")
    }

    err = initServerState()
    if err != nil {
        log.Error("get err when init server state: ", err)
    } else {
        log.Debug("finished to init server state")
    }
}

func initLog() error {
    customFormatter := new(log.TextFormatter)
    customFormatter.TimestampFormat = "2006-01-02 15:04:05.000000000"
    customFormatter.FullTimestamp = true
    customFormatter.DisableTimestamp = false
    log.SetFormatter(customFormatter)
    log.SetOutput(os.Stdout)
    log.SetLevel(config.AppInitLogLevel)
    return nil
}

func initFlag() error {
    flag.UintVar(&config.LogLevel, "l", 4, "the `level` of the log")
    flag.StringVar(&config.AppConfigPath, "f", config.ConfigDefaultName, "set `config` for the program")

    flag.Usage = flagUsage
    flag.Parse()

    if config.AppConfigPath == "" {
        log.Error("the configure file is nil, please input it")
        err := errors.New("the configure is not defined")
        return err
    }

    if !utils.FileExists(config.AppConfigPath) {
        log.WithFields(log.Fields{
            "path": config.AppConfigPath,
            }).Error("the path of configure defined is not exists")
        err := errors.New("the configure file is not exists")
        return err
    }

    return nil
}

func initConfigParse() error {
    var err error

    log.Debug("parsing the conf for loading confif file")
    config.AppIniFile, err = ini.Load(config.AppConfigPath)
    if err != nil {
        return err
    }

    config.RelayInhibitionEventType    = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionEventType).String()
    log.Debugf("load parameter config.RelayInhibitionEventType: %s", config.RelayInhibitionEventType)
    config.RelayInhibitionEventID      = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionEventID).String()
    log.Debugf("load parameter config.RelayInhibitionEventID: %s", config.RelayInhibitionEventID)
    config.RelayInhibitionStatus       = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionStatus).String()
    log.Debugf("load parameter config.RelayInhibitionStatus: %s", config.RelayInhibitionStatus)
    config.RelayInhibitionHostName     = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionHostName).String()
    log.Debugf("load parameter config.RelayInhibitionHostName: %s", config.RelayInhibitionHostName)
    config.RelayInhibitionHostIP       = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionHostIP).String()
    log.Debugf("load parameter config.RelayInhibitionHostIP: %s", config.RelayInhibitionHostIP)
    config.RelayInhibitionEventItem    = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionEventItem).String()
    log.Debugf("load parameter config.RelayInhibitionEventItem: %s", config.RelayInhibitionEventItem)
    config.RelayInhibitionChannel      = config.AppIniFile.Section(config.ConfigSectionRelay).Key(config.ConfigRelayKeyRelayInhibitionChannel).String()
    log.Debugf("load parameter config.RelayInhibitionChannel: %s", config.RelayInhibitionChannel)

    config.HttpListenPort = config.AppIniFile.Section(config.ConfigSectionMain).Key(config.ConfigMainKeyListen).String()
    log.Debugf("load parameter config.HttpListenPort: %s", config.HttpListenPort)
    config.HttpURL        = config.AppIniFile.Section(config.ConfigSectionMain).Key(config.ConfigMainKeyURL).String()
    log.Debugf("load parameter config.HttpURL: %s", config.HttpURL)

    return nil
}

func initIgnore() error {
    var err error

    config.IgnoreUnit = ignore.NewIgnoreUnit()
    for _, subSection := range config.AppIniFile.ChildSections(config.ConfigSectionIgnore) {
        var t map[string][]string
        err = json.Unmarshal([]byte(subSection.Key(config.ConfigIgnoreKeySetting).String()), &t)
        if err != nil {
            return err
        }
        iRole := ignore.IgnoreRole{Key: subSection.Name(), Val: t}
        err = config.IgnoreUnit.AddRole(iRole)
        if err != nil {
            return err
        }
        log.Infof("added ignore role for %s: %v", iRole.Key, iRole.Val)
    }

    return nil
}

func initFormat() error {
    var err error

    config.FormatRegexpHeaderField = config.AppIniFile.Section(config.ConfigSectionFormat).Key(config.ConfigFormatKeyRegexpHeaderField).String()
    log.Debugf("load parameter config.FormatRegexpHeaderField: %s", config.FormatRegexpHeaderField)
    config.FormatRegexpOkHeader = config.AppIniFile.Section(config.ConfigSectionFormat).Key(config.ConfigFormatKeyRegexpOkHeader).String()
    log.Debugf("load parameter config.FormatRegexpOkHeader: %s", config.FormatRegexpOkHeader)
    config.FormatRegexpOkCompileString = config.AppIniFile.Section(config.ConfigSectionFormat).Key(config.ConfigFormatKeyRegexpOkCompile).String()
    log.Debugf("load parameter config.FormatRegexpOkCompileString: %s", config.FormatRegexpOkCompileString)
    config.FormatRegexpProblemHeader = config.AppIniFile.Section(config.ConfigSectionFormat).Key(config.ConfigFormatKeyRegexpProblemHeader).String()
    log.Debugf("load parameter config.FormatRegexpProblemHeader: %s", config.FormatRegexpProblemHeader)
    config.FormatRegexpProblemCompileString = config.AppIniFile.Section(config.ConfigSectionFormat).Key(config.ConfigFormatKeyRegexpProblemCompile).String()
    log.Debugf("load parameter config.FormatRegexpProblemCompileString: %s", config.FormatRegexpProblemCompileString)

    config.FormatRegexpOkCompile, err = regexp.Compile(config.FormatRegexpOkCompileString)
    if err != nil {
        return err
    }
    config.FormatRegexpProblemCompile, err = regexp.Compile(config.FormatRegexpProblemCompileString)
    if err != nil {
        return err
    }

    return nil
}

func initTagconv() error {
    var err error
    config.TagconvCompiles = make(map[string]*regexp.Regexp)

    for _, keyTagconv := range config.AppIniFile.Section(config.ConfigSectionTagconv).Keys() {
        config.TagconvCompiles[keyTagconv.Name()], err = regexp.Compile(keyTagconv.String())
        if err != nil {
            return err
        }
        log.Debug("load tgconv key: ", keyTagconv.Name())
    }

    return nil
}

func initLimitGroup() error {
    var err error

    config.LimitGroup = limit.NewLimitGroup()
    for _, subSection := range config.AppIniFile.ChildSections(config.ConfigSectionRole) {
        err = config.LimitGroup.AddUnit(subSection)
        if err != nil {
            return err
        }
        log.Infof("added an limit unit %s for limit group", subSection.Name())
    }

    return nil
}

func initReport() error {
    var err error

    config.ReportDBType = config.AppIniFile.Section(config.ConfigSectionReport).Key(config.ConfigReportKeyDBType).String()
    log.Debugf("load parameter config.ReportDBType: %s", config.ReportDBType)
    config.ReportDBDsn = config.AppIniFile.Section(config.ConfigSectionReport).Key(config.ConfigReportKeyDBDsn).String()
    config.ReportTableName = config.AppIniFile.Section(config.ConfigSectionReport).Key(config.ConfigReportKeyTable).String()
    log.Debugf("load parameter config.ReportTableName: %s", config.ReportTableName)

    config.ReportDB, err = sql.Open(config.ReportDBType, config.ReportDBDsn)
    if err != nil {
        return err
    }
    err = config.ReportDB.Ping()
    if err != nil {
        return err
    }
    return nil
}

func initServerState() error {
    if state.SState == nil {
        return errors.New("the server state is nil")
    }
    err := state.SState.Reset()
    return err
}

func flagUsage() {
    usageMsg := fmt.Sprintf(`%s
Version: %s
Author: %s
GitCommit: %s
BuildTime: %s
GoVersion: %s
Usage: %s [-l level] [-f config]
Options:
`, AppName, AppVersion, AppAuthor, AppGitCommitHash, AppBuildTime, AppGoVersion, AppName)

    fmt.Fprintf(os.Stderr, usageMsg)
    flag.PrintDefaults()
}

func refreshLog() error {
    s, err := config.AppIniFile.GetSection(config.ConfigSectionLog)
    if err != nil {
        log.WithFields(log.Fields{
            "section": config.ConfigSectionLog,
        }).Error("the section is not exist")
        return err
    }
    config.LogFilePath = s.Key(config.ConfigLogKeyLogFile).String()
    config.LogLevel, err = s.Key(config.ConfigLogKeyLogLevel).Uint()
    if err != nil {
        log.Errorf("the config value of key %s must be uint", config.ConfigLogKeyLogLevel)
        return err
    }

    log.SetLevel(log.Level(config.LogLevel))
    if config.LogFilePath == "" {
        log.SetOutput(os.Stdout)
    } else {
        log.Info("starting to change log mode ......")
        writer, err := rotatelogs.New(
            config.LogFilePath + ".%Y%m%d%H%M",
            rotatelogs.WithLinkName(config.LogFilePath),
            rotatelogs.WithMaxAge(7*24*time.Hour),
        )
        if err != nil {
            log.Error("generate rotatelogs writer failed")
            return err
        }

        src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
        if err != nil {
            log.Error("generate devnull file handler failed")
            return err
        }
        devNullWriter := bufio.NewWriter(src)
        log.SetOutput(devNullWriter)

        lfHook := lfshook.NewHook(
            lfshook.WriterMap{
                log.DebugLevel:     writer,
                log.InfoLevel:      writer,
                log.WarnLevel:      writer,
                log.ErrorLevel:     writer,
                log.FatalLevel:     writer,
                log.PanicLevel:     writer},
            &log.TextFormatter{
                TimestampFormat: "2006-01-02 15:04:05.000000000",
                FullTimestamp:    true,
                DisableTimestamp: false,
                },
        )
        log.AddHook(lfHook)
    }

    log.Info("log re-fresh config finished")
    return nil
}

func reportDBInsert(db *sql.DB, tableName string, resMap map[string]interface{}) error {
    if db == nil {
        return errors.New("the report database is nil, exit report for the database")
    }

    sql := fmt.Sprintf("insert into %s values($1, $2, $3, $4, $5, $6, $7, $8, $9)", tableName)
    if len(resMap) <= 0 {
        err := errors.New("in reportDBInsert, the message counld not conv to the map")
        log.Error(err)
        return err
    }
    nowTime := time.Now().Unix()
    eventID := utils.MapGet(resMap, "EventID", "").(string)
    severity := utils.MapGet(resMap, "Severity", "").(string)
    status_ := utils.MapGet(resMap, "Status", "").(string)
    hostName := utils.MapGet(resMap, "HostName", "").(string)
    eventItem := utils.MapGet(resMap, "EventItem", "").(string)
    eventTime := utils.MapGet(resMap, "EventTime", "").(string)
    recoverTime := utils.MapGet(resMap, "RecoverTime", "").(string)
    details := utils.MapGet(resMap, "Details", "").(string)

    _, err := db.Exec(
        sql,
        nowTime, 
        eventID, 
        severity, 
        status_, 
        hostName, 
        eventItem, 
        eventTime, 
        recoverTime, 
        details, 
    )
    if err != nil {
        log.Error("in reportDBInsert, insert into table get error: ", err)
        return err
    }
    return nil
}

func purgeEnv() {
    if config.ReportDB != nil {
        config.ReportDB.Close()
    }

    log.Info("finished purge the app env")
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    var jData []byte

    data := state.SState.GetState()
    log.Debug("get a new stat request")
    jData, err = json.Marshal(data)
    if err != nil {
        log.Error("get an err when deal with json for state:", err)
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, err.Error())
        return 
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(jData)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    var bodyString string
    var bodyMap map[string]interface{}
    var rRemote string
    var rHeader map[string][]string
    var rRsp string
    var StatusCode int

    log.Debug("get a new accessing request")
    if err = state.SState.IncreaseState(state.RequestSum); err != nil {
        log.Errorf("get an err when increase %s state: %s", state.RequestSum, err)
    }

    bodyString, _, err = utils.BodyToString(r.Body)
    if err != nil {
        log.Error("get an err when conv the request body to string: ", err)
    }
    log.Debug("request's body: ", bodyString)
    log.Debug("requst's header: ", r.Header)

    if r.Method != http.MethodPost {
        log.WithFields(log.Fields{
            "method": r.Method,
        }).Error("the method of request is not expected")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "please POST method")
        return 
    }

    rRemote = r.Header.Get(config.InnerRemoteAddrField)
    if len(rRemote) == 0 {
        err = errors.New("not found InnerRemoteAddrField in the header")
        log.Error(err)
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, err)
        return 
    }
    rHeader, err = achttp.RepairHeader(r.Header, config.RelayIgnoreHeaderFields)
    if err != nil {
        log.Error("get an err when repair header: ", err)
        log.Warn("using the raw http header")
        rHeader = r.Header
    }

    err = nil
    formatFlag := r.Header.Get(config.FormatRegexpHeaderField)
    log.Debugf("get format flag: %s", formatFlag)
    switch formatFlag {
    case config.FormatRegexpOkHeader:
        bodyMap, err = utils.RegexpDealContent(bodyString, config.FormatRegexpOkCompile)
    case config.FormatRegexpProblemHeader:
        bodyMap, err = utils.RegexpDealContent(bodyString, config.FormatRegexpProblemCompile)
    default:
        err = errors.New("not suppot the format flag")
    }
    if err != nil {
        log.Error("cannot format the body string to map: ", err)
        log.Warn("so through send it")
        if err := state.SState.IncreaseState(state.ContentDealFailed); err != nil {
            log.Errorf("get an err when increase %s state: %s", state.ContentDealFailed, err)
        }
        rRsp, StatusCode, err = achttp.SendThrough(rRemote, rHeader, bodyString)
        if err != nil {
            log.Error("get an err when send http request: ", err)
            if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
            }
        } else {
            log.Debugf("get the response [%d] from %s: %s", StatusCode, rRemote, rRsp)
            if StatusCode != 200 {
                if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                    log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
                }
            }
        }
        return 
    } else {
        log.Debug("deal with body string for conving map")
    }
    if len(bodyMap) == 0 {
        log.Error("after format body to map, the length of result is zero")
        log.Warn("so through send it")
        rRsp, StatusCode, err = achttp.SendThrough(rRemote, rHeader, bodyString)
        if err != nil {
            log.Error("get an err when send http request: ", err)
            if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
            }
        } else {
            log.Debugf("get the response [%d] from %s: %s", StatusCode, rRemote, rRsp)
            if StatusCode != 200 {
                if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                    log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
                }
            }
        }
        return
    }

    if yes, iRoleName, _ := config.IgnoreUnit.IsIgnore(bodyMap); yes {
        log.Infof("mean the ignore unit for ignoring the request, with role %s", iRoleName)
        return 
    }

    // conv the inner tag from string to map[string]string
    bodyMap, err = utils.RegexpDealTag(bodyMap, config.TagconvCompiles)
    if err != nil {
        log.Error("get an err when conv tag for body map: ", err)
    }

    lUnit, err := config.LimitGroup.MatchOne(bodyMap)
    if err != nil {
        log.Error("get an err when matach one limit unit from limit group: ", err)
        log.Warn("so through send it")
        rRsp, StatusCode, err = achttp.SendThrough(rRemote, rHeader, bodyMap)
        if err != nil {
            log.Error("get an err when send http request: ", err)
            if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
            }
        } else {
            log.Debugf("get the response [%d] from %s: %s", StatusCode, rRemote, rRsp)
            if StatusCode != 200 {
                if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                    log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
                }
            }
        }

        return 
    }

    dataCurrent := map[string]interface{} {
        "EventType": map[string]string{
            "EventType": config.RelayInhibitionEventType,
        },
        "EventID": config.RelayInhibitionEventID,
        "Severity": r.Header.Get(config.InnerHeaderSeverityField),
        "Status": config.RelayInhibitionStatus,
        "HostName": config.RelayInhibitionHostName,
        "HostIP": config.RelayInhibitionHostIP,
        "EventTime": time.Now().Format("2006-01-02 15:04:05"),
        "EventItem": config.RelayInhibitionEventItem,
        "Channel": config.RelayInhibitionChannel,
    }
    dataDelay := map[string]interface{} {
        "Severity": lUnit.GetSeverity,
        "Details": func() string {
            lName := lUnit.GetName()
            iCount := lUnit.InhibitUnit.Count
            log.Debugf("the limitUnit %s inhibit's count is %d", lName, iCount)
            return fmt.Sprintf("[%s]规则触发抑制，此次抑制了[%d]条", lName, iCount)
        },
    }
    tagRewriteMap := lUnit.GetTagRewriteMap()
    if len(tagRewriteMap) > 0 {
        dataDelay["EventType"] = tagRewriteMap
    }

    lUnitState, err := lUnit.Increase(
        achttp.SendDelayMap,
        rRemote,
        rHeader,
        dataCurrent,
        dataDelay,
    )
    if err != nil {
        log.Errorf("get an err when increase the limit unit %s: %s", lUnit.GetName(), err)
        log.Warn("so through send it")
        rRsp, StatusCode, err = achttp.SendThrough(rRemote, rHeader, bodyMap)
        if err != nil {
            log.Error("get an err when send http request: ", err)
            if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
            }
        } else {
            log.Debugf("get the response [%d] from %s: %s", StatusCode, rRemote, rRsp)
            if StatusCode != 200 {
                if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                    log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
                }
            }
        }
        return 
    }

    switch lUnitState {
    case limit.LimitStateFree:
        log.Debugf("limit unit %s is free now", lUnit.GetName())
        rRsp, StatusCode, err = achttp.SendThrough(rRemote, rHeader, bodyMap)
        if err != nil {
            log.Error("get an err when send http request: ", err)
            if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
            }
        } else {
            if StatusCode != 200 {
                if err := state.SState.IncreaseState(state.ResponseFailed); err != nil {
                    log.Errorf("get an err when increase %s state: %s", state.ResponseFailed, err)
                }
            }
        }
        log.Debugf("get the response [%d] from %s: %s", StatusCode, rRemote, rRsp)

        return 
    case limit.LimitStateWork:
        log.Infof("limit unit %s had been limited", lUnit.GetName())
        err = reportDBInsert(config.ReportDB, config.ReportTableName, bodyMap)
        if err != nil {
            log.Error("get an err when insert to report db: ", err)
        }
        return 
    case limit.LimitStateBorn:
        log.Infof("creating an inhibit on %s", lUnit.GetName())
        err = reportDBInsert(config.ReportDB, config.ReportTableName, bodyMap)
        if err != nil {
            log.Error("get an err when insert to report db: ", err)
        }
        return 
    default:
        err = fmt.Errorf("unknow limit unit state %d from limit unit %s", lUnitState, lUnit.GetName())
        log.Error(err)
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, err)
        return 
    }
}

func main() {
    log.Printf("%s written by %s, the version is %s", AppName, AppAuthor, AppVersion)
    http.HandleFunc(config.HttpURL, mainHandler)
    // hard-code for the state check route
    http.HandleFunc("/state", stateHandler)
    err := http.ListenAndServe(config.HttpListenPort, nil)
    if err != nil {
        log.Error(err)
    }
    purgeEnv()
}