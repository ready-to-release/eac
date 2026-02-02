"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.StableStatusBar = void 0;
const vscode = __importStar(require("vscode"));
const progress_frame_buffer_1 = require("./progress-frame-buffer");
/**
 * Stable status bar with frame buffer integration
 * Provides smooth, jitter-free progress display with fixed-width formatting
 */
class StableStatusBar {
    constructor(context) {
        this.startTime = 0;
        this.currentIcon = "$(robot)";
        this.isActive = false;
        this.buffer = new progress_frame_buffer_1.ProgressFrameBuffer(20);
        // Create status bar item on the right side
        this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.statusBarItem.text = "$(robot) Commit Message AI";
        this.statusBarItem.tooltip = "Commit Message AI is active";
        this.statusBarItem.command = "vscode-ext-commit.callMCP";
        this.statusBarItem.show();
        context.subscriptions.push(this.statusBarItem);
    }
    /** Start generation mode with spinning icon and smooth updates */
    startGeneration() {
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
    stopGeneration() {
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
    addProgress(message) {
        this.buffer.addProgress(message);
    }
    /** Show priority event (e.g., fun text) - auto-clears after duration */
    showEvent(message, durationMs = 5000) {
        this.buffer.pushEvent(message);
        setTimeout(() => this.buffer.clearEvent(), durationMs);
    }
    /** Get elapsed time formatted as "00m00s" or " 00s " (6 chars fixed) */
    getElapsedTime() {
        if (!this.isActive)
            return "  0s  ";
        const elapsedSeconds = Math.floor((Date.now() - this.startTime) / 1000);
        const mins = Math.floor(elapsedSeconds / 60);
        const secs = elapsedSeconds % 60;
        if (mins > 0) {
            return `${String(mins).padStart(2, '0')}m${String(secs).padStart(2, '0')}s`;
        }
        else {
            return ` ${String(secs).padStart(2, '0')}s `;
        }
    }
    /** Update status bar with stable format: [Icon] [Time] Message */
    update() {
        const time = this.getElapsedTime();
        const message = this.buffer.getCurrentFrame();
        // Fixed format: Icon (12ch approx) + Time (6ch) + Message
        this.statusBarItem.text = `${this.currentIcon} ${time} ${message}`;
    }
    /** Get buffer for external access (e.g., for fun text injection) */
    getBuffer() {
        return this.buffer;
    }
    /** Dispose resources */
    dispose() {
        if (this.updateTimer) {
            clearInterval(this.updateTimer);
        }
        this.statusBarItem.dispose();
    }
}
exports.StableStatusBar = StableStatusBar;
//# sourceMappingURL=stable-status-bar.js.map