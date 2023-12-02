package main

import "github.com/gin-gonic/gin"

func GetRouter() {
	router := gin.Default()
	router.Any("/v1/*path")
}
