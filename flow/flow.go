package flow

import (
    "github.com/AcidGo/zabbix-robot/message/msg"
    "github.com/AcidGo/zabbix-robot/slot"
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type FlowStep uint

const (
    Initial FlowStep = iota
    Error
    Cooked
    Pruned
    Filted
    Classifed
    Limited
)

type Flow struct {
    slotGroup       []slot.Slotor
    ch              chan transf.Transfer
    sendCh          chan<- transf.Transfer
}

func NewFlow(ops Options) (err error) {
    return &Flow{
        slotGroup: make([]slot.Slotor, 0),
        ch: make(chan<- transf.Transfer, ops.ChanBufSize),
    }, err
}

func (f *Flow) AppendSlot(s slot.Slotor) (idx int, err error) {
    f.slotGroup = append(f.slotGroup, s)
    idx = len(f.slotGroup) - 1
    return 
}

func (f *Flow)BindSend(c chan<- transf.Transfer) {
    f.sendCh = c
}

func (f *Flow) Infuse(m msg.Message) {
    t := m.ConvToTransfer()
    f.inCh <- t
}

func (f *Flow) InSlot() (c <-chan transf.Transfer) {
    return f.ch
}

func (f *Flow) Intercept() (c <-chan transf.Transfer) {
    c = f.outCh
    return 
}
