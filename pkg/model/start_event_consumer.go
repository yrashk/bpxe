// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package model

import (
	"context"
	"sync"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/event"
	"bpxe.org/pkg/process"
	"bpxe.org/pkg/process/instance"
	"bpxe.org/pkg/tracing"
)

type StartEventConsumer struct {
	process                                *process.Process
	parallel                               bool
	eventInstances, originalEventInstances []event.Instance
	ctx                                    context.Context
	consumptionLock                        sync.Mutex
	tracer                                 *tracing.Tracer
}

func NewStartEventConsumer(
	ctx context.Context,
	tracer *tracing.Tracer,
	process *process.Process,
	startEvent *bpmn.StartEvent, builder event.InstanceBuilder) *StartEventConsumer {
	consumer := &StartEventConsumer{
		ctx:      ctx,
		process:  process,
		parallel: startEvent.ParallelMultiple(),
		tracer:   tracer,
	}
	consumer.eventInstances = make([]event.Instance, len(startEvent.EventDefinitions()))
	for k := range startEvent.EventDefinitions() {
		consumer.eventInstances[k] = builder.NewEventInstance(startEvent.EventDefinitions()[k])
	}
	consumer.originalEventInstances = consumer.eventInstances
	return consumer
}

func (s *StartEventConsumer) ConsumeEvent(ev event.Event) (result event.ConsumptionResult, err error) {
	s.consumptionLock.Lock()
	defer s.consumptionLock.Unlock()

	for i := range s.eventInstances {
		if ev.MatchesEventInstance(s.eventInstances[i]) {
			if !s.parallel {
				goto instantiate
			} else {
				s.eventInstances[i] = s.eventInstances[len(s.eventInstances)-1]
				s.eventInstances = s.eventInstances[0 : len(s.eventInstances)-1]
				if len(s.eventInstances) == 0 {
					s.eventInstances = s.originalEventInstances
					goto instantiate
				}
			}
		}
	}
	result = event.Consumed
	return
instantiate:
	var inst *instance.Instance
	inst, err = s.process.Instantiate(
		instance.WithContext(s.ctx),
		instance.WithTracer(s.tracer),
	)
	if err != nil {
		result = event.ConsumptionError
		return
	}
	result, err = inst.ConsumeEvent(ev)
	if err != nil {
		result = event.ConsumptionError
		return
	}
	return
}
