# Cloner

## Responsibility

Read a structured file (CSV, JSON, or text) and re-clone repositories,
preserving the original folder hierarchy. Also supports cloning a single
repository directly from a Git URL.

## Behavior

### File-based clone

1. Detect file format by extension (`.csv`, `.json`, `.txt`).
2. Parse records from the file.
3. For each record:
   a. Create the relative directory structure under `--target-dir`.
   b. Run `git clone -b <branch> <url> <target-path>`.
4. Log success or failure for each clone operation.
5. Print a summary: N succeeded, M failed.

### Direct URL clone

1. Detect that the source is a URL (`https://`, `http://`, `git@`).
2. Derive the repo name from the URL (or use a custom folder name).
3. Run `git clone <url> <folder>`.
4. Upsert the repo record into the database.
5. Prompt to register with GitHub Desktop.

## Error Handling

- If a clone fails (network, auth, etc.), log the error and continue.
- Do not abort the entire run for a single failure.
- Summary at end lists all failures with reasons.
- For direct URL clone, if the target folder exists, exit with error.

## Input Formats

| Format | Structure                              |
|--------|----------------------------------------|
| CSV    | Standard CSV with headers              |
| JSON   | Array of `ScanRecord` objects          |
| Text   | One `git clone ...` command per line   |
| URL    | Direct HTTPS or SSH git URL            |
