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
- gRPC gateway is a reverse proxy which translates the incoming HTTP req to gRPC req. It is a plugin of protobuf compiler that generates proxy codes from protobuf.
- It translates HTTP JSON calls to gRPC.
    - In process translation: only for unary
    - Seperate proxy server: both unary and streaming
- Write code once, serve both gRPC and HTTP req

    So when we write gRPC protobuf code it generates a gateway and grpc which make it to handle both grpc and http req at once, so when a http req came it goes towards the gateway and gateway convert this in grpc and sends back the required json response, while request comming to grpc it will be handle as per grpc request and binary response \

The concepts of grpc came, when we use the REST api we uses the HTTP req to call the server to get a HTTP res, and for that we need to define the path or url to ask for response so to remove this and call the methods directly with their name the RPC(Remote Call Procedure) was introduced which calls the method directly with their name so no need for the urls. But behind the scene while calling the api using RPC the system uderhood uses the HTTP req and res, so it just creates an abstraction from which we just need to call the method name of the server like createUser(), loginUser().

Now this RPC can be implemented in many way like JSON RPC which sends and receive the req and res in JSON format, and from this google invented the gRPC which is a RPC framework, this was introduced to increase the performance in the req handling, as the grpc uses the protobuf which is a protocol buffers i.e language agnostic serialization framework.

As we know that the server auto serialize and deserialize the incoming req and outgoing res which means any JSON req comming to a server it first get serialize into their environment like for Java we use POJO, for GoLang we use structs, same for deserialization, and this JSON req and res took so much size over network which decrease the performance also it took a particular time for serailization and desrialization.

To solve this protobufs were intoduced in gRPC which uses the binary format for data transfers. 
> JSON data format:
```
{
    "name":"name",
    "email":"email" 
}
```
> Protobuf data format:
```
message User{
    string name = 1;
    string email = 2;
}
```
So protobuf uses the integer for keys which removes the repetation of keys which happens in JSON for a list of data which acquire a big size over network.

## gRPC Client
grpc client are the ones which make the grpc req to grpc server, now to make this req a grpc client requires either the proto files or the server reflection enabled from the grpc server side, without this it will not be able to make the valid call to the grpc server.
### Why curl cannot be used in gRPC?
1. HTTP/2 framing

    As we know that grpc uses HTTP/2 for transfering req and response, now HTTP/2 requires the framing of request in which each message is fixed with a 5-byte header (1 byte for the compression flag and 4 byte for the message length) and atlast the protobuf binary data

2. Binary Protobuf

    curl sends the request in plaintext format it has no way to serialize your input into Protobuf.

    So we use the grpcurl which allow us to make the grpc req to grpc server, but still the grpcurl requires the protobuf schema files or the server grpc reflection enabled to make the request. 