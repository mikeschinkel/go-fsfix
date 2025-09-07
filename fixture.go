// Package fsfix provides testing utilities for creating and managing test fixtures.
// It supports creating temporary file systems, directories, and Git repositories for testing.
package fsfix

import (
	"testing"
)

// Fixture defines the interface for all test fixtures that can create directory structures.
type Fixture interface {
	Dir() string
	RelativePath() string
	createWithParent(*testing.T, Fixture)
}
