// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package model

import (
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/id"
	"bpxe.org/pkg/process"
)

type Model struct {
	Element              *bpmn.Definitions
	processes            []process.Process
	eventConsumersLock   sync.RWMutex
	eventConsumers       []event.Consumer
	idGeneratorBuilder   id.GeneratorBuilder
	eventInstanceBuilder event.InstanceBuilder
}

type Option func(*Model)

func WithIdGenerator(builder id.GeneratorBuilder) Option {
	return func(model *Model) {
		model.idGeneratorBuilder = builder
	}
}

func WithEventInstanceBuilder(builder event.InstanceBuilder) Option {
	return func(model *Model) {
		model.eventInstanceBuilder = builder
	}
}

func New(element *bpmn.Definitions, options ...Option) *Model {
	procs := element.Processes()
	model := &Model{
		Element: element,
	}

	for _, option := range options {
		option(model)
	}

	if model.idGeneratorBuilder == nil {
		model.idGeneratorBuilder = id.DefaultIdGeneratorBuilder
	}

	model.processes = make([]process.Process, len(*procs))
	for i := range *procs {
		model.processes[i] = process.Make(&(*procs)[i], element,
			process.WithIdGenerator(model.idGeneratorBuilder),
			process.WithEventIngress(model), process.WithEventEgress(model),
			process.WithEventInstanceBuilder(model),
		)
	}
	return model
}

func (model *Model) Run() {
	/*for i := range *model.Element.Processes() {
		instantiatingFlowNodes := (*model.Element.Processes())[i].InstantiatingFlowNodes()
	}*/
}

func (model *Model) FindProcessBy(f func(*process.Process) bool) (result *process.Process, found bool) {
	for i := range model.processes {
		if f(&model.processes[i]) {
			result = &model.processes[i]
			found = true
			return
		}
	}
	return
}

func (model *Model) ConsumeEvent(ev event.Event) (result event.ConsumptionResult, err error) {
	model.eventConsumersLock.RLock()
	result, err = event.ForwardProcessEvent(ev, &model.eventConsumers)
	model.eventConsumersLock.RUnlock()
	return
}

func (model *Model) RegisterEventConsumer(ev event.Consumer) (err error) {
	model.eventConsumersLock.Lock()
	model.eventConsumers = append(model.eventConsumers, ev)
	model.eventConsumersLock.Unlock()
	return
}

func (model *Model) NewEventInstance(def bpmn.EventDefinitionInterface) event.Instance {
	if model.eventInstanceBuilder != nil {
		return model.eventInstanceBuilder.NewEventInstance(def)
	} else {
		return event.NewInstance(def)
	}
}
