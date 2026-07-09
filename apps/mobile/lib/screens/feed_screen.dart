import 'package:flutter/material.dart';

import '../services/api_service.dart';

class FeedScreen extends StatefulWidget {
  const FeedScreen({super.key, required this.api});

  final ApiService api;

  @override
  State<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends State<FeedScreen> {
  List<dynamic> _items = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    if (!widget.api.isAuthenticated) {
      setState(() => _loading = false);
      return;
    }
    final data = await widget.api.getFeed();
    if (!mounted) return;
    setState(() {
      _items = data?['items'] as List<dynamic>? ?? [];
      _loading = false;
    });
  }

  String _activityText(Map<String, dynamic> item) {
    final payload = item['payload'] as Map<String, dynamic>? ?? {};
    final name = (item['display_name'] ?? item['username'] ?? 'کاربر').toString();
    final title = (payload['title'] ?? 'عنوان').toString();
    switch (item['activity_type']) {
      case 'episode_watched':
        return '$name — $title S${payload['season_number']}E${payload['episode_number']}';
      case 'movie_watched':
        return '$name فیلم $title را دید';
      case 'show_added':
        return '$name سریال $title را اضافه کرد';
      case 'movie_added':
        return '$name فیلم $title را اضافه کرد';
      default:
        return '$name — $title';
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!widget.api.isAuthenticated) {
      return const Center(child: Text('برای فید وارد شوید'));
    }
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_items.isEmpty) {
      return const Center(child: Text('فید خالی است. دوستان را دنبال کنید.'));
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.separated(
        padding: const EdgeInsets.all(16),
        itemCount: _items.length,
        separatorBuilder: (_, __) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final item = _items[index] as Map<String, dynamic>;
          return Card(
            child: ListTile(
              title: Text(_activityText(item), style: const TextStyle(fontSize: 14)),
              subtitle: Text((item['created_at'] ?? '').toString()),
            ),
          );
        },
      ),
    );
  }
}
