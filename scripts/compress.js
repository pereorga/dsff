#!/usr/bin/env node

import { createReadStream, createWriteStream } from "fs";
import { pipeline } from "stream/promises";
import { createGzip, createBrotliCompress, constants } from "zlib";

const files = process.argv.slice(2);

if (files.length === 0) {
  console.error("Usage: node compress.js <file1> <file2> ...");
  process.exit(1);
}

async function compressFile(inputPath, compressor, extension) {
  const outputPath = `${inputPath}.${extension}`;

  try {
    await pipeline(
      createReadStream(inputPath),
      compressor,
      createWriteStream(outputPath),
    );
    console.log(`✓ Created ${outputPath}`);
  } catch (error) {
    console.error(
      `✗ Error compressing ${inputPath} to ${extension}:`,
      error.message,
    );
    process.exit(1);
  }
}

async function compressFiles() {
  for (const file of files) {
    const gzipOptions = {
      level: constants.Z_BEST_COMPRESSION,
    };

    await compressFile(file, createGzip(gzipOptions), "gz");
    await compressFile(file, createBrotliCompress(), "br");
  }
}

compressFiles().catch(console.error);
