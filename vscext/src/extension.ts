import * as vscode from 'vscode';
import * as path from 'path';

export function activate(context: vscode.ExtensionContext) {
    // Determine the correct binary based on platform and architecture
    let serverCommand: string;
    if (process.platform === 'win32') {
        if (process.arch === 'arm64') {
            serverCommand = context.asAbsolutePath(path.join('server', 'godevmcp-win-arm64.exe'));
        } else {
            serverCommand = context.asAbsolutePath(path.join('server', 'godevmcp-win64.exe'));
        }
    } else if (process.platform === 'darwin') {
        serverCommand = context.asAbsolutePath(path.join('server', 'godevmcp-darwin-arm64'));
    } else {
        // Default to x64 for Linux, x86 is rarely used
        serverCommand = context.asAbsolutePath(path.join('server', 'godevmcp-linux-amd64'));
    }

    const provider: vscode.McpServerDefinitionProvider = {
        provideMcpServerDefinitions(_token: vscode.CancellationToken): vscode.ProviderResult<vscode.McpServerDefinition[]> {
            const serverDefinition: vscode.McpServerDefinition = {
                label: 'Go Development MCP Server',
                command: serverCommand,
                args: ['serve'],
                env: {}
            };
            return [serverDefinition];
        }
    };

    context.subscriptions.push(vscode.lm.registerMcpServerDefinitionProvider('goDevMcpProvider', provider));
}

export function deactivate() {}

