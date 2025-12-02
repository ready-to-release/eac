"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const chai_1 = require("chai");
const progress_frame_buffer_1 = require("../src/progress-frame-buffer");
describe('@L0 @ov ProgressFrameBuffer', () => {
    let buffer;
    beforeEach(() => {
        buffer = new progress_frame_buffer_1.ProgressFrameBuffer();
    });
    describe('initialization', () => {
        it('should start with empty buffer', () => {
            const stats = buffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(0);
            (0, chai_1.expect)(stats.hasPriorityFrame).to.be.false;
        });
        it('should return fallback text when empty', () => {
            const frame = buffer.getCurrentFrame();
            (0, chai_1.expect)(frame).to.be.a('string');
            (0, chai_1.expect)(frame.length).to.be.greaterThan(0);
        });
    });
    describe('addProgress', () => {
        it('should add valid progress messages', () => {
            buffer.addProgress('Loading data');
            const stats = buffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(1);
        });
        it('should reject empty messages', () => {
            buffer.addProgress('');
            const stats = buffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(0);
        });
        it('should reject messages with only special characters', () => {
            buffer.addProgress('---');
            buffer.addProgress('...');
            const stats = buffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(0);
        });
        it('should reject very short messages (less than 3 chars)', () => {
            buffer.addProgress('ab');
            const stats = buffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(0);
        });
        it('should respect max frames limit', () => {
            const smallBuffer = new progress_frame_buffer_1.ProgressFrameBuffer(3);
            smallBuffer.addProgress('Message 1');
            smallBuffer.addProgress('Message 2');
            smallBuffer.addProgress('Message 3');
            smallBuffer.addProgress('Message 4');
            const stats = smallBuffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(3);
            (0, chai_1.expect)(smallBuffer.getCurrentFrame()).to.equal('Message 4');
        });
    });
    describe('getCurrentFrame', () => {
        it('should return latest progress message', () => {
            buffer.addProgress('First message');
            buffer.addProgress('Second message');
            (0, chai_1.expect)(buffer.getCurrentFrame()).to.equal('Second message');
        });
        it('should prioritize event over progress', () => {
            buffer.addProgress('Progress message');
            buffer.pushEvent('Priority event');
            (0, chai_1.expect)(buffer.getCurrentFrame()).to.equal('Priority event');
        });
    });
    describe('pushEvent and clearEvent', () => {
        it('should set priority frame', () => {
            buffer.pushEvent('Important event');
            (0, chai_1.expect)(buffer.getCurrentFrame()).to.equal('Important event');
            (0, chai_1.expect)(buffer.getStats().hasPriorityFrame).to.be.true;
        });
        it('should clear priority and fall back to progress', () => {
            buffer.addProgress('Background progress');
            buffer.pushEvent('Priority event');
            buffer.clearEvent();
            (0, chai_1.expect)(buffer.getCurrentFrame()).to.equal('Background progress');
            (0, chai_1.expect)(buffer.getStats().hasPriorityFrame).to.be.false;
        });
    });
    describe('normalization', () => {
        it('should trim and collapse whitespace', () => {
            buffer.addProgress('  Multiple   spaces  here  ');
            (0, chai_1.expect)(buffer.getCurrentFrame()).to.equal('Multiple spaces here');
        });
        it('should truncate to 40 characters', () => {
            const longMessage = 'A'.repeat(50);
            buffer.addProgress(longMessage);
            (0, chai_1.expect)(buffer.getCurrentFrame().length).to.be.at.most(40);
        });
        it('should remove control characters', () => {
            buffer.addProgress('Valid\x00message\x1F');
            (0, chai_1.expect)(buffer.getCurrentFrame()).to.equal('Validmessage');
        });
    });
    describe('clear', () => {
        it('should clear all frames and priority', () => {
            buffer.addProgress('Some progress');
            buffer.pushEvent('Some event');
            buffer.clear();
            const stats = buffer.getStats();
            (0, chai_1.expect)(stats.bufferSize).to.equal(0);
            (0, chai_1.expect)(stats.hasPriorityFrame).to.be.false;
        });
    });
});
//# sourceMappingURL=progress-frame-buffer.test.js.map