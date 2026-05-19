package tgutil

import "strings"

const Separator = ":"

func Encode(parts ...string) string {
	return strings.Join(parts, Separator)
}

func Decode(data string) []string {
	return strings.Split(data, Separator)
}
