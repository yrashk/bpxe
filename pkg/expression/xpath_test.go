// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

import (
	"testing"

	"bpxe.org/pkg/bpmn"
	"bpxe.org/pkg/data"
	"github.com/stretchr/testify/assert"
)

func TestXPath(t *testing.T) {
	var engine Engine = NewXPath()
	compiled, err := engine.CompileExpression("a > 1")
	assert.Nil(t, err)
	result, err := engine.EvaluateExpression(compiled, map[string]interface{}{
		"a": 2,
	})
	assert.Nil(t, err)
	assert.True(t, result.(bool))
}

type dataObjects map[string]data.ItemAware

func (d dataObjects) FindItemAwareById(id bpmn.IdRef) (itemAware data.ItemAware, found bool) {
	itemAware, found = d[id]
	return
}

func (d dataObjects) FindItemAwareByName(name string) (itemAware data.ItemAware, found bool) {
	itemAware, found = d[name]
	return
}

func TestXPath_getDataObject(t *testing.T) {
	// This funtionality doesn't quite work yet
	t.SkipNow()
	var engine = NewXPath()
	container := data.NewContainer(nil)
	container.Put(data.XMLSource(`<tag attr="val"/>`))
	var objs dataObjects = map[string]data.ItemAware{
		"dataObject": container,
	}
	engine.SetItemAwareLocator(objs)
	compiled, err := engine.CompileExpression("(getDataObject('dataObject')/tag/@attr/string())[1]")
	assert.Nil(t, err)
	result, err := engine.EvaluateExpression(compiled, map[string]interface{}{})
	assert.Nil(t, err)
	assert.Equal(t, "val", result.(string))
}
