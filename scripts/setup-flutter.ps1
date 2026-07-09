param(
  [string]$ApiBaseUrl = "https://showtrack-api.firstdata.ir/api/v1"
)

$ErrorActionPreference = "Stop"
$mobileDir = Join-Path $PSScriptRoot "..\apps\mobile"
Set-Location $mobileDir

if (-not (Get-Command flutter -ErrorAction SilentlyContinue)) {
  Write-Error "Flutter SDK not found in PATH. Install from https://docs.flutter.dev/get-started/install"
}

if (-not (Test-Path "android")) {
  flutter create . --platforms=android --org com.showtrack
}

flutter pub get
flutter run --dart-define=API_BASE_URL=$ApiBaseUrl
