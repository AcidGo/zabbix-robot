package send

import (
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type Sender interface {
    InChan() chan<- transf.Transfer
    BindState(chan<- transf.Transfer)
    Run() error
    throughError(transf.Transfer) error
    throughData(transf.Transfer) error
    throughRaw(transf.Transfer) error
    delayData(transf.Transfer, <-chan struct{}) error
}