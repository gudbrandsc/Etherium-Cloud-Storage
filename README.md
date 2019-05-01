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

 • User can upload his own files
 • User can view their own list of files.
 • Users can not view other users data.(Unless I will be able to enable file
sharing)


## TX fee and data integrity 

TX fees will be paid upon retrieval of a file. If the file matches what the receiver expects then a TX fee based on the file size will be rewarded to the miner. This will give the miners an incentive to keep the integrity of the data, as well as trying to respond as quickly as possible back to the client with the desired data. 

When clients receive the data from the miner, they will verify the integrity of the data and then pay the miner. For this application, we will assume that all clients are behaving, and thereby paying the miner as long as their file is not tampered with.  


## How will the list of files be created and served
The client will be responsible for storing the necessary information about their stored files in a static file. This static file will be written too by the application when a file is stored, and will also be read once the application starts. This file will be extremely important to keep secure as it will have all the information necessary to retrieve clients files. 

For each file that is stored their required information will be stored as follows(**Note:** This might change): filename, key pair, file hash, block hash.

## Limitations 

Size the blockchain will be storing the actual data on the chain the price of storage will be fairly expensive since all miners must store the entire chain which includes the data. 

## Architecture

## API 

## Timeline
