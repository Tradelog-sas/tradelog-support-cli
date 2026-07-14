// Command tradelog is the Tradelog SDK installer.
//
// The client authenticates with their Tradelog API key (the same one used at
// runtime) and the CLI downloads the binary TradelogSupport.xcframework to
// integrate via SwiftPM or CocoaPods — no AWS credentials, no registry setup.
package main

import (
	"fmt"
	"os"

	"github.com/Tradelog-sas/tradelog-support-cli/internal/app"
)

// version is injected at build time (-ldflags "-X main.version=...").
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
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", os.Args[1])
		app.Usage()
		os.Exit(2)
	}
}
