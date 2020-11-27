// Code generated by https://github.com/GoComply/metaschema; DO NOT EDIT.
// Multiplexers are indirect models needed for serialization/deserialization
// as json and xml files differ materially in their structure.
package assessment_common

import (
	"bytes"
	"encoding/json"
)

type ObjectiveMultiplexer []Objective

func (mplex *ObjectiveMultiplexer) UnmarshalJSON(b []byte) error {
	var l []Objective
	switch b[0] {
	case '{':
		var singleton Objective
		if err := json.Unmarshal(b, &singleton); err != nil {
			return err
		}
		l = append(l, singleton)
	default:
		if err := json.Unmarshal(b, &l); err != nil {
			return err
		}
	}
	(*mplex) = l
	return nil
}

func (mplex *ObjectiveMultiplexer) MarshalJSON() ([]byte, error) {
	js := bytes.NewBuffer([]byte{'['})

	empty := true
	for _, v := range *mplex {
		if !empty {
			if err := js.WriteByte(','); err != nil {
				return []byte{}, err
			}
		}
		empty = false

		text, err := json.Marshal(v)
		if err != nil {
			return []byte{}, err
		}
		if _, err = js.Write(text); err != nil {
			return []byte{}, err
		}
	}
	if err := js.WriteByte(']'); err != nil {
		return []byte{}, err
	}
	return js.Bytes(), nil
}
