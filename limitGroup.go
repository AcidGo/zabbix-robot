package main

import (
    "fmt"
    "sort"
    "sync"

    "gopkg.in/ini.v1"
)

type LimitGroup struct {
    sync.Mutex
    size            int
    group           []*LimitUnit
    container       []string
}

func NewLimitGroup() *LimitGroup {
    return &LimitGroup{
        size:       0,
        group:      make([]*LimitUnit)
    }
}

func (limitGroup *LimitGroup) Len() int {
    return len(group)
}

func (limitGroup *LimitGroup) Swap(i, j int) {
    limitGroup.group[i], limitGroup.group[j] = limitGroup.group[j], limitGroup.group[i]
}

func (limitGroup *LimitGroup) Less(i, j int) bool {
    return limitGroup.group[j].weight < limitGroup.group[i].weight
}

func (limitGroup *LimitGroup) AddUnit(section *ini.Section) error {
    var err error

    limitGroup.Lock()
    defer limitGroup.Unlock()

    name := section.Name()
    if idx, ok := Find(limitGroup.container, name); ok {
        return fmt.Errorf("the limitUint %s is conflicting")
    }

    weight, err := section.Key(ConfigRoleKeyWeight).Int()
    if err != nil {
        return err
    }

    limitRole := new(LimitRole)
    err = section.MapTo(limitRole)
    if err != nil {
        return err
    }

    limitUint := NewLimitUnit(name, limitRole)
    limitGroup.group = append(limitGroup.group, limitUint)
    sort.Sort(limitGroup)
    limitGroup.size++

    return nil
}

func (limitGroup *LimitGroup) GetSize() int {
    limitGroup.Lock()
    defer limitGroup.Unlock()

    return limitGroup.size
}

func (limitGroup *LimitGroup) MatchOne(header map[string][]string, content map[string]interface{}) (*LimitUnit, error) {
    for _, limitUint := range limitGroup.group {
        if ok, _ := limitUint.IsMatch(header, content); ok {
            return limitUint, nil
        }
    }

    return nil, fmt.Error("not match one limit unit")
}

