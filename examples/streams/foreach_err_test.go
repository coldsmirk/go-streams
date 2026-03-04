package streams_examples

import (
	"context"
	"errors"
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example_forEachErr demonstrates using ForEachErr for operations that may fail.
func Example_forEachErr() {
	// Simulate a transformer that may fail
	type Model struct {
		ID   int
		Name string
	}

	models := []Model{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	// Transform each model with error handling
	err := streams.FromSlice(models).ForEachErr(func(m Model) error {
		// Simulate transformation that may fail
		if m.ID == 2 {
			return errors.New("transformation failed for ID 2")
		}
		fmt.Printf("Processed: %s\n", m.Name)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Processed: Alice
	// Error: transformation failed for ID 2
}

// Example_forEachErrWithPointer demonstrates using ForEachErr with pointer modifications.
func Example_forEachErrWithPointer() {
	type Model struct {
		ID        int
		Name      string
		Processed bool
	}

	models := []Model{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	// Use ForEachIndexedErr to modify original slice elements
	err := streams.Range(0, len(models)).ForEachIndexedErr(func(i, _ int) error {
		// Access original slice element by pointer
		model := &models[i]

		// Simulate a transformation that may fail
		if model.ID == 2 {
			return fmt.Errorf("failed to process model %d", model.ID)
		}

		model.Processed = true
		fmt.Printf("Processed: %s (ID: %d)\n", model.Name, model.ID)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Processed: Alice (ID: 1)
	// Error: failed to process model 2
}

// Example_forEachErrWithContext demonstrates using ForEachErr with context.
func Example_forEachErrWithContext() {
	type Model struct {
		ID   int
		Name string
	}

	models := []Model{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	ctx := context.Background()

	// Simulate a transformer that uses context
	transformer := func(ctx context.Context, m *Model) error {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return err
		}

		// Simulate transformation
		m.Name = "Processed-" + m.Name
		return nil
	}

	// Transform each model
	err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
		return transformer(ctx, &models[i])
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	for _, m := range models {
		fmt.Printf("%s\n", m.Name)
	}

	// Output:
	// Processed-Alice
	// Processed-Bob
	// Processed-Charlie
}

// Example_forEachIndexedErr demonstrates using ForEachIndexedErr.
func Example_forEachIndexedErr() {
	words := []string{"hello", "world", "go", "streams"}

	err := streams.FromSlice(words).ForEachIndexedErr(func(i int, word string) error {
		if len(word) < 3 {
			return fmt.Errorf("word at index %d is too short: %s", i, word)
		}
		fmt.Printf("[%d] %s\n", i, word)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// [0] hello
	// [1] world
	// Error: word at index 2 is too short: go
}
