package main

import (
    "fmt"
    "sync"

    "gopkg.in/ini.v1"
)

type LimitGroup struct {
    sync.Mutex
    size            int
    group           map[string]*LimitUnit
}

func (limitGroup *LimitGroup) AddUnit(section *ini.Section) error {
    var err error

    limitGroup.Lock()
    defer limitGroup.Unlock()

    name := section.Name()

    if _, ok := limitGroup[name]; ok {
        return fmt.Errorf("the limitUint %s is conflicting")
    }

    limitRole := new(LimitRole)
    err = section.MapTo(limitRole)
    if err != nil {
        return err
    }

    limitGroup[name] = NewLimitUnit(name, limitRole)
    limitGroup.size++

    return nil
}

func (limitGroup *LimitGroup) GetSize() int {
    limitGroup.Lock()
    defer limitGroup.Unlock()

    return limitGroup.size
}

func (limitGroup *LimitGroup) MatchOne(header map[string][]string, content map[string]interface{}) bool {
    
}