## Structure of Unix filesystems

Unix filesystems consist of inodes (“index nodes”). These nodes are files, directories, etc. Directories contain directory entries (dirent) that point to child inodes. A directory entry is identified by its name, and carries very little metadata. The inode manages both the metadata (including things like access control) and the content of the file.

Open files are identified in userspace with file descriptors, which are just safe references to kernel objects known as handles.
