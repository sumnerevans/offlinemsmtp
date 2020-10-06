{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    gobject-introspection
    python3Packages.setuptools
  ];

  propagatedBuildInputs = with pkgs; [
    pkg-config
    python38
    cairo
    poetry
  ];

  buildInputs = with pkgs; [
    libnotify
  ];
}
