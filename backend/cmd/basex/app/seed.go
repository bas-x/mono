package app

import (
	"errors"
	"strings"
)

func parseSeed(raw string) ([32]byte, error) {
	var out [32]byte
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return out, nil
	}
	if len(trimmed) == 0 {
		return out, errors.New("empty seed")
	}
	copy(out[:], []byte(trimmed))
	return out, nil
}
