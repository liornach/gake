package main

import (
	"fmt"
	"log"
)

func main() {
	if err := runAtUserRoot(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Done.")
}
