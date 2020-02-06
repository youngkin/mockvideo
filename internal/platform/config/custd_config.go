package config

import (
	"bufio"
	"io"
	"strings"
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
