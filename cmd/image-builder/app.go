package imagebuilder

import (
	"fmt"
	"log"

	runtime "conveyor.cloud.cranom.tech/pkg/driver-runtime"
)

// Listen for messages from the runtime
func Reconcile(payload string, event string) error {

	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Image Driver Reconciling: %v", payload)

	return nil
}

func Listen() {
	driver := &runtime.Driver{
		Reconcile: Reconcile,
	}

	driverManager := runtime.NewDriverManager(driver, []string{"docker-build-complete"})

	err := driverManager.Run()
	if err != nil {
		fmt.Println("Error running driver manager: ", err)
	}

}
