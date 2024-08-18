package _integration_test

import (
	"bytes"
	"fmt"
	"go-refactor/internal"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Test copies the before directory into temp directory.
// Executes main.go against the directory
// Runs tests to check the files are as expected
func TestIntegration(t *testing.T) {
	const dirTemp = "temp"
	const dirBefore = "before"
	const dirWant = "want"

	err := deleteAllFiles(dirTemp)
	if err != nil {
		t.Fatal(err)
	}

	err = copyDir(dirBefore, dirTemp)
	if err != nil {
		t.Fatal(err)
	}
	defer deleteAllFiles(dirTemp)

	err = internal.Do(dirTemp)
	if err != nil {
		t.Fatal(err)
	}

	err = compareDirs(dirTemp, dirWant)
	if err != nil {
		t.Error(err)
	}
}

func deleteAllFiles(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := dir + "/" + file.Name()
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func copyDir(srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(srcFile, destFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func compareDirs(dir1, dir2 string) error {
	files1, err := os.ReadDir(dir1)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir1, err)
	}

	for _, file1 := range files1 {
		if file1.IsDir() {
			continue
		}

		file1Path := filepath.Join(dir1, file1.Name())
		file2Path := filepath.Join(dir2, file1.Name())

		file1Stat, err := os.Stat(file1Path)
		fmt.Println(file1Stat)
		if os.IsNotExist(err) {
			return fmt.Errorf("File %s does not exist in %s\n", file1.Name(), dir2)
		}

		file2Stat, err := os.Stat(file2Path)
		fmt.Println(file2Stat)
		if os.IsNotExist(err) {
			return fmt.Errorf("File %s does not exist in %s\n", file1.Name(), dir2)
		}

		if file1Stat.Size() != file2Stat.Size() {
			return fmt.Errorf("File %s differs in size: %v %v\n", file1.Name(), file1Stat.Size(), file2Stat.Size())
		}

		content1, err := os.ReadFile(file1Path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file1Path, err)
		}

		content2, err := os.ReadFile(file2Path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file2Path, err)
		}

		if !bytes.Equal(content1, content2) {
			fmt.Printf("File %s differs in content\n", file1.Name())
		}
	}

	return nil
}
