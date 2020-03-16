// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"io"
	"reflect"
	"testing"
)

type Test struct {
	testName   string
	testDir    string
	input      string
	expected   map[string]string
	expectFail bool
}

var (
	configTests []Test
	secretTests []Test
)

func init() {
	configTests = []Test{
		{
			testName:   "SimpleTest",
			input:      "a=1\nb=2\nc=3",
			expected:   map[string]string{"a": "1", "b": "2", "c": "3"},
			expectFail: false,
		},
	}

	secretTests = []Test{
		{
			testName:   "SimpleSecretTest",
			testDir:    "./testdata",
			expected:   map[string]string{"dbuser": "someuser", "dbpassword": "somepassword"},
			expectFail: false,
		},
		{
			testName:   "SecretMissingDirectory",
			testDir:    "someNonExistentDirectory",
			expected:   map[string]string{"dbuser": "someuser", "dbpassword": "somepassword"},
			expectFail: true,
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
	for _, test := range configTests {
		t.Run(test.testName, func(t *testing.T) {
			configs, err := LoadConfig(&MockConfig{contents: test.input})
			if err != nil && test.expectFail == false {
				t.Errorf("TestName: %s, Expected nil error, got %v", test.testName, err)
			}
			if err == nil && test.expectFail == true {
				t.Errorf("TestName: %s, Expected non-nil error", test.testName)
			}
			if !test.expectFail && !reflect.DeepEqual(configs, test.expected) {
				t.Errorf("TestName: %s, expected %v, got %v", test.testName, test.expected, configs)
			}
		})
	}
}

func TestLoadSecrets(t *testing.T) {
	for _, test := range secretTests {
		t.Run(test.testName, func(t *testing.T) {
			secrets, err := LoadSecrets(test.testDir)
			if err != nil && test.expectFail == false {
				t.Errorf("TestName: %s, Expected nil error, got %v", test.testName, err)
			}
			if err == nil && test.expectFail == true {
				t.Errorf("TestName: %s, Expected non-nil error", test.testName)
			}

			if !test.expectFail && !reflect.DeepEqual(secrets, test.expected) {
				t.Errorf("TestName: %s, expected %v, got %v", test.testName, test.expected, secrets)
			}
		})
	}
}
