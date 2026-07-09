import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../screens/discover_screen.dart';
import '../screens/home_screen.dart';
import '../screens/profile_screen.dart';
import '../screens/search_screen.dart';

class ShowTrackApp extends StatefulWidget {
  const ShowTrackApp({super.key, required this.api});

  final ApiService api;

  @override
  State<ShowTrackApp> createState() => _ShowTrackAppState();
}

class _ShowTrackAppState extends State<ShowTrackApp> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final pages = [
      HomeScreen(api: widget.api),
      DiscoverScreen(api: widget.api),
      SearchScreen(api: widget.api),
      ProfileScreen(api: widget.api),
    ];

    return MaterialApp(
      title: 'ShowTrack',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF8B5CF6),
          brightness: Brightness.dark,
        ),
        scaffoldBackgroundColor: const Color(0xFF0A0A0A),
        cardColor: const Color(0xFF141414),
        useMaterial3: true,
      ),
      locale: const Locale('fa', 'IR'),
      supportedLocales: const [Locale('fa', 'IR'), Locale('en', 'US')],
      home: Scaffold(
        body: SafeArea(child: pages[_index]),
        bottomNavigationBar: NavigationBar(
          selectedIndex: _index,
          onDestinationSelected: (value) => setState(() => _index = value),
          destinations: const [
            NavigationDestination(icon: Icon(Icons.home_outlined), label: 'خانه'),
            NavigationDestination(icon: Icon(Icons.explore_outlined), label: 'کشف'),
            NavigationDestination(icon: Icon(Icons.search), label: 'جستجو'),
            NavigationDestination(icon: Icon(Icons.person_outline), label: 'پروفایل'),
          ],
        ),
      ),
    );
  }
}
