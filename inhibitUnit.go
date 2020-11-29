package main

import (
    "sync"
    "time"

    log "github.com/sirupsen/logrus"
)

type Inhibiter interface {
    Increase()
}

type InhibitUnit struct {
    sync.Mutex
    Parent          *LimitUnit
    Finish          chan bool
    Alive           bool
    Interval        time.Duration
    Threshold       int
    Count           int
}

func NewInhibitUnit(parent *LimitUnit, interval time.Duration, threshold int) *InhibitUnit {
    inhibitUnit := &InhibitUnit{
        Parent:     parent,
        Interval:   interval,
        Threshold:  threshold,
        Finish:     make(chan bool),
        Alive:      true,
        Count:      1,
    }

    log.Debug("generating inhibitUnit ticker ......")
    go func() {
        ticker := time.NewTicker(inhibitUnit.Interval)
        for {
            select {
            case <- ticker.C:
                inhibitUnit.Lock()
                if !inhibitUnit.Alive {
                    inhibitUnit.Unlock()
                    return
                }
                close(inhibitUnit.Finish)
                inhibitUnit.Alive = false
                inhibitUnit.Unlock()
                return 
            case <- inhibitUnit.Finish:
                return 
            }
        }
    }()

    return inhibitUnit
}

func (inhibitUnit *InhibitUnit) Increase() {
    inhibitUnit.Lock()
    defer inhibitUnit.Lock()

    if !inhibitUnit.Alive {
        return
    }

    inhibitUnit.Count++
    log.WithFields(log.Fields{
        "name": inhibitUnit.Parent.GetName(),
        "count": inhibitUnit.Count,
    }).Debug("inhibitUnit increased")

    if inhibitUnit.Count >= inhibitUnit.Threshold {
        log.Debug("in InhibitUnit.Increase, close inhibitUnit.Finish ......")
        close(inhibitUnit.Finish)
        inhibitUnit.Alive = false
    }
}