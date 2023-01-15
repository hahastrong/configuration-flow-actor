package cfa

import (
	"errors"
)

const (
	GATEWAY = "gateway"
	START = "start"
	END = "end"
	TASK = "task"
)

type Flow map[string]*Node

type Node struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"` // start, end, gateway, task
	Expr    string    `json:"expr"`
	Next    string    `json:"next"`
	Default string    `json:"default"`
	Task    *TaskParam `json:"task"`
}

type RunNode struct {
	n *Node
	tp Task
	flow Flow
}

func (n *RunNode) Run(ctx *Context) error {
	if n.tp == nil {
		return nil
	}

	ctx.NewTaskResult(n.n.ID)

	err := n.tp.DoTask(ctx)
	return err
}

func (f *Flow) getDefault(ctx *Context, defaultId string) (*RunNode, error) {
	for k, _ := range *f {
		if k == defaultId {
			ctx.NewTaskResult(k)
			return &RunNode{
				n: (*f)[k],
				tp: NewTask((*f)[k].Task, (*f)[k].Task.TaskType),
			}, nil

		}
	}
	return nil, errors.New("failed to get default node")
}

func (f *Flow) getNext(ctx *Context, next string) (*RunNode, error) {
	if next == "" {
		return nil, nil
	}
	for k, _ := range *f {
		if k == next {
			if k == END {
				return &RunNode{
					n: (*f)[k],
				}, nil
			}

			ctx.NewTaskResult(k)
			return &RunNode{
				n: (*f)[k],
				tp: NewTask((*f)[k].Task, (*f)[k].Task.TaskType),
			}, nil

		}
	}
	return nil, errors.New("failed to get next node")
}

func (rn *RunNode) eval(ctx *Context) (bool, error) {
	if rn.n.Type != GATEWAY {
		return true, nil
	}
	// add gateway condition
	return true, nil
}

func (f Flow) getStart() (*RunNode, error) {
	for k, _ := range f {
		if k == START {
			return &RunNode{
				n: f[k],
				tp: NewTask(f[k].Task, START),
			}, nil

		}
	}

	return nil, errors.New("failed to get start node")
}

func (f *Flow) Run(ctx *Context) error {
	n, err := f.getStart()
	if err != nil {
		return err
	}

	for n != nil {
		// gateway
		rv, err := n.eval(ctx)
		if err != nil {
			return err
		}

		if err := n.Run(ctx); err != nil {
			return err
		}

		if rv {
			n, err = f.getNext(ctx, n.n.Next)
		} else {
			n, err = f.getDefault(ctx, n.n.Default)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
