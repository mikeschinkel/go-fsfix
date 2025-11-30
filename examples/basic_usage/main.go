package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/mikeschinkel/go-fsfix"
)

// Example demonstrating basic usage of go-fsfix for creating test file structures
func main() {
	fmt.Println("go-fsfix Basic Usage Example")
	fmt.Printf("=============================%s", "\n\n")

	// Create a root fixture with a descriptive prefix
	tf := fsfix.NewRootFixture("example-project")
	defer tf.Cleanup()

	t := &testing.T{}

	// Add a simple file directly to the root
	configFile := tf.AddFileFixture(t, "config.json", &fsfix.FileFixtureArgs{
		Content: `{"app": "example", "version": "1.0"}`,
	})

	// Add a repository-like structure
	repoFixture := tf.AddRepoFixture(t, "myapp", nil)

	// Add a directory within the repo
	srcDir := repoFixture.AddDirFixture(t, "src", nil)

	// Add files to the src directory
	mainFile := srcDir.AddFileFixture(t, "main.go", &fsfix.FileFixtureArgs{
		Content: `package main

func main() {
	println("Hello from fsfix example!")
}`,
	})

	utilFile := srcDir.AddFileFixture(t, "util.go", &fsfix.FileFixtureArgs{
		Content: `package main

func helper() string {
	return "utility function"
}`,
	})

	// Add a test directory
	testDir := repoFixture.AddDirFixture(t, "test", nil)
	testDir.AddFileFixture(t, "main_test.go", &fsfix.FileFixtureArgs{
		Content: `package main

import "testing"

func TestExample(t *testing.T) {
	t.Log("Example test")
}`,
	})

	// Create all the fixtures (actually write files to disk)
	tf.Create(t)

	// Display the created structure
	fmt.Println("Created temporary test structure:")
	fmt.Printf("  Root directory: %s\n\n", tf.Dir())
	fmt.Printf("  Files created:\n")
	fmt.Printf("    - %s\n", configFile.Filepath)
	fmt.Printf("    - %s\n", mainFile.Filepath)
	fmt.Printf("    - %s\n", utilFile.Filepath)

	// Verify files exist
	fmt.Println("\n  Verifying files exist:")
	for _, path := range []string{
		string(configFile.Filepath),
		string(mainFile.Filepath),
		string(utilFile.Filepath),
	} {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("    ✓ %s\n", path)
		} else {
			fmt.Printf("    ✗ %s (error: %v)\n", path, err)
		}
	}

	// Read and display one of the files
	content, err := os.ReadFile(string(mainFile.Filepath))
	if err != nil {
		fmt.Printf("\nError reading file: %v\n", err)
		return
	}

	fmt.Printf("\n  Content of main.go:\n")
	fmt.Printf("  %s\n", string(content))

	fmt.Println("\nNote: All files will be automatically cleaned up when the program exits")
	fmt.Println("      due to 'defer tf.Cleanup()' at the start of main()")
}
