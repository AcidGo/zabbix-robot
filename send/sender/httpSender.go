package send

import (
    "github.com/AcidGo/zabbix-robot/common"
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type HttpSend struct {
    Logger      *log.Logger
    client      *http.Client
    ch          chan transf.Transfer
    stateCh     chan<- transf.Transfer
}

func NewHttpSend(l *log.Logger, ops send_http.Options) (*HttpSend, error) {
    return &HttpSend{
        Logger:     l,
        client:     &http.Client{},
        ch:         make(chan transf.Transfer, ops.ChanBufSize),
    }, nil
}

func (h *HttpSend) InChan() (chan<- transf.Transfer) {
    return h.ch
}

func (h *HttpSend) BindState(ch chan<- transf.Transfer) {
    h.stateCh = ch
}

func (h *HttpSend) Run() {
    for t := range ch {
        switch t.GetSendMod() {
        case common.ThroughError:
            go h.throughError(t)
        case common.ThroughData:
            go h.throughData(t)
        case common.TrhoughRaw:
            go h.throughRaw(t)
        default:
            err = erros.New("not support the sendMod from transfer")
            t.SetError(err)
            h.Logger.Error(err)
            h.sendError(t)
        }
    }
}

func (h *HttpSend) sendError(t transf.Transfer) {
    t.SetState(common.SendError)
    go func () {
        h.stateCh <- t
    }()
}

func (h *HttpSend) throughData(t transf.Transfer) {
    
}