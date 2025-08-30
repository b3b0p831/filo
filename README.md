```

███████╗██╗██╗      ██████╗
██╔════╝██║██║     ██╔═══██╗
███████╗██║██║     ██║   ██║
██╔════╝██║██║     ██║   ██║
██║     ██║███████╗╚██████╔╝
╚═╝     ╚═╝╚══════╝ ╚═════╝

         先入後出 同期
```
> 先入後出 同期 — *First-in, last-out synchronization*

**FILO (First-In, Last-Out)** is a cross-platform file synchronization tool designed for large datasets like media libraries.  

---

## Overview

FILO monitors a source directory and synchronizes files to a target directory.  
It ensures the most **recent files** are always present on the target, while automatically evicting older files once a configurable fill threshold (`max_fill`) is reached.

Example use case:  
- Source: 2×4 TB external HDDs for a home media server  
- Target: 2 TB SSD for fast playback with Jellyfin/Plex  
- FILO watches the library on the larger drives and keeps the SSD filled with only the most recent media. 


## Features
- Sync newest files from source → target
- Auto-evict oldest files when target approaches `max_fill`
- cross-platform via `fsnotify`
- Priotize files/directories based on Jellyfin/Plex API integration(i.e watch history, favorites, etc)
 


## Config Example

```toml
source_dir = "/mnt/pool"
target_dir = "/mnt/ssd"
max_fill   = 0.92
log_level  = "info"
```