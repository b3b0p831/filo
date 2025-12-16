package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"os/exec"

	"bebop831.com/filo/internal/config"
	"bebop831.com/filo/internal/fs"

	"github.com/BurntSushi/toml"
)

type BuildTreeTest struct {
	name    string
	path    string
	wantErr bool
	check   func(t *testing.T, tree *fs.FileTree)
}

var (
	userHome       string
	test_root      string
	buildTreeTests []BuildTreeTest
)

func init() {
	var err error
	userHome, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	test_root = filepath.Join(userHome, "Dev", "filo_tests")
	buildTreeTests = []BuildTreeTest{
		{
			name: "control",
			path: filepath.Join(test_root, "control"),
			check: func(t *testing.T, tree *fs.FileTree) {
				filePaths := []string{filepath.Join(test_root, "control", "file1.txt")}
				for _, fp := range filePaths {
					if _, ok := tree.Index[fp]; !ok {
						t.Errorf("expected %s in index", fp)
					}
				}
			},
		},
		{
			name:  "large_library",
			path:  filepath.Join(test_root, "large_library"),
			check: nil,
		},
		{
			name:    "empty",
			path:    filepath.Join(test_root, "empty"),
			check:   nil,
			wantErr: false,
		},
		{
			name:    "thisdirdoesnotexist",
			path:    filepath.Join(test_root, "thisdirdoesnotexist"),
			check:   nil,
			wantErr: true,
		},
		{
			name:    "symlinks",
			path:    filepath.Join(test_root, "symlinks"),
			check:   nil,
			wantErr: true,
		},
	}
}

func TestTOMLParse(t *testing.T) {
	tomlFN := "config.toml"
	var cfgTest config.Config
	_, err := toml.DecodeFile(tomlFN, &cfgTest)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.Load()

	t.Log(cfgTest)
	t.Log(cfg)

	if !cfg.Equal(cfgTest) {
		t.Error("mismatch configs")
	} else {
		t.Logf("%s was successfully parsed", tomlFN)
	}

}

func TestBuildTree(t *testing.T) {
	for _, tt := range buildTreeTests {
		t.Run(tt.name, func(t *testing.T) {
			tree := fs.BuildTree(tt.path)
			if tree == nil && !tt.wantErr {
				t.Fatalf("expected non-nil tree, got nil on %s", tt.path)
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
func contentsCheck(targetRoot string, treeIndex map[string]*fs.FileNode) int {
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
			symlink := strings.Split(line, " -> ")

			symNodeInfo, err := os.Lstat(symlink[0])
			if err != nil {
				fmt.Println(err)
				return -1
			}

			if symNodeInfo.Mode()&os.ModeSymlink != 0 {
				fmt.Println("Skipping symlink:", symlink)
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
