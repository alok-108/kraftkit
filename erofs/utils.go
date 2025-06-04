// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package erofs

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/anchore/stereoscope"
	scfile "github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"

	"kraftkit.sh/log"
)

type fInfo struct {
	uid  int
	gid  int
	mode fs.FileMode
}

// unpackTarFileToDirectory extracts the contents of a tar file to a temporary
// directory and returns the path to that directory. It handles directories,
// regular files, symlinks, and hard links.
func unpackTarFileToDirectory(ctx context.Context, source string) (string, map[string]fInfo, error) {
	fInfoMap := make(map[string]fInfo)

	log.G(ctx).Info("unpacking tar file")

	file, err := os.Open(source)
	if err != nil {
		return "", nil, fmt.Errorf("could not open tar file: %w", err)
	}
	defer file.Close()

	targetDir, err := os.MkdirTemp(os.TempDir(), "kraftkit-untar-")
	if err != nil {
		return "", nil, fmt.Errorf("could not create temporary directory: %w", err)
	}

	var tarReader *tar.Reader
	if gzr, err := gzip.NewReader(file); err == nil {
		tarReader = tar.NewReader(gzr)
	} else {
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return "", nil, fmt.Errorf("could not seek to start of tarball: %w", err)
		}

		tarReader = tar.NewReader(file)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return "", nil, err
		}

		// Construct the target file path
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return "", nil, err
			}
		case tar.TypeReg:
			// Create parent directory if necessary (optional)
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return "", nil, err
			}

			// Create and write the file
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return "", nil, err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", nil, err
			}
			outFile.Close()
		case tar.TypeSymlink:
			// Create a symlink
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return "", nil, fmt.Errorf("could not create symlink from %s to %s: %w", targetPath, header.Linkname, err)
			}
		case tar.TypeLink:
			// Link needs to point to exact file
			linkPath := filepath.Join(targetDir, header.Linkname)

			// Create a hard link
			if err := os.Link(linkPath, targetPath); err != nil {
				return "", nil, fmt.Errorf("could not create hard link from %s to %s: %w", targetPath, linkPath, err)
			}
		default:
			log.G(ctx).Warnf("Skipping unsupported file type: %c in %s", header.Typeflag, header.Name)
			continue
		}

		if !strings.HasPrefix(header.Name, "/") {
			header.Name = filepath.Join("/", header.Name)
		}

		fInfoMap[header.Name] = fInfo{
			uid:  header.Uid,
			gid:  header.Gid,
			mode: fs.FileMode(header.Mode),
		}
	}

	return targetDir, fInfoMap, nil
}

func unpackOCIImageToDirectory(ctx context.Context, source string) (string, map[string]fInfo, error) {
	fInfoMap := make(map[string]fInfo)

	log.G(ctx).Info("unpacking oci image")

	image, err := stereoscope.GetImage(ctx, source)
	if err != nil {
		return "", nil, fmt.Errorf("could not load image: %w", err)
	}

	// Create a temporary directory to unpack the image
	targetDir, err := os.MkdirTemp(os.TempDir(), "kraftkit-oci-")
	if err != nil {
		return "", nil, fmt.Errorf("could not create temporary directory: %w", err)
	}

	if err := image.SquashedTree().Walk(func(path scfile.Path, f filenode.FileNode) error {
		if f.Reference == nil {
			log.G(ctx).
				WithField("path", path).
				Debug("skipping: no reference")
			return nil
		}

		info, err := image.FileCatalog.Get(*f.Reference)
		if err != nil {
			return err
		}

		fpath := filepath.Join(targetDir, info.Path)

		// Permissions do not matter as they will be overwritten when
		// creating directories by chown. Just use a default here.
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return fmt.Errorf("could not create directory %s: %w", info.Path, err)
		}

		switch f.FileType {
		case scfile.TypeBlockDevice:
			log.G(ctx).
				WithField("file", path).
				Warn("ignoring block devices")
			return nil

		case scfile.TypeCharacterDevice:
			log.G(ctx).
				WithField("file", path).
				Warn("ignoring char devices")
			return nil

		case scfile.TypeFIFO:
			log.G(ctx).
				WithField("file", path).
				Warn("ignoring fifo files")
			return nil

		case scfile.TypeSymLink:
			log.G(ctx).
				WithField("src", path).
				WithField("link", info.LinkDestination).
				Trace("symlinking")

			if err := os.Symlink(info.LinkDestination, fpath); err != nil {
				return fmt.Errorf("could not create symlink %s: %w", fpath, err)
			}
		case scfile.TypeHardLink:
			log.G(ctx).
				WithField("src", fpath).
				WithField("link", info.LinkDestination).
				Trace("hardlinking")

			dest := filepath.Join(targetDir, info.LinkDestination)
			if err := os.Link(dest, fpath); err != nil {
				return fmt.Errorf("could not create symlink %s: %w", fpath, err)
			}
		case scfile.TypeRegular:
			log.G(ctx).
				WithField("src", path).
				WithField("dst", fpath).
				Trace("copying")

			dfile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("could not open file %s: %w", fpath, err)
			}
			defer dfile.Close()

			reader, err := image.OpenPathFromSquash(path)
			if err != nil {
				return fmt.Errorf("could not open file: %w", err)
			}

			_, err = io.Copy(dfile, reader)
			if err != nil {
				return fmt.Errorf("could not copy file %s: %w", fpath, err)
			}
			if err := dfile.Close(); err != nil {
				return fmt.Errorf("could not close file %s: %w", fpath, err)
			}
		case scfile.TypeDirectory:
			log.G(ctx).
				WithField("dst", fpath).
				Trace("mkdir")

			if err := os.MkdirAll(fpath, info.Mode().Perm()); err != nil {
				return fmt.Errorf("could not create directory %s: %w", fpath, err)
			}

		default:
			log.G(ctx).
				WithField("file", path).
				WithField("type", f.FileType.String()).
				Warn("unsupported file type")
			return nil
		}

		if !strings.HasPrefix(info.Path, "/") {
			info.Path = filepath.Join("/", info.Path)
		}

		fInfoMap[info.Path] = fInfo{
			uid:  info.UserID,
			gid:  info.GroupID,
			mode: info.Mode(),
		}

		return nil
	}, &filetree.WalkConditions{
		LinkOptions: []filetree.LinkResolutionOption{},
		ShouldContinueBranch: func(path scfile.Path, f filenode.FileNode) bool {
			return f.LinkPath == ""
		},
	}); err != nil {
		return "", nil, fmt.Errorf("could not walk image: %w", err)
	}

	return targetDir, fInfoMap, nil
}

// IsErofsFile checks if the given file is an EROFS filesystem.
func IsErofsFile(initrd string) bool {
	fi, err := os.Open(initrd)
	if err != nil {
		return false
	}
	defer fi.Close()

	_, err = Open(fi)
	if err != nil {
		return false
	}

	return true
}
