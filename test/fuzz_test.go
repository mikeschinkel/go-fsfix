package test

import (
	"testing"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-fsfix"
)

// FuzzNewRootFixture tests NewRootFixture with random directory prefixes to ensure it doesn't panic
func FuzzNewRootFixture(f *testing.F) {
	// Seed corpus with various directory prefix patterns
	seeds := []string{
		"test",
		"my-test",
		"test_fixture",
		"test.fixture",
		"test123",
		"",  // Empty prefix
		"a", // Single character
		"very-long-directory-prefix-name-for-testing",
		"test/nested",   // Path separator in prefix
		"test\\windows", // Windows path separator
		"test\x00null",  // Null byte
		"../parent",     // Path traversal attempt
		"..test",        // Starts with dots
		"test..",        // Ends with dots
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, dirPrefix string) {
		// Just ensure NewRootFixture doesn't panic with any input
		_ = fsfix.NewRootFixture(dirPrefix)
	})
}

// FuzzFileFixtureArgs tests file fixture creation with random file names and permissions
func FuzzFileFixtureArgs(f *testing.F) {
	// Seed corpus with various file name patterns
	type seed struct {
		name        string
		permissions int
	}

	seeds := []seed{
		{"test.txt", 0644},
		{"test.go", 0755},
		{"README.md", 0600},
		{"", 0644},  // Empty filename
		{"a", 0644}, // Single character
		{"file with spaces.txt", 0644},
		{"file-with-dashes.txt", 0644},
		{"file_with_underscores.txt", 0644},
		{"file.multiple.dots.txt", 0644},
		{".hidden", 0644},          // Hidden file
		{"..dotdot", 0644},         // Starts with ..
		{"file/nested.txt", 0644},  // Path separator in name
		{"file\x00null.txt", 0644}, // Null byte in name
		{"test.txt", 0000},         // No permissions
		{"test.txt", 0777},         // All permissions
		{"test.txt", -1},           // Negative permissions
		{"test.txt", 9999},         // Invalid permissions
	}

	for _, s := range seeds {
		f.Add(s.name, s.permissions)
	}

	f.Fuzz(func(t *testing.T, name string, permissions int) {
		// Create a minimal root fixture for testing
		rf := fsfix.NewRootFixture("fuzz-test")

		// Test adding file fixture with various inputs
		// Just ensure it doesn't panic
		_ = rf.AddFileFixture(t, dt.RelFilepath(name), &fsfix.FileFixtureArgs{
			Content:     "test content",
			Permissions: permissions,
		})
	})
}

// FuzzFileContent tests file fixtures with random content
func FuzzFileContent(f *testing.F) {
	// Seed corpus with various content types
	seeds := []string{
		"",                         // Empty content
		"a",                        // Single character
		"Hello, World!",            // Simple ASCII
		"Line 1\nLine 2\n",         // Newlines
		"Tab\tseparated\tvalues",   // Tabs
		"Unicode: 你好世界",            // Unicode
		"Mixed\r\nLine\rEndings\n", // Mixed line endings
		string([]byte{0, 1, 2, 3}), // Binary data
		"Very long content " + string(make([]byte, 1024)), // Long content
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, content string) {
		// Create a minimal test structure
		rf := fsfix.NewRootFixture("fuzz-content")

		// Add file with fuzzed content
		_ = rf.AddFileFixture(t, "test.txt", &fsfix.FileFixtureArgs{
			Content: content,
		})

		// Note: We don't call Create() to avoid actually creating files
		// This tests the fixture setup logic without filesystem operations
	})
}
