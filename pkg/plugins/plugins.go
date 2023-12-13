package plugins

import (
	"WarpGPT/pkg/db"
	"WarpGPT/pkg/env"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Component struct {
	Engine *gin.Engine
	Db     db.DB
	Logger *logrus.Logger
	Env    *env.ENV
	Auth   func(arkType int, puid string) (string, error)
}

type Plugin interface {
	Run(com *Component)
}
