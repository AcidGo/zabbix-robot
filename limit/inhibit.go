package limit

import (
    "sync"
    "time"

    log "github.com/sirupsen/logrus"
)

type InhibitUnit struct {
    sync.Mutex
    Parent              *LimitUnit
    Done                chan struct{}
    Alive               bool
    interval            time.Duration
    threshold           int
    Count               int
}

func NewInhibitUnit(parent *LimitUnit, done chan struct{}, interval time.Duration, threshold int) *InhibitUnit {
    iUnit := &InhibitUnit{
        Parent:             parent,
        Done:               done,
        Alive:              true,
        Count:              1,
        interval:           interval,
        threshold:          threshold,
    }

    go func() {
        ticker := time.NewTicker(iUnit.interval)
        for {
            select {
            case <- ticker.C:
                iUnit.Lock()
                if !iUnit.Alive {
                    iUnit.Unlock()
                    return 
                }
                log.Debugf("the inhibit that born from limit unit %s, get ticker signal rotate", iUnit.Parent.GetName())
                close(iUnit.Done)
                iUnit.Alive = false
                iUnit.Unlock()
                return 
            case <- iUnit.Done:
                return 
            }
        }
    }()

    return iUnit
}

func (iUnit *InhibitUnit) Increase() {
    iUnit.Lock()
    defer iUnit.Unlock()

    if !iUnit.Alive {
        return 
    }

    iUnit.Count++
    if iUnit.Count >= iUnit.threshold {
        close(iUnit.Done)
        iUnit.Alive = false
    }
}
