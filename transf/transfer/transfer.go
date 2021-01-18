package transf

import (
    "github.com/AcidGo/zabbix-robot/flow"
    "github.com/AcidGo/zabbix-robot/send/sender"
    "github.com/AcidGo/zabbix-robot/state"
)

type Transfer interface {
    // executing data function
    SetRawData(interface{}) error
    GetRawData() interface{}
    FormatData() error
    GetFormattedData() interface{}
    // executing flow function
    TagFlow(flow.FlowStep)
    ExecutedFlow() []flow.FlowStep
    // executing error fucntion
    SetError(error)
    GetError() error
    // executing send function
    SetSendMod(send.SendMod)
    GetSendMod() send.SendMod
    // executing state function
    SetState(state.StateType)
    GetState() state.StateType
}

type Transf struct {
    rawData             interface{}
    formattedData       interface{}
    executedFlow        []flow.FlowStep
    err                 error
    sendMod             send.SendMod
    stateType           state.StateType
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

func (t *Transf) TagFlow(f flow.FlowStep) {
    t.executedFlow = append(t.executedFlow, f)
}

func (t *Transf) ExecutedFlow() ([]flow.FlowStep) {
    return t.executedFlow
}

func (t *Transf) SetError(e error) {
    t.err = error
}

func (t *Transf) GetError() (error) {
    return t.err
}

func (t *Transf) SetSendMod(s send.SendMod) {
    t.sendMod = s
}

func (t *Transf) GetState() (send.SendMod) {
    return t.sendMod
}

func (t *Transf) SetState(s state.StateType) {
    t.stateType = s
}

func (t *Transf) GetState() (state.StateType) {
    return t.stateType
}