package main

import (
	"github.com/zboyco/bilibili-tools-go/utils"
	"log"
)

func main() {
	log.Println("tools")

	b, err := utils.NewFromLogin("******", "******")
	if err != nil {
		log.Println(err)
		return
	}
	logged, err := b.IsLoggedIn()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Logged:", logged)
}
