package ignore

import (
    "github.com/AcidGo/zabbix-robot/utils"
)

type IgnoreRole struct {
    Key         string
    Val         map[string][]string
}

type IgnoreUnit struct {
    roles       map[string]map[string][]string
}

func NewIgnoreUnit() *IgnoreUnit {
    return &IgnoreUnit{
        roles: make(map[string]map[string][]string),
    }
}

func (ignore *IgnoreUnit) AddRole(iRole IgnoreRole) error {
    ignore.roles[iRole.Key] = iRole.Val
    return nil
}

func (ignore *IgnoreUnit) IsIgnore(data map[string]interface{}) (bool, string, error) {
    for roleK, roleV := range ignore.roles {
        meanNum := 0
        for iKey, iVal := range roleV {
            if item, ok := data[iKey]; ok {
                if itemStr, ok := item.(string); ok {
                    if _, ok := utils.Find(iVal, itemStr); ok {
                        meanNum += 1
                    }
                }
            }
        }
        if meanNum == len(roleV) {
            return true, roleK, nil
        }
    }
    return false, "", nil
}