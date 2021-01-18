package slot

import (
    "github.com/AcidGo/zabbix-robot/log"
    "github.com/AcidGo/zabbix-robot/state"
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

func (s *Slot) BindState(ch chan<- transfer.Transfer) (error) {
    s.stateCh = ch
    return nil
}

func (s *Slot) BindFlow(inCh <-chan transf.Transfer, outCh chan<- transf.Transfer) (error) {
    go func {
        for t := range inCh {
            err := t.GetError()
            switch t.GetState() {
            case state.Unknown:
                if err != nil {
                    s.logger.Errorf("slot get an error from transfer with unknown state, discard the flow task: %v", err)
                    continue
                }
                outCh <- t
            case state.Success:
                s.logger.Debug("slot get a transfer with success state, finish the flow task")
                continue
            default:
                s.logger.Error("slot get a transfer with error state, discard it and push it to state center")
                s.stateCh <- t
                continue
            }
        }
    }()

    return nil
}