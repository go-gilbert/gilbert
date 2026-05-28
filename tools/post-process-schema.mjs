/**
 * Post-processes a JSON Schema file emitted by TypeSpec's `@typespec/json-schema` emitter.
 * Replaces all occurrences of `unevaluatedProperties` with `additionalProperties` to
 * improve compatibility with YAML language servers.
 */

import { readFileSync, writeFileSync } from "fs";

const [inputFile, outputFile] = process.argv.slice(2);

if (!inputFile) {
  console.error("Usage: node post-process-schema.mjs <inputFile> [outputFile]");
  process.exit(1);
}

let data;
try {
  data = JSON.parse(readFileSync(inputFile, "utf8"));
} catch (err) {
  console.error(`Error: Failed to parse '${inputFile}' as JSON: ${err.message}`);
  process.exit(1);
}

function renameKeys(obj) {
  if (Array.isArray(obj)) {
    return obj.map(renameKeys);
  }
  if (obj !== null && typeof obj === "object") {
    const result = {};
    for (const [key, value] of Object.entries(obj)) {
      result[key === "unevaluatedProperties" ? "additionalProperties" : key] =
        renameKeys(value);
    }
    return result;
  }
  return obj;
}

const result = renameKeys(data);
const output = JSON.stringify(result, null, 2) + "\n";

if (outputFile) {
  writeFileSync(outputFile, output);
  console.log(`Patched: ${outputFile}`);
} else {
  process.stdout.write(output);
}
