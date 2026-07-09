import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../services/push_service.dart';
import '../screens/discover_screen.dart';
import '../screens/feed_screen.dart';
import '../screens/home_screen.dart';
import '../screens/profile_screen.dart';
import '../screens/search_screen.dart';

class ShowTrackApp extends StatefulWidget {
  const ShowTrackApp({super.key, required this.api, required this.push});

  final ApiService api;
  final PushService push;

  @override
  State<ShowTrackApp> createState() => _ShowTrackAppState();
}

class _ShowTrackAppState extends State<ShowTrackApp> {
  int _index = 0;

  @override
  void initState() {
    super.initState();
    if (widget.api.isAuthenticated) {
      widget.push.init();
    }
  }

  @override
  Widget build(BuildContext context) {
    const yellow = Color(0xFFFFD60A);

    final pages = [
      HomeScreen(api: widget.api),
      FeedScreen(api: widget.api),
      DiscoverScreen(api: widget.api),
      SearchScreen(api: widget.api),
      ProfileScreen(api: widget.api, push: widget.push),
    ];

    return MaterialApp(
      title: 'ShowTrack',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        colorScheme: ColorScheme.fromSeed(seedColor: yellow, brightness: Brightness.dark),
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
          indicatorColor: yellow.withValues(alpha: 0.2),
          onDestinationSelected: (value) => setState(() => _index = value),
          destinations: const [
            NavigationDestination(icon: Icon(Icons.home_outlined), label: 'خانه'),
            NavigationDestination(icon: Icon(Icons.people_outline), label: 'فید'),
            NavigationDestination(icon: Icon(Icons.explore_outlined), label: 'کشف'),
            NavigationDestination(icon: Icon(Icons.search), label: 'جستجو'),
            NavigationDestination(icon: Icon(Icons.person_outline), label: 'پروفایل'),
          ],
        ),
      ),
    );
  }
}
