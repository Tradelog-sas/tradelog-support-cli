# Integrate the Tradelog support chat into your iOS app

End-to-end guide: install the CLI, download the SDK with your API key, and
integrate the chat via **SwiftPM** or **CocoaPods**.

- ⏱️ ~10 minutes
- 🔑 You only need your Tradelog **API key** and **tenant id** (no AWS credentials)

---

## Requirements

| | |
|---|---|
| macOS | with **Xcode 15+** |
| [Homebrew](https://brew.sh) | to install the CLI |
| iOS deployment target | **15.0** or higher |
| API key + tenant id | get them from the CRM (see step 2) |

---

## Step 1 — Install the CLI

```bash
brew install tradelog-sas/tap/tradelog
```

Verify:

```bash
tradelog version
```

To update it later: `brew upgrade tradelog`.

---

## Step 2 — Get your API key and tenant id

1. Open the **Tradelog CRM**.
2. Go to **API Keys**.
3. Create an API key (or use an existing one) and **copy it** — it starts with `tlk_`.
4. Your **tenant id** (company id) is your account identifier, in the format
   `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.

> The same API key is used to **install** (this CLI) and to **initialize** the SDK
> at runtime.

---

## Step 3 — Download the SDK

> ⚠️ **Run the command from your iOS project's root folder.** The SDK is downloaded
> into `Tradelog/` **relative to where you run the command** — if you run it
> somewhere else, that's where the folder ends up.

```bash
cd /path/to/your-ios-project
tradelog install --api-key tlk_xxxxxxxx --tenant 4be6e386-....
```

Or using environment variables (handy in CI):

```bash
export TRADELOG_API_KEY=tlk_xxxxxxxx
export TRADELOG_TENANT_ID=4be6e386-....
tradelog install
```

This leaves a ready-to-use local package:

```
Tradelog/TradelogSupport/
├── Package.swift              # SwiftPM manifest
├── TradelogSupport.podspec    # (only with --pods)
└── ios_sdk/…/*.xcframework    # binary SDK frameworks
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` | `$TRADELOG_API_KEY` | Your Tradelog API key |
| `--tenant` | `$TRADELOG_TENANT_ID` | Your tenant / company id |
| `--version` | `latest` | SDK version to install |
| `--dest` | `Tradelog` | Destination folder |
| `--pods` | `false` | Also generate a `.podspec` for CocoaPods |

> 💡 Add `Tradelog/` to your `.gitignore` if you don't want to version the
> binaries, or keep it for reproducible builds without re-downloading.

---

## Step 4 — Integrate the SDK

Pick **one** path. Both use the same downloaded files.

### Option A — Swift Package Manager (recommended)

1. In Xcode: **File ▸ Add Package Dependencies…**
2. Bottom left: **Add Local…**
3. Select the `Tradelog/TradelogSupport` folder.
4. Add the **`TradelogSupport`** product to your target.

Or in your own `Package.swift`:

```swift
dependencies: [
    .package(path: "Tradelog/TradelogSupport")
],
targets: [
    .target(name: "MyApp", dependencies: [
        .product(name: "TradelogSupport", package: "TradelogSupport")
    ])
]
```

### Option B — CocoaPods

Install with `--pods` (step 3) and in your `Podfile`:

```ruby
pod 'TradelogSupport', :path => 'Tradelog/TradelogSupport'
```

Then:

```bash
pod install
```

---

## Step 5 — Use the chat in your code

### 1. Initialize the SDK at app startup

```swift
import TradelogSupport

@main
struct MyApp: App {
    init() {
        do {
            try TradeLogSdk.initialize(options: TradeLogSdkOptions(
                apiKey: "tlk_xxxxxxxx",
                tenantId: "4be6e386-....",
                environment: .production,          // or .staging
                enableLogs: true,
                officialModules: [.logger, .userInfo],
                initialCustomerName: "Jane",         // optional
                initialCustomerData: [:],            // optional
                uiConfigurationCacheDurationSeconds: 60
            ))
        } catch {
            print("TradeLog SDK init failed: \(error)")
        }
    }

    var body: some Scene {
        WindowGroup { ContentView() }
    }
}
```

### 2. Present the chat (SwiftUI)

```swift
import SwiftUI
import TradelogSupport

struct ContentView: View {
    @State private var showSupport = false

    var body: some View {
        Button("Open support") { showSupport = true }
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

That's it — the SDK brings up the full support chat.

---

## Updating

**Update the CLI** (via Homebrew):

```bash
brew upgrade tradelog
```

**Update the SDK**:

```bash
tradelog install                 # download the latest version
# or pin a version:
tradelog install --version 2026.508.85
```

Then in Xcode clean and rebuild (SPM) or run `pod install` (CocoaPods).

---

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `invalid api key or tenant` | Check `--api-key` and `--tenant`. The key must be **active** in the CRM. |
| Chat doesn't open / init fails | Make sure `apiKey`, `tenantId` and `environment` match the CRM. |
| `Missing package product 'TradelogSupport'` (SPM) | Clean DerivedData and re-resolve packages in Xcode. |
| Slow download | The SDK is ~100 MB (Flutter engine). It's only downloaded on install/update. |

---

## How it works

`tradelog install` downloads the **`TradelogSupport.xcframework`** (binary) and
leaves it as a local package ready for SPM or CocoaPods.

Your **API key gates the download**, and the SDK re-validates it at runtime before
opening the chat. You never need AWS credentials.
