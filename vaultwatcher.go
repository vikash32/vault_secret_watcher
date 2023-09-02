package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"k8s.io/client-go/kubernetes"
)

func VaultWatcher(clientset *kubernetes.Clientset) {

	// Get Data from Vault
	jsonData := hitVaultURL()

	// parse jsonData
	_, latestPassword := jsonParser(jsonData)

	// Compare prev password and latest passwords value

	ispasswordChanged := comparePass(prevPassword, latestPassword)
	if ispasswordChanged {
		log.Println("Password has been changed Calling now Secret Updation function")
		// TODO logic to update the secret and retrigger the deployment / sts
		secretUpdated(clientset, latestPassword)
		// Updating the prev password value
		prevPassword = latestPassword
	}

}

func hitVaultURL() string {

	// secret_path = "/v1/secret/data/myfirstsecret/path/ui"
	secret_path = os.Getenv("SECRET_PATH")
	vault_addr = os.Getenv("VAULT_ADDR")
	vault_token = os.Getenv("VAULT_TOKEN")

	// Create vault url to call
	vault_path, err := url.JoinPath(vault_addr, secret_path)

	if err != nil {
		log.Println("Error in forming vault path", err.Error())
	}

	// Create an http.Header object to hold custom headers
	headers := make(http.Header)
	headers.Set("x-vault-token", vault_token)

	// Create an HTTP client with custom headers
	client := &http.Client{}

	// Create a GET request with the custom headers
	req, err := http.NewRequest("GET", vault_path, nil)
	if err != nil {
		panic(err)
	}

	req.Header = headers

	// send the GET Request
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	// Read the response body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println("Status Code : ", resp.Status)
	// log.Println("Response Body : ", string(body))
	jsonData := string(body)

	return jsonData
}

func jsonParser(jsonData string) (string, string) {

	// Parse the JSON response into a map
	var response map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		log.Println("jsondata parsing is failing : ", err.Error())
	}

	data, ok := response["data"].(map[string]interface{})

	if ok {
		nestedData, ok := data["data"].(map[string]interface{})
		if ok {
			username, _ := nestedData["username"].(string)
			password, _ := nestedData["password"].(string)
			// log.Println(username, password)
			return username, password

		}
		return "", ""
	}
	return "", ""
}

func comparePass(prevPassword, latestPassword string) bool {
	result := prevPassword != latestPassword
	return result
}
