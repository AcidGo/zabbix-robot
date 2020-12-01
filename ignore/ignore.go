package ignore

type IgnoreRole struct {
    Key         string      `ini:"key"`
    Val         string      `ini:"val"`
}

type IgnoreUnit struct {
    roles       map[string]string
}

func NewIgnoreUnit() *IgnoreUnit {
    return &IgnoreUnit{
        roles: make(map[string]string),
    }
}

func (ignore *IgnoreUnit) AddRole(iRole IgnoreRole) error {
    ignore.roles[iRole.Key] = iRole.Val
    return nil
}

func (ignore *IgnoreUnit) IsIgnore(data map[string]interface{}) (bool, error) {
    for k, v := range data {
        if iRoleVal, ok := ignore.roles[k]; ok {
            if iRoleVal == v {
                return true, nil
            }
        }
    }

    return false, nil
}