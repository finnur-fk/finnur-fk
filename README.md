# Postman Collection Transformer

Perform rapid conversion of JSON structures between Postman Collection Format v1 and v2. The formats are documented at https://schema.postman.com.

## Installation

For CLI usage:

```bash
npm install -g postman-collection-transformer
```

As a library:

```bash
npm install --save postman-collection-transformer
```

## CLI Usage

### Converting Collections

```bash
postman-collection-transformer convert \
  --input ./v1-collection.json \
  --input-version 2.0.0 \
  --output ./v2-collection.json \
  --output-version 1.0.0 \
  --pretty \
  --overwrite
```

Options (`postman-collection-transformer convert -h`):

- `-i, --input <path>`: path to the input Postman collection file
- `-j, --input-version [version]`: version of the input collection format (v1 or v2)
- `-o, --output <path>`: target file path where the converted collection will be written
- `-p, --output-version [version]`: required version to convert to
- `-P, --pretty`: pretty print output
- `--retain-ids`: retain request/folder IDs (collection ID always retained)
- `-w, --overwrite`: overwrite the output file if it exists

### Normalizing v1 Collections

```bash
postman-collection-transformer normalize \
  --input ./v1-collection.json \
  --normalize-version 1.0.0 \
  --output ./v1-norm-collection.json \
  --pretty \
  --overwrite
```

Options (`postman-collection-transformer normalize -h`):

- `-i, --input <path>`: path to the collection JSON file to be normalized
- `-n, --normalize-version <version>`: version to normalize the collection on
- `-o, --output <path>`: target file for normalized collection
- `-P, --pretty`: pretty print output
- `--retain-ids`: retain request/folder IDs (collection ID always retained)
- `-w, --overwrite`: overwrite the output file if it exists

## Library Usage

### Converting Entire Collections

```js
const transformer = require('postman-collection-transformer');
const collection = require('./path/to/collection.json');
const options = {
  inputVersion: '1.0.0',
  outputVersion: '2.0.0',
  retainIds: true
};

transformer.convert(collection, options, (error, result) => {
  if (error) {
    return console.error(error);
  }
  console.log(result);
});
```

### Converting Individual Requests

```js
const transformer = require('postman-collection-transformer');

const objectToConvert = { /* v1 Request or v2 Item */ };
const options = {
  inputVersion: '1.0.0',
  outputVersion: '2.0.0',
  retainIds: true
};

transformer.convertSingle(objectToConvert, options, (err, converted) => {
  console.log(converted);
});
```

### Converting Individual Responses

```js
const transformer = require('postman-collection-transformer');

const objectToConvert = { /* v1 Response or v2 Response */ };
const options = {
  inputVersion: '1.0.0',
  outputVersion: '2.0.0',
  retainIds: true
};

transformer.convertResponse(objectToConvert, options, (err, converted) => {
  console.log(converted);
});
```

### Normalizing v1 Collections (Library)

```js
const transformer = require('postman-collection-transformer');
const collection = require('./path/to/collection.json');
const options = {
  normalizeVersion: '1.0.0',
  mutate: false,
  noDefaults: false,
  prioritizeV2: false,
  retainEmptyValues: false,
  retainIds: true
};

transformer.normalize(collection, options, (error, result) => {
  if (error) {
    return console.error(error);
  }
  console.log(result);
});
```
