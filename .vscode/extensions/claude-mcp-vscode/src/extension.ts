import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';

export function activate(context: vscode.ExtensionContext) {
    console.log('Claude MCP VSCode extension is now active');

    // Register the command
    let disposable = vscode.commands.registerCommand('claude-mcp-vscode.callMCP', async () => {
        try {
            // Get the workspace root
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                vscode.window.showErrorMessage('No workspace folder found');
                return;
            }

            const workspacePath = workspaceFolder.uri.fsPath;
            const agentFilePath = path.join(workspacePath, '.claude', 'agents', 'vscode-ext-commit-button.md');

            // Execute the agent with progress indicator
            const commitMessage = await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Generating semantic commit message",
                cancellable: false
            }, async (progress) => {
                // Simulate progress updates
                progress.report({ increment: 0, message: "Analyzing git changes..." });

                // Track last real progress message
                let lastRealProgress = "";

                // Start the actual agent execution with real progress callback
                const agentPromise = executeAgent(workspacePath, agentFilePath, (realProgress) => {
                    // Update with real progress from Go process
                    lastRealProgress = realProgress;
                    progress.report({ message: realProgress });
                });

                // Simulate progress while waiting (0-300 scale, mapped to 0-100%)
                // Updates every 3s, increments by 14 → reaches "Finalizing..." at ~60s
                // Operation typically completes at ~67s, leaving only ~7s at final stage
                let currentProgress = 0;
                const progressInterval = setInterval(() => {
                    if (currentProgress < 280) {
                        currentProgress += 14;
                        const messages = [
                            "Reading documentation...",
                            "Analyzing commit history...",
                            "Determining module changes...",
                            "Calculating version impacts...",
                            "Generating commit structure...",
                            "Applying semantic conventions...",
                            "Validating module boundaries...",
                            "Computing glob patterns...",
                            "Optimizing message format...",
                            "Formatting commit message...",
                            "Verifying 50/72 rule compliance...",
                            "Finalizing...",
                        ];
                        const messageIndex = Math.floor(currentProgress / 25);
                        // Use real progress if available, otherwise use simulated
                        const displayMessage = lastRealProgress || messages[messageIndex] || "Processing...";
                        progress.report({
                            increment: 4.5, // 20 updates × 4.5% = 90%
                            message: displayMessage
                        });
                    }
                }, 3000); // Update every 3 seconds

                try {
                    const result = await agentPromise;
                    clearInterval(progressInterval);
                    progress.report({ increment: 10, message: "Complete!" });
                    return result;
                } catch (error) {
                    clearInterval(progressInterval);
                    throw error;
                }
            });

            // Get the Git extension API
            const gitExtension = vscode.extensions.getExtension('vscode.git');
            if (!gitExtension) {
                vscode.window.showErrorMessage('Git extension not found');
                return;
            }

            const git = gitExtension.isActive ? gitExtension.exports : await gitExtension.activate();
            const api = git.getAPI(1);

            if (api.repositories.length === 0) {
                vscode.window.showErrorMessage('No Git repository found');
                return;
            }

            // Set the commit message in the first repository
            const repo = api.repositories[0];
            repo.inputBox.value = commitMessage;

            vscode.window.showInformationMessage('✓ Commit message generated successfully!');

        } catch (error) {
            vscode.window.showErrorMessage(`Error: ${error}`);
        }
    });

    context.subscriptions.push(disposable);
}

async function executeAgent(workspacePath: string, agentFilePath: string, onProgress?: (message: string) => void): Promise<string> {
    return new Promise((resolve, reject) => {
        // Execute the agent using the MCP server
        const mcpServerPath = path.join(workspacePath, 'src', 'mcp', 'vscode');

        // No API key needed - MCP server uses Claude Code CLI
        const childProcess = child_process.spawn('go', ['run', '.'], {
            cwd: mcpServerPath,
            stdio: ['pipe', 'pipe', 'pipe'],
            env: process.env
        });

        let outputLines: string[] = [];
        let errorOutput = '';
        let buffer = '';

        // Read stdout line by line for streaming progress
        childProcess.stdout.on('data', (data) => {
            buffer += data.toString();

            // Process complete lines
            let newlineIndex;
            while ((newlineIndex = buffer.indexOf('\n')) !== -1) {
                const line = buffer.substring(0, newlineIndex);
                buffer = buffer.substring(newlineIndex + 1);

                if (line.trim()) {
                    outputLines.push(line);

                    // Try to parse as progress notification
                    try {
                        const parsed = JSON.parse(line);
                        if (parsed.method === '$/progress' && parsed.params && onProgress) {
                            onProgress(parsed.params.message);
                        }
                    } catch (e) {
                        // Not JSON or not a progress notification, that's fine
                    }
                }
            }
        });

        childProcess.stderr.on('data', (data) => {
            errorOutput += data.toString();
        });

        childProcess.on('close', (code) => {
            if (code !== 0 && errorOutput) {
                reject(new Error(`MCP server error: ${errorOutput}`));
                return;
            }

            try {
                // The last line should be the JSON-RPC response
                const lastLine = outputLines[outputLines.length - 1];
                const response = JSON.parse(lastLine);

                if (response.error) {
                    reject(new Error(response.error.message));
                    return;
                }

                // Extract the commit message from the tool result
                if (response.result && response.result.content && response.result.content[0]) {
                    const commitMessage = response.result.content[0].text;
                    resolve(commitMessage);
                } else {
                    reject(new Error('Invalid response format from MCP server'));
                }
            } catch (parseError) {
                reject(new Error(`Failed to parse MCP response: ${parseError}. Output: ${outputLines.join('\n')}`));
            }
        });

        // Send the agent file path as input
        const request = {
            jsonrpc: '2.0',
            id: 1,
            method: 'tools/call',
            params: {
                name: 'execute-agent',
                arguments: {
                    agentFile: agentFilePath
                }
            }
        };

        childProcess.stdin.write(JSON.stringify(request) + '\n');
        childProcess.stdin.end();
    });
}

export function deactivate() {
    console.log('Claude MCP VSCode extension is now deactivated');
}
