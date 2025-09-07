# fsfix Package

The `fsfix` package provides a comprehensive test fixture system for creating temporary file structures, directories, and Git repositories for testing purposes. It enables clean, isolated testing environments with automatic cleanup and flexible content management.

## Overview

This package implements a hierarchical fixture system that allows tests to create complex directory structures, files with specific content, and even simulate Git repositories. All fixtures are automatically cleaned up after test completion, ensuring no test pollution or resource leaks.

## Key Components

### Core Fixtures
- **RootFixture**: Top-level fixture managing temporary directories
- **RepoFixture**: Git repository simulation with project structure
- **DirFixture**: Directory creation and management
- **FileFixture**: Individual file creation with content control

### Interfaces
- **Fixture Interfaces**: Common contracts for all fixture types
- **Cleanup Management**: Automatic resource cleanup and error handling
- **Content Control**: Flexible content specification and templates

### Utilities
- **Fixture Utilities**: Helper functions for common test scenarios
- **Path Management**: Safe path handling and validation
- **Content Generation**: Dynamic content creation for tests

## Architecture

### Hierarchical Structure

Fixtures follow a parent-child relationship:
- **Root**: Creates base temporary directory
- **Repo**: Creates project-like structure within root
- **Dir**: Creates subdirectories within repo or other dirs  
- **File**: Creates files within any parent fixture

### Automatic Cleanup

The package provides guaranteed cleanup:
- **Deferred Cleanup**: Automatic cleanup via defer statements
- **Error Handling**: Cleanup even if tests fail
- **Resource Tracking**: Prevents resource leaks and orphaned files

## Usage

### Basic File Structure

```go
func TestSimpleProject(t *testing.T) {
    // Create root fixture
    tf := fsfix.NewRootFixture("my-test")
    // Delete all test files at end of test function
    defer tf.Cleanup()
    
    // Create project structure
    pf := tf.AddRepoFixture(t, "test-project", nil)
    
    // Add files
    ff := pf.AddFileFixture(t, "main.go", &fsfix.FileFixtureArgs{
    Content: `package main

func main() {
    println("Hello, World!")
}`,
    })
    
    // Creates all fixtures
    tf.Create(t)
    
    // Use tf.Dir() to get root fixture path
    // Use pf.Dir() to get test-project path
    // Use ff.Filepath to get main.go path
}
```

### Complex Project Structure

```go
func TestComplexProject(t *testing.T) {
  	tf := fsfix.NewRootFixture("my-test")
    // Delete all test files at end of test function
    defer tf.Cleanup()

    // Create test data file in root
    tjf := tf.AddFileFixture(t, "test.json", &fsfix.FileFixtureArgs{
      Content: `{"test": true}`,
    })
    
    // Create nested directory structure
    df := tf.AddRepoFixture(t, "internal", nil)
    
    df2 := df.AddDirFixture(t, "widgets", nil)
    df2.AddFileFixture(t, "my-widget.go", &fsfix.FileFixtureArgs{
      ContentFunc: myWidgetContentFunc, // a function defined elsewhere returning content
    })
    
    // Create a file that should be missing for error testing
    mwf := df.AddFileFixture(t, "missing.go", &fsfix.FileFixtureArgs{
      DoNotCreate: true,
    })
    
    // Creates all fixtures
    tf.Create(t)

    // Use pf.Dir() to get project path for testing
    // Use df.Dir() to get path for internal directory 
    // Use df2.Dir() to get path for internal/widgets directory 
    // Use mwf.Filepath to get filepath of missing.go file
}
```

### Git Repository Simulation

```go
func TestRepoProject(t *testing.T) {
    tf := fsfix.NewRootFixture("my-test")
    defer tf.Cleanup()

	// Create repo-like structure
	rf := tf.AddRepoFixture(t, "my-repo", nil)

	// Creates all fixtures
	tf.Create(t)
	// Delete all test files at end of test function
	defer tf.Cleanup()

	// Use rf.GitPath() to get the .git path
}
```
## Fixture Types

### RootFixture
- Creates base temporary directory using system temp
- Manages cleanup of entire fixture hierarchy
- Provides unique directory names to prevent conflicts

### RepoFixture
- Simulates project/repository structure
- Can optionally initialize actual Git repository
- Provides project-level organization for test files

### DirFixture
- Creates subdirectories within parent fixtures
- Supports nested directory structures
- Maintains parent-child relationships

### FileFixture
- Creates individual files with specified content
- Supports missing files for error condition testing
- Provides file path access for test operations

## Content Generation

### Dynamic Content
```go
func myContentFunc(fileNo int) fsfix.ContentFunc {
    return func(ff *fsfix.FileFixture) string {
        return fmt.Sprintf("Text File #%d\n", fileNo)
    }
}

func TestDynamicContent(t *testing.T) {
    tf := fsfix.NewRootFixture("my-test")
    defer tf.Cleanup()

	// Create repo-like structure
	df := tf.AddDirFixture(t, "my-repo", nil)

	ffs := make([]*fsfix.FileFixture, 3)
	for i := range 3 {
		file := fmt.Sprintf("file-%d.txt", i+1)
		// Add typical project files
		ffs[i] = df.AddFileFixture(t, file, &fsfix.FileFixtureArgs{
			ContentFunc: myContentFunc(i + 1),
		})
	}

	// Creates all fixtures
	tf.Create(t)
	// Delete all test files at end of test function
	defer tf.Cleanup()

	// Use ffs[<n>].Filepath to get File #<n>+1 
}
```
### Missing Files
```go
// Create file path but don't create actual file
ff := pf.AddFileFixture("nonexistent.go", &fsfix.FileFixtureArgs{
    DoNotCreate: true,
})

// Use ff.Filepath for filepath of file that should not exist
```

## Testing Integration

### Test Lifecycle
1. **Create**: `NewRootFixture()` creates base fixture
2. **Build**: `AddFileFixture()` and similar methods build structure
3. **Create**: `tf.Create(t)` creates all files and directories  
4. **Test**: Use fixture paths in test operations
5. **Cleanup**: `defer tf.Cleanup()` removes all created resources

### Error Handling

The package provides comprehensive error handling:
- Create errors are reported through the testing framework
- Cleanup errors are logged but don't fail tests
- Path validation prevents unsafe operations

### Isolation

Each test gets its own isolated fixture:
- Unique temporary directories
- No cross-test contamination
- Safe parallel test execution

## Best Practices

### Naming Convention
Use descriptive prefixes for fixture identification:
```go
tf := fsfix.NewRootFixture("read-files-tool-test")
```

### Resource Management
Always use defer for cleanup:
```go
tf := fsfix.NewRootFixture("my-test")
defer tf.Cleanup() // Ensure cleanup even if test panics
```

## Dependencies

No dependencies beyond the Go Standard library

## License

MIT