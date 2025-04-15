# md-html

**md-html** is a lightweight command-line tool to convert Markdown to HTML. It offers additional features such as live previewing.

## Installation

### From source

Install the latest released version:

```go
go install github.com/gnitoahc/md-html
```

### Pre-built binaries

Pre-built binaries are available on the [releases page](https://github.com/GNITOAHC/md-html/releases)

### Homebrew

```bash
brew tap gnitoahc/tap
brew install gnitoahc/tap/md-html
```

## Usage

```
Usage: md-html [options] <filename>
  -o string
        Name of the output file
  -p string
        Port to serve the HTML file (default "8080")
  -v    Version of the program
  -w    Watch for changes in the markdown file
```

For instance, `md-html README.md` will generate a `README.html`, `md-html -w README.md` will watch for changes and regenerate the HTML automatically. Either refresh manually or visit `localhost` for auto live refresh.

To convert markdown to PDF, open the HTML in browser then "Print to PDF" is recommended.
