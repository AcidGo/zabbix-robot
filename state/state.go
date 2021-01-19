package state

import (
    "github.com/AcidGo/zabbix-robot/common"
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type State struct {
    stateCnt        map[common.StateType]int
    ch              chan transf.Transfer
}

func NewState(ops *Options) (*State, error) {
    return &State{
        stateCnt:   make(map[common.StateType]int),
        ch:         make(chan transf.Transfer, ops.StateChanBufSize),
    }, nil
}

func (s *State) StateCh() (chan<- transf.Transfer) {
    return s.ch
}

func (s *State) Run() (error) {
    go func() {
        for t := range s.ch {
            s.stateCnt[t.GetState()]++
        }
    }()

    return nil
}