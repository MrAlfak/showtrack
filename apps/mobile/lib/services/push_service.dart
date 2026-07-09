import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import 'api_service.dart';

class PushService {
  PushService(this._api);

  final ApiService _api;
  bool _initialized = false;

  static bool get isConfigured {
    const projectId = String.fromEnvironment('FIREBASE_PROJECT_ID');
    return projectId.isNotEmpty;
  }

  Future<void> init() async {
    if (_initialized || !isConfigured || !_api.isAuthenticated) return;

    try {
      if (Firebase.apps.isEmpty) {
        await Firebase.initializeApp(
          options: const FirebaseOptions(
            apiKey: String.fromEnvironment('FIREBASE_API_KEY'),
            appId: String.fromEnvironment('FIREBASE_APP_ID'),
            messagingSenderId: String.fromEnvironment('FIREBASE_MESSAGING_SENDER_ID'),
            projectId: String.fromEnvironment('FIREBASE_PROJECT_ID'),
          ),
        );
      }

      final messaging = FirebaseMessaging.instance;
      await messaging.requestPermission();
      final token = await messaging.getToken();
      if (token != null) {
        await _api.registerDevice(token);
      }
      _initialized = true;
    } catch (e) {
      debugPrint('Push init skipped: $e');
    }
  }

  Future<void> reset() async {
    _initialized = false;
  }
}

@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  if (Firebase.apps.isEmpty && PushService.isConfigured) {
    await Firebase.initializeApp(
      options: const FirebaseOptions(
        apiKey: String.fromEnvironment('FIREBASE_API_KEY'),
        appId: String.fromEnvironment('FIREBASE_APP_ID'),
        messagingSenderId: String.fromEnvironment('FIREBASE_MESSAGING_SENDER_ID'),
        projectId: String.fromEnvironment('FIREBASE_PROJECT_ID'),
      ),
    );
  }
}
