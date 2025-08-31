#!/usr/bin/env python3
import os
import sys

def make_tree(root: str, levels: int, files_per_dir: int, level: int = 1, subdirs_per_level: int = 2, file_size: int = 1024):
    """
    Recursively create a directory tree with random files.
    
    :param root: Root directory
    :param levels: Total number of levels to generate
    :param files_per_dir: Number of files per directory
    :param level: Current level (used internally)
    :param subdirs_per_level: Number of subdirectories per directory
    :param file_size: Size of each file in bytes
    """
    os.makedirs(root, exist_ok=True)

    # Create files with random binary data
    for i in range(1, files_per_dir + 1):
        file_path = os.path.join(root, f"file_{level}_{i}.bin")
        with open(file_path, "wb") as f:
            f.write(os.urandom(file_size))

    # Recurse into subdirectories if we haven't reached max depth
    if level < levels:
        for j in range(1, subdirs_per_level + 1):
            subdir = os.path.join(root, f"dir_{level}_{j}")
            make_tree(subdir, levels, files_per_dir, level + 1, subdirs_per_level, file_size)


if __name__ == "__main__":
    if len(sys.argv) < 4:
        print(f"Usage: {sys.argv[0]} <root_dir> <levels> <files_per_dir>")
        sys.exit(1)

    root_dir = sys.argv[1]
    levels = int(sys.argv[2])
    files_per_dir = int(sys.argv[3])

    make_tree(root_dir, levels, files_per_dir)
    print(f"Directory tree created at: {root_dir}")
