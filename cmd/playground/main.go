package main

import (
	"github.com/joho/godotenv"
	"fmt"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("hello from playground")
}