package main

import (
	"nlypage-final/cmd/app"

	_ "time/tzdata"
)

func main() {
	a := app.New()
	a.Start()
}
