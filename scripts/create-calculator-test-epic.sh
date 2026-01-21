#!/bin/bash
# Creates a test epic with 4 calculator tasks for testing tickboard features
# Usage: ./scripts/create-calculator-test-epic.sh

set -e

TK="./tk"

echo "Creating calculator test epic..."

# Create the epic
EPIC_ID=$($TK create "Calculator Test Epic" \
  -d "Test epic with trivial calculator tasks for validating tickboard UI features." \
  -t epic)
echo "Created epic: $EPIC_ID"

# Task 1: Create calculator with Add function
TASK1=$($TK create "Create calculator package with Add function" \
  -d "Create internal/calculator/calc.go with package calculator and an Add(a, b int) int function that returns the sum.

Requirements:
- Package name: calculator
- Function: Add(a, b int) int
- Should include a package-level doc comment

Run: go build ./internal/calculator/..." \
  -t task --parent "$EPIC_ID" --acceptance "File exists, builds without errors")
echo "Created task 1: $TASK1"

# Task 2: Add Subtract and Multiply
TASK2=$($TK create "Add Subtract and Multiply functions" \
  -d "Extend internal/calculator/calc.go with:
- Subtract(a, b int) int - returns a minus b
- Multiply(a, b int) int - returns a times b

Each function should have a doc comment.

Run: go build ./internal/calculator/..." \
  -t task --parent "$EPIC_ID" --blocked-by "$TASK1" --acceptance "Functions exist and build")
echo "Created task 2: $TASK2"

# Task 3: Add unit tests
TASK3=$($TK create "Add unit tests for calculator" \
  -d "Create internal/calculator/calc_test.go with tests for all three functions (Add, Subtract, Multiply).

Test cases to include:
- Add: 2+3=5, 0+0=0, -1+1=0
- Subtract: 5-3=2, 0-0=0, 1-5=-4
- Multiply: 2*3=6, 0*5=0, -2*3=-6

Use table-driven tests.

Run: go test ./internal/calculator/..." \
  -t task --parent "$EPIC_ID" --blocked-by "$TASK2" --acceptance "All tests pass")
echo "Created task 3: $TASK3"

# Task 4: Clean up
TASK4=$($TK create "Clean up calculator test package" \
  -d "Remove the internal/calculator directory that was created for testing.

Verify no calculator directory remains.

Run: ls internal/ | grep -v calculator" \
  -t task --parent "$EPIC_ID" --blocked-by "$TASK3" --acceptance "Calculator directory removed")
echo "Created task 4: $TASK4"

echo ""
echo "Test epic created: $EPIC_ID"
echo "Run with: ./tk run $EPIC_ID"
