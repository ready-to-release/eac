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

    // Create stable status bar
    stableStatusBar = new StableStatusBar(context);
    context.subscriptions.push(stableStatusBar);

    // Initialize git change listener to enable/disable button
    const initializeGitListener = async () => {
        const gitExtension = vscode.extensions.getExtension('vscode.git');
        if (!gitExtension) {
            return;
        }

        // Wait for git extension to activate
        const gitAPI = gitExtension.isActive ? gitExtension.exports : await gitExtension.activate();
        const git = gitAPI.getAPI(1);

        // Function to update button state
        const updateButtonState = () => {
            if (git.repositories.length > 0) {
                const repo = git.repositories[0];
                const hasStagedChanges = (repo.state.indexChanges?.length || 0) > 0;
                vscode.commands.executeCommand('setContext', 'vscode-ext-commit.hasStagedChanges', hasStagedChanges);
            } else {
                vscode.commands.executeCommand('setContext', 'vscode-ext-commit.hasStagedChanges', false);
            }
        };

        // Update immediately
        updateButtonState();

        // Listen for repository changes
        git.repositories.forEach((repo: any) => {
            repo.state.onDidChange(() => updateButtonState());
        });

        // Listen for new repositories
        git.onDidOpenRepository((repo: any) => {
            updateButtonState();
            repo.state.onDidChange(() => updateButtonState());
        });
    };

    // Initialize git listener asynchronously
    initializeGitListener().catch(err => {
        console.error('Failed to initialize git listener:', err);
        // Set to false as fallback
        vscode.commands.executeCommand('setContext', 'vscode-ext-commit.hasStagedChanges', false);
    });

    // Register no-op command for disabled state - now injects fun text into status bar
    let clickCount = 0;
    let noOpDisposable = vscode.commands.registerCommand('vscode-ext-commit.noOp', () => {
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
    let showOutputDisposable = vscode.commands.registerCommand('vscode-ext-commit.showOutput', () => {
        outputChannel.show(true); // true = preserve focus on current editor
    });
    context.subscriptions.push(showOutputDisposable);

    // Register the command
    let disposable = vscode.commands.registerCommand('vscode-ext-commit.callMCP', async () => {
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
        vscode.commands.executeCommand('setContext', 'vscode-ext-commit.isGenerating', true);

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
                vscode.commands.executeCommand('setContext', 'vscode-ext-commit.isGenerating', false);
                return;
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
                }, async (progress) => {
                    // Show initial message in SCM panel
                    const randomMessages = [
                        "🚀 Initializing workflow...",
                        "🔍 Preparing generation...",
                        "⚡ Starting analysis...",
                        "🎯 Launching commit-ai...",
                    ];
                    const initialMsg = randomMessages[Math.floor(Math.random() * randomMessages.length)];
                    progress.report({ message: initialMsg });
                    stableStatusBar.addProgress(initialMsg);

                    // Track current stage for SCM display
                    let lastStage = 'Initializing';

                    // Start the actual commit-ai execution with real progress callback
                    const agentPromise = executeAgent(workspacePath, (realProgress) => {
                        log('[commit-ai Progress] ' + realProgress);

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

            // Check if there are validation warnings in the message
            if (commitMessage.includes('⚠️ **VALIDATION ERRORS**')) {
                vscode.window.showWarningMessage('⚠️ Commit message generated with validation warnings. Please review and fix before committing.');
            } else {
                vscode.window.showInformationMessage('✓ Commit message generated and ready to review!');
            }

        } catch (error) {
            vscode.window.showErrorMessage(`Error: ${error}`);
        } finally {
            // Always reset the flag and status bar, even if there was an error
            isGeneratingCommit = false;
            clickCount = 0; // Reset click counter
            stableStatusBar.stopGeneration();
            vscode.commands.executeCommand('setContext', 'vscode-ext-commit.isGenerating', false);
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

    // Check if there are unstaged changes (working tree changes that aren't staged)
    // This includes both modified files and untracked files
    const hasUnstagedChanges = (state.workingTreeChanges?.length || 0) > 0;

    if (hasUnstagedChanges) {
        return 'You have unstaged changes. Please stage or stash them before generating a commit message.';
    }

    // All validations passed
    return null;
}

async function executeAgent(workspacePath: string, onProgress?: (message: string) => void): Promise<string> {
    return new Promise((resolve, reject) => {
        // Execute commit-ai command directly (replaces old MCP server approach)
        // This leverages the new 7-lever system with auto-cleanup and validation
        const commandsPath = path.join(workspacePath, 'src', 'commands');

        // Call commit-ai which handles everything: generation, cleanup, validation, auto-fix
        const childProcess = child_process.spawn('go', ['run', '.', 'commit-ai'], {
            cwd: commandsPath,
            stdio: ['pipe', 'pipe', 'pipe'],
            env: process.env
        });

        let fullOutput = '';
        let errorOutput = '';

        // Capture all stdout
        childProcess.stdout.on('data', (data) => {
            const text = data.toString();
            fullOutput += text;
            log('[commit-ai output] ' + text);

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
        childProcess.stderr.on('data', (data) => {
            const stderrText = data.toString();
            errorOutput += stderrText;
            log('[commit-ai error] ' + stderrText);
        });

        childProcess.on('close', (code) => {
            if (code !== 0) {
                const errorMsg = errorOutput || 'Command failed with exit code ' + code;
                reject(new Error(`commit-ai error: ${errorMsg}`));
                return;
            }

            // Extract commit message from output
            const commitMessage = extractCommitMessageFromOutput(fullOutput);

            if (!commitMessage) {
                reject(new Error('Failed to extract commit message from output'));
                return;
            }

            resolve(commitMessage);
        });

        // No input needed for commit-ai (it reads git directly)
        childProcess.stdin.end();
    });
}

/**
 * Extracts the commit message from commit-ai output
 * New simplified format:
 * 1. Vanity progress messages (🤖, optional 🔧 if auto-fix)
 * 2. THE COMMIT MESSAGE
 * 3. "---"
 * 4. Verification results
 */
function extractCommitMessageFromOutput(output: string): string | null {
    const lines = output.split('\n');

    // Find the separator "---"
    let separatorIndex = -1;
    for (let i = lines.length - 1; i >= 0; i--) {
        if (lines[i].trim() === '---') {
            separatorIndex = i;
            break;
        }
    }

    if (separatorIndex === -1) {
        // No separator found, something went wrong
        return null;
    }

    // Everything before "---" is the commit message (minus vanity lines at start)
    let messageStartIndex = 0;
    for (let i = 0; i < separatorIndex; i++) {
        const trimmed = lines[i].trim();

        // Skip vanity progress messages
        if (trimmed.startsWith('🤖') ||
            trimmed.startsWith('🔧') ||
            trimmed.startsWith('Discombobulating') ||
            trimmed.startsWith('Reticulating') ||
            trimmed.startsWith('Harmonizing') ||
            trimmed.startsWith('Contemplating') ||
            trimmed.startsWith('Ugh,') ||
            trimmed.startsWith('*Sigh*') ||
            trimmed.startsWith('Fine,') ||
            trimmed === '') {
            continue;
        }

        // Found the start of actual commit message
        messageStartIndex = i;
        break;
    }

    // Extract message from start to separator
    const messageLines = lines.slice(messageStartIndex, separatorIndex);
    const message = messageLines.join('\n').trim();

    return message.length > 0 ? message : null;
}

export function deactivate() {
    console.log('Claude MCP VSCode extension is now deactivated');
}
