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
 • User can upload and delete his own files
 • User can view their own list of files.
 • Users can not view or modify other users data.(Unless I will be able to enable file
sharing)

User will encrypt using their own private key. 
Miner will take the encrypted data and hash it. 
Respond back to the user with the transaction hash. 
Add transaction to blockchain

**How to make sure that the miners do not change the content before putting it in the blockchain**
- Alice will create a data block including: byte array, encrypted(hash(bytearray)),hash(bytearray), public key
- When a transaction is added to a 


## How will the TX fee be decided
Yet to be decided

## How will the list of files be created and served

- User will create a data block to store in the blockchain. 
- User will then use his private key to encrypt the data block.
- The user will then send the datablock to a random node in the network. 
- The node will add the data block to the users data repo list
- When user wants to see the data he can use his private key to decrypt the data

## Architecture

## API 
