package msg

import (
    "io"
    "net/http"

    "github.com/AcidGo/zabbix-robot/transf/zabbix"
)

type HttpMsg struct {
    data            io.ReadCloser
    header          http.Header
}

func NewHttpMsg(h *http.Response) (*HttpMsg, error) {
    return &HttpMsg{
        data:       h.Body,
        header:     h.Header,
    }, nil
}

func (hm *HttpMsg) ConvToTransfer(l *log.Logger) (transf.Transfer) {
    tz, _ := NewTransfZabbix(l)
    tz.SetRawData(hm.data)
    tz.SetMeta(hm.header)
    return tz
}