# Integrar el chat de soporte de Tradelog en tu app iOS

Guía de principio a fin: instalar el CLI, descargar el SDK con tu API key e
integrar el chat por **SwiftPM** o **CocoaPods**.

- ⏱️ ~10 minutos
- 🔑 Solo necesitas tu **API key** y **tenant id** de Tradelog (no credenciales AWS)

---

## Requisitos

| | |
|---|---|
| macOS | con **Xcode 15+** |
| [Homebrew](https://brew.sh) | para instalar el CLI |
| iOS deployment target | **15.0** o superior |
| API key + tenant id | los obtienes en el CRM (ver paso 2) |

---

## Paso 1 — Instalar el CLI

```bash
brew install tradelog-sas/tap/tradelog
```

Verifica:

```bash
tradelog version
```

Para actualizarlo más adelante: `brew upgrade tradelog`.

---

## Paso 2 — Obtener tu API key y tenant id

1. Entra al **CRM de Tradelog**.
2. Ve a **API Keys**.
3. Crea una API key (o usa una existente) y **cópiala** — empieza con `tlk_`.
4. Tu **tenant id** (company id) es el identificador de tu cuenta, con formato
   `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

> La misma API key sirve para **instalar** (este CLI) y para **inicializar** el
> SDK en runtime.

---

## Paso 3 — Descargar el SDK

> ⚠️ **Corre el comando desde la carpeta raíz de tu proyecto iOS.** El SDK se
> descarga en `Tradelog/` **relativo a donde ejecutes el comando** — si lo corres
> en otro lado, ahí quedará el folder.

```bash
cd /ruta/a/tu/proyecto-ios
tradelog install --api-key tlk_xxxxxxxx --tenant 4be6e386-....
```

O usando variables de entorno (útil en CI):

```bash
export TRADELOG_API_KEY=tlk_xxxxxxxx
export TRADELOG_TENANT_ID=4be6e386-....
tradelog install
```

Esto deja un paquete local listo:

```
Tradelog/TradelogSupport/
├── Package.swift              # manifiesto SwiftPM
├── TradelogSupport.podspec    # (solo con --pods)
└── ios_sdk/…/*.xcframework    # frameworks binarios del SDK
```

### Flags

| Flag | Default | Descripción |
|------|---------|-------------|
| `--api-key` | `$TRADELOG_API_KEY` | Tu API key de Tradelog |
| `--tenant` | `$TRADELOG_TENANT_ID` | Tu tenant / company id |
| `--version` | `latest` | Versión del SDK a instalar |
| `--dest` | `Tradelog` | Carpeta destino |
| `--pods` | `false` | Genera también un `.podspec` para CocoaPods |

> 💡 Agrega `Tradelog/` a tu `.gitignore` si no quieres versionar los binarios,
> o consérvalo si prefieres builds reproducibles sin re-descargar.

---

## Paso 4 — Integrar el SDK

Elige **una** vía. Las dos usan los mismos archivos descargados.

### Opción A — Swift Package Manager (recomendado)

1. En Xcode: **File ▸ Add Package Dependencies…**
2. Abajo a la izquierda: **Add Local…**
3. Selecciona la carpeta `Tradelog/TradelogSupport`.
4. Agrega el producto **`TradelogSupport`** a tu target.

O en un `Package.swift` propio:

```swift
dependencies: [
    .package(path: "Tradelog/TradelogSupport")
],
targets: [
    .target(name: "MiApp", dependencies: [
        .product(name: "TradelogSupport", package: "TradelogSupport")
    ])
]
```

### Opción B — CocoaPods

Instala con `--pods` (paso 3) y en tu `Podfile`:

```ruby
pod 'TradelogSupport', :path => 'Tradelog/TradelogSupport'
```

Luego:

```bash
pod install
```

---

## Paso 5 — Usar el chat en tu código

### 1. Inicializa el SDK al arrancar la app

```swift
import TradelogSupport

@main
struct MiApp: App {
    init() {
        do {
            try TradeLogSdk.initialize(options: TradeLogSdkOptions(
                apiKey: "tlk_xxxxxxxx",
                tenantId: "4be6e386-....",
                environment: .production,          // o .staging
                enableLogs: true,
                officialModules: [.logger, .userInfo],
                initialCustomerName: "Juan",        // opcional
                initialCustomerData: [:],            // opcional
                uiConfigurationCacheDurationSeconds: 60
            ))
        } catch {
            print("TradeLog SDK init falló: \(error)")
        }
    }

    var body: some Scene {
        WindowGroup { ContentView() }
    }
}
```

### 2. Presenta el chat (SwiftUI)

```swift
import SwiftUI
import TradelogSupport

struct ContentView: View {
    @State private var showSupport = false

    var body: some View {
        Button("Abrir soporte") { showSupport = true }
            .sheet(isPresented: $showSupport) {
                TradeLogSwiftUIContainer(
                    onCloseRequested:      { showSupport = false },
                    onBackButtonRequested: { showSupport = false }
                )
                .ignoresSafeArea()
            }
    }
}
```

Eso es todo — el SDK levanta el chat de soporte completo.

---

## Actualizar el SDK

```bash
tradelog install                 # baja la última versión
# o fija una versión:
tradelog install --version 2026.508.85
```

Luego en Xcode limpia y reconstruye (SPM) o corre `pod install` (CocoaPods).

---

## Solución de problemas

| Síntoma | Causa / arreglo |
|---|---|
| `api key o tenant inválidos` | Revisa `--api-key` y `--tenant`. La key debe estar **activa** en el CRM. |
| El chat no abre / falla el init | Verifica que `apiKey`, `tenantId` y `environment` coincidan con los del CRM. |
| `Missing package product 'TradelogSupport'` (SPM) | Limpia DerivedData y vuelve a resolver paquetes en Xcode. |
| Descarga lenta | El SDK pesa ~100 MB (motor Flutter). Solo se baja al instalar/actualizar. |

---

## Cómo funciona

`tradelog install` descarga el **`TradelogSupport.xcframework`** (binario) y lo
deja como paquete local listo para SPM o CocoaPods.

Tu **API key controla la descarga** y el SDK la vuelve a validar en runtime antes
de abrir el chat. No necesitas credenciales de AWS en ningún momento.
