package json

import jsoniter "github.com/json-iterator/go"

// Package json provides global JSON encoding and decoding functions used across Lokstra.
//
// By default, it uses json-iterator (which is faster than encoding/json) with a configuration
// compatible with the standard library.
//
// All exported variables behave like their counterparts in the standard encoding/json package,
// and can be overridden globally by application code if needed.
//
// To switch to the standard "encoding/json" implementation:
//
//	import (
//		stdjson "encoding/json"
//		"github.com/primadi/lokstra/core/json"
//	)
//
//	func SwitchToStandarJSON() {
//		json.Marshal = stdjson.Marshal
//		json.Unmarshal = stdjson.Unmarshal
//		json.NewEncoder = stdjson.NewEncoder
//		json.NewDecoder = stdjson.NewDecoder
//		json.MarshalIndent = stdjson.MarshalIndent
//	}

type Unmarshaler interface {
	UnmarshalJSON([]byte) error
}

var (
	Marshal       = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal
	Unmarshal     = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal
	NewEncoder    = jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder
	NewDecoder    = jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder
	MarshalIndent = jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent
)
