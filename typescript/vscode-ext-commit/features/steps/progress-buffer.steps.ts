import { Given, When, Then, Before } from '@cucumber/cucumber';
import { expect } from 'chai';
import { ProgressFrameBuffer } from '../../src/progress-frame-buffer';

interface World {
    buffer: ProgressFrameBuffer;
    currentFrame: string;
}

Before(function (this: World) {
    // Reset state before each scenario
    this.buffer = new ProgressFrameBuffer();
});

// Background
Given('a new progress frame buffer', function (this: World) {
    this.buffer = new ProgressFrameBuffer();
});

Given('a progress buffer with max {int} frames', function (this: World, maxFrames: number) {
    this.buffer = new ProgressFrameBuffer(maxFrames);
});

// When steps
When('I check the current frame', function (this: World) {
    this.currentFrame = this.buffer.getCurrentFrame();
});

When('I add progress {string}', function (this: World, message: string) {
    this.buffer.addProgress(message);
});

When('I add progress with {int} characters', function (this: World, count: number) {
    const message = 'A'.repeat(count);
    this.buffer.addProgress(message);
});

When('I push event {string}', function (this: World, message: string) {
    this.buffer.pushEvent(message);
});

When('I clear the event', function (this: World) {
    this.buffer.clearEvent();
});

When('I clear the buffer', function (this: World) {
    this.buffer.clear();
});

// Then steps
Then('the frame should be a non-empty string', function (this: World) {
    expect(this.currentFrame).to.be.a('string');
    expect(this.currentFrame.length).to.be.greaterThan(0);
});

Then('the buffer size should be {int}', function (this: World, expected: number) {
    const stats = this.buffer.getStats();
    expect(stats.bufferSize).to.equal(expected);
});

Then('the current frame should be {string}', function (this: World, expected: string) {
    expect(this.buffer.getCurrentFrame()).to.equal(expected);
});

Then('the current frame length should be at most {int}', function (this: World, maxLength: number) {
    expect(this.buffer.getCurrentFrame().length).to.be.at.most(maxLength);
});

Then('the buffer should have a priority frame', function (this: World) {
    expect(this.buffer.getStats().hasPriorityFrame).to.be.true;
});

Then('the buffer should not have a priority frame', function (this: World) {
    expect(this.buffer.getStats().hasPriorityFrame).to.be.false;
});
