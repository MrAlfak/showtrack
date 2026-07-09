import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../services/api_service.dart';

class ShowDetailScreen extends StatefulWidget {
  const ShowDetailScreen({super.key, required this.api, required this.tmdbId});

  final ApiService api;
  final String tmdbId;

  @override
  State<ShowDetailScreen> createState() => _ShowDetailScreenState();
}

class _ShowDetailScreenState extends State<ShowDetailScreen> {
  Map<String, dynamic>? show;
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final data = await widget.api.getShow(widget.tmdbId);
    if (mounted) setState(() {
      show = data;
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (show == null) {
      return const Scaffold(body: Center(child: Text('سریال پیدا نشد')));
    }

    final seasons = show!['seasons'] as List<dynamic>? ?? [];
    final inLibrary = show!['in_library'] == true;

    return Scaffold(
      appBar: AppBar(title: Text(show!['title'] as String? ?? '')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: CachedNetworkImage(
              imageUrl: show!['poster_url'] as String? ?? '',
              height: 220,
              width: double.infinity,
              fit: BoxFit.cover,
            ),
          ),
          const SizedBox(height: 16),
          if (widget.api.isAuthenticated)
            FilledButton(
              onPressed: () async {
                final tmdbId = show!['tmdb_id'] as int? ?? int.tryParse(widget.tmdbId);
                if (tmdbId != null && !inLibrary) await widget.api.addShow(tmdbId);
                await _load();
              },
              child: Text(inLibrary ? 'در کتابخانه' : 'افزودن به کتابخانه'),
            ),
          const SizedBox(height: 12),
          Text(show!['overview'] as String? ?? '', style: TextStyle(color: Colors.grey.shade400)),
          const SizedBox(height: 24),
          const Text('فصل‌ها', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          ...seasons.map((season) {
            final episodes = season['episodes'] as List<dynamic>? ?? [];
            return ExpansionTile(
              title: Text(season['name'] as String? ?? 'Season'),
              children: episodes.map((episode) {
                final watched = episode['watched'] == true;
                return ListTile(
                  title: Text('E${episode['episode_number']} · ${episode['name']}'),
                  trailing: watched
                      ? const Icon(Icons.check_circle, color: Colors.green)
                      : IconButton(
                          icon: const Icon(Icons.check),
                          onPressed: () async {
                            await widget.api.markWatched(episode['id'] as int);
                            await _load();
                          },
                        ),
                );
              }).toList(),
            );
          }),
        ],
      ),
    );
  }
}
