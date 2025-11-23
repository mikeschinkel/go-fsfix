// Package fsfix provides testing utilities including file fixtures and mock configurations.
// It supports creating temporary directories, files, and project structures for testing.
package fsfix

import (
	"os"
	"testing"

	"github.com/mikeschinkel/go-dt"
)

// _ is a compile-time check to ensure RootFixture implements the Fixture interface.
var _ Fixture = (*RootFixture)(nil)

// RootFixture manages temporary directories and files for testing purposes.
type RootFixture struct {
	DirPrefix     string         // Prefix for temporary directory names
	tempDir       dt.DirPath     // Path to the temporary directory
	FileFixtures  []*FileFixture // File-level fixtures in the root temp directory
	ChildFixtures []Fixture      // Project-level fixtures (directories with .git)
	cleanupFunc   func()         // Function to clean up resources
	created       bool
	t             *testing.T
}

func (rf *RootFixture) RelativePath() dt.DirPath {
	return "."
}

// ensureCreated forces a failure if called before Create() is called.
func (rf *RootFixture) ensureCreated() {
	if !rf.created {
		rf.t.Fatalf("RootFixture '%s' has not yet been created", rf.DirPrefix)
	}
}

// Dir returns the path to the temporary directory for this root fixture.
func (rf *RootFixture) Dir() dt.DirPath {
	rf.ensureCreated()
	return rf.tempDir
}

// CreateWithParent is not applicable for RootFixture as it is the root of the fixture hierarchy.
func (rf *RootFixture) createWithParent(*testing.T, Fixture) {
	panic("createWithParent is not relevant as RootFixture should be the root")
}

// Create creates the temporary directory and initializes all child fixtures and files.
func (rf *RootFixture) Create(t *testing.T) {
	t.Helper()
	rf.created = true

	// Create temp directory (this can fail, so it belongs in Create)
	var err error
	rf.tempDir, err = dt.MkdirTemp("", rf.DirPrefix+"-*")
	if err != nil {
		t.Errorf("Failed to create temp directory using '%s'; %v", rf.DirPrefix+"-*", err)
	}

	rf.cleanupFunc = func() {
		err := rf.tempDir.RemoveAll()
		if err != nil {
			t.Errorf("Failed to remove temp directory '%s'; %v", rf.tempDir, err)
		}
	}

	// Set up all the project fixtures
	// rf.RemoveFiles(t) // BUG: This removes the directory we just created
	for _, cf := range rf.ChildFixtures {
		cf.createWithParent(t, rf)
	}

	// Set up all the test fixture files (directly in temp directory)
	for _, ff := range rf.FileFixtures {
		ff.Create(t, rf)
	}

}

// NewRootFixture creates a new TestFixture with the specified directory prefix.
func NewRootFixture(dirPrefix string) *RootFixture {
	return &RootFixture{
		DirPrefix:     dirPrefix,
		FileFixtures:  []*FileFixture{},
		ChildFixtures: []Fixture{},
	}
}

// AddRepoFixture adds a project-level fixture (directory with .git) to the TestFixture.
func (rf *RootFixture) AddRepoFixture(t *testing.T, name dt.PathSegments, args *RepoFixtureArgs) *RepoFixture {
	pf := newRepoFixture(t, name, rf, args)
	rf.ChildFixtures = append(rf.ChildFixtures, pf)
	return pf
}

// AddDirFixture adds a directory fixture (directory with optional .git) to the TestFixture.
func (rf *RootFixture) AddDirFixture(t *testing.T, name dt.PathSegments, args *DirFixtureArgs) *DirFixture {
	df := newDirFixture(t, name, rf, args)
	df.Parent = rf
	rf.ChildFixtures = append(rf.ChildFixtures, df)
	return df
}

// AddFileFixture adds a file fixture directly to the TestFixture temp directory
func (rf *RootFixture) AddFileFixture(t *testing.T, name dt.RelFilepath, args *FileFixtureArgs) *FileFixture {
	ff := newFileFixture(t, name, rf, args)
	rf.FileFixtures = append(rf.FileFixtures, ff)
	return ff
}

// TempDir returns the path to the temporary directory created for this fixture.
func (rf *RootFixture) TempDir() dt.DirPath {
	rf.ensureCreated()
	return rf.tempDir
}

// Cleanup removes all temporary files and directories created by this fixture.
func (rf *RootFixture) Cleanup() {
	rf.ensureCreated()
	rf.cleanupFunc()
}

// RemoveFiles safely removes the temporary directory and all its contents.
func (rf *RootFixture) RemoveFiles(t *testing.T) {
	var err error
	var tempDir, rootDir, tmpRoot dt.DirPath
	var rel dt.PathSegments

	t.Helper()
	rf.ensureCreated()

	if rf.tempDir == "" {
		goto end
	}

	// Normalize to a clean path using DirPath helpers.
	tempDir = rf.tempDir.Clean()

	// Require an absolute path; never delete relative dirs.
	if !tempDir.IsAbs() {
		goto end
	}

	// Compute filesystem rootDir for this path.
	// On Unix: VolumeName() == "", so rootDir is "/".
	// On Windows: VolumeName() == "C:", so rootDir is "C:\".
	rootDir = dt.DirPath(tempDir.VolumeName() + dt.VolumeName(os.PathSeparator))

	if tempDir == rootDir {
		// Never delete a filesystem rootDir.
		goto end
	}

	// Work out the OS temp directory as a DirPath.
	tmpRoot = dt.TempDir().Clean()

	// Never delete the entire temp rootDir either.
	if tempDir == tmpRoot {
		goto end
	}

	// Ensure tempDir is *inside* tmpRoot (not a sibling/parent).
	// Adjust the type of rel / conversion if your Rel() returns a DirPath
	// or some other type.
	rel, err = tempDir.Rel(tmpRoot)
	if err != nil {
		// If we can't reason about the relationship, do nothing.
		goto end
	}

	if rel == "." {
		// tempDir == tmpRoot (already guarded above, but belt-and-suspenders).
		goto end
	}

	if rel.HasDotDotPrefix() {
		// tempDir is outside tmpRoot; refuse to delete.
		goto end
	}

	// At this point tempDir is:
	// - absolute,
	// - not a filesystem rootDir,
	// - not the temp rootDir,
	// - and is located *under* the temp rootDir.
	// It's safe to remove.
	err = rf.tempDir.RemoveAll()
	if err != nil {
		t.Fatalf("failed to remove temporary files %q: %v", tempDir, err)
	}

end:
}
