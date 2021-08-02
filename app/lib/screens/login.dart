import 'package:flutter/material.dart';

class Login extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            margin: EdgeInsets.only(bottom: 12),
            child: TextField(
              decoration: InputDecoration(
                  hintText: 'Secret Code',
                  hintStyle: TextStyle(
                    fontSize: 12,
                  )),
            ),
          ),
          ElevatedButton(
            onPressed: () => {},
            child: Text('Login'),
          ),
          Container(
              margin: EdgeInsets.only(top: 24),
              child: Column(
                children: [
                  Text('or'),
                  ElevatedButton(
                    onPressed: () => {},
                    child: Text('SignUp'),
                  )
                ],
              ))
        ],
      ),
    );
  }
}
