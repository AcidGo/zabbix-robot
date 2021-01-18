package state

import (
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type StateType uint

const (
    // unknown state
    Unknown StateType = iota
    // finish flow task sucessfully
    Sucess
    // some errors
    FlowBeginError
    CookError
    PrunError
    FiltError
    ClassifyError
    LimitError
    FlowEndError
)

type State struct {
    stateCnt        map[StateType]int
    ch              chan transf.Transfer
}

func NewState(ops *Options) (*State, error) {
    return &State{
        stateCnt:   make(map[StateType]int),
        ch:         make(chan transf.Transfer, ops.StateChanBufSize),
    }, nil
}

func (s *State) StateCh() (chan<- transf.Transfer) {
    return s.ch
}

func (s *State) Run() (error) {
    go func {
        for t := range s.ch {
            s.stateCnt[t.GetState()]++
        }
    }()

    return nil
}