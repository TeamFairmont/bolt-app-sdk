package main

import (
	"fmt"

	"github.com/TeamFairmont/bolt-app-sdk"
	"github.com/TeamFairmont/boltsdk-go/boltsdk/functions"
	"github.com/jasonlvhit/gocron"
)

var configPath = "./appConfig.json"

func main() {
	//ready gocron.Scheduler
	var scheduler = gocron.Scheduler{}
	//get bolt app config
	appConfig, err := boltAppSdk.LoadConfig(configPath)
	if err != nil {
		fmt.Println(err)
	} //set appFunctionMap
	appFuncMap := map[string]boltAppSdk.AppFunc{
		"/get-config":        boltsdkFunctions.G1etBoltConfigAppFunc,
		"/request/v1/search": SearchAppFunction,
	} //schedule and run apps
	err = boltAppSdk.RunScheduledAppsFromConfig(*appConfig, appFuncMap, &scheduler)
	if err != nil {
		fmt.Println("error in scheduleTest.go: ", err)
	}
	//wait forever
	var forever = make(chan bool)
	<-forever
}

//SearchAppFunction is a sample app function to use the fairmont searches search api actually
func SearchAppFunction(p map[string]interface{}, payloadChan chan map[string]interface{}, respBodyChan chan []byte, doneChan chan bool, args []interface{}) error {
	fmt.Println("SearchAppFunction started")
	//ready empty payload to be send
	for _, v := range args {
		fmt.Println(v)
	}
	fmt.Println(args)
	searchPayload := map[string]interface{}{
		"category":     "undefined",
		"resultOffset": 0,
		"searchKey":    fmt.Sprint(args[0]),
		"storeID":      fmt.Sprint(args[1]),
	}
	fmt.Println("payload created")
	//send payload
	payloadChan <- searchPayload
	fmt.Println("payload sent")
	//receive response from bolt
	respBytes := <-respBodyChan
	fmt.Println(string(respBytes))
	fmt.Println("respBytes created")
	return nil
}
