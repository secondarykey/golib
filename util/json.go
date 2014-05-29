package util

import (
	"encoding/json"
	"io/ioutil"
)

/*
 */
func ReadJsonFile(token interface{}, filename string) error {
	if b, err := ioutil.ReadFile(filename); err != nil {
		return err
	} else {
		return json.Unmarshal(b, token)
	}
}

/*
 */
func WriteJsonFile(token interface{}, filename string) error {
	if b, err := json.Marshal(token); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filename, b, 0666)
	}
}
