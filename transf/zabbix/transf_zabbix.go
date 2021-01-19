package transf_zabbix

import (
    "errors"
    "regexp"

    "github.com/AcidGo/zabbix-robot/lib"
    "github.com/AcidGo/zabbix-robot/log"
    "github.com/AcidGo/zabbix-robot/transf/transfer"
    "github.com/go-playground/validator/v10"
)

type ZabbixAlert struct {
    Channel             string                  `json:"Channel" validate:"required"`
    Details             string                  `json:"Details" validate:"required"`
    EventID             string                  `json:"EventID" validate:"required"`
    EventItem           string                  `jsn:"EventItem" validate:"required"`
    EventTime           string                  `json:"EventTime"`
    HostIP              string                  `json:"HostIP" validate:"required"`
    HostName            string                  `json:"HostName" validate:"required"`
    Status              string                  `json:"Status" validate:"required"`
    Severity            string                  `json:"Severity" validate:"required"`
    Tags                map[string]string       `json:"-"`
}

type TransfZabbix struct {
    transf.Transf
    sync.Mutex
    convFlag            string
    convReMap           map[string]*regexp.Compile
    formattedData       ZabbixAlert
    remoteDst           string
    tagFlag             string
}

func NewTransfZabbix(l *log.Logger, ops Options) (*TransfZabbix, error) {
    return &TransfZabbix{
        Logger:     l,
        tagFlag:    ops.TagFlag,
    }, nil
}

func (tz *TransfZabbix) FormatData() (err error) {
    rd, ok := tz.rawData.(string)
    if !ok {
        return errors.New("the rawData from TransfZabbix is not string type")
    }

    reC, ok := tz.convReMap[convHeaderFlag]
    if !ok {
        return errors.New("not found specified conv flag for formatting data")
    }

    m, err := utils.RegexpGenMap(rd, reC)
    if err != nil {
        return err
    }


}

func (tz *TransfZabbix) AddConvModel(flagVal, reStr string) (err error) {
    if _, ok := tz.convReMap[flagVal]; ok {
        Logger.Warnf("AddConvModel replace multi flag value: %s", flagVal)
    }
    c, err := regexp.Compile(reStr)
    if err != nil {
        return err
    }

    tz.convReMap[flagVal] = c
}