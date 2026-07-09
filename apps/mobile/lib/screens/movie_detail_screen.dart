import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../services/api_service.dart';

class MovieDetailScreen extends StatefulWidget {
  const MovieDetailScreen({super.key, required this.api, required this.tmdbId});

  final ApiService api;
  final String tmdbId;

  @override
  State<MovieDetailScreen> createState() => _MovieDetailScreenState();
}

class _MovieDetailScreenState extends State<MovieDetailScreen> {
  Map<String, dynamic>? movie;
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final data = await widget.api.getMovie(widget.tmdbId);
    if (mounted) setState(() {
      movie = data;
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (movie == null) {
      return const Scaffold(body: Center(child: Text('فیلم پیدا نشد')));
    }

    final inLibrary = movie!['in_library'] == true;
    final watched = movie!['watched'] == true;

    return Scaffold(
      appBar: AppBar(title: Text(movie!['title'] as String? ?? '')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: CachedNetworkImage(
              imageUrl: movie!['poster_url'] as String? ?? '',
              height: 220,
              width: double.infinity,
              fit: BoxFit.cover,
            ),
          ),
          const SizedBox(height: 16),
          if (widget.api.isAuthenticated) ...[
            FilledButton(
              onPressed: () async {
                final tmdbId = movie!['tmdb_id'] as int? ?? int.tryParse(widget.tmdbId);
                if (tmdbId != null && !inLibrary) await widget.api.addMovie(tmdbId);
                await _load();
              },
              child: Text(inLibrary ? 'در کتابخانه' : 'افزودن به کتابخانه'),
            ),
            const SizedBox(height: 8),
            OutlinedButton(
              onPressed: inLibrary
                  ? () async {
                      final movieId = movie!['id'] as int;
                      await widget.api.markMovieWatched(movieId);
                      await _load();
                    }
                  : null,
              child: Text(watched ? 'تماشا شده' : 'علامت تماشا'),
            ),
          ],
          const SizedBox(height: 12),
          Text(movie!['overview'] as String? ?? '', style: TextStyle(color: Colors.grey.shade400)),
        ],
      ),
    );
  }
}
