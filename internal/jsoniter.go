// +build jsoniter

package internal

import "github.com/json-iterator/go"

var (
	json              = jsoniter.ConfigCompatibleWithStandardLibrary
	JsonMarshal       = json.Marshal
	JsonUnmarshal     = json.Unmarshal
	JsonMarshalIndent = json.MarshalIndent
	JsonNewDecoder    = json.NewDecoder
	JsonNewEncoder    = json.NewEncoder
)
