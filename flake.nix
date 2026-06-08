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
        let
          version = self.rev or self.dirtyRev or "dev";
          vendorHash = "sha256-dwZW5PBepzZjfQEyiemLtElIhzPUBaV/wJMk70pzwFs=";

          src = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./cmd
              ./pkg
            ];
          };
        in
        {
          packages.default = pkgs.buildGoModule {
            pname = "md-go-validator";
            inherit version vendorHash src;
            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
            ];
            meta = with lib; {
              description = "Validate code blocks embedded in Markdown and MDX documentation files";
              homepage = "https://github.com/LarsArtmann/md-go-validator";
              license = licenses.mit;
              mainProgram = "md-go-validator";
            };
          };

          apps = {
            default = {
              type = "app";
              program = lib.getExe(config.packages.default);
            };

            test = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "run-test";
                runtimeInputs = [ pkgs.go_1_26 ];
                text = "go test -race -v -coverprofile=coverage.out ./...";
              };
            };

            lint = {
              type = "app";
              program = pkgs.writeShellApplication {
                name = "run-lint";
                runtimeInputs = [ pkgs.go_1_26 pkgs.golangci-lint ];
                text = "golangci-lint run ./...";
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
            };

            ci = pkgs.mkShellNoCC {
              packages = builtins.attrValues {
                inherit (pkgs)
                  go
                  golangci-lint
                  ;
              };
              GOWORK = "off";
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
        md-go-validator = final.callPackage ./package.nix { };
      };
    };
}
