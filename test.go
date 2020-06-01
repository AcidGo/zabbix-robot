package main

import (
    "log"
    "regexp"
)

func main() {
    str := `{"EventType":"EventType:系统","EventID":"24611-64514562","Severity":"Warning","Status":"OK","HostName":"WIN_ComStar资金后台客户端应用服务器_197.1.1.129","HostIP":"197.1.1.129","EventTime":"2020.04.13 17:21:13","EventItem":"Used% Memory","Details":"内存使用率>90%，当前值[89.77 %]","Channel":"Zabbix"}`

    com := `\{\s?"EventType":"(?P<EventType>.*?)",\s?"EventID":"(?P<EventID>.*?)",\s?"Severity":"(?P<Severity>.*?)",\s?"Status":"(?P<Status>.*?)",\s?"HostName":"(?P<HostName>.*?)",\s?"HostIP":"(?P<HostIP>.*?)",\s?"EventTime":"(?P<EventTime>.*?)",\s?"EventItem":"(?P<EventItem>.*?)",\s?"Details":"(?P<Details>.*?)",\s?"Channel":"(?P<Channel>.*?)"\s?\}`

    re, err := regexp.Compile(com)
    if err != nil {
        log.Println(err)
    }
    match := re.FindStringSubmatch(str)
    groupNames := re.SubexpNames()

    // log.Printf("%v, %v, %d, %d", match, groupNames, len(match), len(groupNames))

    result := make(map[string]string)
    if len(match) != len(groupNames) {
        log.Println("fuck")
    } else {
        for i, name := range groupNames {
            if i != 0 && name != "" {
                result[name] = match[i]
            }
        }
    }

    log.Println(result)
}