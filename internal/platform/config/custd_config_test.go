package config

import (
	"reflect"
	"io"
	"testing"
)

type Test struct {
	testName string
	input    string
	expected map[string]string
}

var tests []Test

func init() {
	tests = []Test{
		{
			testName: "SimpleTest",
			input:    "a=1\nb=2\nc=3",
			expected: map[string]string{"a": "1", "b": "2", "c": "3"},
		},
	}
}

type MockConfig struct {
	contents        string
	currentPosition int
}

func (mc *MockConfig) Read(p []byte) (n int, err error) {
	bytesToRead := 0
	if mc.currentPosition >= len(mc.contents)-1 {
		return 0, io.EOF
	}

	if len(p) < len(mc.contents) {
		bytesToRead = len(p)
	} else {
		bytesToRead = len(mc.contents)
	}
	for i := 0; i < bytesToRead; i++ {
		p[i] = mc.contents[i]
		mc.currentPosition++
	}
	return bytesToRead, nil
}

func TestLoadConfig(t *testing.T) {
	for _, test := range tests {
		configs, err := LoadConfig(&MockConfig{contents: test.input})
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
		if !reflect.DeepEqual(configs, test.expected) {
			t.Errorf("expected %v, got %v", test.expected, configs)
		}
	}
}
