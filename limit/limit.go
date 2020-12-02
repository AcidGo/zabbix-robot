package limit

import (
    "errors"
    "fmt"
    "reflect"
    "regexp"
    "sort"
    "strconv"
    "sync"
    "time"

    "gopkg.in/ini.v1"
    log "github.com/sirupsen/logrus"

    "github.com/AcidGo/zabbix-robot/utils"
)


const (
    LimitStateFree   = iota
    LimitStateBorn
    LimitStateWork
)

type LimitState int

type LimitRole struct {
    Severity                string          `ini:"severity,required"`
    Weight                  int             `ini:"weight,min=0,max=100"`
    LimitInterval           int             `ini:"limit_interval,min=1,required"`
    LimitThreshold          int             `ini:"limit_threshold,min=1,required"`
    InhibitInterval         int             `ini:"inhibit_interval,min=1,required"`
    InhibitThreshold        int             `ini:"inhibit_threshold,min=1,required"`
    Field                   string          `ini:"field,string,required"`
    Expression              string          `ini:"expression,string,required"`
    reCompile               *regexp.Regexp  `ini:"-"`
}

type Limiter interface {
    setRole(LimitRole) error
    startTicker()
    inhibitCreate(utils.DelayFunc, ...interface{}) error
    GetName() string
    GetWeight() int
    Match(map[string]interface{}) (bool, error)
    Increase() (LimitState, error)
}

type LimitUnit struct {
    sync.Mutex
    name                string
    weight              int
    interval            time.Duration
    threshold           int
    role                *LimitRole
    inhibitInterval     time.Duration
    inhibitThreshold    int
    inhibitDone         chan struct{}
    InhibitUnit         *InhibitUnit
    delayWorking        sync.Mutex
    count               int
}

func NewLimitUnit(name string) *LimitUnit {
    lUnit := &LimitUnit{
        name: name,
        count: 0,
    }
    return lUnit
}

func (lUnit *LimitUnit) setRole(lRole *LimitRole) error {
    var err error

    lUnit.weight = lRole.Weight
    lUnit.interval = time.Duration(lRole.LimitInterval)*time.Second
    lUnit.threshold = lRole.LimitThreshold
    lUnit.inhibitInterval = time.Duration(lRole.InhibitInterval)*time.Second
    lUnit.inhibitThreshold = lRole.InhibitThreshold

    lRole.reCompile, err = regexp.Compile(lRole.Expression)
    if err != nil {
        return err
    }
    lUnit.role = lRole

    go lUnit.startTicker()

    return nil
}

func (lUnit *LimitUnit) startTicker() {
    ticker := time.NewTicker(lUnit.interval)
    for {
        <- ticker.C
        lUnit.Lock()
        log.Debugf("the limit unit %s get ticker signal rotate", lUnit.GetName())
        lUnit.count = 0
        lUnit.Unlock()
    }
}

func (lUnit *LimitUnit) GetName() string {
    return lUnit.name
}

func (lUnit *LimitUnit) GetSeverity() string {
    return lUnit.role.Severity
}

func (lUnit *LimitUnit) Match(data map[string]interface{}) (bool, error) {
    var err error

    lUnit.Lock()
    defer lUnit.Unlock()

    for key, val := range data {
        if key == lUnit.role.Field {
            var s string
            switch val.(type) {
            case string:
                s = val.(string)
            case int:
                s = strconv.Itoa(val.(int))
            }
            if match := lUnit.role.reCompile.MatchString(s); match {
                return true, err
            }
        }
    }
    return false, err
}

func (lUnit *LimitUnit) Increase(callback utils.DelayFunc, args ...interface{}) (LimitState, error) {
    lUnit.Lock()
    defer lUnit.Unlock()

    if lUnit.InhibitUnit != nil && lUnit.InhibitUnit.Alive {
        lUnit.InhibitUnit.Increase()
        return LimitStateWork, nil
    }

    if lUnit.count < lUnit.threshold {
        lUnit.count++
        return LimitStateFree, nil
    }

    // start born an InhibitUnit for cutting off
    err := lUnit.inhibitCreate(callback, args...)
    return LimitStateBorn, err
}

func (lUnit *LimitUnit) inhibitCreate(callback utils.DelayFunc, args ...interface{}) error {
    lUnit.delayWorking.Lock()
    lUnit.inhibitDone = make(chan struct{})
    lUnit.InhibitUnit = NewInhibitUnit(lUnit, lUnit.inhibitDone, lUnit.inhibitInterval, lUnit.inhibitThreshold)

    go func() {
        defer lUnit.delayWorking.Unlock()
        select {
        case <- lUnit.inhibitDone:
            in := make([]reflect.Value, len(args))
            for i, arg := range args {
                in[i] = reflect.ValueOf(arg)
            }
            f := reflect.ValueOf(callback)
            values := f.Call(in)
            if len(values) > 1 {
                if _, ok := values[1].Interface().(error); ok {
                    log.Error("get an err from callback: ", values[1])
                } else {
                    log.Debug("get the response from callback: ", values[0])
                }
            }
        }
    }()

    lUnit.count = 0
    log.Debugf("created an inhibit from limit unit %s", lUnit.GetName())
    return nil
}

type LimitGroup struct {
    sync.Mutex
    members         []*LimitUnit
    nameSet         []string
}

func NewLimitGroup() *LimitGroup {
    return &LimitGroup{
        members:    make([]*LimitUnit, 0),
        nameSet:    make([]string, 0),
    }
}

func (lGroup *LimitGroup) Len() int {
    return len(lGroup.members)
}

func (lGroup *LimitGroup) Swap(i int, j int) {
    lGroup.members[i], lGroup.members[j] = lGroup.members[j], lGroup.members[i]
}

func (lGroup *LimitGroup) Less(i int, j int) bool {
    return lGroup.members[j].weight < lGroup.members[i].weight
}

func (lGroup *LimitGroup) AddUnit(section *ini.Section) error {
    var err error

    lGroup.Lock()
    defer lGroup.Unlock()

    name := section.Name()
    if _, ok := utils.Find(lGroup.nameSet, name); ok {
        return fmt.Errorf("the limit uint %s is conflicting", name)
    }

    lRole := new(LimitRole)
    err = section.MapTo(lRole)
    if err != nil {
        return err
    }

    lUnit := NewLimitUnit(name)
    lUnit.setRole(lRole)
    lGroup.members = append(lGroup.members, lUnit)
    sort.Sort(lGroup)
    lGroup.nameSet = append(lGroup.nameSet, name)

    return nil
}

func (lGroup *LimitGroup) GetSize() int {
    lGroup.Lock()
    defer lGroup.Unlock()

    return len(lGroup.members)
}

func (lGroup *LimitGroup) MatchOne(data map[string]interface{}) (*LimitUnit, error) {
    for _, lUnit := range lGroup.members {
        if ok, _ := lUnit.Match(data); ok {
            return lUnit, nil
        }
    }

    return nil, errors.New("not match one limit unit on the group")
}