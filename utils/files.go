package utils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

func srcDir(root string) string {
	return JoinPath(root, "./src")
}

func normalizeExt(ext string) string {
	if len(ext) == 0 {
		return ext
	}

	if ext[0] != '.' {
		ext = "." + ext
	}

	return ext
}

func CollectAllFilesWithExt(dir string, ext string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []string
	ext = normalizeExt(ext)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if filepath.Ext(e.Name()) == ext {
			result = append(result, e.Name())
		}
	}

	return result, nil
}

func JoinPath(left, right string) string {
	return path.Join(left, right)
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func JoinCmakeLists(dir string) string {
	return JoinPath(dir, "/CMakeLists.txt")
}

func IsFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, err
}

func ContainCmakeLists(path string) (bool, error) {
	fullPath := JoinCmakeLists(path)
	return IsFileExist(fullPath)
}

func Wd() (string, error) {
	return os.Getwd()
}

func CopyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func Mkdir(fullpath string) error {
	return os.MkdirAll(fullpath, 0755)
}

func toCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, "")
}

func CollectTests(dir string) ([]TestEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	tests := make([]TestEntry, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := path.Ext(name)
		if ext != ".cpp" {
			continue
		}

		base := strings.TrimSuffix(name, ext)

		tests = append(tests, TestEntry{
			Name:   toCamelCase(base),
			Source: name,
		})
	}

	slices.SortFunc(tests, func(a, b TestEntry) int {
		return strings.Compare(a.Source, b.Source)
	})

	return tests, nil
}

type TestEntry struct {
	Name   string
	Source string
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}
