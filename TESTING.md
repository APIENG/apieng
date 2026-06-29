# Testing Guide

## Running Tests

Run all tests:
```bash
go test ./...
```

Run tests for a specific package:
```bash
go test ./internal/db/...
go test ./internal/models/...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests and show coverage:
```bash
go test -cover ./...
```

## Writing Tests

Test files should:
- Be placed in the same package as the code being tested
- End with `_test.go` suffix
- Start each test function with `Test` prefix
- Take `*testing.T` as parameter

### Example Test

```go
func TestMyFunction(t *testing.T) {
    result := MyFunction(input)
    if result != expected {
        t.Fatalf("Expected %v, got %v", expected, result)
    }
}
```

## Current Test Coverage

- `internal/models/` — Password hashing and validation tests
- `internal/db/` — Session storage, validation, and invalidation tests

## TODO: Add More Tests

Future test coverage should include:
- Handler tests (with mock HTTP responses)
- Energy calculation tests
- API measurement tests
- End-to-end integration tests
- Middleware authentication tests
