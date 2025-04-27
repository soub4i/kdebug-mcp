# KDebug - Kubernetes Debugging MCP Server

KDebug is a Kubernetes debugging tool that allows you to interact with your Kubernetes clusters through Claude AI. It uses the Model Control Protocol (MCP) to enable Claude to execute Kubernetes commands on your behalf.

<p align="center" width="100%">

![](demo.gif)

</p>

## Overview

This tool allows you to:
- Inspect Kubernetes resources (pods, services, deployments, etc.)
- View pod logs
- Check node status
- Monitor events
- And more, all through conversational AI

## Installation

### Prerequisites

- Go 1.18+ (if building from source)
- Kubernetes cluster access configured (`~/.kube/config`)
- Claude client with MCP support

### Option 1: Download pre-built binary

1. Go to the GitHub releases page:

```
https://github.com/soub4i/kdebug-mcp/releases
```

2. Download the latest release for your platform (macOS, Linux, Windows)

3. Make the binary executable:

```bash
chmod +x kdebug-mcp
```

4. Move the binary to a location in your PATH:

```bash
mv kdebug-mcp /usr/local/bin/kdebug-mcp
# or 
mv kdebug-mcp ~/bin/kdebug-mcp  # If you have ~/bin in your PATH
```

### Option 2: Build from source

1. Clone the repository:

```bash
git clone https://github.com/soub4i/kdebug-mcp.git
cd kdebug-mcp
```

2. Build the binary:

```bash
go build -o bin/server ./cmd/server/main.go
```

## Configuration

### Configure Claude to use KDebug

Create or edit the Claude MCP configuration file located at:

- macOS: `~/Library/Application Support/com.anthropic.claude/config.json`
- Linux: `~/.config/com.anthropic.claude/config.json`
- Windows: `%APPDATA%\com.anthropic.claude\config.json`

Add the following configuration:

```json
{
    "mcpServers": {
        "kdebug": {
            "command": "/path/to/kdebug-mcp/bin/server"
        }
    }
}
```

Replace `/path/to/kdebug-mcp/bin/server` with the actual path to your KDebug binary.

### Kubernetes Context

KDebug uses your current Kubernetes context from `~/.kube/config`. Make sure your Kubernetes configuration is properly set up.

To switch contexts, you can use:

```bash
kubectl config use-context <context-name>
```

## Usage

1. Start Claude and make sure it's connected to your KDebug MCP server
2. In your conversation with Claude, ask about your Kubernetes resources

Example prompts:
- "what's wrong with my cluster.context: minikube. namespace: default."
- "Show me all pods in the default namespace"
- "What services are running in the kube-system namespace?"
- "Get the logs from pod xyz in namespace abc"
- "List all nodes in my cluster"
- "Check for recent events in the default namespace"

## Available Commands

KDebug provides access to the following Kubernetes resources:

- `nodes`: List all nodes in the cluster
- `pods`: List pods in a namespace or get a specific pod
- `podLogs`: Get logs from a specific pod
- `services`: List services in a namespace or get a specific service
- `deployments`: List deployments in a namespace or get a specific deployment
- `statefulsets`: List stateful sets in a namespace or get a specific stateful set
- `replicasets`: List replica sets in a namespace or get a specific replica set
- `daemonsets`: List daemon sets in a namespace or get a specific daemon set
- `events`: List events in a namespace or related to a specific resource

## Troubleshooting

### Common Issues

1. **Claude can't connect to KDebug**
   - Check that the path in your Claude MCP configuration is correct
   - Ensure the KDebug binary is executable

2. **Permission errors**
   - Make sure your Kubernetes configuration (`~/.kube/config`) has the necessary permissions
   - Try running `kubectl get pods` to verify your Kubernetes access

3. **Context switching**
   - If you have multiple Kubernetes contexts, Claude will ask which context to use
   - Ensure you have access to the context you're trying to use

## Security Considerations

KDebug executes Kubernetes commands with the permissions of your current user. Be mindful that Claude will have the same access to your Kubernetes cluster as you do through your configured kubectl context.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.


