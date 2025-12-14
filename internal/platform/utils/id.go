package util

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Call this during app init (e.g., main.go)
func InitSnowflakeNode(machineID int64) error {
	snowflake.Epoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6

	n, err := snowflake.NewNode(machineID)
	if err != nil {
		return err
	}
	node = n
	return nil
}

func GenerateID() int64 {
	return node.Generate().Int64()
}
