package services

import (
	"archive/zip"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type FS struct {
	archive *zip.Reader
}

var _ fs.FS = (*FS)(nil)

func (f *FS) Root() (fs.Node, fuse.Error) {
	n := &Dir{
		archive: f.archive,
	}
}

type Dir struct {
	archive *zip.Reader
	file    *zip.File // nil for root
}

var _ fs.Node = (*Dir)(nil)

func zipAttr(f *zip.File) fuse.Attr {
	return fuse.Attr{
		Size:   f.UncompressedSize64,
		Mode:   f.Mode(),
		Mtime:  f.ModTime(),
		Ctime:  f.ModTime(),
		Crtime: f.ModTime(),
	}
}

func (d *Dir) Attr() fuse.Attr {
	if d.file == nil {
		// root directory
		return fuse.Attr{Mode: os.ModeDir | 0755}
	}
	return zipAttr(d.file)
}
