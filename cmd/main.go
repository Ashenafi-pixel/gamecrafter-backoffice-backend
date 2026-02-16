package main

import (
	"fmt"
	"log"

	_ "github.com/tucanbit/docs"
	"github.com/tucanbit/initiator"
)

// @title           Game Crafter Backoffice API
// @version         1.0
// @description     This is the API documentation for Game Crafter online casino backoffice system
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	fmt.Println("Starting TucanBIT Online Casino application...")
	fmt.Println("This is API documentation for TucanBIT online casino")
	log.Println("Initializing application components")

	initiator.Initiate()

	select {}
}
