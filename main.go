package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "errors"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
    "reflect"
    "regexp"
    "time"

    lfshook "github.com/rifflock/lfshook"
    log "github.com/sirupsen/logrus"
    "github.com/lestrrat/go-file-rotatelogs"
    "gopkg.in/ini.v1"
)

const (
    ProgramName                 = "zabbix-robot"
    ProgramVersion              = "0.0.3"
    ProgramExecName             = "zabbix_robot"
    ProgramInitLogLevel         = log.DebugLevel

    SeverityNotClassifiedName   = "Not classified"
    SeverityInformationName     = "Information"
    SeverityWarningName         = "Warning"
    SeverityHighName            = "High"
    SeverityDisasterName        = "Disaster"

    ConfigDefaultName           = "zabbix_robot.conf"

    ConfigSectionLog            = "log"
    ConfigLogKeyLogFile         = "logFile"
    ConfigLogKeyLogLevel        = "logLevel"

    ConfigSectionSeverity       = "severity"
    ConfigSeverityKeySamplingIntervalSecond   = "samplingIntervalSecond"
    ConfigSeverityKeySamplingThresholdNum     = "samplingThresholdNum"
    ConfigSeverityKeyInhibitionIntervalSecond = "inhibitionIntervalSecond"
    ConfigSeverityKeyInhibitionThresholdNum   = "inhibitionThresholdNum"

    ConfigSectionRelay          = "relay"
    ConfigRelayKeyRelayInhibitionEventType  = "relayInhibitionEventType"
    ConfigRelayKeyRelayInhibitionEventID    = "relayInhibitionEventID"
    ConfigRelayKeyRelayInhibitionStatus     = "relayInhibitionStatus"
    ConfigRelayKeyRelayInhibitionHostName   = "relayInhibitionHostName"
    ConfigRelayKeyRelayInhibitionHostIP     = "relayInhibitionHostIP"
    ConfigRelayKeyRelayInhibitionEventItem  = "relayInhibitionEventItem"
    ConfigRelayKeyRelayInhibitionChannel    = "relayInhibitionChannel"

    ConfigSectionMain                   = "main"
    ConfigMainKeyURL                    = "url"
    ConfigMainKeyListen                 = "listen"
    ConfigMainKeyFingerPrintKey         = "header_fingerprint_key"
    ConfigMainKeyFingerPrintValue       = "header_fingerprint_value"
    ConfigmainKeyIgnoreStatus           = "ignore_status"
    ConfigMainMsgRegexpOkHeader         = "msg_regexp_ok_header"
    ConfigMainMsgRegexpOkCompile        = "msg_regexp_ok_compile"
    ConfigMainMsgRegexpProblemHeader    = "msg_regexp_problem_header"
    ConfigMainMsgRegexpProblemCompile   = "msg_regexp_problem_compile"

    ConfigSectionTagconv                = "tagconv"
)

var (
    ConfigLogFile       string
    ConfigLogLevel      uint
    ConfigSeverityMap   map[string]map[string]interface{}

    RelayInhibitionEventType    string
    RelayInhibitionEventID      string
    RelayInhibitionStatus       string
    RelayInhibitionHostName     string
    RelayInhibitionHostIP       string
    RelayInhibitionEventItem    string
    RelayInhibitionChannel      string

    MainListen string
    MainURL    string
    MainFingerPrintKey string
    MainFingerPrintValue string

    MainIgnoreStatus map[string][]string

    MsgRegexpOkHeader string
    MsgRegexpOkCompile string
    MsgRegexpProblemHeader string
    MsgRegexpProblemCompile string
    RegexpOkCompile *regexp.Regexp
    RegexpProblemCompile *regexp.Regexp
    RegexpTagconvStrings map[string]string
    RegexpTagconvCompiles map[string]*regexp.Regexp
)

var (
    LogLevel        uint
    ConfigPath      string
)

var unitMap map[string]*LimitUnit
// var ESStrMapping = map[string]string {
//     `"`: `\"`,
//     `\`: `\\`,
//     "\n": "\\n",
//     "\b": "\\b",
//     "\f": "\\f",
//     "\t": "\\t",
//     "\r": "\\r",
// }

// func CookStrRegexp(rawStr string, escapeMapping map[string]string) string {
//     var resString string
//     resTmp := make([]string, 10)
//     for _, r := range rawStr {
//         s := string(r)
//         v, ok := ESStrMapping[s]
//         if ok {
//             resTmp = append(resTmp, v)
//             continue
//         }
//         resTmp = append(resTmp, s)
//     }
//     resString = strings.Join(resTmp, "")
//     return resString
// }

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


func regexpDeal(bodyStatus string, body io.ReadCloser) io.ReadCloser {
    log.Debug("start regexpDeal")
    contents, err := ioutil.ReadAll(body)
    if err != nil {
        log.WithFields(log.Fields{
            "error": err,
            }).Error("in regexpDeal, get error when readall the body")
        return nil
    }

    s := string(contents)
    var data []byte
    // result := make(map[string]string)
    result := make(map[string]interface{})

    if bodyStatus == MsgRegexpOkHeader && RegexpOkCompile != nil {
        match := RegexpOkCompile.FindStringSubmatch(s)
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
    } else if bodyStatus == MsgRegexpProblemHeader && RegexpProblemCompile != nil {
        match := RegexpProblemCompile.FindStringSubmatch(s)
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
            "bodyStatus": bodyStatus,
            }).Error("not match the bodyStatus")
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
                            tagConvRes[matchArr[1]] = matchArr[2]
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
        newReadCloser := ioutil.NopCloser(bytes.NewReader(contents))
        return newReadCloser
    }

    newReadCloser := ioutil.NopCloser(bytes.NewReader(data))
    log.Debug("finish deal with request msg with regexp successfully")
    return newReadCloser
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

func flagUsage() {
    usageMsg := fmt.Sprintf(`
%s version: %s
Author: AcidGo
Usage: %s [-l level] [-f config]
Options:
`, ProgramName, ProgramVersion, ProgramExecName)

    fmt.Fprintf(os.Stderr, usageMsg)
    flag.PrintDefaults()
}

func initFlag() error {
    flag.UintVar(&LogLevel, "l", 4, "the `level` of the log")
    flag.StringVar(&ConfigPath, "f", ConfigDefaultName, "set `config` for the program")

    flag.Usage = flagUsage
    flag.Parse()

    if ConfigPath == "" {
        log.Error("the configure file is nil, please input it")
        err := errors.New("the configure is not defined")
        return err
    }

    if !FileExists(ConfigPath) {
        log.WithFields(log.Fields{
            "path": ConfigPath,
            }).Error("the path of configure defined is not exists")
        err := errors.New("the configure file is not exists")
        return err
    }

    return nil
}

func initUnitMap(cfg *ini.File) error {
    unitMap = make(map[string]*LimitUnit)
    for _, child := range cfg.ChildSections(ConfigSectionSeverity) {
        severityName := strings.Join(strings.Split(child.Name(), ".")[1:], "")
        samplingIntervalSecond, err     := child.Key(ConfigSeverityKeySamplingIntervalSecond).Int()
        samplingThresholdNum, err       := child.Key(ConfigSeverityKeySamplingThresholdNum).Int()
        inhibitionIntervalSecond, err   := child.Key(ConfigSeverityKeyInhibitionIntervalSecond).Int()
        inhibitionThresholdNum, err     := child.Key(ConfigSeverityKeyInhibitionThresholdNum).Int()
        if err != nil {
            return err
        }

        if (samplingIntervalSecond == 0 || samplingIntervalSecond == 0 || 
            inhibitionIntervalSecond == 0 || inhibitionThresholdNum == 0) {
            log.WithFields(log.Fields{
                "severityName": severityName,
                "samplingIntervalSecond": samplingIntervalSecond,
                "samplingThresholdNum": samplingThresholdNum,
                "inhibitionIntervalSecond": inhibitionIntervalSecond,
                "inhibitionThresholdNum": inhibitionThresholdNum,
                }).Error("the params must be int type and positive number")
            return errors.New("the params must be int type and positive number")
        }

        unitMap[severityName] = NewLimitUnit(
            severityName, 
            time.Duration(samplingIntervalSecond)*time.Second, 
            samplingThresholdNum, 
            time.Duration(inhibitionIntervalSecond)*time.Second, 
            inhibitionThresholdNum,
        )

        log.Debug("generate new unit in unitMap: ", unitMap[severityName])
    }

    // init default for unknow section
    severityName := "_default"
    s := cfg.Section(ConfigSectionSeverity)
    samplingIntervalSecond, err     := s.Key(ConfigSeverityKeySamplingIntervalSecond).Int()
    samplingThresholdNum, err       := s.Key(ConfigSeverityKeySamplingThresholdNum).Int()
    inhibitionIntervalSecond, err   := s.Key(ConfigSeverityKeyInhibitionIntervalSecond).Int()
    inhibitionThresholdNum, err     := s.Key(ConfigSeverityKeyInhibitionThresholdNum).Int()

    if err != nil {
        return err
    }

    if (samplingIntervalSecond == 0 || samplingIntervalSecond == 0 || 
        inhibitionIntervalSecond == 0 || inhibitionThresholdNum == 0) {
        log.WithFields(log.Fields{
            "severityName": severityName,
            "samplingIntervalSecond": samplingIntervalSecond,
            "samplingThresholdNum": samplingThresholdNum,
            "inhibitionIntervalSecond": inhibitionIntervalSecond,
            "inhibitionThresholdNum": inhibitionThresholdNum,
            }).Error("the params must be int type and positive number")
        return errors.New("the params must be int type and positive number")
    }

    unitMap[severityName] = NewLimitUnit(
        severityName, 
        time.Duration(samplingIntervalSecond)*time.Second, 
        samplingThresholdNum, 
        time.Duration(inhibitionIntervalSecond)*time.Second, 
        inhibitionThresholdNum,
    )

    log.Debug("generate new _default unit in unitMap: ", unitMap[severityName])

    return nil
}

func initLog(level log.Level) {
    customFormatter := new(log.TextFormatter)
    customFormatter.TimestampFormat = "2006-01-02 15:04:05.000000000"
    customFormatter.FullTimestamp = true
    customFormatter.DisableTimestamp = false
    log.SetFormatter(customFormatter)
    log.SetOutput(os.Stdout)
    log.SetLevel(level)
}

func initRegexp() error {
    var err error
    if MsgRegexpOkCompile == "" {
        RegexpOkCompile = nil
    } else {
        RegexpOkCompile, err = regexp.Compile(MsgRegexpOkCompile)
        if err != nil {
            log.WithFields(log.Fields{
                "error": err,
                "MsgRegexpOkCompile": MsgRegexpOkCompile,
                }).Error("in initRegexp, get an error when complie MsgRegexpOkCompile")
        } else {
            log.WithFields(log.Fields{
                "MsgRegexpOkCompile": MsgRegexpOkCompile,
                }).Debug("in initRegexp, born RegexpOkCompile successfully")
        }
    }

    if MsgRegexpProblemCompile == "" {
        RegexpProblemCompile = nil
    } else {
        RegexpProblemCompile, err = regexp.Compile(MsgRegexpProblemCompile)
        if err != nil {
            log.WithFields(log.Fields{
                "error": err,
                "MsgRegexpProblemCompile": MsgRegexpProblemCompile,
                }).Error("in initRegexp, get an error when complie MsgRegexpProblemCompile")
        } else {
            log.WithFields(log.Fields{
                "MsgRegexpProblemCompile": MsgRegexpProblemCompile,
                }).Debug("in initRegexp, born RegexpProblemCompile successfully")
        }
    }

    RegexpTagconvCompiles = make(map[string]*regexp.Regexp)
    for tagKey, tagCompileString := range RegexpTagconvStrings {
        RegexpTagconvCompiles[tagKey], err = regexp.Compile(tagCompileString)
        if err != nil {
            log.WithFields(log.Fields{
                "error": err,
                "tagKey": tagKey,
                "tagCompile": RegexpTagconvCompiles[tagKey],
                }).Error("in initRegexp, get an error when compile RegexpTagconvCompiles")
        } else {
            log.WithFields(log.Fields{
                "tagKey": tagKey,
                }).Debug("in initRegexp, born tagKey on RegexpTagconvCompiles successfully")
        }
    }

    return err
}

func configParse(path string) (*ini.File, error) {
    log.Debug("into configParse, start load config file: ", path)
    cfg, err := ini.Load(path)
    if err != nil {
        return nil, err
    }

    s, err := cfg.GetSection(ConfigSectionLog)
    if err != nil {
        log.WithFields(log.Fields{
            "section": ConfigSectionLog,
            }).Error("the section not exists")
        return nil, err
    }

    ConfigLogFile = s.Key(ConfigLogKeyLogFile).String()
    ConfigLogLevel, err = s.Key(ConfigLogKeyLogLevel).Uint()
    if err != nil {
        log.WithFields(log.Fields{
            "ConfigLogLevel": ConfigLogLevel,
            }).Error(fmt.Sprintf("the config value of key %s must be uint", ConfigLogKeyLogLevel))
        return nil, err
    }

    RelayInhibitionEventType    = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionEventType).String()
    RelayInhibitionEventID      = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionEventID).String()
    RelayInhibitionStatus       = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionStatus).String()
    RelayInhibitionHostName     = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionHostName).String()
    RelayInhibitionHostIP       = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionHostIP).String()
    RelayInhibitionEventItem    = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionEventItem).String()
    RelayInhibitionChannel      = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionChannel).String()

    MainListen = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyListen).String()
    MainURL    = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyURL).String()

    MainFingerPrintKey = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyFingerPrintKey).String()
    MainFingerPrintValue = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyFingerPrintValue).String()

    err = json.Unmarshal([]byte(cfg.Section(ConfigSectionMain).Key(ConfigmainKeyIgnoreStatus).String()), &MainIgnoreStatus)

    MsgRegexpOkHeader = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpOkHeader).String()
    MsgRegexpOkCompile = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpOkCompile).String()
    MsgRegexpProblemHeader = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpProblemHeader).String()
    MsgRegexpProblemCompile = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpProblemCompile).String()

    RegexpTagconvStrings = make(map[string]string)
    for _, tagKey := range cfg.Section(ConfigSectionTagconv).Keys() {
        RegexpTagconvStrings[tagKey.Name()] = tagKey.String()
    }

    err = configLog()
    if err != nil {
        return nil, err
    }

    return cfg, nil
}

func configLog() error {
    log.SetLevel(log.Level(ConfigLogLevel))
    if ConfigLogFile == "" {
        log.SetOutput(os.Stdout)
    } else {
        log.Info("start to change log mode")
        writer, err := rotatelogs.New(
            ConfigLogFile + ".%Y%m%d%H%M",
            rotatelogs.WithLinkName(ConfigLogFile),
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
                log.InfoLevel:     writer,
                log.WarnLevel:     writer,
                log.ErrorLevel:     writer,
                log.FatalLevel:     writer,
                log.PanicLevel:     writer},
            &log.TextFormatter{
                TimestampFormat: "2006-01-02 15:04:05.000000000",
                FullTimestamp:   true,
                DisableTimestamp: false,
                },
            )
        log.AddHook(lfHook)
    }

    log.Info("log config finished")
    return nil
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
    var s string
    s, r.Body, _ = bodyToString(r.Body)
    log.WithFields(log.Fields{
        "method": r.Method,
        "body": s,
        }).Debug("in router alertHandler, get a new request")

    if r.Method != http.MethodPost {
        log.WithFields(log.Fields{
            "method": r.Method,
            }).Warn("the request method is not expected")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "please POST method")
        return
    }

    rSeverity := r.Header.Get("Severity")
    log.WithFields(log.Fields{
        "Severity": rSeverity,
        }).Debug("get the Severity from the requset header")

    rStatus := r.Header.Get("Status")
    log.WithFields(log.Fields{
        "Status": rStatus,
        }).Debug("get the Status from request header")

    if _, ok := MainIgnoreStatus[rStatus]; ok {
        for _, ignoreStatu := range MainIgnoreStatus[rStatus] {
            if ignoreStatu == rSeverity {
                log.Info("the Severity and Status should be ignored, ignore it")
                return 
            }
        }
    }

    unit, ok := unitMap[rSeverity]
    if !ok {
        log.Warn("the header of request has no hit Severity field, so use the default")
        rSeverity = "_default"
        unit, ok = unitMap[rSeverity]
        if !ok {
            log.Error("cannot use the defualt for the severity")
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprint(w, "cannot use the defualt for the severity")
            return
        }
    }

    rRemote := r.Header.Get("Remote")
    log.WithFields(log.Fields{
        "Remote": rRemote,
        }).Debug("get the Remote from request header")

    if rRemote == "" {
        log.Warn("the header of request has not Remote field")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, "not found the Remote")
        return 
    }

    newHeader := make(map[string][]string)
    for k, v := range r.Header {
        if k == "Severity" || k == "Remote" || k == "Status" {
            continue
        }
        newHeader[k] = v
    }
    newHeader[MainFingerPrintKey] = []string{MainFingerPrintValue}
    log.Debug("newHeader is done")

    unitStatus := unit.LimitIncrease()

    switch unitStatus {
        case LUStatusNotLimit:
            log.Debug("unit is available")
            var newBody io.ReadCloser
            if rStatus != "" {
                newBody = regexpDeal(rStatus, r.Body)
            } else {
                newBody = r.Body
            }
            err := relayPass(rRemote, newHeader, newBody)
            if err != nil {
                log.WithFields(log.Fields{
                    "error": err,
                    }).Error("get error when call relayPass")
                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprint(w, "get error when call relayPass")
                return
            }
            return

        case LUStatusInIU:
            log.Debug("the unit had been limited, check for cutoff")
            log.Debug("still in inhibition")
            return

        case LUStatusCreateIU:
            log.Debug("start to create a IU")
            dataCurrent := map[string]interface{}{
                "EventType": map[string]string{
                    "EventType": RelayInhibitionEventType,
                },
                "EventID": RelayInhibitionEventID,
                "Severity": rSeverity,
                "Status": RelayInhibitionStatus,
                "HostName": RelayInhibitionHostName,
                "HostIP": RelayInhibitionHostIP,
                "EventTime": time.Now().Format("2006-01-02 15:04:05"),
                "EventItem": RelayInhibitionEventItem,
                "Channel": RelayInhibitionChannel,
            }
            dataDelay := map[string]interface{}{
                "Details": func() string {
                        log.Debug("the unit.Inhibition.Count: ", unit.Inhibition.Count)
                        return fmt.Sprintf("触发告警抑制，此次对[%s]级别告警抑制了[%d]条", unit.Name, unit.Inhibition.Count)
                    },
            }
            _ = unit.InhibitionCreate(
                relayDelay,
                rRemote,
                newHeader,
                dataCurrent,
                dataDelay,
                )
            log.Debug("finish born a inhibition")
            return
        default:
            log.WithFields(log.Fields{
                "name": unit.Name,
                "status": unitStatus,
                }).Error("unknow status")
            return
    }

    log.Warn("maybe some error")
}

func init() {
    initLog(ProgramInitLogLevel)

    err := initFlag()
    if err != nil {
        log.Fatal("get error when init flag: ", err)
    }

    cfg, err := configParse(ConfigPath)
    if err != nil {
        log.Fatal("get error when parse config: ", err)
    }
    log.Debug("finish to load config from ", ConfigPath)

    _ = initRegexp()

    err = initUnitMap(cfg)
    if err != nil {
        log.Fatal("get error when init unitmap: ", err)
    }
}

func main() {
    log.Printf("zabbix-robot written by AcidGo, the version is %s", ProgramVersion)
    log.Println("the zabbix-robot is runnig ......")
    http.HandleFunc(MainURL, alertHandler)
    http.ListenAndServe(MainListen, nil)
}