package services

import (
	"archive/zip"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func Mount(path, mountpoint string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return err
	}

	defer archive.Close()

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	defer c.Close()

	fileSys := &FS{
		archive: &archive.Reader,
	}
	if err := fs.Serve(c, fileSys); err != nil {
		return err
	}

	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}
}
