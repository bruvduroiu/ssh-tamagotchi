{
  description = "Nix flake for Janktag";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs@{ self, nixpkgs, flake-utils, ... }: 
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        devServer = pkgs.writeShellScriptBin "dev-server" ''
          source <(cat .env | sed 's/^/export /g')

          go run $@
        '';
      in {
        devShells.default = pkgs.mkShell {

          packages = with pkgs; [
            go
            devServer
          ];
        };
  });
}
