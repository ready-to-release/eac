import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import { StableStatusBar } from './stable-status-bar';

// Global flag to prevent concurrent commit generation
let isGeneratingCommit = false;

// Debug log file path
let debugLogPath: string = '';

// Output channel for extension logs
let outputChannel: vscode.OutputChannel;

// Stable status bar instance
let stableStatusBar: StableStatusBar;

function log(message: string) {
    const timestamp = new Date().toISOString();
    const logLine = `[${timestamp}] ${message}\n`;
    console.log(logLine.trim());
    if (debugLogPath) {
        fs.appendFileSync(debugLogPath, logLine);
    }
    if (outputChannel) {
        outputChannel.appendLine(message);
    }
}

export function activate(context: vscode.ExtensionContext) {
    console.log('✓ Commit Message AI extension activated successfully');

    // Create output channel
    outputChannel = vscode.window.createOutputChannel('Commit Message AI');
    context.subscriptions.push(outputChannel);
    outputChannel.appendLine('✓ Commit Message AI extension activated');

    // Write activation log to file immediately
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (workspaceFolder) {
        const outDir = path.join(workspaceFolder.uri.fsPath, 'out');
        if (!fs.existsSync(outDir)) {
            fs.mkdirSync(outDir, { recursive: true });
        }
        const debugPath = path.join(outDir, 'git-state-debug.log');
        const timestamp = new Date().toISOString();
        fs.writeFileSync(debugPath, `[${timestamp}] Extension activated\n`);
    }

    // Create stable status bar
    stableStatusBar = new StableStatusBar(context);
    context.subscriptions.push(stableStatusBar);

    // Set initial context values (will be updated by git listener)
    vscode.commands.executeCommand('setContext', 'vscode-commit.hasStagedChanges', false);
    vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', false);

    // Initialize git change listener to enable/disable button
    const initializeGitListener = async () => {
        const gitExtension = vscode.extensions.getExtension('vscode.git');
        if (!gitExtension) {
            outputChannel.appendLine('⚠️ Git extension not found');
            return;
        }

        // Wait for git extension to activate
        const gitAPI = gitExtension.isActive ? gitExtension.exports : await gitExtension.activate();
        const git = gitAPI.getAPI(1);

        // Function to update button state
        const updateButtonState = () => {
            const debugInfo: string[] = [];

            if (git.repositories.length > 0) {
                const repo = git.repositories[0];
                const indexChanges = repo.state.indexChanges || [];
                const hasStagedChanges = indexChanges.length > 0;

                // Debug: Log full state
                debugInfo.push(`[Git State Debug]`);
                debugInfo.push(`  Repositories: ${git.repositories.length}`);
                debugInfo.push(`  Index changes: ${indexChanges.length}`);
                debugInfo.push(`  Working tree changes: ${repo.state.workingTreeChanges?.length || 0}`);
                debugInfo.push(`  Has staged changes: ${hasStagedChanges}`);

                if (indexChanges.length > 0) {
                    debugInfo.push(`  Staged files: ${indexChanges.map((c: any) => c.uri.fsPath).join(', ')}`);
                }

                // Log to output channel
                debugInfo.forEach(line => outputChannel.appendLine(line));

                // Also write to debug file
                const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
                if (workspaceFolder) {
                    const debugPath = path.join(workspaceFolder.uri.fsPath, 'out', 'git-state-debug.log');
                    const timestamp = new Date().toISOString();
                    fs.appendFileSync(debugPath, `\n[${timestamp}]\n${debugInfo.join('\n')}\n`);
                }

                vscode.commands.executeCommand('setContext', 'vscode-commit.hasStagedChanges', hasStagedChanges);
            } else {
                debugInfo.push('[Git State] No repositories found');
                outputChannel.appendLine(debugInfo[0]);

                const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
                if (workspaceFolder) {
                    const debugPath = path.join(workspaceFolder.uri.fsPath, 'out', 'git-state-debug.log');
                    const timestamp = new Date().toISOString();
                    fs.appendFileSync(debugPath, `\n[${timestamp}]\n${debugInfo[0]}\n`);
                }

                vscode.commands.executeCommand('setContext', 'vscode-commit.hasStagedChanges', false);
            }
        };

        // Wait for git to initialize and check state multiple times
        // Git state might not be ready immediately, so we retry a few times
        let attempts = 0;
        const maxAttempts = 5;
        const checkInterval = 200; // ms

        const tryUpdateState = async () => {
            attempts++;
            updateButtonState();

            // If no repositories found and we haven't exceeded attempts, try again
            if (git.repositories.length === 0 && attempts < maxAttempts) {
                outputChannel.appendLine(`  Retrying in ${checkInterval}ms... (attempt ${attempts}/${maxAttempts})`);
                await new Promise(resolve => setTimeout(resolve, checkInterval));
                return tryUpdateState();
            }
        };

        await tryUpdateState();

        // Listen for repository changes
        git.repositories.forEach((repo: any) => {
            repo.state.onDidChange(() => {
                outputChannel.appendLine('[Git State] Change detected, updating button state...');
                updateButtonState();
            });
        });

        // Listen for new repositories
        git.onDidOpenRepository((repo: any) => {
            outputChannel.appendLine('[Git State] New repository opened');
            updateButtonState();
            repo.state.onDidChange(() => {
                outputChannel.appendLine('[Git State] Change detected in new repo, updating button state...');
                updateButtonState();
            });
        });
    };

    // Initialize git listener asynchronously
    initializeGitListener().catch(err => {
        console.error('Failed to initialize git listener:', err);
        outputChannel.appendLine(`❌ Failed to initialize git listener: ${err}`);
        // Set to false as fallback
        vscode.commands.executeCommand('setContext', 'vscode-commit.hasStagedChanges', false);
    });

    // Register no-op command for disabled state - now injects fun text into status bar
    let clickCount = 0;
    let noOpDisposable = vscode.commands.registerCommand('vscode-commit.noOp', () => {
        if (isGeneratingCommit) {
            clickCount++;
            const messages = [
                '🔄 Working on it...',
                '🧘 Still working... patience is a virtue!',
                '⏳ Hey, I said working on it!',
                '🤖 Seriously, give me a moment...',
                '😅 Are you testing my patience?',
                '🎮 OK, now you\'re just clicking for fun...',
                '✨ I\'m doing AI magic here! Takes time!',
                '🐌 Each click makes it slower... just kidding!',
                '☕ Brewing the perfect commit message...',
                '🏛️ Rome wasn\'t built in a day!',
                '🎉 You must be really excited! Me too!',
                '👀 Pro tip: Watching the pot doesn\'t make it boil faster',
                '🤖 Computing... beep boop beep...',
                '📊 Nearly there... (not really, same progress)',
                '💝 Your enthusiasm is noted and appreciated!'
            ];
            const messageIndex = Math.min(clickCount - 1, messages.length - 1);

            // Inject fun text into status bar instead of popup
            stableStatusBar.showEvent(messages[messageIndex], 5000);
        } else {
            clickCount = 0; // Reset when not generating
            vscode.window.showInformationMessage('Stage some changes first to generate a commit message');
        }
    });
    context.subscriptions.push(noOpDisposable);

    // Register command to show output channel
    let showOutputDisposable = vscode.commands.registerCommand('vscode-commit.showOutput', () => {
        outputChannel.show(true); // true = preserve focus on current editor
    });
    context.subscriptions.push(showOutputDisposable);

    // Register the command
    let disposable = vscode.commands.registerCommand('vscode-commit.callMCP', async () => {
        // Set up debug log file in /out
        const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
        if (workspaceFolder) {
            const outDir = path.join(workspaceFolder.uri.fsPath, 'out');
            if (!fs.existsSync(outDir)) {
                fs.mkdirSync(outDir, { recursive: true });
            }
            debugLogPath = path.join(outDir, 'vscode-ext-claude-commitension-debug.log');
            // Clear previous log
            fs.writeFileSync(debugLogPath, '');
        }

        log('=== Commit Message AI Button Clicked ===');

        // Check if already generating
        if (isGeneratingCommit) {
            vscode.window.showWarningMessage('A commit message is already being generated. Please wait...');
            return;
        }

        isGeneratingCommit = true;

        // Start stable status bar generation mode
        stableStatusBar.startGeneration();

        // Update command to show it's working (this affects the SCM button too)
        vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', true);

        try {
            // FAIL-EARLY: Validate git state FIRST before doing anything else
            const gitExtension = vscode.extensions.getExtension('vscode.git');
            if (!gitExtension) {
                vscode.window.showErrorMessage('Git extension not found');
                isGeneratingCommit = false;
                return;
            }

            const git = gitExtension.isActive ? gitExtension.exports : await gitExtension.activate();
            const api = git.getAPI(1);

            if (api.repositories.length === 0) {
                vscode.window.showErrorMessage('No Git repository found');
                isGeneratingCommit = false;
                return;
            }

            const repo = api.repositories[0];

            // Validate git state IMMEDIATELY - fail early if invalid
            const validationError = await validateGitState(repo);
            if (validationError) {
                vscode.window.showErrorMessage(validationError);
                isGeneratingCommit = false;
                stableStatusBar.stopGeneration();
                vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', false);
                return;
            }

            // Check if commit input box already has content
            const existingMessage = repo.inputBox.value?.trim();
            if (existingMessage && existingMessage.length > 0) {
                // Stop status bar progress while waiting for user decision
                stableStatusBar.stopGeneration();
                vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', false);

                const overwrite = await vscode.window.showWarningMessage(
                    'Generating a new message will overwrite the existing',
                    { modal: false },
                    'OK'
                );

                if (overwrite !== 'OK') {
                    // User dismissed or cancelled
                    isGeneratingCommit = false;
                    return;
                }

                // User confirmed - restart status bar and continue
                stableStatusBar.startGeneration();
                vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', true);
            }

            // Only proceed with workspace and agent setup if git state is valid
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                vscode.window.showErrorMessage('No workspace folder found');
                isGeneratingCommit = false;
                return;
            }

            const workspacePath = workspaceFolder.uri.fsPath;

            // Execute the agent with simplified progress
            let commitMessage: string;
            try {
                // Simplified SCM progress - just shows stage, details in status bar
                commitMessage = await vscode.window.withProgress({
                    location: vscode.ProgressLocation.SourceControl,
                    title: "Generating commit message",
                    cancellable: false
                }, async (progress: vscode.Progress<{ message?: string; increment?: number }>) => {
                    // Show initial message in SCM panel
                    const randomMessages = [
                        "🚀 Initializing workflow...",
                        "🔍 Preparing generation...",
                        "⚡ Starting analysis...",
                        "🎯 Launching commit message...",
                    ];
                    const initialMsg = randomMessages[Math.floor(Math.random() * randomMessages.length)];
                    progress.report({ message: initialMsg });
                    stableStatusBar.addProgress(initialMsg);

                    // Track current stage for SCM display
                    let lastStage = 'Initializing';

                    // Start the actual commit message execution with real progress callback
                    const agentPromise = executeAgent(workspacePath, (realProgress) => {
                        log('[commit message Progress] ' + realProgress);

                        // Add to status bar buffer
                        if (realProgress && realProgress.length > 0 && realProgress.length < 200) {
                            stableStatusBar.addProgress(realProgress);
                        }

                        // Extract stage from message for SCM display
                        if (realProgress.includes('Generating') || realProgress.startsWith('🤖')) {
                            lastStage = 'Generating';
                            outputChannel.appendLine('🤖 ' + realProgress);
                        } else if (realProgress.includes('Cleaning up') || realProgress.includes('Auto-cleanup')) {
                            lastStage = 'Cleaning up';
                            outputChannel.appendLine('🔧 ' + realProgress);
                        } else if (realProgress.includes('Auto-fix') || realProgress.includes('Feeding errors')) {
                            lastStage = 'Auto-fixing';
                            outputChannel.appendLine('🔧 ' + realProgress);
                        } else if (realProgress.includes('complete')) {
                            lastStage = 'Complete';
                            outputChannel.appendLine('✅ ' + realProgress);
                        } else {
                            // Whimsical progress messages
                            outputChannel.appendLine('   ' + realProgress);
                        }

                        // Update SCM panel with stage info
                        progress.report({ message: lastStage });
                    });

                    try {
                        const result = await agentPromise;
                        progress.report({ message: "Complete!" });
                        return result;
                    } catch (error) {
                        throw error;
                    }
                });
            } catch (error) {
                // Error occurred during agent execution - do NOT show success message
                isGeneratingCommit = false;
                throw error;
            }

            // Only reach here if commit message was successfully generated
            // Set the commit message in the repository input box
            repo.inputBox.value = commitMessage;

            // Check if there are validation errors/warnings in the message
            if (commitMessage.includes('❌')) {
                vscode.window.showErrorMessage('❌ Commit message has validation errors. Please review and fix in the commit box before committing.');
            } else if (commitMessage.includes('⚠️')) {
                vscode.window.showWarningMessage('⚠️ Commit message has validation warnings. Please review in the commit box.');
            } else {
                vscode.window.showInformationMessage('✓ Commit message generated and ready to review!');
            }

        } catch (error) {
            vscode.window.showErrorMessage(`Error: ${error}`);
            // Ensure cleanup happens even if error display fails
            isGeneratingCommit = false;
            clickCount = 0;
            stableStatusBar.stopGeneration();
            vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', false);
        } finally {
            // Final safety net - always reset the flag and status bar
            isGeneratingCommit = false;
            clickCount = 0;
            stableStatusBar.stopGeneration();
            vscode.commands.executeCommand('setContext', 'vscode-commit.isGenerating', false);
            outputChannel.appendLine('[Extension] Generation completed, flag reset');
        }
    });

    context.subscriptions.push(disposable);
}

/**
 * Validates the git repository state before generating commit message
 * @param repo The git repository
 * @returns Error message if validation fails, null if valid
 */
async function validateGitState(repo: any): Promise<string | null> {
    // Get the current repository state
    const state = repo.state;

    // Debug: Log the state to understand what we're working with
    console.log('Git State Debug:', {
        workingTreeChanges: state.workingTreeChanges?.length || 0,
        indexChanges: state.indexChanges?.length || 0,
        mergeChanges: state.mergeChanges?.length || 0,
        untrackedChanges: state.untrackedChanges?.length || 0
    });

    // Check if there are staged changes (indexChanges)
    const hasStagedChanges = (state.indexChanges?.length || 0) > 0;

    if (!hasStagedChanges) {
        return 'No staged changes found. Stage your changes before generating a commit message.';
    }

    // Allow generating commit messages with unstaged changes present
    // The commit message will only describe the staged changes
    return null;
}

async function executeAgent(workspacePath: string, onProgress?: (message: string) => void): Promise<string> {
    return new Promise((resolve, reject) => {
        // Execute commit message command directly (replaces old MCP server approach)
        // This leverages the 7-phase system with auto-cleanup and validation

        // Try pre-built binary first (faster, doesn't require Go in PATH)
        const isWindows = process.platform === 'win32';
        const binaryName = isWindows ? 'eac.exe' : 'eac';
        const binaryPath = path.join(workspacePath, 'out', 'tools', binaryName);

        let command: string;
        let args: string[];
        let cwd: string;

        if (fs.existsSync(binaryPath)) {
            // Use pre-built binary
            command = binaryPath;
            args = ['get', 'commit-message'];
            cwd = workspacePath;
            log(`[commit message] Using pre-built binary: ${binaryPath}`);
        } else {
            // Fall back to go run (requires Go in PATH)
            const cliPath = path.join(workspacePath, 'go', 'cli', 'eac');
            command = 'go';
            args = ['run', '.', 'get', 'commit-message'];
            cwd = cliPath;
            log(`[commit message] Pre-built binary not found, using go run from: ${cliPath}`);
        }

        // Call get commit-message which handles everything: generation, cleanup, validation, auto-fix
        const childProcess = child_process.spawn(command, args, {
            cwd: cwd,
            stdio: ['pipe', 'pipe', 'pipe'],
            env: process.env
        });

        let fullOutput = '';
        let errorOutput = '';

        // Capture all stdout
        childProcess.stdout.on('data', (data: Buffer) => {
            const text = data.toString();
            fullOutput += text;
            log('[commit message output] ' + text);

            // Extract progress indicators for real-time feedback
            const lines = text.split('\n');
            for (const line of lines) {
                const trimmed = line.trim();

                // Detect progress messages (emoji indicators)
                if (trimmed.startsWith('🤖')) {
                    if (onProgress) onProgress('Generating commit message...');
                } else if (trimmed.startsWith('🔧 Auto-cleanup')) {
                    if (onProgress) onProgress('Cleaning up output...');
                } else if (trimmed.startsWith('🔧 Attempting to auto-fix')) {
                    if (onProgress) onProgress('Auto-fixing validation errors...');
                } else if (trimmed.startsWith('🔄 Feeding validation errors')) {
                    if (onProgress) onProgress('Feeding errors to AI...');
                } else if (trimmed.startsWith('✅') && trimmed.includes('complete')) {
                    if (onProgress) onProgress('Generation complete!');
                } else if (trimmed.includes('Harmonizing module boundaries') ||
                          trimmed.includes('Contemplating the WHY') ||
                          trimmed.includes('Reticulating splines')) {
                    // Whimsical progress messages
                    if (onProgress) onProgress(trimmed);
                }
            }
        });

        // Capture stderr for errors
        childProcess.stderr.on('data', (data: Buffer) => {
            const stderrText = data.toString();
            errorOutput += stderrText;
            log('[commit message error] ' + stderrText);
        });

        // Handle spawn errors (e.g., command not found, path doesn't exist)
        childProcess.on('error', (err: Error) => {
            log('[commit message spawn error] ' + err.message);
            reject(new Error(`Failed to start commit message command: ${err.message}. Is 'go' installed and in PATH?`));
        });

        childProcess.on('close', (code: number | null) => {
            log(`[commit message] Process exited with code ${code}`);
            log(`[commit message] stdout length: ${fullOutput.length}, stderr length: ${errorOutput.length}`);

            // Extract commit message regardless of exit code (validation errors are included in output)
            const commitMessage = extractCommitMessageFromOutput(fullOutput);

            if (!commitMessage) {
                // Only fail if we can't extract anything at all
                let errorMsg = '';
                if (errorOutput.trim()) {
                    errorMsg = errorOutput.trim();
                } else if (fullOutput.trim()) {
                    errorMsg = fullOutput.trim();
                } else {
                    errorMsg = `Command exited with code ${code} but produced no output. Check the "Commit Message AI" output panel for details.`;
                }
                reject(new Error(`commit message failed: ${errorMsg}`));
                return;
            }

            // Exit code 1 means validation errors exist, but we still have a message to show
            // The validation errors are already included in the extracted message
            resolve(commitMessage);
        });

        // No input needed for commit message (it reads git directly)
        childProcess.stdin.end();
    });
}

/**
 * Extracts the commit message from commit message command output
 * Format:
 * 1. Whimsical progress messages (random order, shown during generation)
 * 2. ">>>>>>OUTPUT START<<<<<<" marker
 * 3. THE COMMIT MESSAGE (cleaned and validated)
 * 4. "\n---\n"
 * 5. Verification results (errors/warnings) - INCLUDED in output for user to see
 */
function extractCommitMessageFromOutput(output: string): string | null {
    const lines = output.split('\n');

    // Find the OUTPUT START marker
    let outputStartIndex = -1;
    for (let i = 0; i < lines.length; i++) {
        if (lines[i].trim() === '>>>>>>OUTPUT START<<<<<<') {
            outputStartIndex = i;
            break;
        }
    }

    if (outputStartIndex === -1) {
        log('[ERROR] No >>>>>>OUTPUT START<<<<<< marker found in output');
        return null;
    }

    // Extract EVERYTHING after the marker (including validation errors)
    // Start from line after marker, skip any empty lines at the start
    let messageStartIndex = outputStartIndex + 1;
    while (messageStartIndex < lines.length && lines[messageStartIndex].trim() === '') {
        messageStartIndex++;
    }

    // Get all lines from start to end
    const messageLines = lines.slice(messageStartIndex);
    const message = messageLines.join('\n').trim();

    return message.length > 0 ? message : null;
}

export function deactivate() {
    console.log('Claude MCP VSCode extension is now deactivated');
}
