package main

import (
    "errors"
    "reflect"
    "sync"
    "time"

    log "github.com/sirupsen/logrus"
)

const (
    // LU 没有达到限制阶段
    LUStatusNotLimit     = iota
    // LU 达到限制，需要创建 IU 阶段
    LUStatusCreateIU
    // LU 目前存在一个 IU 在做拦截
    LUStatusInIU
)

type DelayFunc func(remote string, header map[string][]string, data map[string]interface{}, delayData map[string]interface{}) error

type LimitUnit struct {
    Name                string
    Interval            time.Duration
    MaxCount            int
    Lock                sync.Mutex
    Count               int
    inhibitionCreating  sync.Mutex
    InhibitionInterval  time.Duration
    InhibitionThreshold int
    Inhibition          *InhibitionUnit
}

func NewLimitUnit(name string, interval time.Duration, maxCnt int, ihInterval time.Duration, inThreshold int) *LimitUnit {
    log.WithFields(log.Fields{
        "name": name,
        }).Debug("into NewLimitUnit for new a instance")

    limitUnit := &LimitUnit{
        Name:                   name,
        Interval:               interval,
        MaxCount:               maxCnt,
        InhibitionInterval:     ihInterval,
        InhibitionThreshold:    inThreshold,
    }

    go func() {
        ticker := time.NewTicker(interval)
        for {
            <- ticker.C
            log.WithFields(log.Fields{
                "name": limitUnit.Name,
            }).Debug("get signal from ticker")
            limitUnit.Lock.Lock()
            limitUnit.Count = 0
            limitUnit.Lock.Unlock()
            log.WithFields(log.Fields{
                "name": limitUnit.Name,
                }).Debug("reset the Count to zero")
        }
    }()

    return limitUnit
}

func (limitUnit *LimitUnit) LimitIncrease() int {
    log.WithFields(log.Fields{
        "name": limitUnit.Name,
        }).Debug("into LimitUnit.Increase")

    limitUnit.Lock.Lock()
    defer limitUnit.Lock.Unlock()

    defer func() {
        log.WithFields(log.Fields{
            "name": limitUnit.Name,
            }).Debug("leave LimitUnit.Increase")
    }()

    // 1. 判断是否具有 IU 在截断
    limitUnit.inhibitionCreating.Lock()
    if limitUnit.Inhibition != nil && limitUnit.Inhibition.Alive {
        log.WithFields(log.Fields{
            "name": limitUnit.Name,
            "count": limitUnit.Count,
            }).Debug("limitUnit had been in IU")
        limitUnit.InhibitionIncrease()
        limitUnit.inhibitionCreating.Unlock()
        return LUStatusInIU
    } else {
        limitUnit.inhibitionCreating.Unlock()
    }

    // 2. 判断 LU 是否到达抑制条件
    if limitUnit.Count < limitUnit.MaxCount {
        log.WithFields(log.Fields{
            "name": limitUnit.Name,
            "count": limitUnit.Count,
            "MaxCount": limitUnit.MaxCount,
            }).Info("limitUnit count is less than MaxCount, pass and increase it")
        limitUnit.Count++
        return LUStatusNotLimit
    }

    // 3. 没有 IU 运行但是 LU 已经达到了抑制条件，需要通知创建 IU
    log.WithFields(log.Fields{
        "name": limitUnit.Name,
        "count": limitUnit.Count,
        }).Info("limitUnit count is more than MaxCount, and there is not InhibitionUnit, need to create IU")
    // 这里主要是因为创建的动作是分开的，所以需要额外
    // time.Sleep(10 * time.Millisecond)
    limitUnit.inhibitionCreating.Lock()
    return LUStatusCreateIU
}

func (limitUnit *LimitUnit) InhibitionCreate(callback DelayFunc, args ...interface{}) error {
    log.WithFields(log.Fields{
        "name": limitUnit.Name,
        }).Debug("into LimitUnit.InhibitionCreate")
    defer func() {
        log.WithFields(log.Fields{
            "name": limitUnit.Name,
            }).Debug("leave LimitUnit.InhibitionCreate")
    }()

    limitUnit.Lock.Lock()
    defer limitUnit.Lock.Unlock()

    limitUnit.Inhibition = NewInhibitionUnit(limitUnit, limitUnit.InhibitionInterval, limitUnit.InhibitionThreshold)

    go func() {
        select {
        case <- limitUnit.Inhibition.Done:
            log.Debug("in InhibitionCreate sub go, receive Done signal")

            in := make([]reflect.Value, len(args))
            for k, arg := range args {
                in[k] = reflect.ValueOf(arg)
            }
            f := reflect.ValueOf(callback)
            _ = f.Call(in)
            limitUnit.Inhibition.Alive = false
        }
    }()

    // 符合逻辑理解，需要保证下一次触发新建 IU 时能够给出 LU 的缓冲
    limitUnit.Count = 0
    limitUnit.inhibitionCreating.Unlock()
    return nil
}

func (limitUnit *LimitUnit) InhibitionIncrease() error {
    inhibition := limitUnit.Inhibition
    if inhibition == nil || !inhibition.Alive {
        log.WithFields(log.Fields{
            "name": limitUnit.Name,
            }).Error("not found InhibitionUnit in the LU")
        return errors.New("not found InhibitionUnit in the LU")
    }
    inhibition.Increase()
    return nil
}


type InhibitionUnit struct {
    Parent          *LimitUnit
    Lock            sync.Mutex
    Interval        time.Duration
    Threshold       int
    Count           int
    Done            chan bool
    Finish          chan bool
    Alive           bool
}

func NewInhibitionUnit(parent *LimitUnit, interval time.Duration, threshold int) *InhibitionUnit {
    inhibition := &InhibitionUnit{
        Parent:         parent,
        Interval:       interval,
        Threshold:      threshold,
        Count:          1,
        Done:           make(chan bool),
        Alive:          true,
    }

    log.Debug("start to go inhibition ticker")
    go func() {
        ticker := time.NewTicker(inhibition.Interval)
        for {
            select {
            case <- ticker.C:
                log.WithFields(log.Fields{
                    "name": inhibition.Parent.Name,
                    }).Debug("into inhibitionUnit.ticker lock")
                inhibition.Lock.Lock()
                if !inhibition.Alive {
                    log.WithFields(log.Fields{
                        "name": inhibition.Parent.Name,
                        }).Debug("in ticker, but found the inhibition.Alive is false, no deal with it")
                    inhibition.Lock.Unlock()
                    return
                }
                log.WithFields(log.Fields{
                    "state": "start",
                    }).Debug("in hibitionUnit.ticker, send to inhibition.Done")
                // inhibition.Done <- true
                close(inhibition.Done)
                log.WithFields(log.Fields{
                    "state": "end",
                    }).Debug("in hibitionUnit.ticker, send to inhibition.Done")
                log.WithFields(log.Fields{
                    "name": inhibition.Parent.Name,
                    }).Debug("because of count more than intervalout, inhibition done")
                inhibition.Lock.Unlock()
                log.WithFields(log.Fields{
                    "name": inhibition.Parent.Name,
                    }).Debug("leave hibitionUnit.ticker lock")
                inhibition.Alive = false
                return
            case <- inhibition.Done:
                log.WithFields(log.Fields{
                    "name": inhibition.Parent.Name,
                    }).Debug("in hibitionUnit.ticker, get the inhibition.Done, exit")
                return
            }
        }
    }()

    return inhibition
}

func (inhibition *InhibitionUnit) Increase() {
    log.WithFields(log.Fields{
        "name": inhibition.Parent.Name,
        }).Debug("into InhibitionUnit.Increase")
    inhibition.Lock.Lock()
    defer inhibition.Lock.Unlock()
    defer func() {
        log.WithFields(log.Fields{
            "name": inhibition.Parent.Name,
            }).Debug("leave InhibitionUnit.Increase")
    }()

    if !inhibition.Alive {
        log.WithFields(log.Fields{
            "name": inhibition.Parent.Name,
            }).Debug("in increase, but found the inhibition.Alive is true, no deal with it")
        return
    }

    inhibition.Count++
    log.WithFields(log.Fields{
        "name": inhibition.Parent.Name,
        "count": inhibition.Count,
        }).Debug("InhibitionUnit increased")

    if inhibition.Count >= inhibition.Threshold {
        log.WithFields(log.Fields{
            "state": "start",
            }).Debug("in InhibitionUnit.Increase, send to inhibition.Done")
        // inhibition.Done <- true
        close(inhibition.Done)
        inhibition.Alive = false
        log.WithFields(log.Fields{
            "state": "end",
            }).Debug("in InhibitionUnit.Increase, send to inhibition.Done")
        log.WithFields(log.Fields{
            "name": inhibition.Parent.Name,
            }).Debug("because of count more than threshold, inhibition done")
    }

}