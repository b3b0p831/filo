package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"os/exec"

	"bebop831.com/filo/util"
)

type BuildTreeTest struct {
	name    string
	path    string
	wantErr bool
	check   func(t *testing.T, tree *util.FileTree)
}

var (
	test_root = filepath.Join(os.Getenv("HOME"), "Dev/filo_tests")
	//test_root = "C:\\Users\\bob\\Dev\\filo_tests"
)

var buildTreeTests = []BuildTreeTest{
	// {
	// 	name: "control",
	// 	path: filepath.Join(test_root, "control"),
	// 	check: func(t *testing.T, tree *util.FileTree) {
	// 		filePaths := []string{filepath.Join(test_root, "control", "file1.txt")}
	// 		for _, fp := range filePaths {
	// 			if _, ok := tree.Index[fp]; !ok {
	// 				t.Errorf("expected %s in index", fp)
	// 			}
	// 		}
	// 	},
	// },
	// {
	// 	name:  "should_pass",
	// 	path:  filepath.Join(test_root, "should_pass"),
	// 	check: nil,
	// },
	// {
	// 	name:  "large_library",
	// 	path:  filepath.Join(test_root, "../large_library"),
	// 	check: nil,
	// },
	{
		name:  "hackerman",
		path:  "/Users/bebop831/Dev/filo_tests/../../../../etc",
		check: nil,
	},
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

			if tree != nil {
				numOfExpectedNodes := contentsCheck(tt.path, tree.Index)
				got := len(tree.Index)

				t.Logf("expected %d nodes, got %d", numOfExpectedNodes, got)
				if got != numOfExpectedNodes {
					t.Errorf("failed contents check for %s", tt.path)
				}
			}

			if tt.check != nil && tree != nil {
				tt.check(t, tree)
			}
		})
	}
}

// This func runs tree command on targetPath, and checks each file exists in Index
// If successful, returns true, len(nodes) returned by tree command (dirs + files - root)
// else returns false, -1
func contentsCheck(targetRoot string, treeIndex map[string]*util.FileNode) int {
	// Define the command and its arguments

	var cmd *exec.Cmd
	var lines []string
	switch runtime.GOOS {
	case "windows":
		// include '-Force' to include hidden files
		treeCmdWindowsNoHidden := "Get-ChildItem -Path %s -Recurse | Select-Object -ExpandProperty FullName"
		//treeCmdWindowsHidden := "Get-ChildItem -Path %s -Recurse -Force | Select-Object -ExpandProperty FullName"
		cmd := exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf(treeCmdWindowsNoHidden, filepath.Clean(targetRoot)))

		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Println(err)
			return -1
		}

		lines = strings.Split(string(output), "\n")
		lines = lines[:len(lines)-1]

	case "linux", "darwin":
		cmd = exec.Command("tree", "-i", "-f", filepath.Clean(targetRoot))
		output, err := cmd.CombinedOutput()
		lines = strings.Split(string(output), "\n")
		if err != nil {
			fmt.Println(err)
			return -1
		}

		if lineLen := len(lines); lineLen < 4 {
			return lineLen
		}

		lines = lines[1:]            // Strip first line from tree output, "Root dir"
		lines = lines[:len(lines)-3] // Strip last 2 lines

	}

	// Execute the command and capture its combined output (stdout and stderr)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if _, ok := treeIndex[line]; !ok {

			//Ignore symlinks
			if strings.Contains(line, " -> ") {
				fmt.Println("Skipping symlink:", strings.Split(line, "->")[0])
				continue
			}

			fmt.Printf("expected %s in %s\n", line, targetRoot)
		}
	}

	return len(lines)
}

// func TestWatchEvents - Will test the dir watch functionality, ensuring that all desired events are captured and handled and others are ignored
//					 	  Should be able to handle errors and race conditions

// func TestFiloSync  -   Will perform the sync after events have been triggered. This should be able to determine the differences between dirs and create, rename, remove etc
//
