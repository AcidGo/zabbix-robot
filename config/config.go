package config

import (
    "database/sql"
    "regexp"

    "github.com/AcidGo/zabbix-robot/ignore"
    "github.com/AcidGo/zabbix-robot/limit"

    "gopkg.in/ini.v1"
    log "github.com/sirupsen/logrus"
)

const(
    // inner
    InnerRemoteAddrField = "Remote"
    InnerHeaderSeverityField = "Severity"

    // config info
    ConfigDefaultName = "zabbix_robot.conf"

    ConfigSectionMain = "main"
    ConfigMainKeyURL                    = "url"
    ConfigMainKeyListen                 = "listen"

    ConfigSectionLog = "log"
    ConfigLogKeyLogFile                             = "log_file"
    ConfigLogKeyLogLevel                            = "log_level"

    ConfigSectionFormat = "format"
    ConfigFormatKeyRegexpHeaderField      = "msg_regexp_header_field"
    ConfigFormatKeyRegexpOkHeader         = "msg_regexp_ok_header"
    ConfigFormatKeyRegexpOkCompile        = "msg_regexp_ok_compile"
    ConfigFormatKeyRegexpProblemHeader    = "msg_regexp_problem_header"
    ConfigFormatKeyRegexpProblemCompile   = "msg_regexp_problem_compile"

    ConfigSectionTagconv = "tagconv"

    ConfigSectionIgnore = "ignore"
    ConfigIgnoreKeySetting = "ignore_setting"

    ConfigSectionRole = "role"
    ConfigRoleKeyWeight                 = "weight"
    ConfigRoleKeyLimitInterval          = "limit_interval"
    ConfigRoleKeyLimitThreshold         = "limit_threshold"
    ConfigRoleKeyInhibitInterval        = "inhibit_interval"
    ConfigRoleKeyInhibitThreshold       = "inhibit_threshold"
    ConfigRoleKeyField                  = "field"
    ConfigRoleKeyExpression             = "expression"

    ConfigSectionReport = "report"
    ConfigReportKeyDBType               = "driver"
    ConfigReportKeyDBDsn                = "dsn"
    ConfigReportKeyTable                = "table"

    ConfigSectionRelay = "relay"
    ConfigRelayKeyRelayInhibitionEventType  = "relay_inhibit_EventType"
    ConfigRelayKeyRelayInhibitionEventID    = "relay_inhibit_EventID"
    ConfigRelayKeyRelayInhibitionStatus     = "relay_inhibit_Status"
    ConfigRelayKeyRelayInhibitionHostName   = "relay_inhibit_HostName"
    ConfigRelayKeyRelayInhibitionHostIP     = "relay_inhibit_HostIP"
    ConfigRelayKeyRelayInhibitionEventItem  = "relay_inhibit_EventItem"
    ConfigRelayKeyRelayInhibitionChannel    = "relay_inhibit_Channel"
)

var (
    // app runtime parameters
    AppConfigPath                       string
    AppInitLogLevel                     log.Level = log.DebugLevel
    AppIniFile                          *ini.File
    AppName                             string
    AppAuthor                           string
    AppVersion                          string
    AppGitCommitHash                    string
    AppBuildTime                        string
    AppGoVersion                        string

    // logger parameters
    LogLevel                            uint
    LogFilePath                         string

    // ignore parameters
    IgnoreUnit                          *ignore.IgnoreUnit

    // format parameters
    FormatRegexpHeaderField             string
    FormatRegexpOkHeader                string
    FormatRegexpOkCompileString         string
    FormatRegexpProblemHeader           string
    FormatRegexpProblemCompileString    string
    FormatRegexpOkCompile               *regexp.Regexp
    FormatRegexpProblemCompile          *regexp.Regexp

    // limit group parameters
    LimitGroup                          *limit.LimitGroup

    // tag conv parameters
    TagconvCompiles                     map[string]*regexp.Regexp

    // relay parameters
    RelayInhibitionEventType    string
    RelayInhibitionEventID      string
    RelayInhibitionStatus       string
    RelayInhibitionHostName     string
    RelayInhibitionHostIP       string
    RelayInhibitionEventItem    string
    RelayInhibitionChannel      string
    RelayIgnoreHeaderFields     []string = []string{"Severity", "Remote", "Status"}

    // report parameters
    ReportDB                *sql.DB
    ReportDBType            string
    ReportDBDsn             string
    ReportTableName         string

    // http server config
    HttpURL                 string = "/alter"
    HttpListenPort          string
    HttpHeaderFieldRemote   string = "Remote"
    HttpFingerPrintKey      string
    HttpFingerPrintValue    string
)