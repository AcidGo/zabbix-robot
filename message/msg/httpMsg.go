package msg

import (
    "io"
    "net/http"
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
    
    NewTransfZabbix()
}

