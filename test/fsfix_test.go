package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-fsfix"
)

func TestSimpleProject(t *testing.T) {
	// Create root fixture
	tf := fsfix.NewRootFixture("my-test")
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

	if tf.DirPrefix != "my-test" {
		t.Errorf("RootFixture.DirPrefix does not equal 'my-test'")
	}

	if !dirExists(t, tf.Dir()) {
		t.Errorf("RootFixture.Dir() doesn't exist: %s", tf.Dir())
	}

	if !dirExists(t, pf.Dir()) {
		t.Errorf("RepoFixture.Dir() doesn't exist: %s", tf.Dir())
	}

	if !dirExists(t, pf.Dir()+"/.git") {
		t.Errorf(".git directory within RepoFixture.Dir() doesn't exist: %s", tf.Dir())
	}

	if ff.Filepath == "" {
		t.Errorf("FileFixture.Filepath not set")
	}

	gotFP := ff.Filepath.Base()
	wantFP := dt.Filename("main.go")
	if gotFP != wantFP {
		t.Errorf("FileFixture.Filepath not set to '%s'; got '%s' instead", wantFP, gotFP)
	}

	gotDP := ff.Filepath.Dir().Base()
	wantDP := dt.PathSegment("test-project")
	if gotDP != wantDP {
		t.Errorf("FileFixture.Filepath's parent dir not set to '%s'; got '%s' instead", wantDP, gotDP)
	}

	if !fileExists(t, ff.Filepath) {
		t.Errorf("FileFixture.Filepath doesn't exist: %s", ff.Filepath)
	}

	// Use tf.Dir() to get root fixture path
	// Use pf.Dir() to get test-project path
	// Use ff.Filepath to get main.go path
}

func TestRepoProject(t *testing.T) {
	tf := fsfix.NewRootFixture("my-test")
	defer tf.Cleanup()

	// Create repo-like structure
	rf := tf.AddRepoFixture(t, "my-repo", nil)

	// Creates all fixtures
	tf.Create(t)
	// Delete all test files at end of test function
	defer tf.Cleanup()

	wantDP := dt.DirPath("my-repo/.git")
	gotDP := rf.RelativeGitPath()
	if wantDP != gotDP {
		t.Errorf("FileFixture.Filepath doesn't contain '%s'; got '%s' instead", gotDP, wantDP)
	}
	// Use rf.GitPath() to get the .git path
}

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
		file := dt.RelFilepath(fmt.Sprintf("file-%d.txt", i+1))
		// Add typical project files
		ffs[i] = df.AddFileFixture(t, file, &fsfix.FileFixtureArgs{
			ContentFunc: myContentFunc(i + 1),
		})
	}

	// Creates all fixtures
	tf.Create(t)
	// Delete all test files at end of test function
	defer tf.Cleanup()

	want := "Text File #2\n"
	gotBB, _ := dt.ReadFile(ffs[1].Filepath)
	if want != string(gotBB) {
		t.Errorf("FileFixture.Filepath doesn't contain '%s'; got '%s' instead", string(gotBB), want)
	}
	// Use ffs[<n>].Filepath to get File #<n>+1
}

func TestComplexProject(t *testing.T) {
	// Create root fixture
	tf := fsfix.NewRootFixture("my-test")
	defer tf.Cleanup()

	// Create test data file in root
	tjf := tf.AddFileFixture(t, "test.json", &fsfix.FileFixtureArgs{
		Content: `{"test": true}`,
	})

	// Create nested directory structure
	df := tf.AddDirFixture(t, "internal", nil)

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
	// Delete all test files at end of test function
	defer tf.Cleanup()

	want := `{"test": true}`
	gotBB, _ := dt.ReadFile(tjf.Filepath)
	if want != string(gotBB) {
		t.Errorf("FileFixture.Fileoath doesn't contain '%s'; got '%s' instead", string(gotBB), want)
	}

	wantDP := dt.DirPath("internal/widgets")
	gotDP := df2.RelativePath()
	if wantDP != gotDP {
		t.Errorf("DirFixture.RelativePath() doesn't contain '%s'; got '%s' instead", gotDP, wantDP)
	}

	if fileExists(t, mwf.Filepath) {
		t.Errorf("File %s was created but should not have been", mwf.Filepath)
	}

	// Use pf.Dir() to get project path for testing
	// Use df.Dir() to get path for internal directory
	// Use df2.Dir() to get path for internal/widgets directory
	// Use mwf.Filepath to get filepath of missing.go file
}

func myWidgetContentFunc(_ *fsfix.FileFixture) string {
	return `package main

type MyWidget struct{
	Name string 
}
`
}

func dirExists(t *testing.T, dp dt.DirPath) bool {
	t.Helper()
	info, err := dt.StatDir(dp)
	return !os.IsNotExist(err) && info.IsDir()
}

func fileExists(t *testing.T, path dt.Filepath) bool {
	t.Helper()
	info, err := dt.StatFile(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
