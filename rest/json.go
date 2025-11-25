package rest

import (
	"encoding/json"
	"io"
)

// JsonMarshalToString marshals a value as a JSON string.
// To replace this with a Jsoniter implementation, set
//
//	rest.JsonMarshalToString = jsoniter.JSON.MarshalToString
var JsonMarshalToString = func(v any) (string, error) {
	bs, err := json.Marshal(v)
	return string(bs), err
}

// JsonUnmarshal unmarshals a value. The decoder UseNumber() option is set;
// replace this function if necessary.
// To replace this with a Jsoniter implementation, set
//
//	rest.JsonUnmarshal = func(r io.Reader, output any) error {
//		decoder := jsoniter.JSON.NewDecoder(r)
//		decoder.UseNumber()
//		return decoder.Decode(output)
//	}
var JsonUnmarshal = func(r io.Reader, output any) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	return decoder.Decode(output)
}
