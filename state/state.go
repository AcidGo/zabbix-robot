package state

import (
    "fmt"
    "sync"
)

const (
    RequestSum          = "request_sum"
    ContentDealFailed   = "content_deal_failed"
    TagDealFailed       = "tag_deal_failed"
)

var (
    SState := newSvrState()
)

type SvrStater interface {
    Reset() error
    GetState() map[string]int
    IncreaseState(string) error
}

type SvrState struct {
    sync.RWMutex
    numRequestSum           int
    numContentDealFailed    int
    numTagDealFailed        int
}

func newSvrState() *SvrState {
    return &SvrState{}
}

func (ss *SvrState) Reset() error {
    ss.Lock()
    defer ss.Unlock()

    ss.numRequestSum = 0
    ss.numContentDealFailed = 0
    ss.numTagDealFailed = 0

    return nil
}

func (ss *SvrState) GetState() map[string]int {
    ss.RLock()
    defer ss.RUnlock()

    return map[string]int{
        RequestSum:         ss.numRequestSum,
        ContentDealFailed:  ss.numContentDealFailed,
        TagDealFailed:      ss.numTagDealFailed,
    }
}

func (ss *SvrState) IncreaseState(k string) error {
    var err error

    ss.Lock()
    defer ss.Unlock()

    switch k {
    case RequestSum:
        ss.numRequestSum += 1
    case ContentDealFailed:
        ss.numContentDealFailed += 1
    case TagDealFailed:
        ss.numTagDealFailed += 1
    default:
        err = fmt.Errorf("not mean supported state on server: %s", k)
    }

    return err
}