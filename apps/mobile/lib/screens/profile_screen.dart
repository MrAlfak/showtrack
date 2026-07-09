import 'package:flutter/material.dart';

import '../services/api_service.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key, required this.api});

  final ApiService api;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _name = TextEditingController();
  bool registerMode = false;
  Map<String, dynamic>? stats;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _name.dispose();
    super.dispose();
  }

  Future<void> _refreshStats() async {
    if (!widget.api.isAuthenticated) return;
    final dashboard = await widget.api.getDashboard();
    if (mounted) setState(() => stats = dashboard?['stats'] as Map<String, dynamic>?);
  }

  @override
  void initState() {
    super.initState();
    _refreshStats();
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
                await _refreshStats();
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
                _stat('تماشاشده', '${stats!['episodes']}'),
                _stat('ساعت', '${stats!['hours']}'),
              ],
            ),
          const SizedBox(height: 16),
          OutlinedButton(
            onPressed: () async {
              await widget.api.clearToken();
              if (mounted) setState(() => stats = null);
            },
            child: const Text('خروج'),
          ),
        ],
      ],
    );
  }

  Widget _stat(String label, String value) {
    return Chip(label: Text('$label: $value'));
  }
}
