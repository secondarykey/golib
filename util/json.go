//Json Utility
package util

import (
	"encoding/json"
	"io/ioutil"
)

//json file{f} changed into a structure{v}
func ReadJsonFile(v interface{}, f string) error {
	if b, err := ioutil.ReadFile(f); err != nil {
		return err
	} else {
		return json.Unmarshal(b, v)
	}
}

//structure{v} changed into a json file{f}
func WriteJsonFile(v interface{}, f string) error {
	if b, err := json.Marshal(v); err != nil {
		return err
	} else {
		return ioutil.WriteFile(f, b, 0666)
	}
}
