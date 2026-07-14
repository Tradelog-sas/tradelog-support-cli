package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractStripRoot unzips zipPath into destDir, dropping the first path segment
// ("tradelog.TradelogSupport/") to leave a clean package. Preserves permissions
// and symlinks (important for the .xcframework bundles).
func extractStripRoot(zipPath, destDir string) (int, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer zr.Close()

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return 0, err
	}

	files := 0
	for _, f := range zr.File {
		rel := stripFirstSegment(f.Name)
		if rel == "" {
			continue
		}

		target := filepath.Join(destAbs, rel)
		// Zip-slip guard: the target must stay inside destDir.
		if !strings.HasPrefix(target, destAbs+string(os.PathSeparator)) && target != destAbs {
			return files, fmt.Errorf("zip entry outside destination: %s", f.Name)
		}

		info := f.FileInfo()
		switch {
		case info.IsDir():
			if err := os.MkdirAll(target, 0o755); err != nil {
				return files, err
			}
		case info.Mode()&os.ModeSymlink != 0:
			if err := writeSymlink(f, target, destAbs); err != nil {
				return files, err
			}
			files++
		default:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return files, err
			}
			if err := writeFile(f, target, info.Mode()); err != nil {
				return files, err
			}
			files++
		}
	}
	return files, nil
}

func stripFirstSegment(name string) string {
	name = strings.TrimPrefix(name, "/")
	i := strings.IndexByte(name, '/')
	if i < 0 {
		return "" // top-level entry (the container folder): ignored
	}
	return name[i+1:]
}

func writeFile(f *zip.File, target string, mode os.FileMode) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func writeSymlink(f *zip.File, target, destAbs string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	dest, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	link := string(dest)

	// Zip-slip via symlink: the link target must stay inside destAbs. Reject
	// absolute paths and any ".." that escapes.
	if filepath.IsAbs(link) {
		return fmt.Errorf("rejected absolute symlink: %s -> %s", f.Name, link)
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(target), link))
	if resolved != destAbs && !strings.HasPrefix(resolved, destAbs+string(os.PathSeparator)) {
		return fmt.Errorf("symlink escapes destination: %s -> %s", f.Name, link)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	_ = os.Remove(target)
	return os.Symlink(link, target)
}
