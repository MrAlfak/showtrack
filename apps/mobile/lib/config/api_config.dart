class ApiConfig {
  static const defaultBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://showtrack-api.firstdata.ir/api/v1',
  );

  static String get baseUrl {
    const localOverride = String.fromEnvironment('API_BASE_URL');
    if (localOverride.isNotEmpty) return localOverride;
    return defaultBaseUrl;
  }
}
