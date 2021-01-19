package work

import (
    "github.com/AcidGo/zabbix-robot/common"
    "github.com/AcidGo/zabbix-robot/log"
    "github.com/AcidGo/zabbix-robot/work/cooker"
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type Cook struct {
    Work
}

func NewCook(l *log.Logger, ops work_cook.Options) (*Cook, error) {
    return &Cook {
        Work: Work {
            Logger:     l,
            readCh:     make(chan transf.Transfer, ops.ReadChanBufSize),
            writeCh:    make(chan transf.Transfer, ops.WriteChanBufSize),
        },
    }, nil
}

func (c *Cook) Run() {
    for t := range c.readCh {
        err := t.FormatData()
        if err != nil {
            c.Logger.Errorf("Cook transfer for formatting data: %v", err)
            t.SetError(err)
            t.SetState(common.CookError)
            t.SetSendMod(common.TrhoughRaw)

            go func () {
                c.sendCh <- t
            } ()
        }

        t.TagFlow(common.Cooked)
        go func () {
            c.writeCh <- t
        } ()
    }
}