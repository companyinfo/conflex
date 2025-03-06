{
  description = "Conflex development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        config = {allowUnfree = true;};
      };
    in {
      formatter = pkgs.alejandra;
      devShells.default = pkgs.mkShell {
        name = "conflex";
        buildInputs = with pkgs; [
          go
          revive
          consul
        ];
      };
    });
}
