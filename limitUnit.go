package main

import (
    "fmt"
    "reflect"
    "regexp"
    "strconv"
    "sync"
    "time"

    log "github.com/sirupsen/logrus"
)

const (
    // Limit 没有到达限制阶段
    LimitFree   = iota
    // Limit 触发了抑制，开始创建 inhibit
    LimitBorn
    // Limit 目前存在一个 inhibit
    LimitWork
)

type DelayFunc func(remote string, header map[string][]string, data map[string]interface{}, delayData map[string]interface{}) error

type LimitRole struct {
    Weight              int     `ini:"weight,min=0,max=100"`
    LimitInterval       int     `ini:"limit_interval,min=1,required"`
    LimitThreshold      int     `ini:"limit_threshold,min=1,required"`
    InhibitInterval     int     `ini:"inhibit_interval,min=1,required"`
    InhibitThreshold    int     `ini:"inhibit_threshold,min=1,required"`
    Space               string  `ini:"space,string,required"`
    Field               string  `ini:"field,string,required"`
    Method              string  `ini:"method,string,required"`
    Description         string  `ini:"description,string,required"`

    reCompile           *regexp.Regexp  `ini:"-"`
}

type Limiter interface {
    GetName() string
    GetWeight() int
    IsMatch(map[string][]string, map[string]interface{}) (bool, error)
    setRole(LimitRole) error
    startTicker()
    inhibitCreate(DelayFunc, ...interface{}) error
    Increase() (int, error)
}

type LimitUnit struct {
    sync.Mutex
    name                string
    weight              int
    interval            time.Duration
    threshold           int
    role                LimitRole
    InhibitInterval     time.Duration
    InhibitThreshold    int
    InhibitFinish       chan bool
    InhibitUnit         *InhibitUnit
    count               int
}

func NewLimitUnit(name string, role *LimitRole) (*LimitUnit, error) {
    var err error

    limitUnit := &LimitUnit{name: name}
    err = limitUnit.setRole(role)
    if err != nil {
        return nil, err
    }

    go limitUnit.startTicker()

    return limitUnit, nil
}

func (limitUnit *LimitUnit) setRole(role *LimitRole) error {
    var err error

    limitUnit.weight = role.Weight
    limitUnit.interval = time.Duration(role.LimitInterval)*time.Second
    limitUnit.threshold = role.LimitThreshold
    limitUnit.InhibitInterval = time.Duration(role.InhibitInterval)*time.Second
    limitUnit.InhibitThreshold = role.InhibitThreshold

    if role.Space != "header" && role.Space != "content" {
        return fmt.Errorf("not support the space on limit role: %s", role.Space)
    }
    if role.Method != "regexp" && role.Method != "noregexp" {
        return fmt.Errorf("not support the method on limit role: %s", role.Method)
    }
    role.reCompile, err = regexp.Compile(role.Description)
    if err != nil {
        return err
    }
    limitUnit.role = role

    return nil
}

func (limitUnit *LimitUnit) startTicker() {
    ticker := time.NewTicker(limitUnit.interval)
    for {
        <- ticker.C
        log.WithFields(log.Fields{
            "name": limitUnit.name,
        }).Debug("get singal from ticker")
        limitUnit.Lock()
        oldCount := limitUnit.count
        limitUnit.count = 0
        limitUnit.Unlock()
        log.WithFields(log.Fields{
            "name": limitUnit.name,
        }).Debugf("reset the count %d to zero", oldCount)
    }
}

func (limitUnit *LimitUnit) GetName() string {
    return limitUnit.name
}

func (limitUnit *LimitUnit) GetWeight() int {
    limitUnit.Lock()
    defer limitUnit.Unlock()

    return limitUnit.weight
}

func (limitUnit *LimitUnit) IsMatch(header map[string][]string, content map[string]interface{}) (bool, error) {
    var err error

    limitUnit.Lock()
    defer limitUnit.Unlock()

    switch limitUnit.role.Field {
    case "header":
        for key, val := range header {
            if key == limitUnit.role.Field && len(val) > 0 {
                for _, subVal := range val {
                    match := limitUnit.role.reCompile.MatchString(subVal)
                    if match {
                        return true, err
                    }
                }
            }
        }
    case "content":
        for key, val := range content {
            if key == limitUnit.role.Field {
                var s string
                switch val.(type) {
                case string:
                    s = val.(string)
                case int:
                    s = strconv.Itoa(val.(int))
                }
                if len(s) > 0 {
                    match := limitUnit.role.reCompile.MatchString(s)
                    if match {
                        return true, err
                    }
                }
            }
        }
    }

    return false, err
}

func (limitUnit *LimitUnit) Increase() (int, error) {
    log.WithFields(log.Fields{
        "name": limitUnit.name,
    }).Debug("into LimitUnit.Increase ......")

    limitUnit.Lock()

    if limitUnit.InhibitUnit != nil && limitUnit.InhibitUnit.Alive {
        log.WithFields(log.Fields{
            "name": limitUnit.name,
            "count": limitUnit.count,
        }).Debug("limitUnit had been working")
        limitUnit.InhibitUnit.Increase()
        limitUnit.Unlock()
        return LimitWork, nil
    }

    if limitUnit.count < limitUnit.threshold {
        log.WithFields(log.Fields{
            "name": limitUnit.name,
            "count": limitUnit.count,
            "threshold": limitUnit.threshold,
        }).Info("count of limitUnit is less than threshold, pass and increase it")
        limitUnit.count++
        limitUnit.Unlock()
        return LimitFree, nil
    }

    log.WithFields(log.Fields{
        "name": limitUnit.name,
        "count": limitUnit.count,
    }).Info("count of limitUnit is more than threshold, and there is no inhibit, need to create it")
    return LimitBorn, nil
}

func (limitUnit *LimitUnit) inhibitCreate(callback DelayFunc, args ...interface{}) error {
    // need handle the lock and unlock it finally
    defer limitUnit.Unlock()

    limitUnit.InhibitUnit = NewInhibitUnit(limitUnit, limitUnit.InhibitInterval, limitUnit.InhibitThreshold)

    go func() {
        select {
        case <- limitUnit.InhibitUnit.Finish:
            log.Debug("receive inhibitUnit finish singal")
            in := make([]reflect.Value, len(args))
            for k, arg := range(args) {
                in[k] = reflect.ValueOf(arg)
            }
            f := reflect.ValueOf(callback)
            err := f.Call(in)
            if err != nil {
                log.WithFields(log.Fields{
                    "name": limitUnit.name,
                }).Error("get an err when call delay callback: ", err)
            }
        }
    }()

    limitUnit.count = 0
    return nil
}