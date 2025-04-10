package imagebuilder

import (
	"encoding/json"
	"fmt"
	"log"

	runtime "github.com/open-ug/conveyor/pkg/driver-runtime"
	cTypes "github.com/open-ug/conveyor/pkg/types"
)

// Listen for messages from the runtime
func Reconcile(payload string, event string) error {

	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Image Driver Reconciling: %v", payload)

	if event == "application" {
		var appMsg cTypes.ApplicationMsg
		err := json.Unmarshal([]byte(payload), &appMsg)
		if err != nil {
			return fmt.Errorf("error unmarshalling application message: %v", err)
		}

		if appMsg.Action == "create" {
			app := appMsg.Payload

			if app.Spec.Source.Type == "git" {
				err := CreateBuildpacksImage(&app)
				if err != nil {
					return fmt.Errorf("error creating buildpacks image: %v", err)
				}

				runtime.BroadCastMessage(
					cTypes.DriverMessage{
						Event:   "buildpack-create-complete",
						Payload: payload,
					},
				)
			}
			return nil
		}
	}

	return nil
}

func Listen() {
	driver := &runtime.Driver{
		Reconcile: Reconcile,
	}

	driverManager := runtime.NewDriverManager(driver, []string{"*"})

	err := driverManager.Run()
	if err != nil {
		fmt.Println("Error running driver manager: ", err)
	}

}
