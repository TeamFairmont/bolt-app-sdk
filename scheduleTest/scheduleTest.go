package main

import (
	"fmt"

	"github.com/TeamFairmont/bolt-app-sdk"
	"github.com/TeamFairmont/boltsdk-go/boltsdk/functions"
	"github.com/jasonlvhit/gocron"
)

var configPath = "./appConfig.json"

func main() {
	var scheduler = gocron.Scheduler{}
	appConfig, err := boltAppSdk.LoadConfig(configPath)
	if err != nil {
		fmt.Println(err)
	}
	appFuncMap := map[string]boltAppSdk.AppFunc{
		"/get-config":        boltsdkFunctions.G1etBoltConfigAppFunc,
		"/request/v1/search": SearchAppFunction,
	}
	boltAppSdk.RunScheduledAppsFromConfig(*appConfig, appFuncMap, &scheduler)

	/*
		var wg = sync.WaitGroup{}
		for command, app := range appConfig.Apps {
			app.AF = appFuncMap[command]
			fmt.Println("running app: ", command)
			go func(wg *sync.WaitGroup, app boltAppSdk.AppCTX) {
				//wg.Add(1)
				//var mmm = sync.RWMutex{}
				//mmm.Lock()
				err = boltAppSdk.ScheduleApp(app.Schedule, func() { fmt.Println("sched func"); boltAppSdk.RunAppCTX(app) }, &scheduler)
				if err != nil {
					fmt.Println("error: ", err)
				}
				//mmm.Unlock()
				//wg.Done()
				fmt.Println("wg.done()")
			}(&wg, app)
		}
		//wg.Wait()
		//fmt.Println("wg.wait")
		<-scheduler.Start()
		fmt.Println("never here")
	*/
}

//SearchAppFunction is a sample app function to use the fairmont searches search api actually
func SearchAppFunction(p map[string]interface{}, payloadChan chan map[string]interface{}, respBodyChan chan []byte, doneChan chan bool, args []interface{}) error {
	fmt.Println("SearchAppFunction started")
	//ready empty payload to be send
	searchPayload := map[string]interface{}{
		"category":     "undefined",
		"resultOffset": "0",
		"searchKey":    "fruit",
		"storeID":      "gourney",
	}
	fmt.Println("payload created")
	//send payload
	payloadChan <- searchPayload
	fmt.Println("payload sent")
	//receive response from bolt
	respBytes := <-respBodyChan
	fmt.Println("respBytes created")
	fmt.Println(string(respBytes))
	return nil
}
