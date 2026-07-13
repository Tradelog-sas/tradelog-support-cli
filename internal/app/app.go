package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Usage imprime la ayuda general.
func Usage() {
	fmt.Print(`tradelog — instalador del SDK de soporte de Tradelog para iOS

USO:
  tradelog install [flags]     Descarga e integra TradelogSupport.xcframework
  tradelog version             Muestra la versión del CLI
  tradelog help                Muestra esta ayuda

INSTALL:
  Autentícate con tu API KEY de Tradelog (la misma que usas en el SDK). El CLI
  descarga el framework binario — no necesitas credenciales de AWS.

  Flags:
    --api-key   API key de Tradelog       (o env TRADELOG_API_KEY)
    --tenant    Tenant / company id        (o env TRADELOG_TENANT_ID)
    --version   Versión del SDK            (default: latest)
    --dest      Carpeta destino            (default: ./Tradelog)
    --pods      Genera también un .podspec para CocoaPods
    --broker-url  Override del broker      (avanzado)

  Ejemplo:
    tradelog install --api-key tlk_xxx --tenant 4be6e386-...

`)
}

// Install ejecuta el flujo completo: auth → descarga → extrae → configura.
func Install(args []string) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	apiKey := fs.String("api-key", os.Getenv("TRADELOG_API_KEY"), "API key de Tradelog")
	tenant := fs.String("tenant", os.Getenv("TRADELOG_TENANT_ID"), "Tenant / company id")
	version := fs.String("version", "latest", "Versión del SDK (o 'latest')")
	dest := fs.String("dest", "Tradelog", "Carpeta destino")
	pods := fs.Bool("pods", false, "Generar .podspec para CocoaPods")
	brokerURL := fs.String("broker-url", DefaultBrokerURL, "URL del broker de tokens")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *apiKey == "" {
		return fmt.Errorf("falta --api-key (o env TRADELOG_API_KEY)")
	}
	if *tenant == "" {
		return fmt.Errorf("falta --tenant (o env TRADELOG_TENANT_ID)")
	}

	fmt.Println("→ Autenticando con tu API key…")
	tok, err := fetchToken(*brokerURL, *tenant, *apiKey)
	if err != nil {
		return err
	}

	resolved, err := resolveVersion(tok.RegistryEndpoint, tok.AuthorizationToken, *version)
	if err != nil {
		return err
	}
	fmt.Printf("→ Versión: %s\n", resolved)

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
	fmt.Printf("→ Extrayendo en %s …\n", pkgDir)
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
		return fmt.Errorf("el paquete extraído no tiene Package.swift (%d archivos)", n)
	}

	xcDir := findXcframeworkDir(pkgDir)
	if *pods {
		if err := writePodspec(pkgDir, xcDir, resolved); err != nil {
			return fmt.Errorf("generando podspec: %w", err)
		}
	}

	printNextSteps(pkgDir, resolved, *pods)
	return nil
}

// findXcframeworkDir devuelve la carpeta (relativa a pkgDir) que contiene los
// .xcframework, para el podspec. Vacío si no se encuentra.
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

// writePodspec genera un TradelogSupport.podspec que vendoriza los xcframeworks.
func writePodspec(pkgDir, xcDir, version string) error {
	if xcDir == "" {
		xcDir = "."
	}
	spec := fmt.Sprintf(`Pod::Spec.new do |s|
  s.name             = 'TradelogSupport'
  s.version          = '%s'
  s.summary          = 'Tradelog support chat SDK (binario).'
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
	fmt.Printf("\n✓ SDK %s instalado en %s\n\n", version, pkgDir)
	fmt.Println("SIGUIENTE — SwiftPM (recomendado):")
	fmt.Println("  Xcode ▸ File ▸ Add Package Dependencies… ▸ Add Local… ▸ selecciona:")
	fmt.Printf("    %s\n", abs)
	fmt.Println("  Luego agrega el producto 'TradelogSupport' a tu target.")
	if pods {
		fmt.Println("\nSIGUIENTE — CocoaPods:")
		fmt.Println("  En tu Podfile:")
		fmt.Printf("    pod 'TradelogSupport', :path => '%s'\n", pkgDir)
		fmt.Println("  Luego: pod install")
	}
	fmt.Println("\nEN CÓDIGO:")
	fmt.Println("  import TradelogSupport")
	fmt.Println("  try TradeLogSdk.initialize(options: TradeLogSdkOptions(apiKey: \"tlk_…\", tenantId: \"…\", environment: .production))")
	fmt.Println("  // Presenta TradeLogSwiftUIContainer(onCloseRequested:onBackButtonRequested:) para abrir el chat")
	fmt.Println()
}
