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

  Future<void> _pickListStatus() async {
    final status = await showModalBottomSheet<String>(
      context: context,
      builder: (ctx) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: ApiService.listStatuses
              .map((s) => ListTile(title: Text(_statusLabel(s)), onTap: () => Navigator.pop(ctx, s)))
              .toList(),
        ),
      ),
    );
    if (status == null) return;
    final showId = show!['id'] as int;
    await widget.api.updateShowStatus(showId, status);
    await _load();
  }

  String _statusLabel(String status) => switch (status) {
        'watching' => 'در حال تماشا',
        'plan_to_watch' => 'برای تماشا',
        'watched' => 'تماشا شده',
        'dropped' => 'رها شده',
        'archived' => 'آرشیو',
        _ => status,
      };

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
    final showId = show!['id'] as int?;

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
          if (widget.api.isAuthenticated) ...[
            Row(
              children: [
                Expanded(
                  child: FilledButton(
                    onPressed: () async {
                      final tmdbId = show!['tmdb_id'] as int? ?? int.tryParse(widget.tmdbId);
                      if (tmdbId != null && !inLibrary) await widget.api.addShow(tmdbId);
                      await _load();
                    },
                    child: Text(inLibrary ? 'در کتابخانه' : 'افزودن به کتابخانه'),
                  ),
                ),
                if (inLibrary) ...[
                  const SizedBox(width: 8),
                  IconButton(
                    onPressed: _pickListStatus,
                    icon: const Icon(Icons.label_outline),
                    tooltip: 'وضعیت',
                  ),
                  IconButton(
                    onPressed: () async {
                      if (showId != null) {
                        await widget.api.removeShow(showId);
                        await _load();
                      }
                    },
                    icon: const Icon(Icons.delete_outline),
                    tooltip: 'حذف',
                  ),
                ],
              ],
            ),
          ],
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
                final epId = episode['id'] as int;
                return ListTile(
                  title: Text('E${episode['episode_number']} · ${episode['name']}'),
                  trailing: watched
                      ? IconButton(
                          icon: const Icon(Icons.check_circle, color: Colors.green),
                          onPressed: () async {
                            await widget.api.unmarkWatched(epId);
                            await _load();
                          },
                        )
                      : IconButton(
                          icon: const Icon(Icons.check),
                          onPressed: () async {
                            await widget.api.markWatched(epId);
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
