package main

import (
	"fmt"
	"gake/logic"
	"log"
)

func main() {
	if err := logic.RunAtUserRoot(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done.")
}
