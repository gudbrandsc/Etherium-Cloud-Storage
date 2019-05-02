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
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	Client_data map[string]MapData
}

type Client_encoded struct {
	PrivateKey  []byte             `json:"privateKey"`
	PublicKey   []byte             `json:"publicKey"`
	Client_data map[string]MapData `json:"data"`
}
type MapData struct {
	BlockHash   string
	blockHeight int32
	FileName    string
	FileHash    string
}

type FileData struct {
	FileData []byte
	FileName string
	FileType string
}
type FileData_encoded struct {
	FileData []byte `json:"FileData"`
	FileName string `json:"FileName"`
	FileType string `json:"FileType"`
}

type StoreFileInfo struct {
	CiphertextData []byte
	DataHash       string
	PublicKey      string
}

type StoreFileInfo_encoded struct {
	CiphertextData []byte `json:"CiphertextData"`
	DataHash       string `json:"DataHash"`
	PublicKey      string `json:"PublicKey"`
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

		data := Client_encoded{}

		json.Unmarshal([]byte(byteValue), &data)
		clientData.privateKey = BytesToPrivateKey(data.PrivateKey)
		clientData.publicKey = BytesToPublicKey(data.PublicKey)
		clientData.Client_data = data.Client_data

		fmt.Println("Loaded existing key pair, and client data")

	} else {
		fmt.Println("Generating private and public key..")
		clientData.privateKey, clientData.publicKey = GenerateKeyPair()
		clientData.Client_data = make(map[string]MapData)
		writeClientInfoToFile(&clientData)

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
	fmt.Println(" -------------------")
	fmt.Println("|\tLocal files\t\t|")
	fmt.Println(" -------------------")

	files, err := ioutil.ReadDir("./testFiles")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println("| " + f.Name())
	}
	fmt.Println(" -------------------")

}

func StoreFile(filename string, clientData *ClientInfo) {
	if _, ok := clientData.Client_data[filename]; !ok {
		fileString := strings.Split(filename, ".")
		fmt.Println(filename)
		fileData := FileData{}
		fileData.FileData = ReadFileFromLocal(filename)
		fileData.FileName = fileString[0]
		fileData.FileType = fileString[1]
		fileDataByteArray := fileDataToJson(fileData)

		sum := sha3.Sum256(fileDataByteArray)
		dataHash := hex.EncodeToString(sum[:])

		storeFileInfo := StoreFileInfo{}
		storeFileInfo.DataHash = dataHash
		storeFileInfo.PublicKey = string(PublicKeyToBytes(clientData.publicKey))
		storeFileInfo.CiphertextData = EncryptWithPublicKey(fileDataByteArray, clientData.publicKey)

		//TODO change this based on respose from miner
		mapData := MapData{"", 1, filename, dataHash}
		clientData.Client_data[filename] = mapData
		writeClientInfoToFile(clientData)

		//fmt.Println(string(storeFileInfoToJson(StoreFileInfo)))
		//RetrieveFile(string(storeFileInfoToJson(storeFileInfo)), clientData)
		//TODO Send file to a miner
		fmt.Println(" ------------------------")
		fmt.Println("|File successfully stored|")
		fmt.Println(" ------------------------")
	} else {
		fmt.Println(" ----------------------------------")
		fmt.Println("|File already stored on block chain|")
		fmt.Println(" ----------------------------------")
	}

}

func RetrieveFile(FileInfo string, clientData *ClientInfo) {
	data := StoreFileInfo_encoded{}
	newFile := StoreFileInfo{}
	newFileData := FileData{}

	json.Unmarshal([]byte(FileInfo), &data)

	newFile.PublicKey = data.PublicKey
	newFile.CiphertextData = data.CiphertextData
	newFile.DataHash = data.DataHash

	jsonString := DecryptWithPrivateKey(newFile.CiphertextData, clientData.privateKey)

	test := FileData{}
	json.Unmarshal([]byte(jsonString), &test)

	newFileData.FileData = test.FileData
	newFileData.FileType = test.FileType
	newFileData.FileName = test.FileName

	WriteFileToLocal(newFileData.FileData, newFileData.FileName, newFileData.FileType)

}

func WriteFileToLocal(data []byte, filename string, extention string) {
	err := ioutil.WriteFile("./outFiles/"+filename+"."+extention, data, 0644)
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

func fileDataToJson(fileData FileData) []byte {
	encodedData := &FileData_encoded{
		FileData: fileData.FileData,
		FileName: fileData.FileName,
		FileType: fileData.FileType,
	}
	result, _ := json.Marshal(encodedData)

	return result
}

func storeFileInfoToJson(storeFileInfo StoreFileInfo) []byte {
	encodedData := &StoreFileInfo_encoded{
		CiphertextData: storeFileInfo.CiphertextData,
		DataHash:       storeFileInfo.DataHash,
		PublicKey:      storeFileInfo.PublicKey,
	}
	result, _ := json.Marshal(encodedData)

	return result
}

func writeClientInfoToFile(clientData *ClientInfo) {

	encodedB := &Client_encoded{
		PrivateKey:  PrivateKeyToBytes(clientData.privateKey),
		PublicKey:   PublicKeyToBytes(clientData.publicKey),
		Client_data: clientData.Client_data,
	}

	clientInfoJson, _ := json.Marshal(encodedB)
	err := ioutil.WriteFile("./clientData/clientInfo.json", clientInfoJson, 0644)
	if err != nil {
		fmt.Println("Unable to write keys to file.")
		os.Exit(1)
	}
}
func ListStoredFiles(clientData *ClientInfo) {
	fmt.Println(" -------------------")
	fmt.Println("|\tStored files\t|")
	fmt.Println(" -------------------")

	for key, _ := range clientData.Client_data {
		fmt.Printf("| %s \n", key)
	}
	fmt.Println(" -------------------")

}
