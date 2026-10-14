## Task 1: Make the release folder

Two commands: one makes the folders, one copies the file. `mkdir` on its own
stops when a folder above the new one is missing.

### Closer

`mkdir -p a/b/c` makes all three folders. `cp FILE FOLDER/` copies FILE into
FOLDER with the same name.

### Answer

```bash
mkdir -p releases/v2/config
cp app/config.yaml releases/v2/config/
```

## Task 2: Back up the app folder

Make `backup` first, then copy the `app` folder itself into it. A folder
needs the copy option that goes into subfolders, and that option takes
hidden files along.

### Closer

`mkdir -p backup`, then `cp -r app backup/`. Avoid `app/*`: the `*` skips
`.env`.

### Answer

```bash
mkdir -p backup
cp -r app backup/
```

## Task 3: Rotate the log

Renaming is moving a file to a new name in the same folder. Then make an
empty file with the old name.

### Closer

`mv OLD NEW` renames; `touch NAME` creates an empty file. Either `cd
app/logs` first, or write `app/logs/...` in every path.

### Answer

```bash
cd app/logs
mv app.log app.log.1
touch app.log
```

## Task 4: Clean up and show the backup

`rm` needs an option to delete a folder and what's in it. For the path,
`pwd` prints where you are as an absolute path, so go there first.

### Closer

`rm -r old-release` deletes the folder. Then `cd backup/app` and `pwd`.

### Answer

```bash
rm -r old-release
cd backup/app
pwd
```
