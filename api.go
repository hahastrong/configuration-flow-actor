package cfa

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

func CFARun(lg *log.Logger, flowName string, request string) ([]byte, error) {
	ctx := &Context{
		Logger: lg,
	}

	if err := ctx.Init(request); err != nil {
		// todo logger
		return nil, err
	}

	//testFlow := `{"start":{"id":"start","type":"start","next":"getHTTPRSP"},"getHTTPRSP":{"id":"getHTTPRSP","type":"task","next":"databuilder","task":{"id":"getHTTPRSP","task_type":"HttpRequest","method":"GET","path":"http://www.weather.com.cn/data/sk/101210101.html"}},"databuilder":{"id":"databuilder","type":"task","next":"end","task":{"id":"databuilder","task_type":"DataBuilder","method":"GET","response":{".":{"data":"__getHTTPRSP:RSP__","action":"expr"}}}}}`
	floeString, err := ioutil.ReadFile("./flow-json/test.json")
	if err != nil {
		return nil, err
	}

	var flow Flow
	err = json.Unmarshal(floeString, &flow)
	if err != nil {
		return nil, err
	}

	ch := make(chan error, 1)
	// asyc run flow
	go func() {
		err := flow.Run(ctx)
		ch <- err
	}()

	// add timeout
	select {
	case err := <- ch:
		if err != nil {
			// 补充 获取返回数据
			return nil, err
		}
	case <- time.After(time.Millisecond * 200000):
			// timeout

	}

	rsp, err := ctx.MarshalResponse()
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
