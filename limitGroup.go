package main

type LimitGroup interface {
    AddUnit(*LimitUnit) error
    MeanOneUnit()
}