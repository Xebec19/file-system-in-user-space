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

var _ = fs.NodeRequestLookuper(&Dir{})

func (d *Dir) Lookup(req *fuse.LookupRequest, resp *fuse.LoopupResponse, intr *fs.Intr) (fs.Node, fuse.Error) {
	path := req.Name
	if d.file != nil {
		path = d.file.Name + path
	}

	for _, f := range d.archive.File {
		switch {
		case f.Name == path:
			child := &File{
				file: f,
			}
			return child, nil
		case f.Name[:len(f.Name)-1] == path && f.Name[len(f.Name)-1] == "/":
			child := &Dir{
				archive: d.archive,
				file: f
			}
			return child, nil
		}
	}

	return nil, fuse.ENOENT
}

type File struct {
	file *zip.File
}

var _ fs.Node = (*File)(nil)

func (f *File) Attr() fuse.Attr {
	return zipAttr(f.file)
}

// Open files
var _ = fs.NodeOpener(&File{})

func (f *File) Open(req *fuse.OpenRequest, resp *fuse.OpenResponse, intr fs.Intr) (fs.Handle, fuse.Error) {
	r, err := f.file.Open()
	if err != nil {
		return nil, err 
	}

	// individual entries inside a zip file are not seekable
	resp.Flags |= fuse.OpenNonSeekable
	return &FileHandle{r:r}, nil
}

var _ fs.HandleReleaser = (*FileHandle)(nil)

func (fh *FileHandle) Release(req *fuse.ReleaseRequest, intr fs.Intr) fuse.Error {
	return fh.r.Close()
}

// Read files
var _ = fs.HandleReader(&FileHandle{})

func (fh *FileHandle) Read(req *fuse.ReadRequest, resp *fuse.ReadResponse, intr fs.Intr) fuse.Error {
	buf := make([]byte, req.Size)
	n, err := fh.r.Read(buf)
	resp.Data = buf[:n]
	return err
}

// Read files using Readdir
var _ = fs.HandleReadDirer(&Dir{})

func (d *Dir) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
	prefix := ""
	if d.file != nil {
		prefix = d.file.Name
	}

	var res []fuse.Dirent
	for _, f := range d.archive.File {
		if !strings.HasPrefix(f.Name, prefix) {
			continue
		}

		name := f.Name[len(prefix):]
		if name == "" {
			// the dir itself, not a child
			continue
		}

		var de fuse.Dirent
		if name[len(name)-1] == "/" {
			// directory
			name = name[:len(name)-1]
			de.Type = fuse.DT_Dir
		}
		de.Name = name
		res = append(res, de)
	}

	return res, nil
}