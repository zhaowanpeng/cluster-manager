package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"zhaowanpeng/cluster-manager/model"

	"golang.org/x/crypto/ssh"
)

func SCP_Files(clients []model.ShellClient, source, destination string, recursive bool) {
	for _, client := range clients {
		if !client.Usable {
			fmt.Printf("Skipping unusable client: %s\n", client.IP)
			continue
		}

		err := scpFileToClient(client, source, destination, recursive)
		if err != nil {
			fmt.Printf("Error copying to %s: %v\n", client.IP, err)
		} else {
			fmt.Printf("Successfully copied to %s\n", client.IP)
		}
	}
}

func scpFileToClient(client model.ShellClient, source, destination string, recursive bool) error {
	addr := fmt.Sprintf("%s:%d", client.IP, client.Port)

	config := &ssh.ClientConfig{
		User:            client.User,
		Auth:            []ssh.AuthMethod{ssh.Password(client.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Prepare the file to copy
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	destinationFile := filepath.Join(destination, fileInfo.Name())

	// Create SCP command
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), fileInfo.Name())
		io.Copy(w, file)
		fmt.Fprint(w, "\x00")
	}()

	err = session.Run(fmt.Sprintf("scp -t %s", destinationFile))
	if err != nil {
		return fmt.Errorf("failed to run scp command: %v", err)
	}

	return nil
}
