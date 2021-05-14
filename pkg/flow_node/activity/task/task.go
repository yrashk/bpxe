// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package task

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/flow/flow_interface"
	"bpxe.org/pkg/flow_node"
	"bpxe.org/pkg/flow_node/activity"
	"bpxe.org/pkg/tracing"
)

type message interface {
	message()
}

type nextActionMessage struct {
	response chan flow_node.Action
}

func (m nextActionMessage) message() {}

type incomingMessage struct {
	index int
}

func (m incomingMessage) message() {}

type cancelMessage struct {
	response chan bool
}

func (m cancelMessage) message() {}

type Task struct {
	flow_node.T
	element        *bpmn.Task
	runnerChannel  chan message
	activeBoundary chan bool
	bodyLock       sync.RWMutex
	body           func(*Task, context.Context) flow_node.Action
	ctx            context.Context
	cancel         context.CancelFunc
}

// SetBody override Task's body with an arbitrary function
//
// Since Task implements Abstract Task, it does nothing by default.
// This allows to add an implementation. Primarily used for testing.
func (node *Task) SetBody(body func(*Task, context.Context) flow_node.Action) {
	node.bodyLock.Lock()
	defer node.bodyLock.Unlock()
	node.body = body
}

func NewTask(startEvent *bpmn.Task) activity.Constructor {
	return func(process *bpmn.Process,
		definitions *bpmn.Definitions,
		eventIngress event.ProcessEventConsumer,
		eventEgress event.ProcessEventSource,
		tracer *tracing.Tracer,
		flowNodeMapping *flow_node.FlowNodeMapping,
		flowWaitGroup *sync.WaitGroup,
	) (node activity.Activity, err error) {
		flowNode, err := flow_node.New(process,
			definitions,
			&startEvent.FlowNode,
			eventIngress, eventEgress,
			tracer, flowNodeMapping,
			flowWaitGroup)
		if err != nil {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		taskNode := &Task{
			T:              *flowNode,
			element:        startEvent,
			runnerChannel:  make(chan message, len(flowNode.Incoming)*2+1),
			activeBoundary: make(chan bool),
			ctx:            ctx,
			cancel:         cancel,
		}
		go taskNode.runner()
		node = taskNode
		return
	}
}

func (node *Task) runner() {
	for {
		msg := <-node.runnerChannel
		switch m := msg.(type) {
		case cancelMessage:
			node.cancel()
			m.response <- true
		case nextActionMessage:
			node.activeBoundary <- true
			go func() {
				var action flow_node.Action
				action = flow_node.FlowAction{SequenceFlows: flow_node.AllSequenceFlows(&node.Outgoing)}
				if node.body != nil {
					node.bodyLock.RLock()
					action = node.body(node, node.ctx)
					node.bodyLock.RUnlock()
				}
				node.activeBoundary <- false
				m.response <- action
			}()
		default:
		}
	}
}

func (node *Task) NextAction(flow_interface.T) chan flow_node.Action {
	response := make(chan flow_node.Action)
	node.runnerChannel <- nextActionMessage{response: response}
	return response
}

func (node *Task) Incoming(index int) {
	node.runnerChannel <- incomingMessage{index: index}
}

func (node *Task) Element() bpmn.FlowNodeInterface {
	return node.element
}

func (node *Task) ActiveBoundary() <-chan bool {
	return node.activeBoundary
}

func (node *Task) Cancel() <-chan bool {
	response := make(chan bool)
	node.runnerChannel <- cancelMessage{response: response}
	return response
}
