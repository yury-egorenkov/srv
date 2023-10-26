## Overview

Extremely simple Go tool that serves files out of a given folder, using a file resolution algorithm similar to GitHub Pages, Netlify, or the default Nginx config. Provides a Go "library" (less than 100 LoC) and an optional CLI tool.

## Why

Useful when making a site that will be served by GitHub Pages or similar.

* Usable from Go code.
* Usable from a shell.
* Not overweight.
* Can serve files from `zip` archives without nesseserity of unpacking.

## Usage

An almost drop-in replacement for `http.FileServer`:

```go
import "github.com/mitranim/srv"

http.ListenAndServe(":<some-port>", srv.FileServer("."))
```

For CLI usage, first install Go: https://golang.org. Then run this:

```sh
go install github.com/mitranim/srv/srv@latest

# Get help
srv -h

# Actually run
srv
```

This will compile the executable into `$GOPATH/bin/srv`. Make sure `$GOPATH/bin` is in your `$PATH` so the shell can discover the `srv` command. For example, my `~/.profile` contains this:

```sh
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
```

Alternatively, you can run the executable using the full path. At the time of writing, `~/go` is the default `$GOPATH` for Go installations. Some systems may have a different one.

```sh
~/go/bin/srv
```

## File Resolution

* For directories including `/`, use `index.html` when available.
* For "clean" links without an extension, try appending `.html`.
* For "not found", try using `404.html` if one exists, falling back on a simple hardcoded message.
* For `zip` archive download if no `/`. If `arch.zip/index.html` it serves `html` and works as it regular folder.

## License

https://unlicense.org

## Misc

I'm receptive to suggestions. If this library _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
