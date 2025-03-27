package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig holds SSH connection parameters
type SSHConfig struct {
	Hosts    []string // List of hostnames or IP addresses
	Port     string
	User     string
	Password string
	KeyPath  string
	Timeout  time.Duration // SSH connection timeout
}

// CommandResult stores the output of a command for a specific host
type CommandResult struct {
	Hostname string
	Output   string
	Error    error
}

// Inventory holds the list of hosts from the inventory file
type Inventory struct {
	Hosts []string `json:"hosts"`
}

// ExecuteRemoteCommands runs commands on multiple hosts
func ExecuteRemoteCommands(config SSHConfig, commands []string) []CommandResult {
	var wg sync.WaitGroup
	results := make(chan CommandResult, len(config.Hosts))

	for _, host := range config.Hosts {
		wg.Add(1)
		go func(currentHost string) {
			defer wg.Done()

			hostResult := CommandResult{Hostname: currentHost}

			var authMethod ssh.AuthMethod
			if config.KeyPath != "" {
				key, err := os.ReadFile(config.KeyPath)
				if err != nil {
					hostResult.Error = fmt.Errorf("read key: %v", err)
					results <- hostResult
					return
				}
				signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					hostResult.Error = fmt.Errorf("parse key: %v", err)
					results <- hostResult
					return
				}
				authMethod = ssh.PublicKeys(signer)
			} else if config.Password != "" {
				authMethod = ssh.Password(config.Password)
			} else {
				hostResult.Error = fmt.Errorf("no auth method")
				results <- hostResult
				return
			}

			clientConfig := &ssh.ClientConfig{
				User:            config.User,
				Auth:            []ssh.AuthMethod{authMethod},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Timeout:         config.Timeout,
			}

			port := config.Port
			if port == "" {
				port = "22"
			}
			addr := fmt.Sprintf("%s:%s", currentHost, port)

			client, err := ssh.Dial("tcp", addr, clientConfig)
			if err != nil {
				hostResult.Error = fmt.Errorf("dial %s: %v", currentHost, err)
				results <- hostResult
				return
			}
			defer client.Close()

			var combinedOutput string
			for _, cmd := range commands {
				session, err := client.NewSession()
				if err != nil {
					hostResult.Error = fmt.Errorf("session: %v", err)
					results <- hostResult
					return
				}

				stdout, err := session.StdoutPipe()
				if err != nil {
					hostResult.Error = fmt.Errorf("stdout pipe: %v", err)
					session.Close()
					results <- hostResult
					return
				}

				err = session.Start(cmd)
				if err != nil {
					combinedOutput += fmt.Sprintf("Command '%s' failed: %v\n", cmd, err)
					session.Close()
					continue
				}

				buf, _ := io.ReadAll(stdout)
				session.Wait()
				session.Close()
				combinedOutput += fmt.Sprintf("Command '%s' output:\n%s\n", cmd, string(buf))
			}

			hostResult.Output = combinedOutput
			results <- hostResult
		}(host)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var commandResults []CommandResult
	for result := range results {
		commandResults = append(commandResults, result)
	}

	return commandResults
}

func main() {
	keyPath := flag.String("key", "", "Path to the SSH private key")
	user := flag.String("user", "", "SSH username")
	inventoryFile := flag.String("inventory", "inventory.json", "Path to inventory file")
	flag.Parse()

	if *user == "" {
		fmt.Fprintln(os.Stderr, "Error: SSH user must be specified with --user")
		os.Exit(1)
	}

	fileContent, err := os.ReadFile(*inventoryFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading inventory: %v\n", err)
		os.Exit(1)
	}

	var inventory Inventory
	err = json.Unmarshal(fileContent, &inventory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing inventory: %v\n", err)
		os.Exit(1)
	}

	config := SSHConfig{
		Hosts:   inventory.Hosts,
		User:    *user,
		KeyPath: *keyPath,
		Timeout: 5 * time.Second,
	}

	commands := []string{
		"uname -a",
		"df -h",
		"uptime",
		"free -h",
		"nproc",
	}

	results := ExecuteRemoteCommands(config, commands)

	for _, result := range results {
		fmt.Printf("\n\033[1;34m========== Host: %s ==========%s\n", result.Hostname, "\033[0m")
		if result.Error != nil {
			fmt.Printf("\033[0;31mError:\033[0m %v\n", result.Error)
		} else {
			fmt.Printf("%s\n", result.Output)
		}
	}
}
