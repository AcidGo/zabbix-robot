package common

type FlowStep uint

const (
    Initial FlowStep = iota
    Error
    Cooked
    Pruned
    Filted
    Classifed
    Limited
)

type StateType uint

const (
    // unknown state
    StateUnknown StateType = iota
    // finish flow task sucessfully
    Success
    // some errors
    FlowBeginError
    SendError
    CookError
    PrunError
    FiltError
    ClassifyError
    LimitError
    FlowEndError
)

type SendMod uint

const (
    SendUnknown = iota
    ThroughError
    ThroughData
    TrhoughRaw
    DelayData
)
