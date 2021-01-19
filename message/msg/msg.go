package msg

import (
    "github.com/AcidGo/zabbix-robot/transf/transfer"
)

type Message interface {
    ConvToTransfer() transf.Transfer
}