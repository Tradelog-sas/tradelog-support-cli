package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Usage prints the general help.
func Usage() {
	fmt.Print(`tradelog — installer for the Tradelog iOS support chat SDK

USAGE:
  tradelog install [flags]     Download and integrate TradelogSupport.xcframework
  tradelog version             Show the CLI version
  tradelog help                Show this help

INSTALL:
  Authenticate with your Tradelog API key (the same one you use in the SDK). The
  CLI downloads the binary framework — no AWS credentials required.

  Flags:
    --api-key   Tradelog API key         (or env TRADELOG_API_KEY)
    --tenant    Tenant / company id       (or env TRADELOG_TENANT_ID)
    --version   SDK version               (default: latest)
    --dest      Destination folder        (default: ./Tradelog)
    --pods      Also generate a .podspec for CocoaPods
    --broker-url  Broker override         (advanced)

  Example:
    tradelog install --api-key tlk_xxx --tenant 4be6e386-...

`)
}

// Install runs the full flow: auth → download → extract → configure.
func Install(args []string) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	apiKey := fs.String("api-key", os.Getenv("TRADELOG_API_KEY"), "Tradelog API key")
	tenant := fs.String("tenant", os.Getenv("TRADELOG_TENANT_ID"), "Tenant / company id")
	version := fs.String("version", "latest", "SDK version (or 'latest')")
	dest := fs.String("dest", "Tradelog", "Destination folder")
	pods := fs.Bool("pods", false, "Generate a .podspec for CocoaPods")
	brokerURL := fs.String("broker-url", DefaultBrokerURL, "Token broker URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *apiKey == "" {
		return fmt.Errorf("missing --api-key (or env TRADELOG_API_KEY)")
	}
	if *tenant == "" {
		return fmt.Errorf("missing --tenant (or env TRADELOG_TENANT_ID)")
	}

	fmt.Println("→ Authenticating with your API key…")
	tok, err := fetchToken(*brokerURL, *tenant, *apiKey)
	if err != nil {
		return err
	}

	resolved, err := resolveVersion(tok.RegistryEndpoint, tok.AuthorizationToken, *version)
	if err != nil {
		return err
	}
	fmt.Printf("→ Version: %s\n", resolved)

	tmp, err := os.CreateTemp("", "tradelog-sdk-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	if err := downloadZip(tok.RegistryEndpoint, tok.AuthorizationToken, resolved, tmpPath); err != nil {
		return err
	}

	pkgDir := filepath.Join(*dest, packageName)
	fmt.Printf("→ Extracting into %s …\n", pkgDir)
	if err := os.RemoveAll(pkgDir); err != nil {
		return err
	}
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return err
	}
	n, err := extractStripRoot(tmpPath, pkgDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(pkgDir, "Package.swift")); err != nil {
		return fmt.Errorf("extracted package has no Package.swift (%d files)", n)
	}

	xcDir := findXcframeworkDir(pkgDir)
	if *pods {
		if err := writePodspec(pkgDir, xcDir, resolved); err != nil {
			return fmt.Errorf("generating podspec: %w", err)
		}
	}

	printNextSteps(pkgDir, resolved, *pods)
	return nil
}

// findXcframeworkDir returns the folder (relative to pkgDir) that contains the
// .xcframework bundles, for the podspec. Empty if none is found.
func findXcframeworkDir(pkgDir string) string {
	found := ""
	_ = filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if info.IsDir() && strings.HasSuffix(info.Name(), ".xcframework") {
			rel, _ := filepath.Rel(pkgDir, filepath.Dir(path))
			found = rel
			return filepath.SkipDir
		}
		return nil
	})
	return found
}

// writePodspec generates a TradelogSupport.podspec that vendors the xcframeworks.
func writePodspec(pkgDir, xcDir, version string) error {
	if xcDir == "" {
		xcDir = "."
	}
	spec := fmt.Sprintf(`Pod::Spec.new do |s|
  s.name             = 'TradelogSupport'
  s.version          = '%s'
  s.summary          = 'Tradelog support chat SDK (binary).'
  s.homepage         = 'https://tradelog.click'
  s.license          = { :type => 'Commercial' }
  s.author           = 'Tradelog'
  s.source           = { :http => 'file:' + __dir__ }
  s.platform         = :ios, '15.0'
  s.vendored_frameworks = '%s/*.xcframework'
end
`, version, xcDir)
	return os.WriteFile(filepath.Join(pkgDir, "TradelogSupport.podspec"), []byte(spec), 0o644)
}

func printNextSteps(pkgDir, version string, pods bool) {
	abs, _ := filepath.Abs(pkgDir)
	fmt.Printf("\n✓ SDK %s installed at %s\n\n", version, pkgDir)
	fmt.Println("NEXT — SwiftPM (recommended):")
	fmt.Println("  Xcode ▸ File ▸ Add Package Dependencies… ▸ Add Local… ▸ select:")
	fmt.Printf("    %s\n", abs)
	fmt.Println("  Then add the 'TradelogSupport' product to your target.")
	if pods {
		fmt.Println("\nNEXT — CocoaPods:")
		fmt.Println("  In your Podfile:")
		fmt.Printf("    pod 'TradelogSupport', :path => '%s'\n", pkgDir)
		fmt.Println("  Then: pod install")
	}
	fmt.Println("\nIN CODE:")
	fmt.Println("  import TradelogSupport")
	fmt.Println("  try TradeLogSdk.initialize(options: TradeLogSdkOptions(apiKey: \"tlk_…\", tenantId: \"…\", environment: .production))")
	fmt.Println("  // Present TradeLogSwiftUIContainer(onCloseRequested:onBackButtonRequested:) to open the chat")
	fmt.Println()
}
