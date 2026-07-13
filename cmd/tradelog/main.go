// Command tradelog is the Tradelog SDK installer.
//
// El cliente se autentica con su API KEY de Tradelog (la misma del runtime) y
// el CLI descarga el TradelogSupport.xcframework binario para integrarlo por
// SwiftPM o CocoaPods — sin credenciales AWS, sin configurar registries.
package main

import (
	"fmt"
	"os"

	"github.com/Tradelog-sas/tradelog-support-cli/internal/app"
)

// version se inyecta en build (-ldflags "-X main.version=...").
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		app.Usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "install":
		if err := app.Install(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "\n✖ %v\n", err)
			os.Exit(1)
		}
	case "version", "--version", "-v":
		fmt.Printf("tradelog %s\n", version)
	case "help", "--help", "-h":
		app.Usage()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %q\n\n", os.Args[1])
		app.Usage()
		os.Exit(2)
	}
}
