package util

import (
	"os"
	"testing"
)

type TestStruct struct {
	num  int32
	name string
}

func TestReadJsonFile(t *testing.T) {
	err := ReadJsonFile(data, "")
}

func TestWriteJsonFile(t *testing.T) {
	err := WriteJsonFile(data, "")
}
