// Package decoder reads independent JSON values from newline-delimited input.
package decoder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Decode returns the nonempty JSON records in order.
func Decode(r io.Reader) ([]json.RawMessage, error) {
	s := bufio.NewScanner(r)
	var records []json.RawMessage
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			return nil, fmt.Errorf("invalid record: %s", line)
		}
		records = append(records, json.RawMessage(line))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
