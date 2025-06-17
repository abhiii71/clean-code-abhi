package bootserver

import (
	"github.com/abhiii71/clean-code-abhi/pkg/config"
	userauth "github.com/abhiii71/clean-code-abhi/pkg/user_auth"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ServerHttp struct {
	engine *gin.Engine
}

func NewServerHttp(userHandler userauth.Handler) *ServerHttp {
	engine := gin.New()
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	userHandler.MountRoutes(engine)

	return &ServerHttp{engine}
}

func (s *ServerHttp) Start(conf config.Config) {
	s.engine.Run(conf.Host + ":" + conf.ServerPort)
}
