# Installation

Install with Go:

```bash
go install github.com/maxencetholomier/knowledge/cmd/kl@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

## Updating

From a local clone (e.g. after making changes):

```bash
go install ./cmd/kl
```

From GitHub, bypassing the Go proxy cache to get the latest commit:

```bash
GOPROXY=direct go install github.com/maxencetholomier/knowledge/cmd/kl@master
```

## Shell completion

Add to your `~/.bashrc`:

```bash
source <(kl completion bash)
```

Or to your `~/.zshrc`:

```zsh
source <(kl completion zsh)
```

For zsh, `compinit` must be loaded before this line (`autoload -U compinit && compinit`).

For fish:

```fish
kl completion fish > ~/.config/fish/completions/kl.fish
```

## Man pages (optional)

```bash
kl man --quiet
sudo cp /tmp/kl-man/*.1 /usr/local/man/man1/
sudo mandb
```

## Joplin setup

In [Joplin](https://github.com/laurent22/joplin), enable [Web Clipper](https://joplinapp.org/help/apps/clipper/#using-the-web-clipper-service):

```
Tools → Options → Web Clipper → Enable Web Clipper
```
