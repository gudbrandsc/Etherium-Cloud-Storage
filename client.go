package main

import (
	"bufio"
	"cs686-blockchain-p3-gudbrandsc/client"
	"fmt"
	"os"
	"strings"
)

func main() {
	clientData := client.LoadUserData()

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
			fmt.Println("---------------------------------------------------------------------")

		} else if strings.Compare("liststored", textArray[0]) == 0 {
			client.ListStoredFiles(&clientData)
		} else if strings.Compare("listlocal", textArray[0]) == 0 {
			client.ListAllLocalFiles()
		} else if strings.Compare("store", textArray[0]) == 0 {
			storeFile(textArray, clientData)
		} else if strings.Compare("get", textArray[0]) == 0 {
			fmt.Println("Send request to BC to get a file ")
		} else {
			fmt.Println("Invalid command, type help for command info")
		}
	}
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
