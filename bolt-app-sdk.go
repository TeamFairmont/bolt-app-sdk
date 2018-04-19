package boltAppSdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TeamFairmont/boltshared/security"
	"github.com/jasonlvhit/gocron"
)

var errChan = make(chan error, 1)

//AppFunc is the app function to be passed in
type AppFunc func(map[string]interface{}, chan map[string]interface{}, chan []byte, chan bool, []interface{}) error

//ConfigCTX will hold the config file information for a Bolt App
type ConfigCTX struct {
	BoltURL  string            `json:"boltURL"`
	UserName string            `json:"userName"`
	PassWord string            `json:"password"`
	Apps     map[string]AppCTX `json:"Apps"`
}

//AppCTX will hold the data to run an AppCTX
type AppCTX struct {
	BoltURL     string  `json:"boltURL"`
	CommandName string  `json:"commandName"`
	UserName    string  `json:"userName"`
	PassWord    string  `json:"password"`
	AF          AppFunc `json:"-"`
	Args        []interface{}
	Schedule    JobSchedule `json:"schedule"`
}

//JobSchedule holds the information to schedule jobs
type JobSchedule struct {
	Every uint64 `json:"every"` //the number of units to be used as the interval
	Unit  string `json:"unit"`  //the unit to be used: seconds, minutes, hours, days, weeks, and day of the weeks
	At    string `json:"at"`    //a 24:00 time to run the function at e.g. gocron.Every(1).Days().At("11:30").Do(thisFunc)
}

//type AppFunc func(...interface{}) error

//RunApp takes a function and handles the bolt communication
func RunApp(boltURL, userName, passWord string, af AppFunc, args ...interface{}) error {
	var payloadChan = make(chan map[string]interface{}) //channel to send and recieve payloads
	var respBodyChan = make(chan []byte, 1)             //channel to send and receive response body
	var doneChan = make(chan bool)                      //channel to signal the app function is done
	var wg = sync.WaitGroup{}
	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		var payload = make(map[string]interface{}) //TODO probably do not need anymore
		//run app function
		err := af(payload, payloadChan, respBodyChan, doneChan, args)
		if err != nil {
			err = errors.New("Error in Bolt App Function: " + err.Error())
			errChan <- err
		}
		wg.Done()
		doneChan <- true //signal that the app function has completed
	}(&wg)

	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		for { //repeat until doneChan has ben sent
			select { //payloadChan is sent from app functions
			case payload := <-payloadChan:
				//marshal the payload into json object
				p, err := json.Marshal(payload)
				if err != nil {
					wg.Done()
					err = errors.New("Error Unmarshalling payload from Bolt App Func: " + err.Error())
					errChan <- err
					return
				}

				// Prepare the payload with hmac encoding
				// Encode the message to send
				hmacToSend, err := security.EncodeHMAC(
					passWord,
					string(p),
					strconv.FormatInt(time.Now().Unix(), 10),
				)
				if err != nil {
					wg.Done()
					err = errors.New("Error encoding payload with HMAC in Bolt App Sdk: " + err.Error())
					errChan <- err
					return
				}

				req, err := http.NewRequest("POST", boltURL, bytes.NewBuffer(hmacToSend))
				if err != nil {
					wg.Done()
					err = errors.New("Error making http request in Bolt App Sdk: " + err.Error())
					errChan <- err
					return
				}

				// The tls.Config settings are set server side, but may also be set client side.
				// InsecureSkipVerify allows self-signed certificates in development environments.
				// This must be set to false for production using a trusted certificate authority.
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
						MinVersion:         tls.VersionTLS12, // Communicate with TLS 1.2 (771)    	PreferServerCipherSuites: true,
						CipherSuites: []uint16{
							tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
						},
					},
					DisableCompression: true, // Compressed TLS is vulnerable to attacks
				}
				timeout := time.Duration(8 * time.Minute)
				client := &http.Client{Timeout: timeout, Transport: tr}

				//set auth and header
				req.SetBasicAuth(userName, "pw ignored")
				req.Close = true //close the request
				resp, err := client.Do(req)
				if err != nil {
					wg.Done()
					err = errors.New("Error in Client.Do in Bolt App Sdk: " + err.Error())
					errChan <- err
					return
				}
				//read response body
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					wg.Done()
					err = errors.New("Error reading response body in Bolt App Sdk: " + err.Error())
					errChan <- err
					return
				}
				respBodyChan <- body
				//explicitly close the body
				resp.Body.Close()
			case <-doneChan: //doneChan signal recieved, end the process
				wg.Done()
				return
			}
		}
	}(&wg)
	wg.Wait()
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

//RunAppCTX is a wrapper for RunApp so it can use a struct
func RunAppCTX(appCTX AppCTX) error {
	err := RunApp(appCTX.BoltURL+appCTX.CommandName, appCTX.UserName, appCTX.PassWord, appCTX.AF, appCTX.Args)
	if err != nil {
		return err
	}
	return nil
}

//ScheduleApp is a wrapper that will RunApp on a passed in interval //currently interval is an int value
func ScheduleApp(js JobSchedule, task func()) error {
	if js.At == "" {
		switch strings.ToLower(js.Unit) {
		case "seconds":
			gocron.Every(js.Every).Seconds().Do(task)
		case "minutes":
			gocron.Every(js.Every).Minutes().Do(task)
		case "hours":
			gocron.Every(js.Every).Hours().Do(task)
		case "days":
			gocron.Every(js.Every).Days().Do(task)
		case "weeks":
			gocron.Every(js.Every).Weeks().Do(task)
		case "monday":
			gocron.Every(js.Every).Monday().Do(task)
		case "tuesday":
			gocron.Every(js.Every).Tuesday().Do(task)
		case "wednesday":
			gocron.Every(js.Every).Wednesday().Do(task)
		case "thursday":
			gocron.Every(js.Every).Thursday().Do(task)
		case "friday":
			gocron.Every(js.Every).Friday().Do(task)
		default:
			fmt.Println("unit did not match anything in the switch")
		}
	} else { // At field is populated
		switch strings.ToLower(js.Unit) {
		case "seconds":
			gocron.Every(js.Every).Seconds().At(js.At).Do(task)
		case "minutes":
			gocron.Every(js.Every).Minutes().At(js.At).Do(task)
		case "hours":
			gocron.Every(js.Every).Hours().At(js.At).Do(task)
		case "days":
			gocron.Every(js.Every).Days().At(js.At).Do(task)
		case "weeks":
			gocron.Every(js.Every).Weeks().At(js.At).Do(task)
		case "monday":
			gocron.Every(js.Every).Monday().At(js.At).Do(task)
		case "tuesday":
			gocron.Every(js.Every).Tuesday().At(js.At).Do(task)
		case "wednesday":
			gocron.Every(js.Every).Wednesday().At(js.At).Do(task)
		case "thursday":
			gocron.Every(js.Every).Thursday().At(js.At).Do(task)
		case "friday":
			gocron.Every(js.Every).Friday().At(js.At).Do(task)
		default:
			fmt.Println("unit did not match anything in the switch")
		}
	}
	return nil
}

//LoadConfig loads the configuration file
func LoadConfig(cfgPath string) (*ConfigCTX, error) {
	// load the config file
	var cfg = ConfigCTX{}
	configFile, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(configFile), &cfg)
	if err != nil {
		return &cfg, err
	}
	return &cfg, nil
}
