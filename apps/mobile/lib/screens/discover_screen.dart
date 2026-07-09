import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../widgets/show_card.dart';

class DiscoverScreen extends StatefulWidget {
  const DiscoverScreen({super.key, required this.api});

  final ApiService api;

  @override
  State<DiscoverScreen> createState() => _DiscoverScreenState();
}

class _DiscoverScreenState extends State<DiscoverScreen> with SingleTickerProviderStateMixin {
  late final TabController _tabs;
  List<dynamic> shows = [];
  List<dynamic> movies = [];
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 2, vsync: this);
    _load();
  }

  Future<void> _load() async {
    setState(() => loading = true);
    final results = await Future.wait([widget.api.trending(), widget.api.trendingMovies()]);
    shows = results[0];
    movies = results[1];
    if (mounted) setState(() => loading = false);
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
          child: Text('کشف', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
        ),
        TabBar(
          controller: _tabs,
          tabs: const [Tab(text: 'سریال'), Tab(text: 'فیلم')],
        ),
        Expanded(
          child: loading
              ? const Center(child: CircularProgressIndicator())
              : TabBarView(
                  controller: _tabs,
                  children: [
                    _grid(shows, 'tv'),
                    _grid(movies, 'movie'),
                  ],
                ),
        ),
      ],
    );
  }

  Widget _grid(List<dynamic> items, String mediaType) {
    return GridView.builder(
      padding: const EdgeInsets.all(16),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 0.62,
        crossAxisSpacing: 12,
        mainAxisSpacing: 12,
      ),
      itemCount: items.length,
      itemBuilder: (_, index) => ShowCard(
        api: widget.api,
        item: {...items[index] as Map<String, dynamic>, 'media_type': mediaType},
      ),
    );
  }
}
