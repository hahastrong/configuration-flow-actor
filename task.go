package cfa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fastjson"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

type Task interface {
	DoTask(ctx *Context) error
}

type TaskParam struct {
	ID          string                `json:"id"`
	TaskType    string                `json:"task_type"`
	Path        string                `json:"path"`
	Method      string                `json:"method"`
	ContentType string                `json:"content_type"`
	Request     map[string]*ParamNode `json:"request"` // 后续在更改为可以替换的
	Response    map[string]*ParamNode `json:"response"`
}

type ParamNode struct {
	Data      string            `json:"data"`
	Type      string            `json:"type"`       // string, number, bool, array
	Action    string            `json:"action"`     // data parse method, expr, data, iNfunc
	TargetIdx map[string]string `json:"target_idx"` // to replace target dst item
	SourceIdx map[string]string `json:"source_idx"` // to replace source dst item
}

func (p *ParamNode) exec(ctx *Context, k string) error {
	if p.Action == "expr" {
		v, _ := ctx.GetValue(p.Data)
		err := ctx.SetValue(k, v)
		if err != nil {
			return err
		}
	}

	if p.Action == "data" {
		if p.Type == "string" {
			tmp := ctx.A.NewString(p.Data)
			v := []*fastjson.Value{tmp}
			err := ctx.SetValue(k, v)
			if err != nil {
				return err
			}
		}
		if p.Type == "bool" {
			var v []*fastjson.Value
			if p.Data == "true" {
				v = append(v, ctx.A.NewTrue())
			} else if p.Data == "false" {
				v = append(v, ctx.A.NewFalse())
			} else {
				return errors.New(fmt.Sprintf("unexpected value, type: bool, value: %s", p.Data))
			}

			err := ctx.SetValue(k, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	HTTPREQUEST = "HttpRequest"
	DATABUILDER = "DataBuilder"
	YTBDOWNLOADER = "YtbDownloader"
)

type HttpRequest struct {
	tp *TaskParam
}

type DataBuilder struct {
	tp *TaskParam
}

type YtbDownloader struct {
	tp *TaskParam
}

type Start struct {
	tp *TaskParam
}

type End struct {
	tp *TaskParam
}

func (t *HttpRequest) DoTask(ctx *Context) error {
	for k, v := range t.tp.Request {
		dst := fmt.Sprintf("__%s:REQ__%s", t.tp.ID, k)
		err := v.exec(ctx, dst)
		if err != nil {
			return err
		}
	}

	actionRequest := ctx.MarshalActionRequest(t.tp.ID)
	var req  = new(Request)
	err := json.Unmarshal([]byte(actionRequest), req)
	if err != nil {
		return err
	}

	start := time.Now().UnixMilli()
	defer func() {
		fmt.Println(time.Now().UnixMilli() - start)
	}()
	if t.tp.Method == "GET" {

		// 添加参数
		if req.Get != nil {
			params := url.Values{}
			for k, v := range req.Get {
				params.Add(k, fmt.Sprintf("%v", v))
			}

			t.tp.Path += "?" + params.Encode()
		}

		client := &http.Client{}
		request, err := http.NewRequest(http.MethodGet, t.tp.Path, nil)
		if err != nil {
			// 添加日志打印
			return err
		}

		// 需要自定义请求头
		// 添加header
		rsp, err := client.Do(request)
		if err != nil {
			return err
		}

		rspBody, err := ioutil.ReadAll(rsp.Body)
		_ = rsp.Body.Close()

		err = ctx.SetHeaders(t.tp.ID, rsp.Header)
		if err != nil {
			return err
		}

		// 后续修改为 根据 rsp 中的 header 来判断
		if t.tp.ContentType == "json" {
			err = ctx.SetRsp(t.tp.ID, rspBody)
			if err != nil {
				return err
			}
		}
	}

	if t.tp.Method == "POST" {
		// 添加参数
		var body io.Reader
		if req.Post != nil {
			reqPost, err := json.Marshal(req.Post)
			if err != nil {
				return err
			}
			body = bytes.NewReader(reqPost)
		}

		client := &http.Client{}
		request, err := http.NewRequest(http.MethodPost, t.tp.Path, body)
		if err != nil {
			// 添加日志打印
			return err
		}

		// 需要自定义请求头
		// 添加header
		rsp, err := client.Do(request)
		if err != nil {
			return err
		}

		rspBody, err := ioutil.ReadAll(rsp.Body)
		_ = rsp.Body.Close()

		err = ctx.SetHeaders(t.tp.ID, rsp.Header)
		if err != nil {
			return err
		}

		// 后续修改为 根据 rsp 中的 header 来判断
		if t.tp.ContentType == "json" {
			err = ctx.SetRsp(t.tp.ID, rspBody)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *DataBuilder) DoTask(ctx *Context) error {
	for k, v := range t.tp.Response {
		dst := fmt.Sprintf("__RESPONSE__%s", k)
		err := v.exec(ctx, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *YtbDownloader) DoTask(ctx *Context) error {
	for k, v := range t.tp.Request {
		dst := fmt.Sprintf("__%s:REQ__%s", t.tp.ID,k)
		err := v.exec(ctx, dst)
		if err != nil {
			return err
		}
	}

	actionRequest := ctx.MarshalActionRequest(t.tp.ID)
	var req  = new(YtbParams)
	err := json.Unmarshal([]byte(actionRequest), req)
	if err != nil {
		return err
	}

	if req.Url == "" {
		return errors.New("miss url params")
	}

	// generate shell script
	var args []string
	cmd := "/usr/local/bin/yt-dlp"

	args = append(args, "-N")
	args = append(args, "12")

	args = append(args, "--yes-playlist")

	outPutFormat := "mp4"
	if req.Format != "" {
		outPutFormat = req.Format
	}
	args = append(args, "--merge-output-format")
	args = append(args, outPutFormat)
	args = append(args, req.Url)

	fmt.Println(args)

	go func(url string) {
		// add a goroutine to exec the task
		command := exec.Command(cmd, args...)
		command.Dir = "/Users/hahastrong/"

		err = command.Start()
		fmt.Printf("download video success, url: %s", url)
	}(req.Url)

	return nil
}

func (t *Start) DoTask(ctx *Context) error {
	return nil
}

func (t *End) DoTask(ctx *Context) error {
	return nil
}

func NewTask(tp *TaskParam, taskType string) Task {
	switch taskType {
	case START:
		return &Start{tp: tp}
	case END:
		return &Start{tp: tp}
	case HTTPREQUEST:
		return &HttpRequest{tp: tp}
	case DATABUILDER:
		return &DataBuilder{tp: tp}
	case YTBDOWNLOADER:
		return &YtbDownloader{tp: tp}
	default:
		return nil
	}
}
