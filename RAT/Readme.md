#this code is a test of copilot ai, everything should be working thoug.
# Option 3 can be used to deploy a vulnerable software, lets call that BYOS
# For example libreoffice 6.1.0.1 can be transfered executed with a payload to trigger this https://github.com/rapid7/metasploit-framework/blob/master/documentation/modules/exploit/multi/fileformat/libreoffice_macro_exec.md
# Remote administration tool

This project contains two Go programs that enable secure communication between a server and a client using TLS and AES-GCM encryption. The server can send commands, executables to run in memory, or compressed folders to the client. The client receives these messages, decrypts any transmitted files, and processes them accordingly.

## Files Included

- **server.go**  
  Implements the server application which:
  - Listens on a TLS-encrypted TCP connection.
  - Presents a menu with three options for the operator:
    1. Send a command for execution.
    2. Send an executable to run in memory.
    3. Compress a folder (including subfolders), send it as a ZIP file, and instruct the client to extract and execute a specified tool.
  - Encrypts file contents using AES-256 (with GCM mode) before sending to the client.

- **client.go**  
  Implements the client application which:
  - Connects to the server via a TLS-encrypted connection.
  - Reads incoming instructions from the server.
  - Supports the **ARC:** command for decompressing a ZIP archive, extracting files to a temporary directory, and executing a specified executable with provided arguments.
  - Decrypts any transmitted file data using AES-256 (with GCM mode).

## How It Works

### TLS Configuration
- Both server and client use TLS for secure communication.
- The server loads certificate and key files (`server.crt` and `server.key`). Ensure these files are generated and placed in the same directory as the server executable.

### AES Encryption and Decryption
- A 32-byte key (`password`) is used for AES-256 encryption.
- **On the Server:**  
  The file data is encrypted before sending. A random nonce is generated (using the GCM's nonce size), then the nonce and ciphertext are concatenated and hex-encoded.
- **On the Client:**  
  The received hex data is converted back into bytes, the nonce is separated from the ciphertext, and the data is decrypted.

### Command Handling
- **Simple Command (CMD:)**  
  Not fully implemented in the client example. It sends a command string to the client.
  
- **In-Memory Execution (MEM:)**  
  Intended to send an executable along with optional arguments to be executed from memory.

- **Archive and Execute (ARC:)**  
  The server compresses a folder into a ZIP file, encrypts it, and sends a command formatted as:  
  `ARC:<executable-name>:<executable-arguments>:<hexEncoded(nonce|encryptedData)>`
  The client then:
  - Decrypts the archive.
  - Extracts the archive contents into a temporary folder.
  - Executes the specified executable with given arguments (setting the extraction folder as its working directory).

## Running the Applications

### Prerequisites
- Go installed on your machine.
- TLS certificate and key files for the server (`server.crt` and `server.key`).

### Build the Server
```bash
go build -o server server.go
```

### Build the Client
```bash
go build -o client client.go
```

### Running the Server
Ensure your certificate and key files are in the same directory as the server executable, then run:
```bash
./server
```
The server listens on `192.168.1.20:443` by default. Modify the `service` variable in the code if needed.

### Running the Client
Run the client with the server address as the first argument:
```bash
./client 192.168.1.20:443
```
The client establishes a secure connection to the server and waits for instructions.

## Notes
- **Security:** Although the AES key is hardcoded in this example, in a production system, you should manage keys securely.
- **Error Handling:** Both programs include basic error handling to log failures and print errors. Further enhancements may be required for robust deployments.
- **Extensibility:** The command protocol (using `CMD:`, `MEM:`, and `ARC:` prefixes) can be extended to support additional functionalities or more sophisticated command parsing.

Feel free to modify and expand this tool based on your requirements.
