package main

import (
    "database/sql"

    "gopkg.in/ini.v1"
    log "github.com/sirupsen/logrus"
)

const (
    ConfigSectionLog            = "log"
    ConfigLogKeyLogFile         = "logFile"
    ConfigLogKeyLogLevel        = "logLevel"

    ConfigSectionRelay = "relay"
    ConfigRelayKeyRelayInhibitionEventType  = "relayInhibitionEventType"
    ConfigRelayKeyRelayInhibitionEventID    = "relayInhibitionEventID"
    ConfigRelayKeyRelayInhibitionStatus     = "relayInhibitionStatus"
    ConfigRelayKeyRelayInhibitionHostName   = "relayInhibitionHostName"
    ConfigRelayKeyRelayInhibitionHostIP     = "relayInhibitionHostIP"
    ConfigRelayKeyRelayInhibitionEventItem  = "relayInhibitionEventItem"
    ConfigRelayKeyRelayInhibitionChannel    = "relayInhibitionChannel"

    ConfigSectionMain = "main"
    ConfigMainKeyURL                    = "url"
    ConfigMainKeyListen                 = "listen"
    ConfigMainKeyFingerPrintKey         = "header_fingerprint_key"
    ConfigMainKeyFingerPrintValue       = "header_fingerprint_value"
    ConfigmainKeyIgnoreStatus           = "ignore_status"
    COnfigMainMsgRegexpHeaderField      = "msg_regexp_header_field"
    ConfigMainMsgRegexpOkHeader         = "msg_regexp_ok_header"
    ConfigMainMsgRegexpOkCompile        = "msg_regexp_ok_compile"
    ConfigMainMsgRegexpProblemHeader    = "msg_regexp_problem_header"
    ConfigMainMsgRegexpProblemCompile   = "msg_regexp_problem_compile"

    ConfigSectionTagconv = "tagconv"

    ConfigSectionReport = "report"
    ConfigReportKeyDBType               = "driver"
    ConfigReportKeyDBDsn                = "dsn"
    ConfigReportKeyTable                = "table"

    ConfigSectionRole = "role"
    ConfigRoleKeyWeight                 = "weight"
    ConfigRoleKeyLimitInterval          = "limit_interval"
    ConfigRoleKeyLimitThreshold         = "limit_threshold"
    ConfigRoleKeyInhibitInterval        = "inhibit_interval"
    ConfigRoleKeyInhibitThreshold       = "inhibit_threshold"
    ConfigRoleKeySpace                  = "space"
    ConfigRoleKeyField                  = "field"
    ConfigRoleKeyMethod                 = "method"
    ConfigRoleKeyDescription            = "descript"

    InnerRemoteAddrField                = "Remote"
)

var (
    // program initialization parameters
    programName             string
    programAuthor           string
    programVersion          string
    programGitCommitHash    string
    programBuildTime        string
    programGoVersion        string
    programInitLogLevel     log.Level = log.DebugLevel
    programIniFile          *ini.File
    programConfigPath       string

    // http server config
    HttpURL                 string = "/alter"
    HttpListenPort          string
    HttpHeaderFieldRemote   stirng = "Remote"
    HttpFingerPrintKey      string
    HttpFingerPrintValue    string

    // ignore parameters
    IgnoreSelector map[string][]string

    // logger parameters
    logLevel                uint
    logFilePath             string

    // regexp parameters
    RegexpHeaderField string
    RegexpOkHeader string
    RegexpOkCompileString string
    RegexpProblemHeader string
    RegexpProblemCompileString string
    RegexpOkCompile *regexp.Regexp
    RegexpProblemCompile *regexp.Regexp

    // tagconv parameters
    TagconvRegexpStrings map[string]string
    TagconvRegexpCompiles map[string]*regexp.Regexp

    // report parameters
    ReportDB                *sql.DB
    ReportDBType            string
    ReportDBDsn             string
    ReportTableName         string

    // relay parameters
    RelayInhibitionEventType    string
    RelayInhibitionEventID      string
    RelayInhibitionStatus       string
    RelayInhibitionHostName     string
    RelayInhibitionHostIP       string
    RelayInhibitionEventItem    string
    RelayInhibitionChannel      string
    RelayIgnoreHeaderFields     []string = []string{"Severity", "Remote", "Status"}

    // role parameters
    limitGroup                  *LimitGroup
)