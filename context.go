package cfa

import (
	"errors"
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"net/http"
	"os"
	"strings"
)

type Context struct {
	Logger *log.Logger
	request *fastjson.Value
	response *fastjson.Value
	taskResult map[string]*TaskResult
	localVar *fastjson.Value
	A *fastjson.Arena

	// flow process result
	Exit bool
	CmdDir string
}

type TaskResult struct {
	request *fastjson.Value
	response *fastjson.Value
	headers http.Header // 这个字段还需要后续在修改下
	responseByte []byte // binary, text
}

func (c *Context) Init(request string) error {
	c.request, _ = fastjson.Parse(request)

	c.A = new(fastjson.Arena)

	c.response = c.A.NewObject()

	c.localVar = c.A.NewObject()

	c.taskResult = make(map[string]*TaskResult)

	c.Exit = true

	c.CmdDir, _ = os.Getwd()

	return nil
}

func (c *Context) SetCmdDir(dir string) {
	c.CmdDir = dir
}

func (c *Context) NewTaskResult(id string) {
	c.taskResult[id] = &TaskResult{
		response: c.A.NewObject(),
		request: c.A.NewObject(),
		headers: make(http.Header),
	}
}

func (c *Context) SetRsp(id string, rsp []byte) error {
	rspValue, err := fastjson.Parse(string(rsp))
	if err != nil {
		return err
	}
	if _, ok := c.taskResult[id]; !ok {
		return errors.New("failed to get taskResult value")
	}
	c.taskResult[id].response = rspValue
	return nil
}

func (c *Context) SetHeaders(id string, headers http.Header) error {
	if _, ok := c.taskResult[id]; !ok {
		return errors.New("failed to get taskResult value")
	}
	c.taskResult[id].headers = headers
	return nil
}

func (c *Context) MarshalResponse() ([]byte, error) {
	var rspByte []byte
	rspByte = c.response.MarshalTo(rspByte)

	//var rsp = make(map[string]interface{})
	//err := json.Unmarshal(rspByte, &rsp)
	//if err != nil {
	//	return nil, err
	//}

	return rspByte, nil
}

func IsTaskTsp(source string) bool {
	return strings.Contains(source, ":RSP__")
}

func (c *Context) SetResponse(v []*fastjson.Value) {
	if len(v) == 0 {
		return
	}
	c.response = v[0]
}

func (c *Context) MarshalActionRequest(id string) string {
	actionResult, ok := c.taskResult[id]
	if !ok {
		// 待补充
		return ""
	}
	var req []byte
	req = actionResult.request.MarshalTo(req)
	return string(req)
}

func (c *Context) MarshalActionResponse(id string) string {
	actionResult, ok := c.taskResult[id]
	if !ok {
		// 待补充
		return ""
	}
	var req []byte
	req = actionResult.response.MarshalTo(req)
	return string(req)
}



func getTaskId(source string) string {
	idList := strings.Split(source, "__")
	if len(idList) < 3 {
		return ""
	}
	idString := strings.Replace(idList[1], ":RSP", "", 1)
	return strings.Replace(idString, ":REQ", "", 1)
}

func (c *Context) GetRequest() *fastjson.Value {
	return c.request
}

func (c *Context) getActionResponse(id string) (*fastjson.Value, error) {
	action, ok := c.taskResult[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("there is not exist actionResult, id: %s", id))
	}
	return action.response, nil
}

func (c *Context) getActionRequest(id string) (*fastjson.Value, error) {
	action, ok := c.taskResult[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("there is not exist actionResult, id: %s", id))
	}
	return action.request, nil
}

func getValue(v *fastjson.Value, dst string) (*fastjson.Value, error) {
	if len(dst) == 0 {
		return v, nil
	}
	return nil, errors.New("failed to get value")
}
