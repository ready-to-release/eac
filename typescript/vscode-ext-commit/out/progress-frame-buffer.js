"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ProgressFrameBuffer = void 0;
/**
 * Slim FIFO buffer for progress messages with stable frame display
 * ~20 frames/minute, priority event override, always shows current stable frame
 */
class ProgressFrameBuffer {
    constructor(maxFrames = 20) {
        this.frames = [];
        this.priorityFrame = "";
        this.maxFrames = maxFrames;
    }
    /** Push high-priority event (overrides current display) */
    pushEvent(message) {
        this.priorityFrame = this.normalize(message);
    }
    /** Add progress message to FIFO buffer */
    addProgress(message) {
        const normalized = this.normalize(message);
        // Final verification: only add valid messages
        if (this.isValidMessage(normalized)) {
            this.frames.push(normalized);
            if (this.frames.length > this.maxFrames)
                this.frames.shift();
        }
    }
    /** Get current stable frame (priority > latest progress > fallback) */
    getCurrentFrame() {
        if (this.priorityFrame)
            return this.priorityFrame;
        if (this.frames.length > 0)
            return this.frames[this.frames.length - 1];
        return this.getFallbackText();
    }
    /** Clear priority event, falls back to progress buffer */
    clearEvent() {
        this.priorityFrame = "";
    }
    /** Normalize text: trim, collapse whitespace, limit to 40 chars, remove problematic patterns */
    normalize(text, maxLen = 40) {
        // Remove control characters and clean up whitespace
        let cleaned = text
            .replace(/[\x00-\x1F\x7F]/g, '') // Remove control chars
            .trim()
            .replace(/\s+/g, ' '); // Collapse whitespace
        // Remove common noise patterns that might appear
        cleaned = cleaned
            .replace(/^[a-z]+-[a-z]+\s+/, '') // Remove patterns like "gen-something "
            .replace(/\([^)]*\)\s*-?\s*$/g, ''); // Remove trailing parentheses with times
        // Cut at 40 chars
        return cleaned.slice(0, maxLen);
    }
    /** Final verification: check if message is valid for display */
    isValidMessage(message) {
        // Must have content
        if (!message || message.length === 0)
            return false;
        // Must have at least one letter or number
        if (!/[a-zA-Z0-9]/.test(message))
            return false;
        // Must not be just noise (only special chars/punctuation)
        if (/^[^a-zA-Z0-9]+$/.test(message))
            return false;
        // Must not be suspiciously short (less than 3 chars)
        if (message.length < 3)
            return false;
        // Passed all checks
        return true;
    }
    /** Random nice text when buffer is empty */
    getFallbackText() {
        const texts = ["Ready", "Idle", "Waiting for changes...", "Standing by..."];
        return texts[Math.floor(Math.random() * texts.length)];
    }
    /** Clear all frames */
    clear() {
        this.frames = [];
        this.priorityFrame = "";
    }
    /** Get buffer stats */
    getStats() {
        return {
            bufferSize: this.frames.length,
            maxFrames: this.maxFrames,
            hasPriorityFrame: !!this.priorityFrame
        };
    }
}
exports.ProgressFrameBuffer = ProgressFrameBuffer;
//# sourceMappingURL=progress-frame-buffer.js.map