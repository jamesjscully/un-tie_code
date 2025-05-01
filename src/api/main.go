package main

import (
	"github.com/jamesjscully/un-tie_code/src/api/app"
)

func main() {
	// Create and configure the application
	application := app.NewApplication()
	
	// Start the application
	application.Start()
}
