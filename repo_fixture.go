// Package fsfix provides testing utilities for creating and managing test fixtures.
// It supports creating temporary file systems, directories, and Git repositories for testing.
package fsfix

import (
	"testing"
	"time"

	"github.com/mikeschinkel/go-dt"
)

// _ is a compile-time check to ensure RepoFixture implements the Fixture interface.
var _ Fixture = (*RepoFixture)(nil)

// RepoFixture represents a project directory fixture with optional Git repository.
type RepoFixture struct {
	*DirFixture
	created bool
	t       *testing.T
}

// RepoFixtureArgs contains arguments for creating a RepoFixture.
type RepoFixtureArgs struct {
	Files        []*FileFixture // Files to create within this project
	Permissions  int            // Directory permissions
	ModifiedTime time.Time      // Modification time for the directory
	Parent       Fixture        // Parent test fixture
}

// newRepoFixture creates a new repository fixture with the specified name and arguments.
func newRepoFixture(t *testing.T, name dt.PathSegments, args *RepoFixtureArgs) *RepoFixture {
	if args == nil {
		args = &RepoFixtureArgs{}
	}
	if args.Permissions == 0 {
		args.Permissions = 0755
	}
	return &RepoFixture{
		t: t,
		DirFixture: newDirFixture(t, name, &DirFixtureArgs{
			ModifiedTime: args.ModifiedTime,
			Permissions:  args.Permissions,
			Parent:       args.Parent,
		}),
	}
}

func (rf *RepoFixture) RelativePath() dt.DirPath {
	return dt.DirPathJoin(rf.Parent.RelativePath(), rf.Name)
}

func (rf *RepoFixture) GitPath() dt.DirPath {
	return dt.DirPathJoin(rf.Dir(), ".git")
}

func (rf *RepoFixture) RelativeGitPath() dt.DirPath {
	return dt.DirPathJoin(rf.RelativePath(), ".git")
}

// ensureCreated forces a failure if called before Create() is called.
func (rf *RepoFixture) ensureCreated() {
	rf.t.Helper()
	if !rf.created {
		rf.t.Fatalf("RepoFixture '%s' has not yet been created", rf.Name)
	}
}

// CreateWithParent creates the repository structure and files for this fixture with the specified parent.
func (rf *RepoFixture) createWithParent(t *testing.T, parent Fixture) {
	t.Helper()
	rf.created = true
	rf.DirFixture.createWithParent(t, parent)

	// Create .git directory to simulate making it a valid repo
	// TODO: Maybe we could shell out to `git init` here if anyone ever needs that
	gitDir := dt.DirPathJoin(rf.dir, ".git")
	err := dt.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Errorf("Failed to create .git directory within %s; %v", rf.dir, err)
	}
}

// MakeDir creates a path relative to this repository fixture.
func (rf *RepoFixture) MakeDir(fp string) dt.DirPath {
	rf.ensureCreated()
	return dt.DirPathJoin(rf.dir, fp)
}

// AddRepoFixture adds a sub-repository fixture to this repository fixture.
func (rf *RepoFixture) AddRepoFixture(t *testing.T, name dt.PathSegments, args *RepoFixtureArgs) *RepoFixture {
	child := newRepoFixture(t, name, args)
	child.Parent = rf
	rf.ChildFixtures = append(rf.ChildFixtures, child)
	return child
}

// AddDirFixture adds a directory fixture to this repository fixture.
func (rf *RepoFixture) AddDirFixture(t *testing.T, name dt.PathSegments, args *DirFixtureArgs) *DirFixture {
	child := newDirFixture(t, name, args)
	child.Parent = rf
	rf.ChildFixtures = append(rf.ChildFixtures, child)
	return child
}

// AddFileFixture adds a file fixture to a project fixture
func (rf *RepoFixture) AddFileFixture(t *testing.T, name dt.RelFilepath, args *FileFixtureArgs) *FileFixture {
	child := newFileFixture(t, name, args)
	child.Parent = rf
	rf.FileFixtures = append(rf.FileFixtures, child)
	return child
}

// AddFileFixtures adds multiple files at once using defaults if as
// FileFixtureArgs when one of args is passed just as a string(string) and it
// gets its content from ContentFunc, or a FileFixtureArgs is passed which much
// include Name.
func (rf *RepoFixture) AddFileFixtures(t *testing.T, defaults *FileFixtureArgs, args ...any) {
	for _, f := range args {
		switch ffa := f.(type) {
		case dt.RelFilepath:
			rf.AddFileFixture(t, ffa, defaults)
		case *FileFixtureArgs:
			if ffa.Name == "" {
				t.Fatalf("Name not set for file fixure being added to project fixture '%s'", rf.Name)
			}
			rf.AddFileFixture(t, ffa.Name, ffa)
		default:
			t.Fatalf("Invalid type '%T' passed for file fixure being added to project fixture: '%v'", f, f)
		}
	}
}
