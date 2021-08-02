/*
Copyright (c) 2021 T-Systems International GmbH, SAP SE or an SAP affiliate company. All right reserved
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package patch

const (
	addOperation     = "add"
	removeOperation  = "remove"
	replaceOperation = "replace"
	copyOperation    = "copy"
	moveOperation    = "move"
)

// Operation is an operation of a JSON patch https://tools.ietf.org/html/rfc6902.
type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	From  string      `json:"from"`
	Value interface{} `json:"value,omitempty"`
}

// AddPatchOperation returns an add JSON patch operation.
func AddPatchOperation(path string, value interface{}) Operation {
	return Operation{
		Op:    addOperation,
		Path:  path,
		Value: value,
	}
}

// RemovePatchOperation returns a remove JSON patch operation.
func RemovePatchOperation(path string) Operation {
	return Operation{
		Op:   removeOperation,
		Path: path,
	}
}

// ReplacePatchOperation returns a replace JSON patch operation.
func ReplacePatchOperation(path string, value interface{}) Operation {
	return Operation{
		Op:    replaceOperation,
		Path:  path,
		Value: value,
	}
}

// CopyPatchOperation returns a copy JSON patch operation.
func CopyPatchOperation(from, path string) Operation {
	return Operation{
		Op:   copyOperation,
		Path: path,
		From: from,
	}
}

// MovePatchOperation returns a move JSON patch operation.
func MovePatchOperation(from, path string) Operation {
	return Operation{
		Op:   moveOperation,
		Path: path,
		From: from,
	}
}
