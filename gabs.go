/*
Copyright (c) 2014 Ashley Jeffs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package gabs implements a simplified wrapper around json parsing an unknown structure
package gabs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

/*
Container - an internal structure that holds a reference to the core interface map of the parsed
json. Use this container to move context.
*/
type Container struct {
	object interface{}
}

/*
Search - Attempt to find and return an object within the JSON structure by specifying the hierarchy of
field names to locate the target.
*/
func (g *Container) Search(hierarchy ...string) *Container {
	var object interface{}

	object = g.object
	for target := 0; target < len(hierarchy); target++ {
		if mmap, ok := object.(map[string]interface{}); ok {
			object = mmap[hierarchy[target]]
		} else {
			return &Container{nil}
		}
	}

	return &Container{object}
}

/*
S - Shorthand method, does the same thing as Search.
*/
func (g *Container) S(hierarchy ...string) *Container {
	return g.Search(hierarchy...)
}

/*
Data - Return the contained data as an interface{}.
*/
func (g *Container) Data() interface{} {
	return g.object
}

/*
Set - Set the value for an object within the JSON structure by specifying the new value and the
hierarchy of field names to locate the target.
*/
func (g *Container) Set(value interface{}, hierarchy ...string) error {
	nParents := len(hierarchy)
	if nParents <= 0 {
		return errors.New("must specify at least one target parent")
	}

	parent := g.Search(hierarchy[:nParents-1]...).Data()

	if mmap, ok := parent.(map[string]interface{}); ok {
		mmap[hierarchy[nParents-1]] = value
	} else {
		return errors.New("target object was not found in structure")
	}

	return nil
}

/*
Push - Push a value onto a JSON array.
*/
func (g *Container) Push(value interface{}, hierarchy ...string) error {
	nParents := len(hierarchy)
	if nParents <= 0 {
		return errors.New("must specify at least one target parent")
	}

	parent := g.Search(hierarchy[:nParents-1]...).Data()

	if mmap, ok := parent.(map[string]interface{}); ok {
		target := mmap[hierarchy[nParents-1]]
		if array, ok := target.([]interface{}); ok {
			mmap[hierarchy[nParents-1]] = append(array, value)
		} else {
			return errors.New("target object was not an array")
		}
	} else {
		return errors.New("target object was not found in structure")
	}

	return nil
}

/*
CreateObject - Create a new JSON object. Returns a container of the new object.
*/
func (g *Container) CreateObject(name string) *Container {
	if mmap, ok := g.Data().(map[string]interface{}); ok {
		mmap[name] = map[string]interface{}{}
		return &Container{mmap[name]}
	}

	return &Container{nil}
}

/*
C - Shorthand method for CreateObject.
*/
func (g *Container) C(name string) *Container {
	return g.CreateObject(name)
}

/*
CreateArray - Create a new JSON array.
*/
func (g *Container) CreateArray(name string) error {
	if mmap, ok := g.Data().(map[string]interface{}); ok {
		mmap[name] = []interface{}{}
		return nil
	}
	return errors.New("container was not a valid object")
}

/*
A - Shorthand method for CreateArray.
*/
func (g *Container) A(name string) error {
	return g.CreateArray(name)
}

/*
String - Converts the contained object back to a JSON formatted string.
*/
func (g *Container) String() string {
	if bytes, err := json.Marshal(g.object); err == nil {
		return string(bytes)
	}
	return "{}"
}

/*
ParseJson - Convert a string into a representation of the parsed JSON.
*/
func ParseJson(sample []byte) (*Container, error) {
	var gabs Container

	if err := json.Unmarshal(sample, &gabs.object); err != nil {
		return nil, err
	}

	if _, ok := gabs.object.(map[string]interface{}); ok {
		return &gabs, nil
	}

	return nil, errors.New("json appears to contain no data.")
}

/*
ParseJsonFile - Read a file and convert into a representation of the parsed JSON.
*/
func ParseJsonFile(path string) (*Container, error) {
	if len(path) > 0 {
		if cBytes, err := ioutil.ReadFile(path); err == nil {
			if container, err := ParseJson(cBytes); err == nil {
				return container, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return nil, errors.New("file path was invalid")
}