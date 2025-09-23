# FILO Testing

Filo is a fist-in, last-out file sync tool used to sync large files and media libraries. One of the main use cases was allowing my media library (+4TB) to be selectivly synced to my SSD which is only 2TB based on user defined criteria and first come first serve.


In order to accomplish this, filo performs the following functions:

1. Check data usage on src and tgt dirs
2. Build file tree of src and tgt dirs
3. Create map of nodes in src dir, missing from tgt dir (i.e what files are missing in tgt from src?)
    - This process will take some time as it will recursivly compare children starting at the src and tgt dirs defiend in config.toml
    - Then it will compare file content
4. Using this map, perform inital sync. 
5. Once both 


#### A *FileTree is a snapshot in time
we take one in the beggninng to perform initial sync
Everytime changes are detected, enough time has elapsed, AND FILO_LOCK is unlocked(i.e sync not in progress) a new *FileTree is built.


### TODO: 
    1. Need to create atomic structure or a filo db
    2. "Scanned 10GB in 33.168 seconds"
    3. Implement FILO_LOCK, currently sync doesn't check if a sync is currently active. FILO_LOCK will lock filo from syncing changes until the current sync is completed.
    