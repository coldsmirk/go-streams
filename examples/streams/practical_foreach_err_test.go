package streams_examples

import (
	"context"
	"fmt"

	streams "github.com/coldsmirk/go-streams"
)

// Example_practicalForEachErr demonstrates the real-world use case:
// Converting traditional error-handling loops to stream operations.
func Example_practicalForEachErr() {
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

	// Simulate a transformer function
	transform := func(ctx context.Context, m *Model) error {
		if m.ID == 2 {
			return fmt.Errorf("transformation failed for model ID %d", m.ID)
		}
		m.Name = "Transformed-" + m.Name
		return nil
	}

	// Original code (traditional loop):
	/*
		for i := range models {
			if err := transform(ctx, &models[i]); err != nil {
				return err
			}
		}
	*/

	// New stream-based approach:
	err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
		return transform(ctx, &models[i])
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("All models transformed successfully")
	}

	// Check results
	fmt.Printf("Model 0: %s\n", models[0].Name)

	// Output:
	// Error: transformation failed for model ID 2
	// Model 0: Transformed-Alice
}

// Example_practicalForEachErrSuccess demonstrates successful transformation of all elements.
func Example_practicalForEachErrSuccess() {
	type Model struct {
		ID   int
		Name string
	}

	models := []Model{
		{ID: 1, Name: "Alice"},
		{ID: 3, Name: "Charlie"},
		{ID: 4, Name: "Dave"},
	}

	ctx := context.Background()

	// Simulate a transformer function
	transform := func(ctx context.Context, m *Model) error {
		m.Name = "Transformed-" + m.Name
		return nil
	}

	// Transform all models
	err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
		return transform(ctx, &models[i])
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print all transformed models
	for _, m := range models {
		fmt.Printf("%s\n", m.Name)
	}

	// Output:
	// Transformed-Alice
	// Transformed-Charlie
	// Transformed-Dave
}

// Example_forEachErrWithValidation demonstrates validation before processing.
func Example_forEachErrWithValidation() {
	type User struct {
		ID    int
		Email string
		Age   int
	}

	users := []User{
		{ID: 1, Email: "alice@example.com", Age: 25},
		{ID: 2, Email: "invalid-email", Age: 30},
		{ID: 3, Email: "charlie@example.com", Age: 22},
	}

	// Validate and process users
	err := streams.Range(0, len(users)).ForEachErr(func(i int) error {
		user := &users[i]

		// Validation logic
		if user.Email == "" || user.Email == "invalid-email" {
			return fmt.Errorf("invalid email for user %d", user.ID)
		}

		if user.Age < 18 {
			return fmt.Errorf("user %d is underage", user.ID)
		}

		// Processing logic (modify in-place)
		fmt.Printf("Processing user: %s\n", user.Email)
		return nil
	})

	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}

	// Output:
	// Processing user: alice@example.com
	// Validation error: invalid email for user 2
}

// Example_forEachIndexedErrBatchProcessing demonstrates batch processing with index tracking.
func Example_forEachIndexedErrBatchProcessing() {
	items := []string{"item1", "item2", "item3", "item4", "item5"}

	err := streams.FromSlice(items).ForEachIndexedErr(func(idx int, item string) error {
		// Simulate batch processing where we track the index
		if idx == 3 {
			return fmt.Errorf("processing failed at index %d for item %s", idx, item)
		}

		fmt.Printf("[%d] Processed: %s\n", idx, item)
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// [0] Processed: item1
	// [1] Processed: item2
	// [2] Processed: item3
	// Error: processing failed at index 3 for item item4
}
