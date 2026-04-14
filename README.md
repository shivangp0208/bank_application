## Session Management:
As we know that we have token based authentication in our project but the token which we are generating are short lived access token for 15min which means user needs to log in every 15min which is very frustating for user, so to solve this we use the session mananagement and refresh token where we created a session table in db to store the session details of user and create a refresh token which is a long lived token which acts a key to renew the access token without user logging in

## gRPC
grpc is a remote procedure call framework which helps to communicate between different server written in any language
1. The client can execute a remote procedure on the server
2. The remote interaction code is handled by gRPC
3. The API & data structure code is automatically generated
4. supported multiple programming languages
### How it works?
1. Define API & data structure
    > The RPC and it's request/response structure are defined using protobuf which conatins the info about req and resp format and service name
2. Generate gRPC stubs
    > Generate code for the server and client in the language of your choice.
3. Implement the server 
    > Implement the RPC handler on the server side.
4. Use the client 
    > Use the generated client stubs to call the RPC on the server.
### 4 Types of gRPC
1. Unary gRPC: In this the client will sent a single req on which the server will work and give the response back immediately
2. Client streaming gRPC: In this client will send a streams of req to server and the client is expected to receive a single response from server
3. Server streaming gRPC: In this client will send a single req on which server will responds with stream of response
4. Biderctional streaming gRPC: In this client will send a stream of req and server will responds with providing stream of response
### gRPC Gateway
- gRPC gateway serves both GRPC and HTTP requests at once.
- it is a plugin of protobuf compiler that generates proxy codes from protobuf.
- It translates HTTP JSON calls to gRPC.
    - In process translation: only for unary
    - Seperate proxy server: both unary and streaming
- Write code once, serve both gRPC and HTTP req

    So when we write gRPC protobuf code it generates a gateway and grpc which make it to handle both grpc and http req at once, so when a http req came it goes towards the gateway and gateway convert this in grpc and sends back the required json response, while request comming to grpc it will be handle as per grpc request and binary response 