# ShowTrack Mobile (Flutter)

Android client for ShowTrack — connects to the self-hosted Go API.

## Prerequisites

- [Flutter SDK](https://docs.flutter.dev/get-started/install) 3.22+
- Android Studio or Android SDK

## First-time setup

From repo root:

```powershell
cd apps/mobile
flutter create . --platforms=android --org com.showtrack
flutter pub get
```

`flutter create` generates the `android/` folder. The `lib/` source is already included.

## Run (local API)

```powershell
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

Use your machine LAN IP instead of `10.0.2.2` on a physical device.

## Run (production)

```powershell
flutter run --dart-define=API_BASE_URL=https://showtrack-api.firstdata.ir/api/v1
```

## Build APK

```powershell
flutter build apk --release --dart-define=API_BASE_URL=https://showtrack-api.firstdata.ir/api/v1
```

Output: `build/app/outputs/flutter-apk/app-release.apk`

## Features

- Home dashboard (library + trending)
- Discover (TV + movies tabs)
- Search and add to library
- Show/movie detail with watch progress
- Profile auth + stats (Persian UI)

## Push notifications

Register FCM token via `POST /devices` after integrating `firebase_messaging` (optional next step).
