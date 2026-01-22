package snowflake

import (
	"strconv"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/fx"
)

var Module = fx.Module("snowflake",
	fx.Provide(NewNode),
)

// Node wraps snowflake.Node to abstract dependency
type Node struct {
	*snowflake.Node
}

func NewNode() (*Node, error) {
	// For now, we hardcode node ID to 1. In production, this should come from config/env
	// to ensure uniqueness across distributed service instances.
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	return &Node{node}, nil
}

// GenerateID returns a new snowflake ID as int64
func (n *Node) GenerateID() int64 {
	return n.Generate().Int64()
}

// ParseID parses a string ID into an int64
func ParseID(id string) (int64, error) {
	nid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return nid, nil
}
