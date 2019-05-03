# Decentralized Cloud Storage

Cloud storage is controlled by a few super large providers. Raising questions about data
protection, privacy, licensing, control and ownership of data. To pay for cloud storage is
expensive in the long run. Though uptime is pretty decent, there’s a chance of downtime due to
outages and DDos attack. Also, is your private data in the cloud really private? They are often
duplicated across multiple data centers and backup.
By moving cloud storage over to a blockchain your data will not only be a part of a large
decentralized network, due to the underlying technology of blockchain such as encryption keys
and one way hash function they can ensure total privacy, cheaper prices and higher availability
rate.

**Functionalities/Success**

  * Client can upload his own files
  * Client can retrieve his own files
  * Client can view their own list of files.
  * Clients can not view other users data.(Unless I will be able to enable file
sharing)


## TX fee and data integrity 

TX fees will be paid upon retrieval of a file. If the file matches what the receiver expects then a TX fee based on the file size will be rewarded to the miner. This will give the miners an incentive to keep the integrity of the data, as well as trying to respond as quickly as possible back to the client with the desired data. 

When clients receive the data from the miner, they will verify the integrity of the data and then pay the miner. For this application, we will assume that all clients are behaving, and thereby paying the miner as long as their file are not tampered with. The client will also be responsible for creating their own key pair(public and private key), that they will use to encrypt their data, and that will also be used for signature verification on retrieval. 
The blockchain will costist of two MPT's; 
 * One too store the transaction history for all the clients and miners. 
 * One to store the data(files) in each block. This mpt will remove all its entries once a block is created to ensure that we don't have to store the same file n times where n is the number of blocks after the first block containing the data. 


## How will the list of files be created and served
The client will be responsible for storing the necessary information about their stored files in a static file. This static file will be written too by the application when a file is stored, and will also be read once the application starts. This file will be extremely important to keep secure as it will have all the information necessary to retrieve clients files. 

For each file that is stored their required information will be stored as follows(**Note:** This might change): filename, key pair, file hash, block hash.

## Limitations 

Since the blockchain will be storing the actual data on the blockchain the price of storage will be fairly expensive since all miners must store the entire chain which includes the data. 

## Architecture
![alt text](https://github.com/usfcs686/cs686-blockchain-p3-gudbrandsc/blob/master/img/IMG_9183.jpg "Architecture")

## API 
**NOTE:** These methods are not final and may change during the development process.  Also the dataHash, and json body may change to also include the TimeStamp to allow a user to upload the same file.  

  **/store**  
  **Method:** POST  
  **Request:** 
  ```JavaScript
  {
   "EncryptedByteString": "String", // Using AES encryption
   "PublicKey" :  "String", // RSA Public key
   "Signature" :  "String", //Encrypted with RSA public key 
   "dataHash"  :  "String"
  }
```
  **Response:** 
  ```json
  {
   "BlockHash"       :  "String",
   "BlockHeight"     :  "Integer"
  }
```

  **Description:** Post method that will store a users file to the blockchain. The signature is used to ensure data integrity on retrieval, and is encrypted using RSA private key. The data will be encrypted using AES encryption. Response will contain information about where the data is stored.
 
**/retrieve/{filehash}/{blockHeight}/{blockHash}/**  
**Method:** GET  
**Response:** If the file exist return it to the client togheter with the TX fee amount for the file retrieval. If the file does not exist then return 404 not found.
  ```json
{
"TXfee": "integer",
"data":{
   "EncryptedByteString": "String",
   "DataHash"  :  "String",
   "PublicKey" :  "String",
   "Signature" :  "String"
   }
}
```
**Description:** Get method that allows a client to retrieve a file stored on the BlockChain  

**/payTx**  
**Method:** POST  
  **Request:** 
  ```json
  {
   "Amount": "Float",
   "publicKey": "String",
   "signature": "String"
  }
```
**Description:** Method to pay a miner for a file retrieval, the amount will be based on the response amount from the retrieve request.



## Timeline

| Date        | Milestone      
| ------------- |:-------------|
| 04/17/19      | <ul><li> [x] Finish project proposal.</li></ul>  |
| 04/24/19      | <ul><li> [x] Research exisitng DFS, and their architecture/implimentation </li><li> [ ] Implement basic application UI </li></ul>      | 
| 05/01/19      | <ul><li> [x] Finish Midpoint milestone   <ul><li>[x] Define API</li><li>[x] Define TX fee</li><li>[x] Define Data integrity</li><li>[x] Define Application architecture</li><li>[x] Finish project timeline </li></ul></li><li> [x] Implement basic application UI </li><li> [x] Client key pair generator </li><li> [x] Client storing and retrieval functions </li></ul>      |
| 05/05/19      | <ul><li> [x] Implement storing API </li><li> [x] Implement retrieval API </li><li> [ ] Implement TX API </li> </ul> |
| 05/10/19      | <ul><li> [ ] Finish testing and bug fixing </li></ul>      |
| 05/16/19      | Finished demo video, and project description.      |

**NOTE:** Some of these dates might seem random, but they are based on my current calendar and the days I have allocated to work on this project. This is a sunny day scenario where all goes well during the development so the final timeline might look a bit different.
