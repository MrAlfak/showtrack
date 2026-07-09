import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import 'app.dart';
import 'services/api_service.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final api = ApiService(http.Client());
  await api.loadToken();
  runApp(ShowTrackApp(api: api));
}
