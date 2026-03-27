package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

	// registering all custom made validators in gin
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validatorMap := NewRequestValidator().Validator()
		for tag, fn := range validatorMap {
			v.RegisterValidation(tag, fn)
		}
	}

	server.SetupRoute()
	return server
}

// SetupRoute func helps us to setup routest easily and initailize the
// gin.Engine back to Server struct
func (s *Server) SetupRoute() {

	router := gin.Default()

	// defining accounts routes
	router.POST("/api/v1/accounts", s.CreateAccount)
	router.GET("/api/v1/accounts/:id", s.GetAccountByID)
	router.GET("/api/v1/accounts", s.GetAllAccount)
	router.PUT("/api/v1/accounts/:id", s.UpdateAccount)
	router.DELETE("/api/v1/accounts/:id", s.DeleteAccount)

	// defining transfer routes
	router.POST("/api/v1/transfer", s.TransferMoney)

	// defining users routes
	router.POST("/api/v1/users", s.CreateUser)

	s.router = router
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
