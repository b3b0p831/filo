![image](/imgs/filo)
> 先入後出 同期 — *First-in, last-out synchronization*


**FILO (First-In, Last-Out)** is a cross-platform file synchronization tool designed for large datasets like media libraries.  

---

## Overview

FILO monitors a source directory and synchronizes files to a target directory.  
It ensures the most **recent files** are always present on the target, while automatically evicting older files once a configurable fill threshold (`max_fill`) is reached.

FILO watches the files on the source directory and keeps the target directory filled with only the most recent media and other criteria. 



## Features
- Sync newest files from source → target
- Auto-evict oldest files when target approaches `max_fill`
- cross-platform via `fsnotify`
- Priotize files/directories based on Jellyfin/Plex API integration(i.e watch history, favorites, etc)
 


## Config Example

```toml
source_dir = "/mnt/pool" # filo will watch this directory for changes
target_dir = "/mnt/ssd"  # When changes are detected, they will be synced here such that the contents match that of source_dir
max_fill   = 0.92        # filo cannot perform an action such that results in target_dir_fill > max_fill 
log_level  = "info"      # info, debug, warn
sync_delay = "5m"        # 1s, 5m, 10h (Default=30s)
```
