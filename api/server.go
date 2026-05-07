package api

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	db "github.com/shivangp0208/bank_application/db/sqlc"
	"github.com/shivangp0208/bank_application/token"
	"github.com/shivangp0208/bank_application/util"
	"github.com/shivangp0208/bank_application/worker"
)

type Server struct {
	Store        db.Store
	TokenMaker   token.Maker
	Config       *util.Config
	Router       *gin.Engine
	TaskProducer worker.TaskProducer
}

func NewServer(store db.Store, config util.Config, producer worker.TaskProducer) (*Server, error) {
	jwtMaker, err := token.NewJwtMaker(config.AccessTokenSecretKey)
	if err != nil {
		return nil, err
	}

	// defining a server with sb configuration
	server := &Server{
		Config:       &config,
		Store:        store,
		TokenMaker:   jwtMaker,
		TaskProducer: producer,
	}

	// registering all custom made validators in gin
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validatorMap := NewRequestValidator().Validator()
		for tag, fn := range validatorMap {
			v.RegisterValidation(tag, fn)
		}
	}

	server.SetupRoute()
	return server, nil
}

// SetupRoute func helps us to setup routest easily and initailize the
// gin.Engine back to Server struct
func (s *Server) SetupRoute() {

	router := gin.New()

	router.Use(logger.SetLogger(logger.WithLogger(func(ctx *gin.Context, l zerolog.Logger) zerolog.Logger {
		return *myLogger
	})))

	// unauthorized routes
	router.POST("/api/v1/users", s.CreateUser)
	router.POST("/api/v1/users/login", s.LoginUser)
	router.POST("/api/v1/token/renew", s.RenewUserSession)

	authRouter := router.Group("/").Use(authMiddleware(s.TokenMaker))

	// authorized routes
	// defining users routes
	authRouter.GET("/api/v1/users/:username", s.GetUser)
	authRouter.GET("/api/v1/users", s.GetAllUser)
	authRouter.PATCH("/api/v1/users/:username", s.UpdateUser)
	// as the above ones are for both admin level and user level api because in above api we have handle the case of authorization so that one user should not be able to other user's info
	// so in this api i am using token for getting the username not depending on path url
	authRouter.PATCH("/api/v1/users/me/password", s.UpdateUserPassword)

	// defining accounts routes
	authRouter.POST("/api/v1/accounts", s.CreateAccount)
	authRouter.GET("/api/v1/accounts/:id", s.GetAccountByID)
	authRouter.GET("/api/v1/accounts", s.GetAllAccount)
	authRouter.PUT("/api/v1/accounts/:id", s.UpdateAccount)
	authRouter.DELETE("/api/v1/accounts/:id", s.DeleteAccount)

	// defining transfer routes
	authRouter.POST("/api/v1/transfer", s.TransferMoney)

	// defining entries routes
	authRouter.GET("/api/v1/accounts/:id/entries", s.GetAllEntryForAccountID)
	authRouter.GET("/api/v1/entries", s.GetAllEntries)

	s.Router = router
}

func (s *Server) Start(address string) error {
	return s.Router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
