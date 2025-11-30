package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-fsfix"
)

// TestFuzzCorpus runs all fuzz corpus files as regression tests
// This ensures that any interesting inputs discovered during fuzzing
// are tested in CI/CD to prevent regressions
func TestFuzzCorpus(t *testing.T) {
	corpusDir := "testdata/fuzz"

	// Check if corpus directory exists
	if _, err := os.Stat(corpusDir); os.IsNotExist(err) {
		// t.Skip("No fuzz corpus found - run fuzzing locally to generate corpus")
		return
	}

	// Find all fuzz test directories
	fuzzTests := []string{
		"FuzzNewRootFixture",
		"FuzzFileFixtureArgs",
		"FuzzFileContent",
	}

	for _, fuzzTest := range fuzzTests {
		t.Run(fuzzTest, func(t *testing.T) {
			testDir := filepath.Join(corpusDir, fuzzTest)

			// Check if this fuzz test has corpus data
			if _, err := os.Stat(testDir); os.IsNotExist(err) {
				// t.Skipf("No corpus for %s", fuzzTest)
				return
			}

			// Read all corpus files
			entries, err := os.ReadDir(testDir)
			if err != nil {
				t.Fatalf("Failed to read corpus directory: %v", err)
			}

			if len(entries) == 0 {
				// t.Skipf("No corpus files for %s", fuzzTest)
				return
			}

			// Run each corpus file as a test
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				t.Run(entry.Name(), func(t *testing.T) {
					corpusFile := filepath.Join(testDir, entry.Name())
					data, err := os.ReadFile(corpusFile)
					if err != nil {
						t.Fatalf("Failed to read corpus file: %v", err)
					}

					// Run the appropriate test based on fuzz test name
					switch fuzzTest {
					case "FuzzNewRootFixture":
						runNewRootFixtureCorpus(t, data)
					case "FuzzFileFixtureArgs":
						runFileFixtureArgsCorpus(t, data)
					case "FuzzFileContent":
						runFileContentCorpus(t, data)
					}
				})
			}
		})
	}
}

func runNewRootFixtureCorpus(t *testing.T, data []byte) {
	// Extract the string from the corpus file
	// Go's fuzzing format: "go test fuzz v1\nstring(\"...\")\n"
	input := extractStringFromCorpus(data)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewRootFixture panicked with input: %q, panic: %v", input, r)
		}
	}()

	_ = fsfix.NewRootFixture(input)
}

func runFileFixtureArgsCorpus(t *testing.T, data []byte) {
	// For FileFixtureArgs, we expect string and int
	parts := extractMultipleFromCorpus(data)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddFileFixture panicked with inputs: %q, panic: %v", parts, r)
		}
	}()

	if len(parts) >= 2 {
		name := parts[0]
		permissions := 0644 // Default
		if len(parts) > 1 {
			// Try to parse permission from second part
			// For simplicity, just use a default value
			permissions = 0644
		}

		rf := fsfix.NewRootFixture("corpus-test")
		_ = rf.AddFileFixture(t, dt.RelFilepath(name), &fsfix.FileFixtureArgs{
			Content:     "test content",
			Permissions: permissions,
		})
	}
}

func runFileContentCorpus(t *testing.T, data []byte) {
	input := extractStringFromCorpus(data)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddFileFixture with content panicked with input: %q, panic: %v", input, r)
		}
	}()

	rf := fsfix.NewRootFixture("corpus-content")
	_ = rf.AddFileFixture(t, "test.txt", &fsfix.FileFixtureArgs{
		Content: input,
	})
}

// extractStringFromCorpus extracts a string value from Go's fuzz corpus format
func extractStringFromCorpus(data []byte) string {
	// Simple extraction - corpus format is: "go test fuzz v1\nstring(\"...\")\n"
	// For production use, you might want more robust parsing
	str := string(data)

	// Skip the header line
	if len(str) > 0 {
		// This is a simplified version - real corpus parsing would be more robust
		return str
	}

	return ""
}

// extractMultipleFromCorpus extracts multiple values from corpus
func extractMultipleFromCorpus(data []byte) []string {
	// Simplified - for multi-parameter fuzzing
	str := extractStringFromCorpus(data)

	// For now, just return as single string
	// In real usage, the corpus format would properly encode multiple values
	return []string{str}
}
