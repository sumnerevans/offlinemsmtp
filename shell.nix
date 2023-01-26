{ pkgs ? import <nixpkgs> {} }: with pkgs; let
  py = python38.override {
    packageOverrides = self: super: {
      pycairo = super.pycairo.overridePythonAttrs (
        oldAttrs: rec {
          version = "1.20.0";
          src = oldAttrs.src.override {
            inherit version;
            sha256 = "5695a10cb7f9ae0d01f665b56602a845b0a8cb17e2123bfece10c2e58552468c";
          };
        }
      );
    };
  };
in
pkgs.mkShell {
  nativeBuildInputs = [
    gobject-introspection
    python3Packages.setuptools
    wrapGAppsHook
  ];

  buildInputs = [
    libnotify
    rnix-lsp
  ];

  propagatedBuildInputs = with python3Packages; [
    cairo
    msmtp
    pass
    pkg-config
    (
      python3.withPackages (
        ps: with ps; [
          pygobject3
          pycairo
          pkgconfig
        ]
      )
    )
  ];

  shellHook = ''
    export SOURCE_DATE_EPOCH=315532800
  '';
}
