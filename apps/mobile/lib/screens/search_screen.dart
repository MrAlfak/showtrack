import 'package:flutter/material.dart';

import '../services/api_service.dart';

class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key, required this.api});

  final ApiService api;

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  final _controller = TextEditingController();
  List<dynamic> results = [];
  bool loading = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _search(String query) async {
    if (query.length < 2) {
      setState(() => results = []);
      return;
    }
    setState(() => loading = true);
    final data = await widget.api.search(query);
    if (mounted) setState(() {
      results = data;
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('جستجو', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          const SizedBox(height: 12),
          TextField(
            controller: _controller,
            decoration: const InputDecoration(
              hintText: 'Breaking Bad، Matrix...',
              border: OutlineInputBorder(),
            ),
            onChanged: _search,
          ),
          const SizedBox(height: 16),
          if (loading) const LinearProgressIndicator(),
          Expanded(
            child: ListView.separated(
              itemCount: results.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (_, index) {
                final item = results[index] as Map<String, dynamic>;
                final mediaType = item['media_type'] as String? ?? 'tv';
                return ListTile(
                  leading: CircleAvatar(
                    backgroundImage: NetworkImage(item['poster_url'] as String? ?? ''),
                  ),
                  title: Text(item['title'] as String? ?? ''),
                  subtitle: Text(mediaType),
                  trailing: widget.api.isAuthenticated && mediaType != 'person'
                      ? IconButton(
                          icon: const Icon(Icons.add),
                          onPressed: () async {
                            final id = item['id'] as int;
                            if (mediaType == 'movie') {
                              await widget.api.addMovie(id);
                            } else {
                              await widget.api.addShow(id);
                            }
                            if (context.mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(content: Text('به کتابخانه اضافه شد')),
                              );
                            }
                          },
                        )
                      : null,
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
