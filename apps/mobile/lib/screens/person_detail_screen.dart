import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../services/api_service.dart';
import 'movie_detail_screen.dart';
import 'show_detail_screen.dart';

class PersonDetailScreen extends StatefulWidget {
  const PersonDetailScreen({super.key, required this.api, required this.tmdbId});

  final ApiService api;
  final String tmdbId;

  @override
  State<PersonDetailScreen> createState() => _PersonDetailScreenState();
}

class _PersonDetailScreenState extends State<PersonDetailScreen> {
  Map<String, dynamic>? person;
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final data = await widget.api.getPerson(widget.tmdbId);
    if (mounted) setState(() {
      person = data;
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (person == null) {
      return const Scaffold(body: Center(child: Text('بازیگر پیدا نشد')));
    }

    final credits = person!['credits'] as List<dynamic>? ?? [];

    return Scaffold(
      appBar: AppBar(title: Text(person!['name'] as String? ?? '')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Center(
            child: ClipRRect(
              borderRadius: BorderRadius.circular(80),
              child: CachedNetworkImage(
                imageUrl: person!['profile_url'] as String? ?? '',
                width: 120,
                height: 120,
                fit: BoxFit.cover,
              ),
            ),
          ),
          const SizedBox(height: 16),
          Text(person!['biography'] as String? ?? 'بیوگرافی موجود نیست.',
              style: TextStyle(color: Colors.grey.shade400)),
          const SizedBox(height: 24),
          const Text('فیلم‌نامه', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          ...credits.take(20).map((credit) {
            final item = credit as Map<String, dynamic>;
            final mediaType = item['media_type'] as String? ?? 'tv';
            return ListTile(
              leading: ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: CachedNetworkImage(
                  imageUrl: item['poster_url'] as String? ?? '',
                  width: 40,
                  height: 56,
                  fit: BoxFit.cover,
                ),
              ),
              title: Text(item['title'] as String? ?? ''),
              subtitle: Text(item['character'] as String? ?? ''),
              onTap: () {
                final tmdbId = '${item['tmdb_id']}';
                final screen = mediaType == 'movie'
                    ? MovieDetailScreen(api: widget.api, tmdbId: tmdbId)
                    : ShowDetailScreen(api: widget.api, tmdbId: tmdbId);
                Navigator.of(context).push(MaterialPageRoute(builder: (_) => screen));
              },
            );
          }),
        ],
      ),
    );
  }
}
