// +build !jsoniter

package internal

import "encoding/json"

var (
	JsonMarshal       = json.Marshal
	JsonUnmarshal     = json.Unmarshal
	JsonMarshalIndent = json.MarshalIndent
	JsonNewDecoder    = json.NewDecoder
	JsonNewEncoder    = json.NewEncoder
)
