{
  description = "Pangbox server for PangYa.";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in rec {
        packages = rec {
          default = pkgs.buildGoModule {
            name = "pangbox";
            src = self;
            vendorHash = "sha256-FuW0wTQFA1l2HJHut/MSQlbcGb4A9mc+PamofREdIAM=";
          };
        };
        devShell = pkgs.mkShell {
          inputsFrom = with packages; [ default ];
        };
      }
    );
}
