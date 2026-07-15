{
  description = "Validate code blocks embedded in Markdown and MDX documentation files";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-finding-src = {
      url = "git+ssh://git@github.com/LarsArtmann/go-finding?ref=refs/tags/v1.2.0";
      flake = false;
    };
  };

  outputs =
    inputs@{ self, flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        {
          # The package definition lives in package.nix (single source of truth).
          # We pass `self` so the build gets the git-derived version.
          packages.default = pkgs.callPackage ./package.nix {
            inherit self;
            inherit (inputs) go-finding-src;
          };

          apps = {
            default = {
              type = "app";
              program = lib.getExe config.packages.default;
            };

            test = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "run-test";
                runtimeInputs = [ pkgs.go_1_26 ];
                text = ''
                  export GOEXPERIMENT=jsonv2
                  go test -race -v -coverprofile=coverage.out ./...
                '';
              };
            };

            lint = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "run-lint";
                runtimeInputs = [
                  pkgs.go_1_26
                  pkgs.golangci-lint
                ];
                text = ''
                  export GOEXPERIMENT=jsonv2
                  golangci-lint run ./...
                '';
              };
            };
          };

          devShells = {
            default = pkgs.mkShell {
              packages = builtins.attrValues {
                inherit (pkgs)
                  go
                  gopls
                  golangci-lint
                  goreleaser
                  ;
              };
              GOWORK = "off";
              GOEXPERIMENT = "jsonv2";
            };

            ci = pkgs.mkShellNoCC {
              packages = builtins.attrValues {
                inherit (pkgs)
                  go
                  golangci-lint
                  ;
              };
              GOWORK = "off";
              GOPRIVATE = "github.com/larsartmann/*";
              GOEXPERIMENT = "jsonv2";
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;
            test = config.packages.default.overrideAttrs (_: {
              doCheck = true;
            });
          };

          treefmt = {
            programs.gofmt.enable = true;
            programs.nixfmt.enable = true;
          };
        };

      flake.overlays.default = final: _prev: {
        # Note: overlays cannot access the flake's git-derived version (no flake
        # `self` in scope here), so the package version falls back to "dev".
        # Consumers that need the real version should use `packages.default`.
        md-go-validator = final.callPackage ./package.nix { };
      };
    };
}
