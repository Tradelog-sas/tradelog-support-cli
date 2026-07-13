package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractStripRoot descomprime zipPath dentro de destDir, quitando el primer
// segmento de ruta ("tradelog.TradelogSupport/") para dejar el paquete limpio.
// Preserva permisos y symlinks (importante para los .xcframework).
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
		// Defensa zip-slip: el destino debe quedar dentro de destDir.
		if !strings.HasPrefix(target, destAbs+string(os.PathSeparator)) && target != destAbs {
			return files, fmt.Errorf("entrada de zip fuera de destino: %s", f.Name)
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
		return "" // entrada de nivel raíz (la carpeta contenedora): se ignora
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

	// Defensa zip-slip vía symlink: el destino del link debe quedar dentro de
	// destAbs. Rechazamos rutas absolutas y cualquier ".." que escape.
	if filepath.IsAbs(link) {
		return fmt.Errorf("symlink absoluto rechazado: %s -> %s", f.Name, link)
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(target), link))
	if resolved != destAbs && !strings.HasPrefix(resolved, destAbs+string(os.PathSeparator)) {
		return fmt.Errorf("symlink escapa el destino: %s -> %s", f.Name, link)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	_ = os.Remove(target)
	return os.Symlink(link, target)
}
