param(
  [string]$ApiBaseUrl = "https://showtrack-api.firstdata.ir/api/v1",
  [switch]$PersistMirror
)

$ErrorActionPreference = "Stop"
$mobileDir = Join-Path $PSScriptRoot "..\apps\mobile"
Set-Location $mobileDir

# Myket mirror — https://maven.myket.ir/services/flutter.html
$env:PUB_HOSTED_URL = "https://pub.myket.ir"
$env:FLUTTER_STORAGE_BASE_URL = "https://pub.myket.ir"

if ($PersistMirror) {
  [Environment]::SetEnvironmentVariable("PUB_HOSTED_URL", "https://pub.myket.ir", "User")
  [Environment]::SetEnvironmentVariable("FLUTTER_STORAGE_BASE_URL", "https://pub.myket.ir", "User")
  Write-Host "Mirror env vars saved for current user (restart IDE/terminal)."
}

if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
  $flutterBin = "C:\src\flutter\bin\flutter.bat"
  if (Test-Path $flutterBin) {
    $env:Path = "C:\src\flutter\bin;" + $env:Path
  } else {
    Write-Error "Flutter SDK not found. Install from https://pub.myket.ir/flutter_infra_release/releases/stable/windows/"
  }
}

if (-not (Test-Path "android")) {
  flutter create . --platforms=android --org com.showtrack
}

flutter pub get
flutter build apk --dart-define=API_BASE_URL=$ApiBaseUrl
Write-Host "APK: build\app\outputs\flutter-apk\app-release.apk"
