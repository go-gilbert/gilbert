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
  --option @typespec/json-schema.emitter-output-dir=./docs/spec/json-schema \
  --list-files
```

With this option, the generated schema is written to:

```text
docs/spec/json-schema/WorkflowFile.json
```

`WorkflowFile` is the only TypeSpec model decorated with `@jsonSchema`, so helper types are emitted into the same JSON Schema document under `$defs`. This keeps the schema unified and avoids external `*.json` references.

## YAML Language Server Compatibility

TypeSpec emits map-like records with `unevaluatedProperties`:

```json
{
  "type": "object",
  "unevaluatedProperties": {
    "$ref": "#/$defs/Task"
  }
}
```

This is valid JSON Schema 2020-12, but YAML language server support is weaker for this keyword. In practice, it may show hover information for top-level properties like `tasks` and `inputs`, but fail to provide hover/completion for arbitrary nested keys inside those maps.

YAML language server handles `additionalProperties` better for map-like objects:

```json
{
  "type": "object",
  "additionalProperties": {
    "$ref": "#/$defs/Task"
  }
}
```

After generating the schema, copy each `unevaluatedProperties` value to `additionalProperties`. Keeping both fields preserves the original 2020-12 semantics while making the schema more useful to YAML language server.

Example post-processing command:

```sh
node -e 'const fs=require("fs"); const p="tsp-output/@typespec/json-schema/gilbert.schema.json"; const doc=JSON.parse(fs.readFileSync(p,"utf8")); function walk(x){ if(!x||typeof x!=="object") return; if(Object.prototype.hasOwnProperty.call(x,"unevaluatedProperties") && !Object.prototype.hasOwnProperty.call(x,"additionalProperties")) x.additionalProperties=x.unevaluatedProperties; for(const v of Object.values(x)) walk(v); } walk(doc); fs.writeFileSync(p, JSON.stringify(doc,null,4)+"\n");'
```

Then reference it from YAML:

```yaml
# yaml-language-server: $schema=http://localhost:8000/json-schema/gilbert.schema.json
```
