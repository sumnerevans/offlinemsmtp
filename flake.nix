{
  description = "Offline queue daemon for msmtp";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs { system = system; };
          nativeBuildInputs = with pkgs; [
            gobject-introspection
            python3Packages.setuptools
            wrapGAppsHook
          ];
        in
        rec {
          devShells = {
            default = pkgs.mkShell {
              inherit nativeBuildInputs;

              buildInputs = with pkgs; [
                cairo
                libnotify
                msmtp
                pass
                pkg-config
                pre-commit
                rnix-lsp

                python3
                python3Packages.pygobject3
                python3Packages.pycairo
                python3Packages.pkgconfig
              ];
            };

            shellHook = ''
              export SOURCE_DATE_EPOCH=315532800
            '';
          };
        }
      ));
}
