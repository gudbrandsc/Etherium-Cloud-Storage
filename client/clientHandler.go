package client

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type ClientInfo struct {
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	aesSecret   []byte
	Client_data map[string]MapData
	Gcoin       float32
}

type Client_encoded struct {
	PrivateKey  []byte             `json:"privateKey"`
	PublicKey   []byte             `json:"publicKey"`
	AesSecret   []byte             `json:"AesSecret"`
	Client_data map[string]MapData `json:"data"`
	Gcoin       float32            `json:"Gcoin"`
}

type MapData struct {
	BlockHash   string
	BlockHeight int32
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
	Signature      []byte
	PublicKey      string
	DataHash       string
}

type StoreFileInfo_encoded struct {
	CiphertextData []byte `json:"CiphertextData"`
	Signature      []byte `json:"Signature"`
	PublicKey      string `json:"PublicKey"`
	DataHash       string `json:"DataHash"`
}

type store_resp struct {
	BlockHash   string `json:"BlockHash"`
	BlockHeight int32  `json:"BlockHeight"`
}
type file_retrieval_resp struct {
	TXfee int32         `json:"TXfee"`
	Data  StoreFileInfo `json:"Data"`
}

var FIRST_NODE_ADDR = "http://localhost:6686"

func (client *ClientInfo) GetPrivateKey() *rsa.PrivateKey {
	return client.privateKey
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
		clientData.aesSecret = data.AesSecret
		clientData.Gcoin = data.Gcoin

		fmt.Println("Loaded existing key pair, and client data")

	} else {
		fmt.Println("Generating private and public key..")
		clientData.privateKey, clientData.publicKey = GenerateKeyPair()
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			fmt.Println("Unable to create Aes key")
			os.Exit(1)
		}
		fmt.Println(key)
		clientData.aesSecret = key
		clientData.Client_data = make(map[string]MapData)
		clientData.Gcoin = 1000.00
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

		//Create signature from private key, using json string of FileData
		signature, err := CreateSignature(clientData.privateKey, fileDataToJson(fileData))
		if err != nil {
			fmt.Println("failed to create signature. Abort")
			return
		}

		hashed := sha3.Sum256(fileDataByteArray)
		storeFileInfo := StoreFileInfo{}
		storeFileInfo.Signature = signature
		storeFileInfo.PublicKey = string(PublicKeyToBytes(clientData.publicKey))
		storeFileInfo.DataHash = hex.EncodeToString(hashed[:])
		storeFileInfo.CiphertextData = AesEncrypt(clientData.aesSecret, fileDataByteArray)

		if sendStoreRequest(storeFileInfo, filename, clientData) {
			writeClientInfoToFile(clientData)
			fmt.Println(" ------------------------")
			fmt.Println("|File successfully stored|")
			fmt.Println(" ------------------------")
		} else {
			fmt.Println("Error on store request")
		}

	} else {
		fmt.Println(" ----------------------------------")
		fmt.Println("|File already stored on block chain|")
		fmt.Println(" ----------------------------------")
	}

}

func RetrieveFile(fileName string, clientData *ClientInfo) {
	minerResp := file_retrieval_resp{}
	newFile := StoreFileInfo{}
	newFileData := FileData{}

	if _, ok := clientData.Client_data[fileName]; !ok {
		fmt.Println("File is not stored on block chain")
		return
	}

	fileInfo, err := sendRetrieveRequest(fileName, clientData)
	if err {
		fmt.Println("Unable to retrieve file")
		return
	}

	json.Unmarshal([]byte(fileInfo), &minerResp)
	newFile.PublicKey = minerResp.Data.PublicKey
	newFile.CiphertextData = minerResp.Data.CiphertextData
	newFile.Signature = minerResp.Data.Signature
	txFeeAmount := minerResp.TXfee
	fmt.Println(txFeeAmount)

	jsonString := AesDecrypt(clientData.aesSecret, newFile.CiphertextData)

	fileDataEncoded := FileData_encoded{}
	json.Unmarshal([]byte(jsonString), &fileDataEncoded)

	newFileData.FileData = fileDataEncoded.FileData
	newFileData.FileType = fileDataEncoded.FileType
	newFileData.FileName = fileDataEncoded.FileName
	fileDataByteArray := fileDataToJson(newFileData)

	if VerifySignature(newFile.Signature, clientData.privateKey, fileDataByteArray) {
		WriteFileToLocal(newFileData.FileData, newFileData.FileName, newFileData.FileType)
		//TODO pay miner
	} else {
		fmt.Println("Signature not valid -> Do not pay")
	}

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
		Signature:      storeFileInfo.Signature,
		PublicKey:      storeFileInfo.PublicKey,
		DataHash:       storeFileInfo.DataHash,
	}
	result, _ := json.Marshal(encodedData)

	return result
}

func writeClientInfoToFile(clientData *ClientInfo) {

	encodedB := &Client_encoded{
		PrivateKey:  PrivateKeyToBytes(clientData.privateKey),
		PublicKey:   PublicKeyToBytes(clientData.publicKey),
		AesSecret:   clientData.aesSecret,
		Client_data: clientData.Client_data,
		Gcoin:       clientData.Gcoin,
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

func sendStoreRequest(body StoreFileInfo, filename string, clientData *ClientInfo) bool {

	url := FIRST_NODE_ADDR + "/store"
	jsonString := storeFileInfoToJson(body)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonString))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp := store_resp{}
		json.Unmarshal([]byte(bodyBytes), &resp)
		mapData := MapData{resp.BlockHash, resp.BlockHeight, filename, body.DataHash}
		clientData.Client_data[filename] = mapData
		return true

	}
	return false
}

func sendRetrieveRequest(filename string, clientData *ClientInfo) ([]byte, bool) {
	data := clientData.Client_data[filename]
	url := fmt.Sprintf("%s%s%d%s", FIRST_NODE_ADDR+"/retrieve/", data.FileHash+"/", data.BlockHeight, "/"+data.BlockHash)

	resp, err := http.Get(url)
	if err != nil {
		return nil, true
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, true
		}

		return bodyBytes, false
	}
	return nil, true

}
