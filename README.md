# SSH Command Executor

A Go-based tool for executing commands on multiple remote hosts via SSH. This tool allows you to run commands in parallel across multiple servers and collect their outputs.

## Features

- Parallel command execution across multiple hosts
- Support for both password and SSH key authentication
- Configurable connection timeout
- JSON-based inventory management
- Detailed command output collection
- Error handling and reporting per host

## Prerequisites

- Go 1.x or later
- SSH access to target hosts
- Either SSH key or password authentication credentials

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd <repository-name>
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
go build
```

## Usage

### Inventory File

Create a JSON file named `inventory.json` (or specify a different name) with the list of hosts:

```json
{
    "hosts": [
        "host1.example.com",
        "host2.example.com",
        "192.168.1.100"
    ]
}
```

### Running Commands

Basic usage with SSH key authentication:
```bash
./routine --user username --key /path/to/private/key
```

With password authentication:
```bash
./routine --user username
```

Using a custom inventory file:
```bash
./routine --user username --inventory custom_inventory.json
```

### Command Line Options

- `--user`: SSH username (required)
- `--key`: Path to SSH private key file (optional)
- `--inventory`: Path to inventory file (default: "inventory.json")

## Default Commands

The tool executes the following commands on each host by default:
- `uname -a`: System information
- `df -h`: Disk usage
- `uptime`: System uptime
- `free -h`: Memory usage
- `nproc`: Number of processors

## Output Format

The output is color-coded and organized by host:
- Host headers are displayed in blue
- Error messages are displayed in red
- Command outputs are displayed in the default terminal color

## Error Handling

- Connection errors are reported per host
- Command execution errors are captured and included in the output
- The tool continues execution even if some hosts fail

## Security Notes

- The tool uses `ssh.InsecureIgnoreHostKey()` for host key verification
- For production use, consider implementing proper host key verification
- Store sensitive credentials securely and use appropriate file permissions

## License

TBD