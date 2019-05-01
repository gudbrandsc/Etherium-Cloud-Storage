package client

import (
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"os"
	"strings"
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
	signature    fileData
	publicKey    []byte
	fileDataHash string
	blockHeight  int32
}

type fileData struct {
	FileData []byte
	FileName string
	FileType string
}
type fileData_encoded struct {
	FileData []byte `json:"FileData"`
	FileName string `json:"FileName"`
	FileType string `json:"FileType"`
}

type storeFileInfo struct {
	CiphertextData []byte
	DataHash       string
	PublicKey      string
}

type storeFileInfo_encoded struct {
	CiphertextData []byte `json:"CiphertextData"`
	DataHash       string `json:"DataHash"`
	PublicKey      string `json:"PublicKey"`
}

func GetPrivateKey(cf ClientInfo) *rsa.PrivateKey {
	return cf.privateKey
}

func GetPublicKey(cf ClientInfo) *rsa.PublicKey {
	return cf.publicKey
}

func LoadUserData() ClientInfo {
	clientData := ClientInfo{}

	if Exists("./clientData/clientInfo.json") {
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

func Exists(path string) bool {
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

func StoreFile(filename string, publicKey []byte) {
	fileString := strings.Split(filename, ".")
	fileData := fileData{}
	fileData.FileData = ReadFileFromLocal(filename)
	fileData.FileName = fileString[0]
	fileData.FileType = fileString[1]
	fileDataByteArray := fileDataToJson(fileData)

	sum := sha3.Sum256(fileDataByteArray)
	dataHash := hex.EncodeToString(sum[:])

	storeFileInfo := storeFileInfo{}
	storeFileInfo.DataHash = dataHash
	storeFileInfo.PublicKey = string(publicKey)
	storeFileInfo.CiphertextData = EncryptWithPublicKey(fileDataByteArray, BytesToPublicKey(publicKey))

	fmt.Println(string(storeFileInfoToJson(storeFileInfo)))
}

func WriteFileToLocal(data []byte) {
	err := ioutil.WriteFile("./outFiles/test2.jpg", data, 0644)
	if err != nil {
		panic(err)
	}
}

func ReadFileFromLocal(name string) []byte {
	data, err := ioutil.ReadFile("./testFiles/" + name) // _ returns [] byte
	if err != nil {
		fmt.Println("File reading error", err)
	}
	return data
}

func fileDataToJson(fileData fileData) []byte {
	encodedData := &fileData_encoded{
		FileData: fileData.FileData,
		FileName: fileData.FileName,
		FileType: fileData.FileType,
	}
	result, _ := json.Marshal(encodedData)

	return result
}

func storeFileInfoToJson(storeFileInfo storeFileInfo) []byte {
	encodedData := &storeFileInfo_encoded{
		CiphertextData: storeFileInfo.CiphertextData,
		DataHash:       storeFileInfo.DataHash,
		PublicKey:      storeFileInfo.PublicKey,
	}
	result, _ := json.Marshal(encodedData)

	return result

}
