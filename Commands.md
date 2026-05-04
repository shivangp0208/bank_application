## Mockgen Command
`mockgen -package mockdb -destination db/mock/store.go github.com/shivangp0208/bank_application/db/sqlc Store`

## Proto Command
`protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative --grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative --openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank --experimental_allow_proto3_optional=true proto/*.proto`

## Cleanup Users
`DELETE FROM sessions;`
`DELETE FROM verify_emails;`
`DELETE FROM users;`