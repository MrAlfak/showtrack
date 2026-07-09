# One-time: point Flutter SDK included Gradle build at Myket mirror
# https://maven.myket.ir/services/flutter.html

$ErrorActionPreference = "Stop"

$flutterRoot = $null
if (Get-Command flutter -ErrorAction SilentlyContinue) {
  $flutterRoot = (flutter --version --machine 2>$null | ConvertFrom-Json).flutterRoot
}
if (-not $flutterRoot) {
  $candidates = @("C:\src\flutter", "$env:USERPROFILE\flutter", "$env:LOCALAPPDATA\flutter")
  foreach ($c in $candidates) {
    if (Test-Path "$c\packages\flutter_tools\gradle\settings.gradle.kts") {
      $flutterRoot = $c
      break
    }
  }
}
if (-not $flutterRoot) {
  Write-Error "Flutter SDK not found. Set PATH or install from https://pub.myket.ir/flutter_infra_release/releases/stable/windows/"
}

$settingsFile = Join-Path $flutterRoot "packages\flutter_tools\gradle\settings.gradle.kts"
$content = @"
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.PREFER_SETTINGS)
    repositories {
        maven { url = uri("https://maven.myket.ir/") }
    }
}
"@

Set-Content -Path $settingsFile -Value $content -Encoding UTF8
Write-Host "Patched: $settingsFile"
