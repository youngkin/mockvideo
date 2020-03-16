// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"bufio"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/juju/errors"
)

// LoadConfig loads the custd service configuration and returns a map of key/value pairs or an error
func LoadConfig(configData io.Reader) (map[string]string, error) {
	config := make(map[string]string)
	lineReader := bufio.NewScanner(configData)
	for lineReader.Scan() {
		line := lineReader.Text()
		keyVal := strings.Split(line, "=")
		key := keyVal[0]
		val := keyVal[1]
		config[key] = val
	}

	return config, nil
}

// LoadSecrets loads the custd service's secrets and returns a map of key/value pairs or an error
func LoadSecrets(secretsDir string) (map[string]string, error) {
	secrets := make(map[string]string)

	secretFiles := []string{"dbuser", "dbpassword"}

	for _, fileName := range secretFiles {
		content, err := ioutil.ReadFile(filepath.Join(secretsDir, fileName))
		if err != nil {
			return nil, errors.Annotatef(err, "Secrets file %s could not be read", filepath.Join(secretsDir, fileName))
		}

		secrets[fileName] = string(content)
	}

	return secrets, nil
}
