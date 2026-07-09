import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../screens/movie_detail_screen.dart';
import '../screens/show_detail_screen.dart';

class ShowCard extends StatelessWidget {
  const ShowCard({super.key, required this.api, required this.item, this.compact = false});

  final ApiService api;
  final Map<String, dynamic> item;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final mediaType = item['media_type'] as String? ?? 'tv';
    final tmdbId = '${item['tmdb_id'] ?? item['id']}';
    final title = item['title'] as String? ?? '';
    final poster = item['poster_url'] as String? ?? '';

    return GestureDetector(
      onTap: () {
        final screen = mediaType == 'movie'
            ? MovieDetailScreen(api: api, tmdbId: tmdbId)
            : ShowDetailScreen(api: api, tmdbId: tmdbId);
        Navigator.of(context).push(MaterialPageRoute(builder: (_) => screen));
      },
      child: SizedBox(
        width: compact ? 112 : null,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            AspectRatio(
              aspectRatio: 2 / 3,
              child: ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: CachedNetworkImage(
                  imageUrl: poster,
                  fit: BoxFit.cover,
                  placeholder: (_, __) => Container(color: Colors.grey.shade900),
                  errorWidget: (_, __, ___) => Container(color: Colors.grey.shade900),
                ),
              ),
            ),
            if (!compact) ...[
              const SizedBox(height: 8),
              Text(title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13)),
            ],
          ],
        ),
      ),
    );
  }
}
