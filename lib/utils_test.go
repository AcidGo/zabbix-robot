package utils

import (
    "regexp"
    "reflect"
    "testing"

    "github.com/AcidGo/zabbix-robot/transf/transfer"
    "github.com/google/go-cmp/cmp"
)

func TestRegexpGenMap(t *testing.T) {
    tests := []struct {
        content     string
        compile     *regexp.Regexp
        res         map[string]interface{}
    } {
        {
            `prefix;;;Channel:::res1;;;Details::: res 2;;;suffix`,
            regexp.MustCompile(`Channel:::(?P<Channel>.*?);;;Details:::(?P<Details>.*?);;;`),
            map[string]interface{}{"Channel": "res1", "Details": " res 2"},
        },
    }

    for _, test := range tests {
        res, err := RegexpGenMap(test.content, test.compile)
        if err != nil {
            t.Errorf("RegexpGenMap(%v, %v): expected: %v, got err: %v", test.content, test.compile, test.res, err)
        }
        if cmp.Diff(res, test.res) != "" {
            t.Errorf("RegexpGenMap(%v, %v): expected: %v, got: %v", test.content, test.compile, test.res, res)
        }
    }
}

func TestMapToStruct(t *testing.T) {
    tests := []struct {
        m   map[string]interface{}
        v   interface{}
        res *transf.ZabbixAlert
    } {
        {
            map[string]interface{} {
                "Channel": "val-Channel",
                "Details": "val-Details",
                "EventID": "val-EventID",
                "EventItem": "val-EventItem",
                "EventTime": "val-EventTime",
                "HostIP": "val-HostIP",
                "HostName": "val-HostName",
                "Status": "val-Status",
                "Severity": "val-Severity",
                "Tags": map[string]string {"t1": "r1", "t2": "r2"},
            },
            transf.ZabbixAlert{},
            &transf.ZabbixAlert{
                Channel: "val-Channel",
                Details: "val-Details",
                EventID: "val-EventID",
                EventItem: "val-EventItem",
                EventTime: "val-EventTime",
                HostIP: "val-HostIP",
                HostName: "val-HostName",
                Status: "val-Status",
                Severity: "val-Severity",
                Tags: map[string]string {"t1": "r1", "t2": "r2"},
            },
        },
    }

    for _, test := range tests {
        err := MapToStruct(test.m, &test.v)
        if err != nil {
            t.Errorf("MapToStruct(%v, %v): got err: %v", test.m, test.v, err)
        }

        if reflect.DeepEqual(test.v, test.res) {
            t.Errorf("MapToStruct(%v, %v): expected: %v, got: %v", test.m, test.v, test.res, test.v)
        }
    }
}