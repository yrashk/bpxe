// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package expression

import (
	"bytes"
	"reflect"

	"bpxe.org/pkg/data"
	"bpxe.org/pkg/errors"
	"github.com/ChrisTrenkamp/xsel/exec"
	"github.com/ChrisTrenkamp/xsel/grammar"
	"github.com/ChrisTrenkamp/xsel/parser"
	"github.com/ChrisTrenkamp/xsel/store"
	"github.com/Chronokeeper/anyxml"
)

// XPath language engine
//
// Implementation details and limitations as per https://github.com/antchfx/xpath
type XPath struct {
	itemAwareLocator data.ItemAwareLocator
}

func (engine *XPath) SetItemAwareLocator(itemAwareLocator data.ItemAwareLocator) {
	engine.itemAwareLocator = itemAwareLocator
}

func MakeXPath() XPath {
	return XPath{}
}

func NewXPath() *XPath {
	engine := MakeXPath()
	return &engine
}

func (engine *XPath) CompileExpression(source string) (result CompiledExpression, err error) {
	compiled, err := grammar.Build(source)
	if err == nil {
		result = &compiled
	}
	return
}

func (engine *XPath) EvaluateExpression(e CompiledExpression,
	datum interface{},
) (result Result, err error) {
	if expression, ok := e.(*grammar.Grammar); ok {
		// Here, in order to save some prototyping type,
		// instead of implementing `parser.Parser` for `interface{}`,
		// we use it over `interface{}` serialized as XML.
		// This is not very efficient but does the job for now.
		// Eventually, a direct implementation of `parser.Parser`
		// over `interface{}` should be developed to optimize this path.

		var serialized []byte
		serialized, err = anyxml.Xml(datum)
		if err != nil {
			result = nil
			return
		}
		p := parser.ReadXml(bytes.NewBuffer(serialized))

		contextSettings := func(c *exec.ContextSettings) {
			if engine.itemAwareLocator != nil {
				c.FunctionLibrary[exec.Name("", "getDataObject")] = engine.getDataObject()
			}
		}

		var cursor store.Cursor
		cursor, err = store.CreateInMemory(p)
		if err != nil {
			return
		}
		var res exec.Result
		res, err = exec.Exec(cursor, expression, contextSettings)
		if err != nil {
			return
		}
		switch r := res.(type) {
		case exec.String:
			result = r.String()
		case exec.Bool:
			result = r.Bool()
		case exec.Number:
			result = r.Number()
		case exec.NodeSet:
			result = r
		}
	} else {
		err = errors.InvalidArgumentError{
			Expected: "CompiledExpression to be *github.com/ChrisTrenkamp/xsel/grammar#Grammar",
			Actual:   reflect.TypeOf(e),
		}
	}
	return
}

var asXMLType = reflect.TypeOf(new(data.AsXML)).Elem()

func (engine *XPath) getDataObject() func(context exec.Context, args ...exec.Result) (exec.Result, error) {
	return func(context exec.Context, args ...exec.Result) (exec.Result, error) {
		var name string
		switch len(args) {
		case 0:
			return nil, errors.InvalidArgumentError{Expected: "at least one argument", Actual: "none"}
		case 2:
			return nil, errors.NotSupportedError{
				What:   "two-argument getDataObject",
				Reason: "BPXE doesn't support sub-processes yet",
			}
		case 1:
			name = args[0].String()
		}
		itemAware, found := engine.itemAwareLocator.FindItemAwareByName(name)
		if !found {
			return exec.NodeSet{}, nil
		}
		item := <-itemAware.Get()
		switch item := item.(type) {
		case string:
			return exec.String(item), nil
		case float64:
			return exec.Number(item), nil
		case bool:
			return exec.Bool(item), nil
		default:
			// Until we have own data type to represent XML nodes, we'll piggy-back
			// on xsel's parser and datum.AsXML interface. This is not very efficient,
			// but should do for now
			if reflect.TypeOf(item).Implements(asXMLType) {
				p := parser.ReadXml(bytes.NewReader(item.(data.AsXML).AsXML()))
				cursor, err := store.CreateInMemory(p)
				if err != nil {
					return nil, err
				}
				return exec.NodeSet{cursor}, nil
			} else {
				return nil, errors.InvalidArgumentError{
					Expected: "XML-serializable value (string, float64 or Node)",
					Actual:   item,
				}
			}
		}
	}
}
