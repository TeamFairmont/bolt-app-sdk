package boltAppSdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//AppFunc is the app function to be passed in
type AppFunc func(map[string]interface{}) error

//RunApp takes a function and handles the bolt communication.
func RunApp(boltURL, userName, passWord string, af AppFunc) error {
	var payload = make(map[string]interface{})
	//run app function
	err := af(payload)
	if err != nil {
		fmt.Println(`error in app funbction `, err)
		return err
	}
	//mashal the payload into json object
	p, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("eeror Marshalling json ", err)
		return err
	}

	req, err := http.NewRequest("POST", boltURL, bytes.NewBuffer(p))
	if err != nil {
		fmt.Println("error making http request", err)
		return err
	}
	//set auth and header
	req.SetBasicAuth(userName, passWord)
	req.Header.Set("Content-Type", "application/json")
	//send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error doing client: ", err)
		return err
	}
	defer resp.Body.Close()
	//read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_ = body

	return err
}
