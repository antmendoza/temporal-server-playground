package loadgen

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const TaskQueue = "load-gen"

type Params struct {
	NumActivities     int
	NumChildWorkflows int
}

func ParentWorkflow(ctx workflow.Context, p Params) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Start activities concurrently
	actFutures := make([]workflow.Future, p.NumActivities)
	for i := 0; i < p.NumActivities; i++ {
		actFutures[i] = workflow.ExecuteActivity(ctx, SimpleActivity)
	}

	// Start child workflows concurrently
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{})
	childFutures := make([]workflow.Future, p.NumChildWorkflows)
	for i := 0; i < p.NumChildWorkflows; i++ {
		childFutures[i] = workflow.ExecuteChildWorkflow(childCtx, ChildWorkflow)
	}

	// Wait for all
	for _, f := range actFutures {
		if err := f.Get(ctx, nil); err != nil {
			return err
		}
	}
	for _, f := range childFutures {
		if err := f.Get(ctx, nil); err != nil {
			return err
		}
	}

	return nil
}

func ChildWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(ctx, SimpleActivity).Get(ctx, nil)
}
