{
  lib,
  buildGoModule,
  self ? { },
  go-finding-src ? null,
}:
let
  version = self.shortRev or self.dirtyShortRev or "dev";
  vendorHash = "sha256-I7oN6zZueidT9TfytKNvbMhSkj6y0WLySQEzujRvnw0=";

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
buildGoModule {
  pname = "md-go-validator";
  inherit version vendorHash src;
  proxyVendor = true;
  GOEXPERIMENT = "jsonv2";

  # When go-finding-src is passed (top-level flake build), inject a replace
  # directive so the derivation compiles against the flake input's source
  # rather than the go.mod version. This enables local iteration on go-finding
  # without publishing a new module version.
  #
  # IMPORTANT: bumping go-finding requires a coordinated 3-place update:
  #   1. go.mod / go.sum
  #   2. flake.nix go-finding-src input ref
  #   3. flake.lock re-lock
  # Forgetting any one produces a split-brain: `go build` (go.mod) and
  # `nix build` (flake input via replace) disagree.
  #
  # The overlay path (flake.overlays.default) calls this WITHOUT go-finding-src,
  # so the replace is skipped there and go.mod is honored as-is.
  postPatch =
    if go-finding-src != null then
      ''
        if ! grep -q 'replace github.com/larsartmann/go-finding => ' go.mod; then
          echo 'replace github.com/larsartmann/go-finding => ${go-finding-src}' >> go.mod
        fi
      ''
    else
      null;
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
    platforms = platforms.all;
    maintainers = [
      {
        name = "Lars Artmann";
        email = "git@lars.software";
        github = "LarsArtmann";
        githubId = 23587853;
      }
    ];
  };
}
