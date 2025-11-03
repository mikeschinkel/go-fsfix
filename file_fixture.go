// Package fsfix provides testing utilities for creating and managing test fixtures.
// It supports creating temporary file systems, directories, and Git repositories for testing.
package fsfix

import (
	"os"
	"testing"
	"time"

	"github.com/mikeschinkel/go-dt"
)

// FileFixture represents a file fixture that can be created in test environments.
type FileFixture struct {
	Filepath       dt.Filepath
	Name           dt.RelFilepath
	Content        string
	ContentFunc    ContentFunc
	Permissions    int
	DirPermissions int
	ModifiedTime   time.Time
	DoNotCreate    bool
	Parent         Fixture
	created        bool
	t              *testing.T
}

type ContentFunc func(ff *FileFixture) string

// FileFixtureArgs contains arguments for creating a FileFixture.
type FileFixtureArgs struct {
	Name           dt.RelFilepath
	Content        string
	ContentFunc    ContentFunc
	ModifiedTime   time.Time
	Permissions    int
	DirPermissions int
	DoNotCreate    bool
	Parent         Fixture
}

// newFileFixture creates a new file fixture with the specified name and arguments.
func newFileFixture(t *testing.T, name dt.RelFilepath, args *FileFixtureArgs) *FileFixture {
	if args == nil {
		args = &FileFixtureArgs{}
	}
	if args.Permissions == 0 {
		args.Permissions = 0644
	}
	if args.DirPermissions == 0 {
		args.DirPermissions = 0755
	}
	return &FileFixture{
		Name:           name,
		Content:        args.Content,
		ContentFunc:    args.ContentFunc,
		Permissions:    args.Permissions,
		DirPermissions: args.DirPermissions,
		ModifiedTime:   args.ModifiedTime,
		DoNotCreate:    args.DoNotCreate,
		Parent:         args.Parent,
		t:              t,
	}
}

func (ff *FileFixture) RelativePath() dt.Filepath {
	return dt.FilepathJoin(ff.Parent.RelativePath(), ff.Name)
}

// ensureCreated forces a failure if called before Create() is called.
func (ff *FileFixture) ensureCreated() {
	ff.t.Helper()
	if !ff.created {
		ff.t.Fatalf("FileFixture '%s' has not yet been created", ff.Name)
	}
}

// Create creates the file within the specified parent fixture's directory.
func (ff *FileFixture) Create(t *testing.T, pf Fixture) {
	t.Helper()
	ff.created = true
	ff.Parent = pf
	ff.Filepath = dt.FilepathJoin(pf.Dir(), ff.Name)
	ff.createFile(t)
}

// createFile handles the common file creation logic
func (ff *FileFixture) createFile(t *testing.T) {
	var err error
	t.Helper()
	// Skip file creation if it's marked as DoNotCreate
	if ff.DoNotCreate {
		goto end
	}

	if ff.Permissions == 0 {
		t.Errorf("File permissions not set for %s", ff.Filepath)
	}

	err = ff.Filepath.Dir().MkdirAll(os.FileMode(ff.DirPermissions))
	if err != nil {
		t.Errorf("Failed to create test file directory %s", ff.Filepath.Dir())
	}

	if ff.ContentFunc != nil {
		ff.Content = ff.ContentFunc(ff)
	}

	err = dt.WriteFile(ff.Filepath, []byte(ff.Content), os.FileMode(ff.Permissions))
	if err != nil {
		t.Errorf("Failed to create test file %s", ff.Filepath)
	}

	// Set modification time if specified
	if !ff.ModifiedTime.IsZero() {
		err = dt.ChangeFileTimes(ff.Filepath, ff.ModifiedTime, ff.ModifiedTime)
		if err != nil {
			t.Errorf("Failed to set modification time for %s", ff.Filepath)
		}
	}
end:
}
