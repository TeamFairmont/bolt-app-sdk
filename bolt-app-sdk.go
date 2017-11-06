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
type AppFunc func(map[string]interface{}, []interface{}) error

//RunApp takes a function and handles the bolt communication
func RunApp(boltURL, userName, passWord string, af AppFunc, args ...interface{}) error {
	var payload = make(map[string]interface{})
	//run app function
	err := af(payload, args)
	if err != nil {
		fmt.Println(`error in app function: `, err)
		return err
	}

	//marshal the payload into json object
	p, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
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
		return err
	}

	req, err := http.NewRequest("POST", boltURL, bytes.NewBuffer(hmacToSend))
	if err != nil {
		fmt.Println("error making http request: ", err)
		return err
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
	client := &http.Client{Transport: tr}

	//set auth and header
	req.SetBasicAuth(userName, "pw ignored")
	req.Close = true //close the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error in client.Do: ", err)
		return err
	}
	//read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_ = body
	//explicitly close the body
	resp.Body.Close()
	return err
}
