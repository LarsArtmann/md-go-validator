{
  lib,
  buildGoModule,
  self ? { },
}:
let
  version = self.shortRev or self.dirtyShortRev or "dev";
  vendorHash = "sha256-o+/CwY92p/ic8EH2i4rhXzPrFrHTYRzH/4SK5KKeSyg=";

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
