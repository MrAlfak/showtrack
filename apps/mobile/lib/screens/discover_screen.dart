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
  List<dynamic> genres = [];
  int? selectedGenre;
  List<dynamic> items = [];
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 2, vsync: this);
    _tabs.addListener(() {
      if (!_tabs.indexIsChanging) {
        setState(() => selectedGenre = null);
        _load();
        _loadGenres();
      }
    });
    _loadGenres();
    _load();
  }

  Future<void> _loadGenres() async {
    final type = _tabs.index == 0 ? 'tv' : 'movie';
    final data = await widget.api.getGenres(type: type);
    if (mounted) setState(() => genres = data);
  }

  Future<void> _load() async {
    setState(() => loading = true);
    final type = _tabs.index == 0 ? 'tv' : 'movie';
    final data = await widget.api.discover(type: type, genreId: selectedGenre);
    if (mounted) setState(() {
      items = data;
      loading = false;
    });
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    const yellow = Color(0xFFFFD60A);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
          child: Text('کشف', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
        ),
        TabBar(
          controller: _tabs,
          indicatorColor: yellow,
          labelColor: yellow,
          tabs: const [Tab(text: 'سریال'), Tab(text: 'فیلم')],
        ),
        SizedBox(
          height: 48,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            children: [
              Padding(
                padding: const EdgeInsets.only(right: 8),
                child: FilterChip(
                  label: const Text('همه'),
                  selected: selectedGenre == null,
                  onSelected: (_) {
                    setState(() => selectedGenre = null);
                    _load();
                  },
                ),
              ),
              ...genres.map((g) {
                final map = g as Map<String, dynamic>;
                final id = map['id'] as int;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(map['name'] as String? ?? ''),
                    selected: selectedGenre == id,
                    onSelected: (_) {
                      setState(() => selectedGenre = id);
                      _load();
                    },
                  ),
                );
              }),
            ],
          ),
        ),
        Expanded(
          child: loading
              ? const Center(child: CircularProgressIndicator())
              : GridView.builder(
                  padding: const EdgeInsets.all(16),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    childAspectRatio: 0.62,
                    crossAxisSpacing: 12,
                    mainAxisSpacing: 12,
                  ),
                  itemCount: items.length,
                  itemBuilder: (_, index) {
                    final type = _tabs.index == 0 ? 'tv' : 'movie';
                    return ShowCard(
                      api: widget.api,
                      item: {...items[index] as Map<String, dynamic>, 'media_type': type},
                    );
                  },
                ),
        ),
      ],
    );
  }
}
