package main

import (
	"encoding/json"
	"fmt"

	"github.com/TeamFairmont/bolt-app-sdk"
	"github.com/TeamFairmont/boltsdk-go/boltsdk/functions"
	"github.com/amazingfly/cv3go"
	"github.com/jasonlvhit/gocron"
)

var configPath = "./appConfig.json"

func main() {
	appConfigCTX, err := boltAppSdk.LoadConfig(configPath)
	if err != nil {
		fmt.Println(err)
	}
	//boltsdkFunctions.GetBoltConfigAppFunc(p, payloadChan, respBodyChan, doneChan, args)
	//var js = boltAppSdk.JobSchedule{Every: 10, Unit: "seconds"}
	//var appCTX = boltAppSdk.AppCTX{BoltURL: "http://localhost:8888/get-config", UserName: "searchgroup", PassWord: "open_to_public", AF: boltsdkFunctions.GetBoltConfigAppFunc}
	//appCTX.Schedule = js
	appFuncMap := map[string]boltAppSdk.AppFunc{
		"/get-config": boltsdkFunctions.GetBoltConfigAppFunc,
	}
	for command, app := range appConfigCTX.Apps {
		app.AF = appFuncMap[command]
		err = boltAppSdk.ScheduleApp(app.Schedule, func() { boltAppSdk.RunAppCTX(app) })
		if err != nil {
			fmt.Println("error: ", err)
		}
	}
	/*
		js.Unit = "minutes"
		js.Every = 1
		boltAppSdk.ScheduleApp(js, func() { fmt.Println("minutes: ", time.Now()) })
	*/
	<-gocron.Start()
	fmt.Println("this never happens?")
	//var appConfigCTX = boltAppSdk.ConfigCTX{BoltURL: "http://localhost:8888", UserName: "searchgroup", PassWord: "open_to_public"}
	appConfigCTX.Apps = make(map[string]boltAppSdk.AppCTX)
	//appConfigCTX.Apps["/get-config"] = appCTX
	b, err := json.MarshalIndent(&appConfigCTX, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	cv3go.PrintToFile(b, "./appConfig.json")
}
