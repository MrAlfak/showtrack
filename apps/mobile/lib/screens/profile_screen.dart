import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../services/push_service.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key, required this.api, required this.push});

  final ApiService api;
  final PushService push;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _name = TextEditingController();
  bool registerMode = false;
  Map<String, dynamic>? stats;
  List<dynamic> library = [];
  String libraryTab = 'watching';

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _name.dispose();
    super.dispose();
  }

  Future<void> _refresh({String? tab}) async {
    if (!widget.api.isAuthenticated) return;
    final activeTab = tab ?? libraryTab;
    final dashboard = await widget.api.getDashboard();
    final lib = await widget.api.getLibrary(listStatus: activeTab);
    if (mounted) {
      setState(() {
        stats = dashboard?['stats'] as Map<String, dynamic>?;
        library = [
          ...(lib?['shows'] as List<dynamic>? ?? []),
          ...(lib?['movies'] as List<dynamic>? ?? []),
        ];
      });
    }
  }

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        const Text('پروفایل', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
        const SizedBox(height: 16),
        if (!widget.api.isAuthenticated) ...[
          TextField(controller: _name, decoration: const InputDecoration(labelText: 'نام نمایشی')),
          TextField(controller: _email, decoration: const InputDecoration(labelText: 'ایمیل')),
          TextField(controller: _password, obscureText: true, decoration: const InputDecoration(labelText: 'رمز')),
          const SizedBox(height: 12),
          FilledButton(
            onPressed: () async {
              final response = registerMode
                  ? await widget.api.register(_email.text, _password.text, _name.text)
                  : await widget.api.login(_email.text, _password.text);
              final token = response?['token'] as String?;
              if (token != null) {
                await widget.api.saveToken(token);
                await widget.push.init();
                await _refresh();
                if (mounted) setState(() {});
              }
            },
            child: Text(registerMode ? 'ثبت‌نام' : 'ورود'),
          ),
          TextButton(
            onPressed: () => setState(() => registerMode = !registerMode),
            child: Text(registerMode ? 'حساب دارید؟ ورود' : 'حساب ندارید؟ ثبت‌نام'),
          ),
        ] else ...[
          if (stats != null)
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                _stat('سریال', '${stats!['shows']}'),
                _stat('فیلم', '${stats!['movies']}'),
                _stat('قسمت', '${stats!['episodes']}'),
                _stat('ساعت', '${stats!['hours']}'),
                _stat('پیاپی', '${stats!['streak']}d'),
              ],
            ),
          const SizedBox(height: 16),
          if (PushService.isConfigured)
            Card(
              child: ListTile(
                leading: const Icon(Icons.notifications_outlined),
                title: const Text('اعلان قسمت جدید'),
                subtitle: const Text('بعد از ورود، دستگاه ثبت می‌شود'),
                trailing: IconButton(
                  icon: const Icon(Icons.refresh),
                  onPressed: () => widget.push.init(),
                ),
              ),
            ),
          const SizedBox(height: 16),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: ApiService.listStatuses.map((tab) {
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(_tabLabel(tab)),
                    selected: libraryTab == tab,
                    onSelected: (_) async {
                      setState(() => libraryTab = tab);
                      await _refresh(tab: tab);
                    },
                  ),
                );
              }).toList(),
            ),
          ),
          const SizedBox(height: 8),
          ...library.map((item) {
            final map = item as Map<String, dynamic>;
            return ListTile(
              title: Text(map['title'] as String? ?? ''),
              subtitle: Text(map['media_type'] as String? ?? 'tv'),
              trailing: Text('${(map['progress'] as num?)?.round() ?? 0}%'),
            );
          }),
          const SizedBox(height: 16),
          OutlinedButton(
            onPressed: () async {
              await widget.api.clearToken();
              await widget.push.reset();
              if (mounted) setState(() {
                stats = null;
                library = [];
              });
            },
            child: const Text('خروج'),
          ),
        ],
      ],
    );
  }

  Widget _stat(String label, String value) => Chip(label: Text('$label: $value'));

  String _tabLabel(String tab) => switch (tab) {
        'watching' => 'تماشا',
        'plan_to_watch' => 'بعداً',
        'watched' => 'دیده',
        'dropped' => 'رها',
        'archived' => 'آرشیو',
        _ => tab,
      };
}
