package cli

import (
	"encoding/json"
	"strconv"
)

func marshalJSON(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
