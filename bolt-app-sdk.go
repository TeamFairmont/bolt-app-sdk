package boltAppSdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/TeamFairmont/boltshared/security"
)

//AppFunc is the app function to be passed in
type AppFunc func(map[string]interface{}, chan map[string]interface{}, chan []byte, chan bool, []interface{}) error

//type AppFunc func(...interface{}) error

//RunApp takes a function and handles the bolt communication
func RunApp(boltURL, userName, passWord string, af AppFunc, args ...interface{}) error {
	var payloadChan = make(chan map[string]interface{}) //channel to send and recieve payloads
	var respBodyChan = make(chan []byte, 1)             //channel to send and receive response body
	var doneChan = make(chan bool)                      //channel to signal the app function is done
	go func() {
		var payload = make(map[string]interface{}) //TODO probably do not need anymore
		//run app function
		err := af(payload, payloadChan, respBodyChan, doneChan, args)
		if err != nil {
			fmt.Println(`error in app function: `, err)
			return
		}
		doneChan <- true //signal that the app function has completed
	}()

	//
	//
	//
	go func() {
		//var payload = make(map[string]interface{})
		for { //repeat until doneChan has ben sent
			select { //payloadChan is sent from app functions
			case payload := <-payloadChan:
				//marshal the payload into json object
				p, err := json.Marshal(payload)
				if err != nil {
					fmt.Println("error marshalling json: ", err)
					return //err
				}

				// Prepare the payload with hmac encoding
				// Encode the message to send
				hmacToSend, err := security.EncodeHMAC(
					passWord,
					string(p),
					strconv.FormatInt(time.Now().Unix(), 10),
				)
				if err != nil {
					fmt.Println(`error hmac encoding payload: `, err)
					return //err
				}

				req, err := http.NewRequest("POST", boltURL, bytes.NewBuffer(hmacToSend))
				if err != nil {
					fmt.Println("error making http request: ", err)
					return //err
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
					fmt.Println("error in client.Do: ", err)
					return //err
				}
				//read response body
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					return //err
				}
				respBodyChan <- body
				//explicitly close the body
				resp.Body.Close()
			case <-doneChan: //doneChan signal recieved, end the process
				return
			}
		}
	}()
	return nil
}

//ScheduleApp is a wrapper that will RunApp on a passed in interval //currently interval is an int value
func ScheduleApp(hour int, boltURL, userName, passWord string, af AppFunc, args ...interface{}) error {
	for true {
		//TODO location should be loaded from config
		utc, err := time.LoadLocation("America/New_York")
		if err != nil {
			fmt.Println("error loading location", err)
		}
		var t = time.Now().In(utc).Hour()
		if t == hour {
			err = RunApp(boltURL, userName, passWord, af, args)
			if err != nil {
				fmt.Println("error running bolt app: ", err)
			}
		} else if t > hour { //wait 24 hours minus the difference
			time.Sleep(time.Duration(24-(t-hour)) * time.Hour)
		} else { //if t < hour // wait the difference
			time.Sleep(time.Duration(hour-t) * time.Hour)
		}
	}

	return nil
}
