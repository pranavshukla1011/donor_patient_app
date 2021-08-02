import 'package:flutter/material.dart';
import './screens/login.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter Demo',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: MyHomePage(title: 'Donors and Patients'),
    );
  }
}

class MyHomePage extends StatelessWidget {
  final String title;

  MyHomePage({required this.title});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: Container(
        child: Login(),
        padding: EdgeInsets.symmetric(
          horizontal: 32,
        ),
      ),
    );
  }
}
