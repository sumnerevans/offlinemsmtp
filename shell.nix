{ pkgs ? import <nixpkgs> {} }: with pkgs;
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

  propagatedBuildInputs = with python38Packages; [
    cairo
    msmtp
    pass
    pkg-config
    poetry
    (
      python38.withPackages (
        ps: with ps; [
          pygobject3
          pycairo
          watchdog
          pkgconfig
        ]
      )
    )
  ];

  POETRY_VIRTUALENVS_IN_PROJECT = 1;

  shellHook = ''
    set -x
    export SOURCE_DATE_EPOCH=315532800
    export NIX_SHELL_LAST_UPDATED=$(date +%s)

    # An update happened to the shell.nix, so remove and reinstall everything in the virtualenv
    rm -rf .venv
    poetry install
    set +x
  '';

  # hook for gobject-introspection doesn't like strictDeps
  # https://github.com/NixOS/nixpkgs/issues/56943
  strictDeps = false;
}
