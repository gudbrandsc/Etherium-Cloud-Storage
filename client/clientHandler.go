package client

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type ClientInfo struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type client_encoded struct {
	PrivateKey []byte `json:"privateKey"`
	PublicKey  []byte `json:"publicKey"`
}

type fileInfo struct {
	filename       string
	blockHash      string
	ciphertextData []byte
	fileType       string
}

func GetPrivateKey(cf ClientInfo) *rsa.PrivateKey {
	return cf.privateKey
}

func GetPublicKey(cf ClientInfo) *rsa.PublicKey {
	return cf.publicKey
}

func LoadUserData() ClientInfo {
	clientData := ClientInfo{}

	if exists("./clientData/clientInfo.json") {
		jsonFile, err := os.Open("./clientData/clientInfo.json")
		if err != nil {
			fmt.Println("Unable to read user data")
			os.Exit(1)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		data := client_encoded{}

		json.Unmarshal([]byte(byteValue), &data)
		clientData.privateKey = BytesToPrivateKey(data.PrivateKey)
		clientData.publicKey = BytesToPublicKey(data.PublicKey)

		fmt.Println("Loaded existing key pair")

	} else {
		fmt.Println("Generating private and public key..")
		clientData.privateKey, clientData.publicKey = GenerateKeyPair()

		encodedB := &client_encoded{
			PrivateKey: PrivateKeyToBytes(clientData.privateKey),
			PublicKey:  PublicKeyToBytes(clientData.publicKey),
		}

		clientInfoJson, _ := json.Marshal(encodedB)

		err := ioutil.WriteFile("./clientData/clientInfo.json", clientInfoJson, 0644)
		if err != nil {
			fmt.Println("Unable to write keys to file.")
			os.Exit(1)
		}

	}
	return clientData
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func ListAllLocalFiles() {
	files, err := ioutil.ReadDir("./testFiles")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
}

func WriteFileToLocal(data []byte) {
	err := ioutil.WriteFile("./outFiles/test2.jpg", data, 0644)
	if err != nil {
		panic(err)
	}
}

func ReadFileFromLocal(name string) {
	_, err := ioutil.ReadFile("./testFiles/" + name) // _ returns [] byte
	if err != nil {
		fmt.Println("File reading error", err)
	}
}
