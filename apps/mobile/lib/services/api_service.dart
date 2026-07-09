import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import '../config/api_config.dart';

class ApiService {
  ApiService(this._client);

  final http.Client _client;
  String? _token;

  static const _tokenKey = 'showtrack.token';

  Future<void> loadToken() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString(_tokenKey);
  }

  Future<void> saveToken(String token) async {
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
  }

  Future<void> clearToken() async {
    _token = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
  }

  bool get isAuthenticated => _token != null && _token!.isNotEmpty;

  Map<String, String> _headers({bool auth = false}) {
    final headers = {'Content-Type': 'application/json'};
    if (auth && _token != null) {
      headers['Authorization'] = 'Bearer $_token';
    }
    return headers;
  }

  Future<Map<String, dynamic>?> getJson(String path, {bool auth = false}) async {
    final response = await _client.get(
      Uri.parse('${ApiConfig.baseUrl}$path'),
      headers: _headers(auth: auth),
    );
    if (response.statusCode >= 400) return null;
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<List<dynamic>> getList(String path, {bool auth = false}) async {
    final data = await getJson(path, auth: auth);
    if (data == null) return [];
    final results = data['results'];
    if (results is List) return results;
    final items = data['items'];
    if (items is List) return items;
    return [];
  }

  Future<Map<String, dynamic>?> postJson(
    String path,
    Map<String, dynamic> body, {
    bool auth = false,
  }) async {
    final response = await _client.post(
      Uri.parse('${ApiConfig.baseUrl}$path'),
      headers: _headers(auth: auth),
      body: jsonEncode(body),
    );
    if (response.statusCode >= 400) return null;
    if (response.body.isEmpty) return {};
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>?> deleteJson(String path, {bool auth = false}) async {
    final response = await _client.delete(
      Uri.parse('${ApiConfig.baseUrl}$path'),
      headers: _headers(auth: auth),
    );
    if (response.statusCode >= 400) return null;
    if (response.body.isEmpty) return {};
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>?> login(String email, String password) async {
    return postJson('/auth/login', {'email': email, 'password': password});
  }

  Future<Map<String, dynamic>?> register(String email, String password, String displayName) async {
    return postJson('/auth/register', {
      'email': email,
      'password': password,
      'display_name': displayName,
    });
  }

  Future<List<dynamic>> search(String query) => getList('/search?q=${Uri.encodeQueryComponent(query)}');

  Future<List<dynamic>> trending() => getList('/trending');

  Future<List<dynamic>> trendingMovies() => getList('/trending/movies');

  Future<Map<String, dynamic>?> getShow(String id) => getJson('/shows/$id', auth: isAuthenticated);

  Future<Map<String, dynamic>?> getMovie(String id) => getJson('/movies/$id', auth: isAuthenticated);

  Future<Map<String, dynamic>?> getDashboard() => getJson('/me/dashboard', auth: true);

  Future<Map<String, dynamic>?> addShow(int tmdbId) =>
      postJson('/shows', {'tmdb_id': tmdbId}, auth: true);

  Future<Map<String, dynamic>?> addMovie(int tmdbId) =>
      postJson('/movies', {'tmdb_id': tmdbId}, auth: true);

  Future<Map<String, dynamic>?> markWatched(int episodeId) =>
      postJson('/episodes/$episodeId/watched', {}, auth: true);

  Future<Map<String, dynamic>?> markMovieWatched(int movieId) =>
      postJson('/movies/$movieId/watched', {}, auth: true);

  Future<Map<String, dynamic>?> registerDevice(String fcmToken) => postJson(
        '/devices',
        {'token': fcmToken, 'platform': 'android', 'app_version': '1.0.0'},
        auth: true,
      );
}
