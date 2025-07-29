package main

import (
	"fmt"
	"log"

	_ "github.com/joshjones612/egyptkingcrash/docs"
	"github.com/joshjones612/egyptkingcrash/initiator"
)

func main() {
	fmt.Println("Starting EgyptKingCrash application...")
	log.Println("Initializing application components")

	initiator.Initiate()

	select {}
}
