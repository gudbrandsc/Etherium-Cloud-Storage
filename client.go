package main

import (
	"./p3/data"
	"bufio"
	"cs686-blockchain-p3-gudbrandsc/client"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

/*
Encrypt signature with private key
encrypt data with syntetic public key -> That way only I can see it
*/

func main() {

	clientData := client.LoadUserData()
	text := []byte("Hello world")
	signature, _ := client.CreateSignature(clientData.GetPrivateKey(), text)
	if client.VerifySignature(signature, clientData.GetPrivateKey(), text) {
		fmt.Println("Valid")
	}
	id, _ := strconv.ParseInt(os.Args[1], 10, 32)
	Peers := data.NewPeerList( /*Register()*/ int32(id), 32) // Uses port number as ID since TA server is down
	updatePeerList(Peers)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome, to GudbrandCoin")
	fmt.Println("---------------------")
	for {

		fmt.Println("Enter command:")

		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		textArray := strings.Split(text, " ")

		if strings.Compare("help", textArray[0]) == 0 {

			fmt.Println("---------------------------------------------------------------------")
			fmt.Println("|\t\t Command \t\t|\t\t\t\t Description \t\t\t\t|")
			fmt.Println("---------------------------------------------------------------------")
			fmt.Println("| liststored \t\t\t|\t Show all stored files \t\t\t\t\t|")
			fmt.Println("| listlocal \t\t\t|\t Show all local files \t\t\t\t\t|")
			fmt.Println("| store <filename path> |\t Store a file on the BlockChain \t\t|")
			fmt.Println("| get   <filename>\t\t|\t Retrieve a file from the BlockChain \t|")
			fmt.Println("| balance   \t\t\t|\t Show your balance \t\t\t\t\t\t|")
			fmt.Println("| updateminers   \t\t|\t Update list of available miners \t\t|")
			fmt.Println("---------------------------------------------------------------------")

		} else if strings.Compare("liststored", textArray[0]) == 0 {
			client.ListStoredFiles(&clientData)
		} else if strings.Compare("listlocal", textArray[0]) == 0 {
			client.ListAllLocalFiles()
		} else if strings.Compare("store", textArray[0]) == 0 {
			storeFile(textArray, clientData)
		} else if strings.Compare("updateminers", textArray[0]) == 0 {
			updatePeerList(Peers)
		} else if strings.Compare("get", textArray[0]) == 0 {
			if len(textArray) != 3 {
				fmt.Print("Invalid command please use format: get <filename>\n")
			} else {
				client.RetrieveFile(getRandomPeer(textArray[2]), textArray[1], &clientData)
			}
		} else if strings.Compare("balance", textArray[0]) == 0 {
			fmt.Println(" -------------------")
			fmt.Print("|\t Balance: ")
			fmt.Print(client.GetBalance(&clientData))
			fmt.Println("\t|")
			fmt.Println(" -------------------")

		} else {
			fmt.Println("Invalid command, type help for command info")
		}
	}
}

func getRandomPeer(v string) string {
	if v == "s" {
		return "http://localhost:6687"
	}
	return "http://localhost:6686"

}

func storeFile(textArray []string, clientData client.ClientInfo) {
	if len(textArray) != 2 {
		fmt.Print("Invalid command please use format: store <filename path>\n")
	} else {
		if client.Exists("./testFiles/" + textArray[1]) {
			client.StoreFile(textArray[1], &clientData)
		} else {
			fmt.Println("File does not exit")
		}
	}
}

func updatePeerList(peers data.PeerList) {
	url := "http://localhost:6686/getallpeers"
	resp, _ := http.Get(url)

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		peers.InjectPeerMapJson(string(bodyBytes), "http://localhost:"+os.Args[1])
		peers.Add("http://localhost:6686", 10)
		fmt.Println(peers.Show())
	} else {
		fmt.Println("unable to update get miners list")
	}

}
