package transf

import (
    "errors"
    "regexp"
    "sync"

    "github.com/AcidGo/zabbix-robot/common"
    "github.com/AcidGo/zabbix-robot/lib"
    "github.com/AcidGo/zabbix-robot/log"
    "github.com/AcidGo/zabbix-robot/transf/zabbix"
    "github.com/go-playground/validator/v10"
)

type ZabbixAlert struct {
    Channel             string                  `json:"Channel" validate:"required"`
    Details             string                  `json:"Details" validate:"required"`
    EventID             string                  `json:"EventID" validate:"required"`
    EventItem           string                  `json:"EventItem" validate:"required"`
    EventTime           string                  `json:"EventTime"`
    HostIP              string                  `json:"HostIP" validate:"required"`
    HostName            string                  `json:"HostName" validate:"required"`
    Status              string                  `json:"Status" validate:"required"`
    Severity            string                  `json:"Severity" validate:"required"`
    Tags                map[string]string       `json:"Tags"`
}

type TransfZabbix struct {
    Transf
    sync.Mutex
    convFlag            string
    convReMap           map[string]*regexp.Regexp
    formattedData       *ZabbixAlert
    remoteDst           string
    tagFlag             string
    tagRe               *regexp.Regexp
}

func NewTransfZabbix(l *log.Logger, ops transf_zabbix.Options) (*TransfZabbix, error) {
    r, err := regexp.Compile(ops.TagReStr)
    if err != nil {
        return nil, err
    }

    return &TransfZabbix{
        Transf: Transf{
            Logger:         l,
            rawData:        nil,
            err:            nil,
            executedFlow:   make([]common.FlowStep, 0),
            sendMod:        common.SendUnknown,
            stateType:      common.StateUnknown,
        },
        convFlag:   ops.ConvFlag,
        convReMap:  make(map[string]*regexp.Regexp),
        tagFlag:    ops.TagFlag,
        tagRe:      r,
    }, nil
}

func (tz *TransfZabbix) FormatData() (error) {
    rd, ok := tz.rawData.(string)
    if !ok {
        return errors.New("the rawData from TransfZabbix is not string type")
    }

    reC, ok := tz.convReMap[tz.convFlag]
    if !ok {
        return errors.New("not found specified conv flag for formatting data")
    }

    m, err := utils.RegexpGenMap(rd, reC)
    if err != nil {
        return err
    }

    s, ok := m[tz.tagFlag].(string)
    if !ok {
        return errors.New("cannot get valid tag content with string type")
    }

    tagM, err := utils.ExtractTags(s, tz.tagRe)
    if err != nil {
        return err
    }

    var zAlert ZabbixAlert
    err = utils.MapToStruct(m, &zAlert)
    if err != nil {
        return err
    }

    zAlert.Tags = tagM

    validate := validator.New()
    err = validate.Struct(zAlert)
    if err != nil {
        return err
    }

    tz.formattedData = &zAlert

    return nil
}

func (tz *TransfZabbix) AddConvModel(flagVal, reStr string) (error) {
    if _, ok := tz.convReMap[flagVal]; ok {
        tz.Logger.Warnf("AddConvModel replace multi flag value: %s", flagVal)
    }
    c, err := regexp.Compile(reStr)
    if err != nil {
        return err
    }

    tz.convReMap[flagVal] = c

    return nil
}