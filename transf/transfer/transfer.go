package transf

import (
    "github.com/AcidGo/zabbix-robot/common"
    "github.com/AcidGo/zabbix-robot/log"
)

type Transfer interface {
    // executing data function
    SetRawData(interface{}) error
    GetRawData() interface{}
    FormatData() error
    GetFormattedData() interface{}
    // executing flow function
    TagFlow(common.FlowStep)
    ExecutedFlow() []common.FlowStep
    // executing error fucntion
    SetError(error)
    GetError() error
    // executing send function
    SetSendMod(common.SendMod)
    GetSendMod() common.SendMod
    GetSendDst() interface{}
    // executing state function
    SetState(common.StateType)
    GetState() common.StateType
    // executing meta function
    SetMeta(map[string]interface{})
    GetMeta() map[string]interface{}
}

type Transf struct {
    Logger              *log.Logger
    rawData             interface{}
    formattedData       interface{}
    executedFlow        []common.FlowStep
    err                 error
    sendMod             common.SendMod
    stateType           common.StateType
}

func (t *Transf) SetRawData(i interface{}) (error) {
    t.rawData = i
    return nil
}

func (t *Transf) GetRawData() (interface{}) {
    return t.rawData
}

func (t *Transf) GetFormattedData() (interface{}) {
    return t.formattedData
}

func (t *Transf) TagFlow(f common.FlowStep) {
    t.executedFlow = append(t.executedFlow, f)
}

func (t *Transf) ExecutedFlow() ([]common.FlowStep) {
    return t.executedFlow
}

func (t *Transf) SetError(e error) {
    t.err = e
}

func (t *Transf) GetError() (error) {
    return t.err
}

func (t *Transf) SetSendMod(s common.SendMod) {
    t.sendMod = s
}

func (t *Transf) GetSendMod() (common.SendMod) {
    return t.sendMod
}

func (t *Transf) SetState(s common.StateType) {
    t.stateType = s
}

func (t *Transf) GetState() (common.StateType) {
    return t.stateType
}