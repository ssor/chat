package redis

import (
	redigo "github.com/ssor/redigo/redis"
)

type Script struct {
	*redigo.Script
}

func NewScript(keyCount int, src string) *Script {
	return &Script{redigo.NewScript(keyCount, src)}
}
