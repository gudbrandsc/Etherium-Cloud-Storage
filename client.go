package main

import (
	"bufio"
	"cs686-blockchain-p3-gudbrandsc/client"
	"fmt"
	"os"
	"strings"
)

var clientData = client.LoadUserData()

func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome, enter your command:")
	fmt.Println("---------------------")
	for {
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("help", text) == 0 {

			fmt.Println("---------------------------------------------------------------------")
			fmt.Println("|\t\t Command \t\t|\t\t\t\t Description \t\t\t\t|")
			fmt.Println("---------------------------------------------------------------------")
			fmt.Println("| liststored \t\t\t|\t Show all stored files \t\t\t\t\t|")
			fmt.Println("| listlocal \t\t\t|\t Show all local files \t\t\t\t\t|")
			fmt.Println("| store <filename path> |\t Store a file on the BlockChain \t\t|")
			fmt.Println("| get   <filename>\t\t|\t Retrieve a file from the BlockChain \t|")
			fmt.Println("---------------------------------------------------------------------")

		} else if strings.Compare("liststored", text) == 0 {
			fmt.Println("Show list of files")
		} else if strings.Compare("listlocal", text) == 0 {
			client.ListAllLocalFiles()
		} else if strings.Compare("store", text) == 0 {
			fmt.Println("Send request to BC to store a file")
		} else if strings.Compare("get", text) == 0 {
			fmt.Println("Send request to BC to get a file ")
		} else {
			fmt.Println("Invalid command, type help for command info")
		}
	}
}
