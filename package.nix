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
