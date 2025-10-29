package main

import (
	"fmt"
	"log"

	_ "github.com/tucanbit/docs"
	"github.com/tucanbit/initiator"
)

func main() {
	fmt.Println("Starting TucanBIT Online Casino application...")
	fmt.Println("This is API documentation for TucanBIT online casino")
	log.Println("Initializing application components")

	initiator.Initiate()

	select {}
}
