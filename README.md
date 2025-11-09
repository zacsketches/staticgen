# staticgen

A lightweight static site generator for Go templates.

## Installation

Download the latest binary from [releases](https://github.com/zacsketches/staticgen/releases).

## Usage

```bash
# Render templates with default settings
./staticgen

# Custom source/output directories  
./staticgen -src ./templates -out ./public

# Custom page glob pattern
./staticgen -glob "pages/*.template.html"

# Custom build timestamp
./staticgen -timestamp "2024-01-01 12:00:00 CST"
```

## Template Structure

```
src/
├── _includes/          # Shared partials
├── _layouts/          # Layout templates  
└── pages/            # Page templates
    └── *.template.html
```

## License

MIT
