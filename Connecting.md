## How the connection flow works
1. Get an app token through some other node
2. Send a request from the client with the connection token
3. Send a "setup" action to space-node with your account data (encrypted) -> UDP connection details
4. Just start sending voice packets, the server will register you automatically