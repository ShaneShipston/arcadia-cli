# Arcadia CLI

A command line utility for installing blocks

## Usage

Install a block

```bash
arc install [block_name]
```

Check program version

```bash
arc version
```

## Manifest Format

| Index         | Purpose                         | Required |
|---------------|---------------------------------|----------|
| name          | Friendly Block name             | true     |
| key           | Block file name                 | true     |
| type          | block, component, section, base | true     |
| acfgroup      | ACF group name                  | false    |
| modifications | Which files need to be modified | false    |

Append can be formatted in a couple ways. The below sample indicates that the @import should be added to the end of style.scss

```json
"src/scss/style.scss": "@import \"blocks/call_to_action\";"
```

Additionally if there are multiple pieces of code that need to be altered. You can include an array of modifications to be performed on a per file basis. For example, the first one will append to the bottom of the file and the second will append after the target.

```json
"src/scss/style.scss": [
    {
        "code": "@import \"blocks/call_to_action\";"
    },
    {
        "code": "@import \"blocks/call_to_action\";",
        "target": "@import \"blocks/accent_image\";"
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
    "append": {
        "src/scss/style.scss": "@import \"blocks/call_to_action\";"
    }
}
```

## Changelog

### Unreleased

## Roadmap

- [ ] Install Blocks
- [ ] Windows Install
- [ ] MacOS Install
- [ ] Documentation
- [ ] Install Components
- [ ] Install Foundations
- [ ] Install Sections
- [ ] Install Plugins
- [ ] Search
