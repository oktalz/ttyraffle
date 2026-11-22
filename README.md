```
┌┬┐┌┬┐┬ ┬  ┬─┐┌─┐┌─┐┌─┐┬  ┌─┐
 │  │ └┬┘  ├┬┘├─┤├┤ ├┤ │  ├┤
 ┴  ┴  ┴   ┴└─┴ ┴└  └  ┴─┘└─┘
```

tool for selecting random person/item from list in terminal

Entries are read from a text file, one per line. Blank or whitespace-only lines are skipped.
Duplicate lines count as separate entries. Empty or unreadable files report an error.
Edits apply only to the current session and never overwrite the file.

## Installation
Use the following command to download and install this tool:
```sh
go install github.com/oktalz/ttyraffle@latest
```

## Binaries
  prebuilt binaries can be found on [releases](https://github.com/oktalz/ttyraffle/releases) page

## Running

![Example](demo.gif)

```sh
ttyraffle file.txt
```

if more txt files exist, start the program and pick correct one.
if one txt file exists it will be autoselected
```sh
ttyraffle
```

## Configuration

Optional settings are loaded from `~/.ttyraffle/.env`.
Move an existing `~/.terminal-of-fortune/.env` there when upgrading.

## Controls

The race list scrolls to accommodate entries that do not fit on screen.

| Key | Action |
| --- | --- |
| `s` | Start, or skip the current winner without excluding them |
| `c`, Space, Enter | Continue, excluding the previous winner |
| `x` | Reset, restoring all entries |
| `i` | Insert an entry for this session |
| `d` | Delete an entry for this session |
| `r` | Shuffle entries |
| ↑ / ↓, Page Up / Page Down | Scroll the race list |
| Arrow keys, Enter | Navigate and confirm selections |
| Esc | Cancel adding or deleting |
| Ctrl+C | Exit from any screen |
| `q` | Exit, except while entering a name |
