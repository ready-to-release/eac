#!/usr/bin/env node
/**
 * Mermaid Batch Renderer
 *
 * Renders multiple mermaid diagrams from a manifest file.
 * Supports parallel rendering with configurable concurrency.
 *
 * Usage: node batch-render.js --manifest /work/manifest.json
 *
 * Manifest format:
 * {
 *   "diagrams": [
 *     { "input": "/work/diagram1.mmd", "output": "/staging/cache/abc.svg" }
 *   ],
 *   "concurrency": 4,
 *   "theme": "dark"
 * }
 */

const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

// Configuration
const DEFAULT_CONCURRENCY = 4;
const MAX_RETRIES = 3;
const RETRY_DELAY_MS = 500;
const MERMAID_CONFIG = '/etc/mermaid/mermaid-config.json';
const PUPPETEER_CONFIG = '/etc/mermaid/puppeteer-config.json';

/**
 * Parse command line arguments
 */
function parseArgs() {
  const args = process.argv.slice(2);
  const result = { manifest: null };

  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--manifest' && args[i + 1]) {
      result.manifest = args[i + 1];
      i++;
    }
  }

  return result;
}

/**
 * Read and parse manifest file
 */
function readManifest(manifestPath) {
  if (!fs.existsSync(manifestPath)) {
    throw new Error(`Manifest file not found: ${manifestPath}`);
  }

  const content = fs.readFileSync(manifestPath, 'utf8');
  const manifest = JSON.parse(content);

  if (!manifest.diagrams || !Array.isArray(manifest.diagrams)) {
    throw new Error('Manifest must contain a "diagrams" array');
  }

  return manifest;
}

/**
 * Render a single diagram using mmdc
 */
function renderDiagram(diagram, theme) {
  return new Promise((resolve, reject) => {
    const args = [
      '-i', diagram.input,
      '-o', diagram.output,
      '-c', diagram.config || MERMAID_CONFIG,
      '-p', PUPPETEER_CONFIG,
      '--quiet'
    ];

    if (theme) {
      args.push('-t', theme);
    }

    // Ensure output directory exists
    const outputDir = path.dirname(diagram.output);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    const proc = spawn('mmdc', args, {
      stdio: ['ignore', 'pipe', 'pipe']
    });

    let stdout = '';
    let stderr = '';

    proc.stdout.on('data', (data) => {
      stdout += data.toString();
    });

    proc.stderr.on('data', (data) => {
      stderr += data.toString();
    });

    proc.on('close', (code) => {
      if (code === 0) {
        resolve({ success: true, output: diagram.output });
      } else {
        reject(new Error(`mmdc failed with code ${code}: ${stderr || stdout}`));
      }
    });

    proc.on('error', (err) => {
      reject(new Error(`Failed to spawn mmdc: ${err.message}`));
    });
  });
}

/**
 * Render with retry logic
 */
async function renderWithRetry(diagram, theme, maxRetries = MAX_RETRIES) {
  let lastError;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await renderDiagram(diagram, theme);
    } catch (err) {
      lastError = err;
      const errStr = err.message || '';

      // Check for retryable errors (Chromium/file lock issues)
      const isRetryable =
        errStr.includes('EBUSY') ||
        errStr.includes('Cannot access') ||
        errStr.includes('text file busy') ||
        errStr.includes('Protocol error') ||
        errStr.includes('Target closed') ||
        errStr.includes('Session closed');

      if (!isRetryable || attempt === maxRetries) {
        throw lastError;
      }

      // Wait before retry with exponential backoff
      const delay = RETRY_DELAY_MS * Math.pow(2, attempt - 1);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }

  throw lastError;
}

/**
 * Render batch with concurrency limit using simple chunking
 */
async function renderBatch(diagrams, concurrency, theme) {
  const results = {
    success: 0,
    failed: 0,
    errors: []
  };

  if (diagrams.length === 0) {
    console.log('[mermaid] No diagrams to render');
    return results;
  }

  const total = diagrams.length;
  console.log(`[mermaid] Starting batch render: ${total} diagrams, concurrency: ${concurrency}`);

  // Process in chunks for concurrency control
  for (let i = 0; i < diagrams.length; i += concurrency) {
    const chunk = diagrams.slice(i, i + concurrency);
    const promises = chunk.map(diagram =>
      renderWithRetry(diagram, theme)
        .then(() => ({ success: true, diagram }))
        .catch(err => ({ success: false, diagram, error: err.message }))
    );

    const chunkResults = await Promise.all(promises);

    for (const result of chunkResults) {
      if (result.success) {
        results.success++;
      } else {
        results.failed++;
        results.errors.push({
          input: result.diagram.input,
          output: result.diagram.output,
          error: result.error
        });
        console.error(`[mermaid] Error rendering ${result.diagram.input}: ${result.error}`);
      }
    }

    // Report progress after each chunk
    const completed = Math.min(i + concurrency, total);
    const pct = Math.round((completed / total) * 100);
    console.log(`[mermaid] Progress: ${completed}/${total} (${pct}%)`);
  }

  return results;
}

/**
 * Main entry point
 */
async function main() {
  const args = parseArgs();

  if (!args.manifest) {
    console.error('Usage: node batch-render.js --manifest <path>');
    process.exit(1);
  }

  try {
    console.log(`[mermaid] Reading manifest: ${args.manifest}`);
    const manifest = readManifest(args.manifest);

    const concurrency = manifest.concurrency || DEFAULT_CONCURRENCY;
    const theme = manifest.theme || null;

    console.log(`[mermaid] Diagrams: ${manifest.diagrams.length}, Theme: ${theme || 'default'}`);

    const startTime = Date.now();
    const results = await renderBatch(manifest.diagrams, concurrency, theme);
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);

    console.log(`[mermaid] Completed in ${elapsed}s: ${results.success} success, ${results.failed} failed`);

    if (results.failed > 0) {
      console.error('[mermaid] Failed diagrams:');
      for (const err of results.errors) {
        console.error(`  - ${err.input}: ${err.error}`);
      }
      process.exit(1);
    }

    process.exit(0);
  } catch (err) {
    console.error(`[mermaid] Fatal error: ${err.message}`);
    process.exit(1);
  }
}

main();
