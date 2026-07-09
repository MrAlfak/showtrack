# ShowTrack Mobile (Flutter Android)

## Setup

```powershell
# از ریشه پروژه — mirror مایکت برای ایران
.\scripts\setup-flutter.ps1

cd apps\mobile
flutter pub get
```

## Run (dev)

```powershell
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
```

`10.0.2.2` = localhost از داخل emulator اندروید.

## Build APK

```powershell
flutter build apk --release `
  --dart-define=API_BASE_URL=https://showtrack-api.firstdata.ir/api/v1
```

خروجی: `build/app/outputs/flutter-apk/app-release.apk`

## Push (FCM) — اختیاری

Firebase Console → پروژه → Android app (`com.showtrack.showtrack_mobile`):

1. `google-services.json` را در `android/app/` بگذار
2. build با dart-define:

```powershell
flutter build apk --release `
  --dart-define=API_BASE_URL=https://showtrack-api.firstdata.ir/api/v1 `
  --dart-define=FIREBASE_PROJECT_ID=your-project `
  --dart-define=FIREBASE_API_KEY=... `
  --dart-define=FIREBASE_APP_ID=... `
  --dart-define=FIREBASE_MESSAGING_SENDER_ID=...
```

بدون این مقادیر، اپ کار می‌کند ولی push ثبت نمی‌شود.

## Features

- ورود / ثبت‌نام
- کتابخانه با وضعیت‌های TV Time
- mark/unmark قسمت و فیلم
- حذف از کتابخانه
- صفحه بازیگر
- تم تیره + زرد TV Time
