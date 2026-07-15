# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

Enter the Nix development shell (provides Go, golangci-lint, goreleaser):

```bash
nix develop
```

Then run the core checks:

```bash
go test ./... -race
golangci-lint run ./...
nix flake check
```

## Documentation Website

The documentation website lives in `website/` and uses Astro + Starlight + Tailwind v4.
Full setup and commands are documented in `website/flake.nix`.

```bash
cd website
nix develop          # dev shell with Node.js + Firebase Tools
nix run .#dev        # local dev server
nix run .#build      # production build
nix run .#preview    # preview production build locally
nix run .#deploy     # deploy to Firebase Hosting
```

The site auto-deploys on push to `master` via `.github/workflows/website.yml`.

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
