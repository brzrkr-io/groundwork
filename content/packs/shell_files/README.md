# Files and folders: paths, mkdir, cp, mv and rm

Deploys are mostly files moving around: a release folder gets created,
config is copied in, an old log is renamed, an old release is deleted. This
exercise covers the commands for that, and how to say *which* file you mean.

**New here:** a practice sandbox built by `setup.sh`, absolute and relative
paths (`/`, `.`, `..`, `~`), hidden files and `ls -a`, `mkdir -p`, `cp`,
`cp -r`, `mv`, `rm`, `rm -r`, `touch`, and `cd` inside a script.

## Learn

### Your sandbox

These scripts change files, so they work on a practice folder, the
*sandbox*, and never on anything that matters. Build it from the exercise
folder, then look around:

```bash
bash setup.sh      # build (or rebuild, from scratch) sandbox/
cd sandbox
ls
# app  old-release
```

Each script runs **inside** the sandbox. Try one like this (`..` is
explained below):

```bash
cd sandbox
bash ../1_make_release.sh
```

Each test does the same: it runs `setup.sh` in a fresh folder, then your
scripts **in order**, from 1 up to the one it checks, and looks at the
files. Made a mess? Run `bash setup.sh` again. The Git exercises later work
exactly like this.

### Paths: absolute and relative

A *path* names a file or folder. There are two kinds:

| Kind | Starts with | Example | Means |
|---|---|---|---|
| *absolute* | `/` | `/etc/nginx/nginx.conf` | the same file, wherever you are |
| *relative* | anything else | `nginx/nginx.conf` | starting from the folder you're in now |

`pwd` prints the folder you're in, always as an absolute path. Three short
names help with relative paths:

| Name | Means | Example |
|---|---|---|
| `.` | this folder | `./deploy.sh` |
| `..` | the folder above | `cd ..`, `../notes.txt` |
| `~` | your home folder | `cd ~/projects` |

So from inside `sandbox/`, `../1_make_release.sh` means "go up one folder,
then `1_make_release.sh`": the script next to the sandbox.

> **Why is it written like this?** **Linux requires** nothing more: `.`
> and `..` exist in every folder. **Bash** turns `~` into your home folder
> before running the command. **Convention:** relative paths when you type,
> absolute paths in jobs that cron or systemd start, because they don't
> start in the folder you'd expect.

### Hidden files

A name that starts with a dot, like `.env` or `.gitignore`, is *hidden*.
Plain `ls` skips it, and so does the `*` wildcard. `ls -a` shows everything:

```bash
ls nginx
# nginx.conf
ls -a nginx
# .  ..  .htpasswd  nginx.conf
```

Config and secrets often hide this way, which is exactly why they get
forgotten in copies and committed by accident.

### Make, copy, move, delete

| Command | Does |
|---|---|
| `mkdir nginx` | make a folder |
| `mkdir -p nginx/sites/live` | make a folder **and** any missing folders above it |
| `cp nginx.conf nginx/` | copy a file into a folder |
| `cp -r nginx backup/` | copy a whole folder, everything inside it (hidden files too) |
| `mv nginx.conf nginx.conf.old` | rename (or move) |
| `touch access.log` | create an empty file (or update an existing one's time) |
| `rm nginx.conf.old` | delete a file |
| `rm -r nginx-old` | delete a folder and everything in it |

When the last argument of `cp` or `mv` is an existing folder, the file or
folder goes **inside** it, keeping its name. `cp -r nginx backup/` makes
`backup/nginx/` when `backup/` exists. When it doesn't, `cp -r` makes
`backup` itself the copy, with nginx's files directly in it.

`cp -r nginx/* backup/nginx/` looks the same but isn't: `*` skips hidden
files, so `.htpasswd` would be left behind.

> **Why is it written like this?** **Linux requires** nothing to undo a
> delete, because there is nothing: `rm` has no trash can. **Convention:**
> run `ls` on what you're about to delete first, and never put `rm -r` on a
> path you haven't checked. `mv` replaces an existing file silently, so
> look before moving onto a name, too.

### cd inside a script

A script runs as its own process with its own current folder. A `cd` in it
moves the script, not your terminal: when it ends, you're where you were.
Every script here starts in the sandbox, whatever the one before it did.

```bash
cd nginx/sites    # the rest of this script works in nginx/sites
ls ..             # lists nginx
```

## Your task

The web app's next release needs a folder, a backup, a fresh log and some
cleaning up. Run `bash setup.sh`, then `cd sandbox` and `ls -a app` to see
what's there.

### 1. Make the release folder

**Goal:** in `1_make_release.sh`, create the folder `releases/v2/config`
and copy `app/config.yaml` into it.

Afterwards:

```text
$ cat releases/v2/config/config.yaml
service: web
port: 8080
replicas: 3
```

`app/config.yaml` stays where it is.

**Think it through:**
- Neither `releases` nor `releases/v2` exists yet. Which `mkdir` option
  makes them on the way?
- Copying into a folder keeps the file's name. What's the last argument?

### 2. Back up the app folder

**Goal:** in `2_back_up_app.sh`, copy the whole `app` folder, hidden files
included, so the copy is `backup/app`.

Afterwards:

```text
$ ls -a backup/app
.  ..  .env  config.yaml  logs
```

`backup/app/logs/app.log` is there too, and `app` itself is unchanged.

**Think it through:**
- Should `backup` exist before you copy, so the copy lands inside it?
- Which way of copying includes `.env`, and which skips it?

### 3. Rotate the log

**Goal:** in `3_rotate_log.sh`, rename `app/logs/app.log` to `app.log.1`
(in the same folder), then create a new, empty `app/logs/app.log`.

Afterwards:

```text
$ ls app/logs
app.log  app.log.1
$ cat app/logs/app.log.1
09:14 GET /health 200
09:15 GET /orders 500
```

The new `app.log` is empty. You can `cd app/logs` first and use the short
names, or write the paths out in full: both pass.

**Think it through:**
- Which command renames a file?
- Which command creates an empty file?

### 4. Clean up and show the backup

**Goal:** in `4_clean_up.sh`, delete the `old-release` folder and
everything in it, then print the absolute path of `backup/app`.

```bash
cd sandbox
bash ../4_clean_up.sh
# /home/sam/groundwork/exercises/07_shell/145_shell_files/sandbox/backup/app
```

Your path will differ: it's wherever your sandbox is. Print nothing else.

**Think it through:**
- Plain `rm` refuses folders. Which option deletes one with its contents?
- Which command always prints an absolute path? Where do you need to be
  when you run it?

## Learn more

- [The Linux Command Line: Navigation](https://linuxcommand.org/lc3_lts0020.php)
- [The Linux Command Line: Manipulating Files](https://linuxcommand.org/lc3_lts0050.php)
- [GNU coreutils: cp invocation](https://www.gnu.org/software/coreutils/manual/html_node/cp-invocation.html)
