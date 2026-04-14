## Mockgen Command
`mockgen -package mockdb db/mock/store.go github.com/shivangp0208/bank_application/db/sqlc Store`

## Proto Command
`protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative proto/*.proto`