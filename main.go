package main

import (
	"WarpGPT/pkg/tools"
	"fmt"
)

func main() {

	auth := tools.NewAuthenticator("egxtjjqjm@hotmail.com", "FBMjdg257FBMjdg257")
	err := auth.Begin()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	println(auth.GetAccessToken())
	println(auth.GetRefreshToken())
}
