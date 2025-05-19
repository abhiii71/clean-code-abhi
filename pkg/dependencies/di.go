package di

import (
	bootserver "github.com/abhiii71/clean-code-abhi/pkg/boot"
	"github.com/abhiii71/clean-code-abhi/pkg/config"
	"github.com/abhiii71/clean-code-abhi/pkg/db"
	userauth "github.com/abhiii71/clean-code-abhi/pkg/user_auth"
)

func InitializeEvents(conf config.Config) (*bootserver.ServerHttp, error) {
	DB := db.ConnectPGDB(conf)

	userRepository := userauth.NewRepository(DB)
	userService := userauth.NewService(userRepository)
	userHandler := userauth.NewHandler(userService)

	serverHttp := bootserver.NewServerHttp(*userHandler)

	return serverHttp, nil

}
