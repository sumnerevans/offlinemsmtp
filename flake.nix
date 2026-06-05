{
  description = "offlinemsmtp";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
    }:
    (flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" ];
      perSystem =
        {
          pkgs,
          system,
          ...
        }:
        {
          _module.args.pkgs = import inputs.nixpkgs { inherit system; };

          packages = rec {
            default = offlinemsmtp;
            offlinemsmtp = pkgs.buildGoModule {
              pname = "offlinemsmtp";
              version = "unstable";
              src = self;
              subPackages = [ "cmd/offlinemsmtp" ];
              vendorHash = "sha256-ilgfbfeeRKEAekxDPTL3RTSWhpJAO8sCBB/Chn2W7dU=";
            };
          };

          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              msmtp
              pre-commit
            ];
          };
        };
    });
}
