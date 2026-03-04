package context_examples

import (
	"context"
	"errors"
	"fmt"
	"time"

	streams "github.com/coldsmirk/go-streams"
)

// Example_forEachCtxWithError demonstrates ForEachCtx with error handling.
func Example_forEachCtxWithError() {
	ctx := context.Background()

	err := streams.ForEachCtx(ctx, streams.Of(1, 2, 3, 4, 5), func(v int) error {
		if v == 3 {
			return errors.New("error at 3")
		}
		fmt.Println(v)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// 1
	// 2
	// Error: error at 3
}

// Example_forEachCtxWithCancellation demonstrates context cancellation.
func Example_forEachCtxWithCancellation() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	s := streams.Generate(func() int {
		time.Sleep(20 * time.Millisecond)
		return 1
	})

	var count int
	err := streams.ForEachCtx(ctx, s, func(v int) error {
		count++
		fmt.Printf("Processed: %d\n", count)
		return nil
	})

	if err != nil {
		fmt.Printf("Context error: %v\n", errors.Is(err, context.DeadlineExceeded))
	}

	// Output:
	// Processed: 1
	// Processed: 2
	// Context error: true
}

// Example_reduceCtxWithError demonstrates ReduceCtx with error handling.
func Example_reduceCtxWithError() {
	ctx := context.Background()

	result, err := streams.ReduceCtx(
		ctx,
		streams.Of(10, 5, 2, 0, 1),
		100,
		func(a, b int) (int, error) {
			if b == 0 {
				return a, errors.New("division by zero")
			}
			return a / b, nil
		},
	)

	fmt.Printf("Result: %d\n", result)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Result: 1
	// Error: division by zero
}

// Example_reduceCtxSuccess demonstrates successful reduction with error-aware reducer.
func Example_reduceCtxSuccess() {
	ctx := context.Background()

	result, err := streams.ReduceCtx(
		ctx,
		streams.Of(1, 2, 3, 4, 5),
		0,
		func(a, b int) (int, error) {
			return a + b, nil
		},
	)

	fmt.Printf("Sum: %d, Error: %v\n", result, err)

	// Output:
	// Sum: 15, Error: <nil>
}

// Example_practicalDataProcessing demonstrates a practical data processing pipeline.
func Example_practicalDataProcessing() {
	type Record struct {
		ID   int
		Name string
	}

	records := []Record{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: ""},  // Invalid: empty name
		{ID: 4, Name: "Dave"},
	}

	// Validate and process records
	err := streams.FromSlice(records).
		ForEachErr(func(r Record) error {
			if r.Name == "" {
				return fmt.Errorf("record %d has empty name", r.ID)
			}
			fmt.Printf("Processed: %s\n", r.Name)
			return nil
		})

	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}

	// Output:
	// Processed: Alice
	// Processed: Bob
	// Validation error: record 3 has empty name
}

// Example_complexPipelineWithContext demonstrates a complex pipeline with error handling.
func Example_complexPipelineWithContext() {
	// Simulate processing data with potential errors
	process := func(v int) (int, error) {
		if v == 0 {
			return 0, errors.New("cannot process zero")
		}
		return v * 2, nil
	}

	s := streams.Of(1, 2, 3, 0, 5)

	// Use MapErrTo to transform with error handling
	results := streams.MapErrTo(s, process)

	// Collect results
	values, err := streams.CollectResults(results)

	fmt.Printf("Processed values: %v\n", values)
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
	}

	// Output:
	// Processed values: [2 4 6]
	// Error occurred: cannot process zero
}

// Example_mixedContextAndErrorHandling demonstrates using both context and error handling.
func Example_mixedContextAndErrorHandling() {
	ctx := context.Background()

	var processed []int

	err := streams.ForEachCtx(ctx, streams.Of(1, 2, 3, 4, 5), func(v int) error {
		if v == 4 {
			return errors.New("processing failed at 4")
		}
		processed = append(processed, v)
		return nil
	})

	fmt.Printf("Processed: %v\n", processed)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Processed: [1 2 3]
	// Error: processing failed at 4
}

// Example_reduceCtxWithContextCancellation demonstrates ReduceCtx with context cancellation.
func Example_reduceCtxWithContextCancellation() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	s := streams.Generate(func() int {
		time.Sleep(20 * time.Millisecond)
		return 1
	})

	result, err := streams.ReduceCtx(ctx, s, 0, func(a, b int) (int, error) {
		return a + b, nil
	})

	fmt.Printf("Partial sum: %d\n", result)
	if err != nil {
		fmt.Printf("Timeout: %v\n", errors.Is(err, context.DeadlineExceeded))
	}

	// Output:
	// Partial sum: 2
	// Timeout: true
}
