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

// packageScope and packageName identify the Swift package in the registry.
const (
	packageScope = "tradelog"
	packageName  = "TradelogSupport"
)

// releaseURL builds the package resource URL in the Swift registry.
func releaseURL(registryEndpoint string) string {
	return strings.TrimRight(registryEndpoint, "/") + "/" + packageScope + "/" + packageName
}

// resolveVersion returns the requested version, or the latest if "" / "latest".
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
		return "", fmt.Errorf("could not list versions: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry responded %d while listing versions", resp.StatusCode)
	}

	var payload struct {
		Releases map[string]json.RawMessage `json:"releases"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("unreadable version list: %w", err)
	}
	if len(payload.Releases) == 0 {
		return "", fmt.Errorf("no published SDK versions found")
	}

	latest := ""
	for v := range payload.Releases {
		if latest == "" || compareVersions(v, latest) > 0 {
			latest = v
		}
	}
	return latest, nil
}

// compareVersions compares "2026.508.85" numerically per component. >0 if a>b.
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

// downloadZip downloads the release archive to the given path, showing progress.
func downloadZip(registryEndpoint, token, version, dst string) error {
	url := releaseURL(registryEndpoint) + "/" + version + ".zip"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.swift.registry.v1+zip")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry responded %d while downloading %s", resp.StatusCode, version)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	pw := &progressWriter{total: resp.ContentLength, label: "  downloading"}
	if _, err := io.Copy(io.MultiWriter(f, pw), resp.Body); err != nil {
		return fmt.Errorf("writing download: %w", err)
	}
	pw.done()
	return nil
}

// progressWriter prints a simple download progress line.
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
