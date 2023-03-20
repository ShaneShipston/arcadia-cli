# Arcadia CLI

A command line utility for installing blocks

## Usage

### Install blocks

```bash
arc install [blocks]
```

As as example, if we wanted to install both FAQ and Call to Action you could run:

```bash
arc install faq call-to-action
```

### Check program version

```bash
arc version
```

### Unpack archives

If you've already downloaded a bunch of blocks and added the zip files to your theme you can unpack them to install

```bash
arc unpack
```

### Self Updating

```bash
arc update
```

### Display a list of available commands

```bash
arc help
```

## Manifest Format

| Index         | Purpose                         | Required |
|---------------|---------------------------------|----------|
| name          | Friendly Block name             | true     |
| key           | Block file name                 | true     |
| contents      | block, component, section, base | true     |
| acfgroup      | ACF group name                  | false    |
| modifications | Which files need to be modified | false    |

Append can be formatted in a couple ways. The below sample indicates that the @import should be added to the end of style.scss

```json
"src/scss/style.scss": "@import \"blocks/call_to_action\";"
```

Additionally if there are multiple pieces of code that need to be altered. You can include an array of modifications to be performed on a per file basis. For example, the first one will append to the bottom of the file and the second will prepend the target. The mode can consist of either "append", "prepend", "replace" or "remove"

```json
"src/scss/style.scss": [
    {
        "code": "@import \"blocks/call_to_action\";"
    },
    {
        "code": "@import \"blocks/call_to_action\";",
        "target": "@import \"blocks/accent_image\";",
        "mode": "prepend"
    }
]
```

### Blocks

Below is the example code for a simple block

```json
{
    "name": "Call to Action",
    "key": "call_to_action",
    "type": "block",
    "acfgroup": "group_610ca48968b82",
    "modifications": {
        "src/scss/style.scss": "@import \"blocks/call_to_action\";"
    }
}
```

## Building

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o bin/arc-windows-amd64.exe arc.go

# MacOS
GOOS=darwin GOARCH=amd64 go build -o bin/arc-macos-amd64 arc.go

# Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o bin/arc-macos-arm arc.go
```

## Changelog

### 0.2.0

- Install Foundations

### 0.1.0

- Install blocks
- Install Components
- Unpack command
- Windows build
- MacOS Install
- Documentation
- Update check

## Roadmap

- Match indentation when doing modifications
- Install Sections
- Install Plugins
- Unpack nested zip files
- Search
