package nomad

import (
	"fmt"
	"strings"

	"github.com/hashicorp/nomad/api"
)

type Client struct {
	client *api.Client
}

func NewClient() (*Client, error) {
	// Assumes NOMAD_ADDR and NOMAD_TOKEN are set in env or defaults to localhost
	cfg := api.DefaultConfig()
	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{client: c}, nil
}

func (c *Client) DeployInstance(cfg JobConfig) error {
	job, err := GenerateJob(cfg)
	if err != nil {
		return err
	}

	_, _, err = c.client.Jobs().Register(job, nil)
	return err
}

func (c *Client) StopInstance(orgID int64) error {
	jobName := fmt.Sprintf("railzway-org-%d", orgID)
	_, _, err := c.client.Jobs().Deregister(jobName, true, nil)
	return err
}

func (c *Client) GetInstanceStatus(orgID int64) (string, error) {
	jobName := fmt.Sprintf("railzway-org-%d", orgID)
	allocs, _, err := c.client.Jobs().Allocations(jobName, false, nil)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return "not_found", nil
		}
		return "", err
	}

	if len(allocs) == 0 {
		return "pending", nil
	}

	alloc := latestAllocation(allocs)
	if alloc == nil {
		return "pending", nil
	}

	status := strings.ToLower(strings.TrimSpace(alloc.ClientStatus))
	if status != "running" {
		return status, nil
	}

	if allocationReady(alloc) {
		return "running", nil
	}

	return "pending", nil
}

func latestAllocation(allocs []*api.AllocationListStub) *api.AllocationListStub {
	var latest *api.AllocationListStub
	for _, alloc := range allocs {
		if alloc == nil {
			continue
		}
		if latest == nil {
			latest = alloc
			continue
		}
		if alloc.ModifyIndex > latest.ModifyIndex {
			latest = alloc
			continue
		}
		if alloc.ModifyIndex == latest.ModifyIndex && alloc.CreateIndex > latest.CreateIndex {
			latest = alloc
		}
	}
	return latest
}

func allocationReady(alloc *api.AllocationListStub) bool {
	if alloc == nil {
		return false
	}
	if alloc.DesiredStatus != "" && strings.ToLower(alloc.DesiredStatus) != api.AllocDesiredStatusRun {
		return false
	}
	if len(alloc.TaskStates) == 0 {
		return false
	}
	for _, state := range alloc.TaskStates {
		if state == nil {
			return false
		}
		if state.Failed {
			return false
		}
		if strings.ToLower(strings.TrimSpace(state.State)) != "running" {
			return false
		}
	}
	return true
}
