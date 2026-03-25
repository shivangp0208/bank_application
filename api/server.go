package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/shivangp0208/bank_application/db/sqlc"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	// defining a server with sb configuration
	server := &Server{
		store: store,
	}
	router := gin.Default()

	// defining routes
	router.POST("/api/v1/accounts", server.CreateAccount)
	router.GET("/api/v1/accounts/:id", server.GetAccountByID)
	router.GET("/api/v1/accounts", server.GetAllAccount)
	router.PUT("/api/v1/accounts/:id", server.UpdateAccount)
	router.DELETE("/api/v1/accounts/:id", server.DeleteAccount)

	server.router = router
	return server
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
