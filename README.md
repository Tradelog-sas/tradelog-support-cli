# TradelogCLI

Installer for the **Tradelog iOS support chat SDK**. Clients authenticate with
their **Tradelog API key** (the same one used at runtime) and the CLI downloads
the binary `TradelogSupport.xcframework`, ready to integrate via **SwiftPM** or
**CocoaPods** — no AWS credentials, no registry setup.

> 📖 **Full integration guide:** [docs/getting-started.md](docs/getting-started.md)

## Installation

```bash
brew install tradelog-sas/tap/tradelog
```

Update to the latest version:

```bash
brew upgrade tradelog
```

## Usage

```bash
tradelog install --api-key tlk_xxx --tenant 4be6e386-...
# or with environment variables:
export TRADELOG_API_KEY=tlk_xxx
export TRADELOG_TENANT_ID=4be6e386-...
tradelog install
```

> Run it from your iOS project's root folder — the SDK is downloaded into
> `Tradelog/` relative to your current directory.

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` | `$TRADELOG_API_KEY` | Tradelog API key |
| `--tenant` | `$TRADELOG_TENANT_ID` | Tenant / company id |
| `--version` | `latest` | SDK version |
| `--dest` | `Tradelog` | Destination folder |
| `--pods` | `false` | Also generate a `.podspec` |
| `--broker-url` | prod broker | Override (advanced) |

The command leaves a local Swift package at `Tradelog/TradelogSupport/`.

### SwiftPM

Xcode ▸ **File ▸ Add Package Dependencies… ▸ Add Local…** ▸ select
`Tradelog/TradelogSupport` ▸ add the `TradelogSupport` product to your target.

### CocoaPods (`--pods`)

```ruby
pod 'TradelogSupport', :path => 'Tradelog/TradelogSupport'
```

## In code

```swift
import TradelogSupport

try TradeLogSdk.initialize(options: TradeLogSdkOptions(
    apiKey: "tlk_…", tenantId: "…", environment: .production))

// Present the chat:
TradeLogSwiftUIContainer(
    onCloseRequested: { … },
    onBackButtonRequested: { … })
```

## How it works

`tradelog install` downloads the `TradelogSupport.xcframework` (binary) and leaves
it as a local package ready for SPM/CocoaPods. Your API key gates the download and
the SDK re-validates it at runtime. No AWS credentials required.

## Development

```bash
make build       # local binary
make test        # vet + tests
make release     # cross-compile macOS arm64/amd64 + tar.gz + shasums (Homebrew)
```
