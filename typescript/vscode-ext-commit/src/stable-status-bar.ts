import * as vscode from 'vscode';
import { ProgressFrameBuffer } from './progress-frame-buffer';

/**
 * Stable status bar with frame buffer integration
 * Provides smooth, jitter-free progress display with fixed-width formatting
 */
export class StableStatusBar {
  private buffer: ProgressFrameBuffer;
  private statusBarItem: vscode.StatusBarItem;
  private updateTimer: NodeJS.Timeout | undefined;
  private startTime: number = 0;
  private currentIcon: string = "$(robot)";
  private isActive: boolean = false;

  constructor(context: vscode.ExtensionContext) {
    this.buffer = new ProgressFrameBuffer(20);

    // Create status bar item on the right side
    this.statusBarItem = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Right,
      100
    );
    this.statusBarItem.text = "$(robot) Commit Message AI";
    this.statusBarItem.tooltip = "Commit Message AI is active";
    this.statusBarItem.command = "vscode-ext-commit.callMCP";
    this.statusBarItem.show();

    context.subscriptions.push(this.statusBarItem);
  }

  /** Start generation mode with spinning icon and smooth updates */
  startGeneration(): void {
    this.isActive = true;
    this.startTime = Date.now();
    this.currentIcon = "$(sync~spin)";
    this.statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.prominentBackground');
    this.statusBarItem.command = undefined; // Make non-clickable
    this.buffer.clear();

    // Start smooth update loop at 50ms (20fps)
    this.updateTimer = setInterval(() => {
      this.update();
    }, 50);
  }

  /** Stop generation mode, return to idle */
  stopGeneration(): void {
    this.isActive = false;
    this.currentIcon = "$(robot)";
    this.statusBarItem.backgroundColor = undefined;
    this.statusBarItem.command = "vscode-ext-commit.callMCP"; // Make clickable
    this.buffer.clear();

    if (this.updateTimer) {
      clearInterval(this.updateTimer);
      this.updateTimer = undefined;
    }

    // Reset to default text
    this.statusBarItem.text = "$(robot) Commit Message AI";
  }

  /** Add progress message to buffer */
  addProgress(message: string): void {
    this.buffer.addProgress(message);
  }

  /** Show priority event (e.g., fun text) - auto-clears after duration */
  showEvent(message: string, durationMs: number = 5000): void {
    this.buffer.pushEvent(message);
    setTimeout(() => this.buffer.clearEvent(), durationMs);
  }

  /** Get elapsed time formatted as "00m00s" or " 00s " (6 chars fixed) */
  private getElapsedTime(): string {
    if (!this.isActive) return "  0s  ";

    const elapsedSeconds = Math.floor((Date.now() - this.startTime) / 1000);
    const mins = Math.floor(elapsedSeconds / 60);
    const secs = elapsedSeconds % 60;

    if (mins > 0) {
      return `${String(mins).padStart(2, '0')}m${String(secs).padStart(2, '0')}s`;
    } else {
      return ` ${String(secs).padStart(2, '0')}s `;
    }
  }

  /** Update status bar with stable format: [Icon] [Time] Message */
  private update(): void {
    const time = this.getElapsedTime();
    const message = this.buffer.getCurrentFrame();

    // Fixed format: Icon (12ch approx) + Time (6ch) + Message
    this.statusBarItem.text = `${this.currentIcon} ${time} ${message}`;
  }

  /** Get buffer for external access (e.g., for fun text injection) */
  getBuffer(): ProgressFrameBuffer {
    return this.buffer;
  }

  /** Dispose resources */
  dispose(): void {
    if (this.updateTimer) {
      clearInterval(this.updateTimer);
    }
    this.statusBarItem.dispose();
  }
}
