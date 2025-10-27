// Package fsfix provides testing utilities for creating and managing test fixtures.
// It supports creating temporary file systems, directories, and Git repositories for testing.
package fsfix

import (
	"os"
	"testing"
	"time"

	"github.com/mikeschinkel/go-dt"
)

// _ is a compile-time check to ensure DirFixture implements the Fixture interface.
var _ Fixture = (*DirFixture)(nil)

// DirFixture represents a dir directory fixture with optional Git repository.
type DirFixture struct {
	Name          dt.PathSegments // Name of the dir directory
	FileFixtures  []*FileFixture  // Files to create within this dir
	ChildFixtures []Fixture       // Subdirectories or Projects to create within this dir
	ModifiedTime  time.Time       // Modification time for the dir directory
	Permissions   int             // Directory permissions (e.g., 0755)
	dir           dt.DirPath      // Full path to the created directory
	Parent        Fixture         // Parent test fixture
	created       bool
	t             *testing.T
}

func (df *DirFixture) RelativePath() dt.DirPath {
	return dt.DirPathJoin(df.Parent.RelativePath(), df.Name)
}

// ensureCreated forces a failure if called before Create() is called.
func (df *DirFixture) ensureCreated() {
	df.t.Helper()
	if !df.created {
		df.t.Fatalf("DirFixture '%s' has not yet been created", df.Name)
	}
}

// Dir returns the full path to the directory fixture.
func (df *DirFixture) Dir() dt.DirPath {
	df.ensureCreated()
	return df.dir
}

// DirFixtureArgs contains arguments for creating a DirFixture.
type DirFixtureArgs struct {
	Files        []*FileFixture // Files to create within this dir
	Permissions  int            // Directory permissions
	ModifiedTime time.Time      // Modification time for the directory
	Parent       Fixture        // Parent test fixture
}

// newDirFixture creates a new directory fixture with the specified name and arguments.
func newDirFixture(t *testing.T, name dt.PathSegments, args *DirFixtureArgs) *DirFixture {
	if args == nil {
		args = &DirFixtureArgs{}
	}
	if args.Permissions == 0 {
		args.Permissions = 0755
	}
	return &DirFixture{
		Name:         name,
		Parent:       args.Parent,
		FileFixtures: args.Files,
		ModifiedTime: args.ModifiedTime,
		Permissions:  args.Permissions,
		t:            t,
	}
}

// MakeDir creates a path relative to this directory fixture.
func (df *DirFixture) MakeDir(fp string) dt.DirPath {
	df.ensureCreated()
	return dt.DirPathJoin(df.dir, fp)
}

// CreateWithParent creates the directory structure and files for this fixture with the specified parent.
func (df *DirFixture) createWithParent(t *testing.T, pf Fixture) {
	t.Helper()
	df.created = true

	// Create a single dir directory with .git
	df.dir = dt.DirPathJoin(pf.Dir(), df.Name)
	if df.Permissions == 0 {
		t.Errorf("File permissions not set for %s", df.dir)
	}
	err := dt.MkdirAll(df.dir, os.FileMode(df.Permissions))
	if err != nil {
		t.Errorf("Failed to create testing directory %s", df.dir)
	}
	for _, file := range df.FileFixtures {
		file.Create(t, df)
	}
	for _, child := range df.ChildFixtures {
		child.createWithParent(t, df)
	}
}

// AddDirFixture adds a subdirectory fixture to this directory fixture.
func (df *DirFixture) AddDirFixture(t *testing.T, name dt.PathSegments, args *DirFixtureArgs) *DirFixture {
	cf := newDirFixture(t, name, args)
	cf.Parent = df
	df.ChildFixtures = append(df.ChildFixtures, cf)
	return cf
}

// AddRepoFixture adds a repository fixture to this directory fixture.
func (df *DirFixture) AddRepoFixture(t *testing.T, name dt.PathSegments, args *RepoFixtureArgs) *RepoFixture {
	cf := newRepoFixture(t, name, args)
	cf.Parent = df
	df.ChildFixtures = append(df.ChildFixtures, cf)
	return cf
}

// AddFileFixture adds a file fixture to a dir fixture
func (df *DirFixture) AddFileFixture(t *testing.T, name dt.RelFilepath, args *FileFixtureArgs) *FileFixture {
	ff := newFileFixture(t, name, args)
	ff.Parent = df
	df.FileFixtures = append(df.FileFixtures, ff)
	return ff
}

// AddFileFixtures adds multiple files at once using defaults if as
// FileFixtureArgs when one of args is passed just as a string(string) and it
// gets its content from ContentFunc, or a FileFixtureArgs is passed which must
// include Name.
func (df *DirFixture) AddFileFixtures(t *testing.T, defaults *FileFixtureArgs, args ...any) {
	for _, f := range args {
		switch ffa := f.(type) {
		case dt.Filename:
			df.AddFileFixture(t, dt.RelFilepath(ffa), defaults)
		case dt.RelFilepath:
			df.AddFileFixture(t, ffa, defaults)
		case *FileFixtureArgs:
			if ffa.Name == "" {
				t.Fatalf("Name not set for file fixure being added to dir fixture '%s'", df.Name)
			}
			df.AddFileFixture(t, ffa.Name, ffa)
		default:
			t.Fatalf("Invalid type '%T' passed for file fixure being added to dir fixture: '%v'", f, f)
		}
	}
}
