package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	loadgen "temporal-server-playground/loadgen"
)

func main() {
	c, err := client.Dial(client.Options{
		MetricsHandler: loadgen.NewMetricsHandler("temporal-worker"),
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client:", err)
	}
	defer c.Close()

	w := worker.New(c, loadgen.TaskQueue, worker.Options{
		MaxConcurrentWorkflowTaskPollers:       1000,
		MaxConcurrentActivityTaskPollers:       1000,
		MaxConcurrentActivityExecutionSize:     1000,
		MaxConcurrentWorkflowTaskExecutionSize: 1000,
	})
	w.RegisterWorkflow(loadgen.ParentWorkflow)
	w.RegisterWorkflow(loadgen.ChildWorkflow)
	w.RegisterActivity(loadgen.SimpleActivity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("worker stopped:", err)
	}
}
