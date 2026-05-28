# JSON Schema Build

`docs/spec/schema.tsp` is the TypeSpec source used to generate the editor-facing JSON Schema for `gilbert.yml`.

## Generate Schema

Run:

```sh
npx tsp compile docs/spec/schema.tsp \
  --emit=@typespec/json-schema \
  --option @typespec/json-schema.file-type=json \
  --list-files
```

This emits:

```text
tsp-output/@typespec/json-schema/WorkflowFile.json
```

The generated file can be renamed to:

```text
tsp-output/@typespec/json-schema/gilbert.schema.json
```

To write the generated schema into a custom directory, set the JSON Schema emitter output directory:

```sh
npx tsp compile docs/spec/schema.tsp \
  --emit=@typespec/json-schema \
  --option @typespec/json-schema.file-type=json \
  --option @typespec/json-schema.emitter-output-dir='{cwd}/docs/spec/json-schema' \
  --list-files
```

With this option, the generated schema is written to:

```text
docs/spec/json-schema/WorkflowFile.json
```

`WorkflowFile` is the only TypeSpec model decorated with `@jsonSchema`, so helper types are emitted into the same JSON Schema document under `$defs`. This keeps the schema unified and avoids external `*.json` references.

## Post-Processing

TypeSpec emits map-like records with `unevaluatedProperties`:

```json
{
  "type": "object",
  "unevaluatedProperties": {
    "$ref": "#/$defs/Task"
  }
}
```

This is valid JSON Schema 2020-12, but YAML language servers have weak support for `unevaluatedProperties` and fail to provide hover/completion for arbitrary nested keys inside those maps. The `additionalProperties` keyword is handled better by YAML language servers.

To fix this, run the post-processing script to replace `unevaluatedProperties` with `additionalProperties` throughout the schema:

```sh
node tools/post-process-schema.mjs <inputFile> [outputFile]
```

- `<inputFile>`: the generated JSON Schema file
- `[outputFile]`: optional output path; if omitted, writes to stdout

After generating and patching the schema, reference it from YAML:

```yaml
# yaml-language-server: $schema=http://localhost:8000/json-schema/gilbert.schema.json
```