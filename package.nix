{
  lib,
  buildGoModule,
  self ? { },
}:
let
  version = self.rev or self.dirtyRev or "dev";
  vendorHash = "sha256-gszC1DS4vvxPQxUWIOk6TlDMxSc6Djva5b/5cbCg+l0=";

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
}
