// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package initrd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"kraftkit.sh/cpio"
	"kraftkit.sh/erofs"
)

func compressFiles(output string, writer *cpio.Writer, reader *os.File) error {
	if writer != nil {
		err := writer.Close()
		if err != nil {
			return fmt.Errorf("could not close CPIO writer: %w", err)
		}
	}

	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("could not seek to start of file: %w", err)
	}

	fw, err := os.OpenFile(output+".gz", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("could not open initramfs file: %w", err)
	}

	gw := gzip.NewWriter(fw)

	if _, err := io.Copy(gw, reader); err != nil {
		return fmt.Errorf("could not compress initramfs file: %w", err)
	}

	err = gw.Close()
	if err != nil {
		return fmt.Errorf("could not close gzip writer: %w", err)
	}

	err = fw.Close()
	if err != nil {
		return fmt.Errorf("could not close compressed initramfs file: %w", err)
	}

	if err := os.Remove(output); err != nil {
		return fmt.Errorf("could not remove uncompressed initramfs: %w", err)
	}

	if err := os.Rename(output+".gz", output); err != nil {
		return fmt.Errorf("could not rename compressed initramfs: %w", err)
	}

	return nil
}

func convertToErofs(output string, source string, isTar, allRoot bool) error {
	// Extract the archive to a temporary directory
	if isTar {
		file, err := os.Open(source)
		if err != nil {
			return fmt.Errorf("could not open tar file: %w", err)
		}
		defer file.Close()

		targetDir, err := os.MkdirTemp(os.TempDir(), "kraftkit-untar-")
		if err != nil {
			return fmt.Errorf("could not create temporary directory: %w", err)
		}
		defer os.RemoveAll(targetDir)

		tarReader := tar.NewReader(file)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break // End of tar archive
			}
			if err != nil {
				return err
			}

			// Construct the target file path
			targetPath := filepath.Join(targetDir, header.Name)

			switch header.Typeflag {
			case tar.TypeDir:
				// Create directory
				if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			case tar.TypeReg:
				// Create parent directory if necessary (optional)
				if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
					return err
				}

				// Create and write the file
				outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
				if err != nil {
					return err
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					outFile.Close()
					return err
				}
				outFile.Close()
			case tar.TypeSymlink:
				// Create a symlink
				if err := os.Symlink(header.Linkname, targetPath); err != nil {
					return fmt.Errorf("could not create symlink from %s to %s: %w", targetPath, header.Linkname, err)
				}
			case tar.TypeLink:
				// Link needs to point to exact file
				linkPath := filepath.Join(targetDir, header.Linkname)

				// Create a hard link
				if err := os.Link(linkPath, targetPath); err != nil {
					return fmt.Errorf("could not create hard link from %s to %s: %w", targetPath, linkPath, err)
				}
			default:
				log.Printf("Skipping unsupported file type: %c in %s", header.Typeflag, header.Name)
			}
		}

		source = targetDir
	}

	// Open writer for the output file
	writer, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("could not open output file: %w", err)
	}

	err = erofs.Create(writer, os.DirFS(source), erofs.WithAllRoot(allRoot))
	if err != nil {
		return fmt.Errorf("could not create EROFS filesystem: %w", err)
	}
	writer.Close()

	return nil
}

func isCpioFile(initrd string) (bool, error) {
	fi, err := os.Open(initrd)
	if err != nil {
		return false, fmt.Errorf("could not open file: %w", err)
	}
	defer fi.Close()

	reader := cpio.NewReader(fi)

	_, _, err = reader.Next()
	if err != nil {
		return false, nil
	}

	return true, nil
}

func isErofsFile(initrd string) (bool, error) {
	fi, err := os.Open(initrd)
	if err != nil {
		return false, fmt.Errorf("could not open file: %w", err)
	}
	defer fi.Close()

	_, err = erofs.Open(fi)
	if err != nil {
		return false, nil
	}

	return true, nil
}
