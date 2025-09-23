package testing

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bebop831.com/filo/util"
)

var (
	test_root = "/Users/bebop831/Dev/filo_tests/"
	test_dirs = []string{
		//Control
		filepath.Join(test_root, "control"),
		//Normal Pass Test Cases
		filepath.Join(test_root, "should_pass/t1"),
		filepath.Join(test_root, "should_pass/t2"),
		filepath.Join(test_root, "should_pass/t3"),
		// Normal Fail Test Cases
		filepath.Join(test_root, "should_err/t1"),
		filepath.Join(test_root, "should_err/t2"),
		filepath.Join(test_root, "should_err/t3"),
		// Malicous Test Cases
		filepath.Join(test_root, "hackerman/t1"),
		filepath.Join(test_root, "hackerman/t2"),
		filepath.Join(test_root, "hackerman/t3"),
	}
)

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
	for _, currentPath := range test_dirs {
		tmpTree := util.BuildTree(currentPath)

		if tmpTree == nil {
			t.Errorf("expected non-nil tree for %s", currentPath)
			continue
		}

		// Example: control dir should have exactly 5 nodes in index
		if strings.Contains(currentPath, "control") {
			expectedCount := 5
			if len(tmpTree.Index) != expectedCount {
				t.Errorf("expected %d nodes in index, got %d", expectedCount, len(tmpTree.Index))
			}

			// Check that a known file exists
			file1Path := filepath.Join(currentPath, "file1.txt")
			if _, ok := tmpTree.Index[file1Path]; !ok {
				t.Errorf("expected node for %s not found in index", file1Path)
			}

			// Check that subdirA has 2 children
			subdirAPath := filepath.Join(currentPath, "subdirA")
			subdirNode, ok := tmpTree.Index[subdirAPath]
			if !ok {
				t.Errorf("expected node for %s not found in index", subdirAPath)
			} else if len(subdirNode.Children) != 2 {
				t.Errorf("expected subdirA to have 2 children, got %d", len(subdirNode.Children))
			}
		}
		fmt.Println(currentPath)
	}
}
