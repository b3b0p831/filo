package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"os/exec"

	"bebop831.com/filo/util"
)

type BuildTreeTest struct {
	name      string
	path      string
	wantNodes int
	wantErr   bool
	check     func(t *testing.T, tree *util.FileTree)
}

var (
	test_root = filepath.Join(os.Getenv("HOME"), "Dev/filo_tests")
)

var buildTreeTests = []BuildTreeTest{
	{
		name:      "control",
		path:      filepath.Join(test_root, "control"),
		wantNodes: 5,
		check: func(t *testing.T, tree *util.FileTree) {
			file1Path := filepath.Join(test_root, "control", "file1.txt")
			if _, ok := tree.Index[file1Path]; !ok {
				t.Errorf("expected %s in index", file1Path)
			}
			subdirAPath := filepath.Join(test_root, "control", "subdirA")
			if len(tree.Index[subdirAPath].Children) != 2 {
				t.Errorf("subdirA children mismatch")
			}
		},
	},
	{
		name:      "should_pass",
		path:      filepath.Join(test_root, "should_pass"),
		wantNodes: 0, //Shoud be len(tree.Index) - 1
		check: func(t *testing.T, tree *util.FileTree) {

			for k, v := range tree.Index {
				if !util.IsApprovedPath(v.Entry.Name()) {
					t.Errorf("unapproved file '%s' detected in filetree", k)
				}
			}

		},
	},
	{
		name:      "large_library",
		path:      filepath.Join(test_root, "large_library"),
		wantNodes: 6357, //Shoud be len(tree.Index) - 1
		check: func(t *testing.T, tree *util.FileTree) {

			for k, v := range tree.Index {
				if !util.IsApprovedPath(v.Entry.Name()) {
					t.Errorf("unapproved file '%s' detected in filetree", k)
				}
			}

		},
	},
	// {
	// 	name:      "hackerman",
	// 	path:      filepath.Join(test_root, "hackerman"),
	// 	wantNodes: 6357, //Shoud be len(tree.Index) - 1
	// 	check: func(t *testing.T, tree *util.FileTree) {

	// 	},
	// },
}

func TestGetTimeInterval(t *testing.T) {
	test_data := map[string]time.Duration{
		// Valid basics
		"1s":  time.Second,
		"10m": 10 * time.Minute,
		"2h":  2 * time.Hour,

		// Boundary numbers
		"0s":                    0,
		"999999999s":            999999999 * time.Second, // huge, may overflow
		"0001s":                 time.Second,
		"18446744073709551615s": 0, // should fail

		// Weird suffixes
		"1sec":     0,
		"2minutes": 0,
		"3hour":    0,
		"5ss":      0,
		"10hh":     0,

		// Nonnumeric junk
		"5s5":  0,
		"12m3": 0,
		"５s":   0,
		"1💥s":  0,

		// Whitespace / control chars
		" 1s":      0,
		"1s ":      0,
		"1\tm":     0,
		"1\nh":     0,
		"1\u200Bs": 0,

		// Signs and decimals
		"-10s": 0,
		"+5m":  0,
		"1.5h": 0,

		// Case traps
		"1S":   0,
		"1M":   0,
		"1H":   0,
		"10Ms": 0,

		// Empty / random junk
		"":    0,
		" ":   0,
		"abc": 0,
		"💣":   0,
		"123": 0,
		"s":   0,
	}

	for timeStr, expectedTimeVal := range test_data {
		currentTimeVal, err := util.GetTimeInterval(timeStr)
		if err != nil && expectedTimeVal != 0 {
			t.Error(err)
		}

		if currentTimeVal != expectedTimeVal {
			t.Errorf("util.GetTimeInterval(%s) != %v\n", timeStr, expectedTimeVal)
		}
	}
}

func TestBuildTree(t *testing.T) {
	for _, tt := range buildTreeTests {
		t.Run(tt.name, func(t *testing.T) {
			tree := util.BuildTree(tt.path)
			if tree == nil && !tt.wantErr {
				t.Fatalf("expected non-nil tree, got nil")
			}

			if tree != nil && tt.wantNodes > 0 {
				if got := len(tree.Index); got != tt.wantNodes {
					t.Errorf("expected %d nodes, got %d", tt.wantNodes, got)
				}
			}

			if !contentsCheck(tt.path, tree.Index) {
				t.Errorf("failed contents check for %s", tt.path)
			}

			if tt.check != nil && tree != nil {
				tt.check(t, tree)
			}
		})
	}
}

// This func runs tree command on targetPath, and checks each file exists in Index
func contentsCheck(targetRoot string, treeIndex map[string]*util.FileNode) bool {
	// Define the command and its arguments
	cmd := exec.Command("tree", "-i", "-f", targetRoot)

	// Execute the command and capture its combined output (stdout and stderr)
	output, err := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")
	if err != nil {
		return false
	}

	lines = lines[1:]            // Strip first line from tree output, "Root dir"
	lines = lines[:len(lines)-3] // Strip last 2 lines

	for _, line := range lines {
		if _, ok := treeIndex[line]; !ok {

			//Ignore symlinks
			if strings.Contains(line, " -> ") {
				fmt.Println("Skipping symlink:", strings.Split(line, "->")[0])
				fmt.Println(ok)
				continue
			}

			fmt.Printf("expected %s in %s\n", line, targetRoot)
			fmt.Println(ok)
			return false
		}
	}

	return true
}

// func TestWatchEvents - Will test the dir watch functionality, ensuring that all desired events are captured and handled and others are ignored
//					 	  Should be able to handle errors and race conditions

// func TestFiloSync  -   Will perform the sync after events have been triggered. This should be able to determine the differences between dirs and create, rename, remove etc
//
