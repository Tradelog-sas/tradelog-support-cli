# TradelogCLI

Instalador del **SDK de soporte de Tradelog para iOS**. El cliente se autentica
con su **API key de Tradelog** (la misma del runtime) y el CLI descarga el
`TradelogSupport.xcframework` binario listo para integrar por **SwiftPM** o
**CocoaPods** — sin credenciales AWS, sin configurar registries.

> 📖 **Guía completa de integración:** [docs/getting-started.md](docs/getting-started.md)

## Instalación

```bash
brew install tradelog-sas/tap/tradelog
```

## Uso

```bash
tradelog install --api-key tlk_xxx --tenant 4be6e386-...
# o con variables de entorno:
export TRADELOG_API_KEY=tlk_xxx
export TRADELOG_TENANT_ID=4be6e386-...
tradelog install
```

Flags:

| Flag | Default | Descripción |
|------|---------|-------------|
| `--api-key` | `$TRADELOG_API_KEY` | API key de Tradelog |
| `--tenant` | `$TRADELOG_TENANT_ID` | Tenant / company id |
| `--version` | `latest` | Versión del SDK |
| `--dest` | `Tradelog` | Carpeta destino |
| `--pods` | `false` | Genera también un `.podspec` |
| `--broker-url` | broker prod | Override (avanzado) |

El comando deja un Swift package local en `Tradelog/TradelogSupport/`.

### SwiftPM

Xcode ▸ **File ▸ Add Package Dependencies… ▸ Add Local…** ▸ selecciona
`Tradelog/TradelogSupport` ▸ agrega el producto `TradelogSupport` a tu target.

### CocoaPods (`--pods`)

```ruby
pod 'TradelogSupport', :path => 'Tradelog/TradelogSupport'
```

## En código

```swift
import TradelogSupport

try TradeLogSdk.initialize(options: TradeLogSdkOptions(
    apiKey: "tlk_…", tenantId: "…", environment: .production))

// Presenta el chat:
TradeLogSwiftUIContainer(
    onCloseRequested: { … },
    onBackButtonRequested: { … })
```

## Cómo funciona

```
tradelog install
  → broker (Basic auth: tenant + api key)      ← valida la api key
  → token de CodeArtifact (efímero)
  → GET <registry>/tradelog/TradelogSupport/<version>.zip  (Bearer)
  → extrae los .xcframework a ./Tradelog/TradelogSupport
```

La api key controla la descarga (gate de install). En runtime, el SDK vuelve a
validar la api key. Sin AWS del lado del cliente.

## Desarrollo

```bash
make build       # binario local
make test        # vet + tests
make release     # cross-compila macOS arm64/amd64 + tar.gz + shasums (Homebrew)
```
