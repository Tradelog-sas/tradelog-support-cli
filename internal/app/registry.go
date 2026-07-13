package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// packageScope y packageName identifican el paquete Swift en CodeArtifact.
const (
	packageScope = "tradelog"
	packageName  = "TradelogSupport"
)

// releaseURL arma la URL del recurso del paquete en el registry Swift.
func releaseURL(registryEndpoint string) string {
	return strings.TrimRight(registryEndpoint, "/") + "/" + packageScope + "/" + packageName
}

// resolveVersion devuelve la versión pedida, o la más reciente si es "" / "latest".
func resolveVersion(registryEndpoint, token, want string) (string, error) {
	if want != "" && want != "latest" {
		return want, nil
	}

	req, _ := http.NewRequest(http.MethodGet, releaseURL(registryEndpoint), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.swift.registry.v1+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo listar versiones: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry respondió %d al listar versiones", resp.StatusCode)
	}

	var payload struct {
		Releases map[string]json.RawMessage `json:"releases"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("lista de versiones ilegible: %w", err)
	}
	if len(payload.Releases) == 0 {
		return "", fmt.Errorf("no hay versiones publicadas del SDK")
	}

	latest := ""
	for v := range payload.Releases {
		if latest == "" || compareVersions(v, latest) > 0 {
			latest = v
		}
	}
	return latest, nil
}

// compareVersions compara "2026.508.85" numéricamente por componente. >0 si a>b.
func compareVersions(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var na, nb int
		if i < len(pa) {
			na, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			nb, _ = strconv.Atoi(pb[i])
		}
		if na != nb {
			if na > nb {
				return 1
			}
			return -1
		}
	}
	return 0
}

// downloadZip baja el archivo del release al path dado, mostrando progreso.
func downloadZip(registryEndpoint, token, version, dst string) error {
	url := releaseURL(registryEndpoint) + "/" + version + ".zip"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.swift.registry.v1+zip")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("descarga falló: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry respondió %d al descargar %s", resp.StatusCode, version)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	pw := &progressWriter{total: resp.ContentLength, label: "  descargando"}
	if _, err := io.Copy(io.MultiWriter(f, pw), resp.Body); err != nil {
		return fmt.Errorf("escribiendo descarga: %w", err)
	}
	pw.done()
	return nil
}

// progressWriter imprime progreso simple de descarga.
type progressWriter struct {
	total   int64
	written int64
	label   string
	lastPct int
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n := len(b)
	p.written += int64(n)
	if p.total > 0 {
		pct := int(p.written * 100 / p.total)
		if pct != p.lastPct {
			p.lastPct = pct
			fmt.Fprintf(os.Stderr, "\r%s… %d%% (%d MB)", p.label, pct, p.written>>20)
		}
	} else {
		fmt.Fprintf(os.Stderr, "\r%s… %d MB", p.label, p.written>>20)
	}
	return n, nil
}

func (p *progressWriter) done() { fmt.Fprintln(os.Stderr) }
