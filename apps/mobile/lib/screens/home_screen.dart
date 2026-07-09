import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../widgets/show_card.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key, required this.api});

  final ApiService api;

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  List<dynamic> trending = [];
  List<dynamic> library = [];
  bool loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => loading = true);
    final results = await Future.wait([
      widget.api.trending(),
      if (widget.api.isAuthenticated) widget.api.getDashboard() else Future.value(null),
    ]);
    trending = results[0] as List<dynamic>;
    final dashboard = results.length > 1 ? results[1] as Map<String, dynamic>? : null;
    library = dashboard?['library'] as List<dynamic>? ?? [];
    if (mounted) setState(() => loading = false);
  }

  @override
  Widget build(BuildContext context) {
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text('شوتِرک', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          const SizedBox(height: 4),
          Text('خوش آمدید', style: TextStyle(color: Colors.grey.shade400)),
          const SizedBox(height: 24),
          const Text('ادامه تماشا', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          if (loading)
            const Center(child: CircularProgressIndicator())
          else if (library.isEmpty)
            Text('کتابخانه خالی است.', style: TextStyle(color: Colors.grey.shade500))
          else
            GridView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                childAspectRatio: 0.62,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
              ),
              itemCount: library.length.clamp(0, 6),
              itemBuilder: (_, index) => ShowCard(
                api: widget.api,
                item: library[index] as Map<String, dynamic>,
              ),
            ),
          const SizedBox(height: 24),
          const Text('پرطرفدار', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          SizedBox(
            height: 168,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: trending.length,
              separatorBuilder: (_, __) => const SizedBox(width: 12),
              itemBuilder: (_, index) => ShowCard(
                api: widget.api,
                compact: true,
                item: {...trending[index] as Map<String, dynamic>, 'media_type': 'tv'},
              ),
            ),
          ),
        ],
      ),
    );
  }
}
