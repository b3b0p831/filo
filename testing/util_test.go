package testing

import (
	"path/filepath"
	"testing"
	"time"

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
	test_root = "/Users/bebop831/Dev/filo_tests/"
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
		name:      "Large Library",
		path:      "/Users/bebop831/Dev/serv",
		wantNodes: 5, //Shoud be len(tree.Index) - 1
		check: func(t *testing.T, tree *util.FileTree) {

			for k, v := range tree.Index {
				if !util.IsAppovedPath(v.Entry.Name()) {
					t.Errorf("unapproved file '%s' detected in filetree", k)
				}
			}

			// file2test := "One.Battle.After.Another.2025.1080p.WEBRip.10Bit.DDP5.1.x265-NeoNoir.mkv"
			// file1Path := filepath.Join("/Users/bebop831/Dev/serv/movies", file2test)
			// if _, ok := tree.Index[file1Path]; !ok {
			// 	t.Errorf("expected %s in index", file1Path)
			// }

		},
	},
	{
		name:      "should_pass",
		path:      filepath.Join(test_root, "should_pass"),
		wantNodes: 12, //Shoud be len(tree.Index) - 1
		check: func(t *testing.T, tree *util.FileTree) {

			for k, v := range tree.Index {
				if !util.IsAppovedPath(v.Entry.Name()) {
					t.Errorf("unapproved file '%s' detected in filetree", k)
				}
			}


			// file2test := "One.Battle.After.Another.2025.1080p.WEBRip.10Bit.DDP5.1.x265-NeoNoir.mkv"
			// file1Path := filepath.Join("/Users/bebop831/Dev/serv/movies", file2test)
			// if _, ok := tree.Index[file1Path]; !ok {
			// 	t.Errorf("expected %s in index", file1Path)
			// }

		},
	},
		{
		name:      "hackerman",
		path:      filepath.Join(test_root, "hackerman"),
		wantNodes: 6357, //Shoud be len(tree.Index) - 1
		check: func(t *testing.T, tree *util.FileTree) {

			for k, v := range tree.Index {
				if !util.IsAppovedPath(v.Entry.Name()) {
					t.Errorf("unapproved file '%s' detected in filetree", k)
				}
			}


			// file2test := "One.Battle.After.Another.2025.1080p.WEBRip.10Bit.DDP5.1.x265-NeoNoir.mkv"
			// file1Path := filepath.Join("/Users/bebop831/Dev/serv/movies", file2test)
			// if _, ok := tree.Index[file1Path]; !ok {
			// 	t.Errorf("expected %s in index", file1Path)
			// }

		},
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
			if tree != nil && tt.wantNodes > 0 {
				if got := len(tree.Index); got != tt.wantNodes {
					t.Errorf("expected %d nodes, got %d", tt.wantNodes, got)
				}
			}
			if tt.check != nil && tree != nil {
				tt.check(t, tree)
			}
		})
	}
}

// func TestWatchEvents - Will test the dir watch functionality, ensuring that all desired events are captured and handled and others are ignored
//					 	  Should be able to handle errors and race conditions

// func TestFiloSync  -   Will perform the sync after events have been triggered. This should be able to determine the differences between dirs and create, rename, remove etc
// 