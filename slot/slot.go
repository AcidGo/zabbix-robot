package slot

import (
    "github.com/AcidGo/zabbix-robot/common"
    "github.com/AcidGo/zabbix-robot/log"
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type Slotor interface {
    BindFlow(<-chan transf.Transfer, chan<- transf.Transfer) error
    BindState(chan<- transf.Transfer) error
}

type Slot struct {
    Logger          *log.Logger
    inSlotCh        <-chan transf.Transfer
    outSlotCh       chan<- transf.Transfer
    stateCh         chan<- transf.Transfer
}

func NewSlot(l *log.Logger) (*Slot, error) {
    return &Slot{
        Logger: l,
    }, nil
}

func (s *Slot) BindState(ch chan<- transf.Transfer) (error) {
    s.stateCh = ch
    return nil
}

func (s *Slot) BindFlow(inCh <-chan transf.Transfer, outCh chan<- transf.Transfer) (error) {
    go func() {
        for t := range inCh {
            err := t.GetError()
            switch t.GetState() {
            case common.Unknown:
                if err != nil {
                    s.Logger.Errorf("slot get an error from transfer with unknown state, discard the flow task: %v", err)
                    continue
                }
                outCh <- t
            case common.Success:
                s.Logger.Debug("slot get a transfer with success state, finish the flow task")
                continue
            default:
                s.Logger.Error("slot get a transfer with error state, discard it and push it to state center")
                s.stateCh <- t
                continue
            }
        }
    }()

    return nil
}