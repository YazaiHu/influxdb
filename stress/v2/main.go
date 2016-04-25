package stress

import (
	"fmt"
	"log"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/stress/v2/ponyExpress"
	"github.com/influxdata/influxdb/stress/v2/stressql"
)

// RunStress takes a configFile and kicks off the stress test
func RunStress(file string) {

	// Spin up the Client
	s := ponyExpress.NewStoreFront()

	// Parse the file into Statements
	stmts, err := stressql.ParseStatements(file)

	// Log parse errors and quit if found
	if err != nil {
		log.Fatalf("Parsing Error\n  error: %v\n", err)
	}

	// Run all statements
	for _, stmt := range stmts {
		stmt.Run(s)
	}

	// Clear out the batch of unsent response points
	resp := blankResponse()
	s.ResultsChan <- resp
	resp.Tracer.Wait()
	// close s.ResultsChan

	// Compile all Reports
	for _, stmt := range stmts {
		fmt.Println(stmt.Report(s))
	}
}

func blankResponse() ponyExpress.Response {
	// Points must have at least one field
	fields := map[string]interface{}{"done": true}
	// Make a 'blank' point
	p, err := influx.NewPoint("done", make(map[string]string), fields, time.Now())
	// Panic on error
	if err != nil {
		log.Fatalf("Error creating blank response point\n  error: %v\n", err)
	}
	// Add a tracer to prevent program from returning too early
	tracer := ponyExpress.NewTracer(make(map[string]string))
	// Add to the WaitGroup
	tracer.Add(1)
	// Make a new response with the point and the tracer
	resp := ponyExpress.NewResponse(p, tracer)
	return resp
}
