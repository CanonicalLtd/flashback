package reset

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/CanonicalLtd/flashback/audit"
	"github.com/CanonicalLtd/flashback/core"
)

// restoreWritable restores a backup of the files to the writable partition
// We don't use an image as we'd need to regenerate the encryption key
func restoreWritable() error {
	audit.Println("Backup writable partition to the restore partition")
	// Mount the writable path
	if err := core.Mount(core.PartitionTable.Writable, core.WritablePath); err != nil {
		return err
	}

	// Mount the restore path
	err := core.Mount(core.PartitionTable.Restore, core.RestorePath)
	if err != nil {
		return err
	}

	// Open the tar file
	tarfile, err := os.Open(core.BackupImageWritable)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	// Open the gzip reader
	gr, err := gzip.NewReader(tarfile)
	if err != nil {
		return err
	}
	defer gr.Close()

	// Open the tar reader
	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			// Unmount the writable partition
			_ = core.Unmount(core.WritablePath)
			_ = core.Unmount(core.RestorePath)
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// Target location where the dir/file should be created
		target := filepath.Join(core.WritablePath, header.Name)

		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}

	}
}
