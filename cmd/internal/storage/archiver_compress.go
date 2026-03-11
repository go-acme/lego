package storage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func compressDirectory(name, dir, comment string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	writer := zip.NewWriter(file)

	defer func() { _ = writer.Close() }()

	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}

	defer func() { _ = root.Close() }()

	err = writer.SetComment(comment)
	if err != nil {
		return err
	}

	return writer.AddFS(root.FS())
}

func compressFiles(dest string, files []string, comment string) error {
	file, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	writer := zip.NewWriter(file)

	defer func() { _ = writer.Close() }()

	err = writer.SetComment(comment)
	if err != nil {
		return err
	}

	for _, file := range files {
		f, err := writer.Create(filepath.Base(file))
		if err != nil {
			return err
		}

		err = addFileToZip(f, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(f io.Writer, file string) error {
	in, err := os.Open(file)
	if err != nil {
		return err
	}

	defer func() { _ = in.Close() }()

	_, err = io.Copy(f, in)
	if err != nil {
		return err
	}

	return nil
}
