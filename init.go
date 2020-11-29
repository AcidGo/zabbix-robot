package main

package (
    "flag"
    "os"
    "time"

    log "github.com/sirupsen/logrus"
    "gopkg.in/ini.v1"
    "github.com/lestrrat/go-file-rotatelogs"
)


func initLog() error {
    customFormatter := new(log.TextFormatter)
    customFormatter.TimestampFormat = "2006-01-02 15:04:05.000000000"
    customFormatter.FullTimestamp = true
    customFormatter.DisableTimestamp = false
    log.SetFormatter(customFormatter)
    log.SetOutput(os.Stdout)
    log.SetLevel(programInitLogLevel)
    return nil
}

func initFlag() error {
    flag.UintVar(&logLevel, "l", 4, "the `level` of the log")
    flag.StringVar(&programConfigPath, "f", ConfigDefaultName, "set `config` for the program")

    flag.Usage = flagUsage
    flag.Parse()

    if programConfigPath == "" {
        log.Error("the configure file is nil, please input it")
        err := errors.New("the configure is not defined")
        return err
    }

    if !FileExists(programConfigPath) {
        log.WithFields(log.Fields{
            "path": programConfigPath,
            }).Error("the path of configure defined is not exists")
        err := errors.New("the configure file is not exists")
        return err
    }

    return nil
}

func initConfigParse() errro {
    var err error

    log.Debug("parsing the conf for loading confif file")
    programIniFile, err = ini.Load(programConfigPath)
    if err != nil {
        return err
    }

    s, err := programIniFile.GetSection(ConfigSectionLog)
    if err != nil {
        log.WithFields(log.Fields{
            "section": ConfigSectionLog,
        }).Error("the section is not exist")
        return err
    }
    logFilePath = s.Key(ConfigLogKeyLogFile).String()
    logLevel, err = s.Key(ConfigLogKeyLogLevel).Uint()
    if err != nil {
        log.Errorf("the config value of key %s must be uint", ConfigLogKeyLogLevel)
        return err
    }
    err = configLog()
    if err != nil {
        return err
    }

    RelayInhibitionEventType    = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionEventType).String()
    RelayInhibitionEventID      = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionEventID).String()
    RelayInhibitionStatus       = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionStatus).String()
    RelayInhibitionHostName     = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionHostName).String()
    RelayInhibitionHostIP       = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionHostIP).String()
    RelayInhibitionEventItem    = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionEventItem).String()
    RelayInhibitionChannel      = cfg.Section(ConfigSectionRelay).Key(ConfigRelayKeyRelayInhibitionChannel).String()

    httpListenPort = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyListen).String()
    HttpURL    = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyURL).String()

    HttpFingerPrintKey = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyFingerPrintKey).String()
    HttpFingerPrintValue = cfg.Section(ConfigSectionMain).Key(ConfigMainKeyFingerPrintValue).String()

    err = json.Unmarshal([]byte(cfg.Section(ConfigSectionMain).Key(ConfigmainKeyIgnoreStatus).String()), &IgnoreStatus)
    if err != nil {
        log.Errorf("unmarshal %s from %s is failed", ConfigmainKeyIgnoreStatus, ConfigSectionMain)
        return nil
    }

    RegexpOkHeader = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpOkHeader).String()
    RegexpOkCompileString = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpOkCompile).String()
    RegexpProblemHeader = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpProblemHeader).String()
    RegexpProblemCompileString = cfg.Section(ConfigSectionMain).Key(ConfigMainMsgRegexpProblemCompile).String()

    TagconvRegexpStrings = make(map[string]string)
    for _, tagKey := range cfg.Section(ConfigSectionTagconv).Keys() {
        TagconvRegexpStrings[tagKey.Name()] = tagKey.String()
    }

    ReportDBType = cfg.Section(ConfigSectionReport).Key(ConfigReportKeyDBType).String()
    ReportDBDsn = cfg.Section(ConfigSectionReport).Key(ConfigReportKeyDBDsn).String()
    ReportTableName = cfg.Section(ConfigSectionReport).Key(ConfigReportKeyTable).String()

    return nil
}

func initRegexp() error {
    var err error

    if RegexpOkCompileString == "" {
        RegexpOkCompile = nil
    } else {
        RegexpOkCompile, err = regexp.Compile(RegexpOkCompileString)
        if err != nil {
            log.WithFields(log.Fields{
                "error": err,
                "RegexpOkCompileString": RegexpOkCompileString,
                }).Error("in initRegexp, get an error when complie RegexpOkCompileString")
        } else {
            log.WithFields(log.Fields{
                "RegexpOkCompileString": RegexpOkCompileString,
                }).Debug("in initRegexp, born RegexpOkCompile successfully")
        }
    }

    if RegexpProblemCompileString == "" {
        RegexpProblemCompile = nil
    } else {
        RegexpProblemCompile, err = regexp.Compile(RegexpProblemCompileString)
        if err != nil {
            log.WithFields(log.Fields{
                "error": err,
                "RegexpProblemCompileString": RegexpProblemCompileString,
                }).Error("in initRegexp, get an error when complie RegexpProblemCompileString")
        } else {
            log.WithFields(log.Fields{
                "RegexpProblemCompileString": RegexpProblemCompileString,
                }).Debug("in initRegexp, born RegexpProblemCompile successfully")
        }
    }

    TagconvRegexpCompiles = make(map[string]*regexp.Regexp)
    for tagKey, tagCompileString := range TagconvRegexpStrings {
        TagconvRegexpCompiles[tagKey], err = regexp.Compile(tagCompileString)
        if err != nil {
            log.WithFields(log.Fields{
                "error": err,
                "tagKey": tagKey,
                "tagCompile": TagconvRegexpCompiles[tagKey],
                }).Error("in initRegexp, get an error when compile TagconvRegexpCompiles")
        } else {
            log.WithFields(log.Fields{
                "tagKey": tagKey,
                }).Debug("in initRegexp, born tagKey on TagconvRegexpCompiles successfully")
        }
    }

    return err
}

func initLimitUnitGroup() errror {
    var err error
    limitGroup = &LimitGroup{}

    for _, subSction := range programIniFile.ChildSections(ConfigSectionRole) {
        err = limitGroup.AddUnit(subSction)
        if err != nil {
            return err
        }
    }
    return nil
}

func initReport() error {
    var err error

    ReportDB, err = sql.Open(ReportDBType, ReportDBDsn)
    if err != nil {
        return err
    }
    err = ReportDB.Ping()
    if err != nil {
        return err
    }
    return nil
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
`, programName, programVersion, programAuthor, programGitCommitHash, programBuildTime, programGoVersion, programName)

    fmt.Fprintf(os.Stderr, usageMsg)
    flag.PrintDefaults()
}

func configLog() error {
    log.SetLevel(log.Level(logLevel))
    if logFilePath == "" {
        log.SetOutput(os.Stdout)
    } else {
        log.Info("starting to change log mode ......")
        writer, err := rotatelogs.New(
            logFilePath + ".%Y%m%d%H%M",
            rotatelogs.WithLinkName(logFilePath),
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

        lfHook := lfshook.NewHock(
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