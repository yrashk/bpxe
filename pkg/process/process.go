// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package process

import (
	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/process/instance"
)

type Process struct {
	Element              *bpmn.Process
	Definitions          *bpmn.Definitions
	instances            []*instance.Instance
	EventIngress         event.Consumer
	EventEgress          event.Source
	idGeneratorBuilder   id.GeneratorBuilder
	eventInstanceBuilder event.InstanceBuilder
}

type Option func(*Process)

func WithIdGenerator(builder id.GeneratorBuilder) Option {
	return func(process *Process) {
		process.idGeneratorBuilder = builder
	}
}

func WithEventIngress(consumer event.Consumer) Option {
	return func(process *Process) {
		process.EventIngress = consumer
	}
}

func WithEventEgress(source event.Source) Option {
	return func(process *Process) {
		process.EventEgress = source
	}
}

func WithEventInstanceBuilder(builder event.InstanceBuilder) Option {
	return func(process *Process) {
		process.eventInstanceBuilder = builder
	}
}

func Make(element *bpmn.Process, definitions *bpmn.Definitions, options ...Option) Process {
	process := Process{
		Element:     element,
		Definitions: definitions,
		instances:   make([]*instance.Instance, 0),
	}

	for _, option := range options {
		option(&process)
	}

	if process.idGeneratorBuilder == nil {
		process.idGeneratorBuilder = id.DefaultIdGeneratorBuilder
	}

	if process.eventInstanceBuilder == nil {
		process.eventInstanceBuilder = event.DefaultInstanceBuilder{}
	}

	if process.EventIngress == nil && process.EventEgress == nil {
		fanOut := event.NewFanOut()
		process.EventIngress = fanOut
		process.EventEgress = fanOut
	}
	return process
}

func New(element *bpmn.Process, definitions *bpmn.Definitions, options ...Option) *Process {
	process := Make(element, definitions, options...)
	return &process
}

func (process *Process) Instantiate(options ...instance.Option) (inst *instance.Instance, err error) {
	options = append([]instance.Option{
		instance.WithIdGenerator(process.idGeneratorBuilder),
		instance.WithEventInstanceBuilder(process.eventInstanceBuilder),
		instance.WithEventEgress(process.EventEgress),
		instance.WithEventIngress(process.EventIngress),
	}, options...)
	inst, err = instance.NewInstance(process.Element, process.Definitions, options...)
	if err != nil {
		return
	}

	return
}
