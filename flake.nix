{
  description = "gh-dash";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    {
      overlays.default = final: prev: {
        gh-dash = self.packages.${prev.system}.default;
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        gh-dash = pkgs.buildGoModule {
          pname = "gh-dash";
          version = "v4.10.0";
          src = ./.;
          vendorHash = "sha256-lqmz+6Cr9U5IBoJ5OeSN6HKY/nKSAmszfvifzbxG7NE=";
        };
      in
      {
        packages = {
          default = gh-dash;
          inherit gh-dash;
        };

        devShells.default = pkgs.mkShell {
          name = "gh-dash";
          inherit (gh-dash) nativeBuildInputs buildInputs;
          packages = with pkgs; [
            fd
            goimports-reviser
            golangci-lint
            golangci-lint-langserver
            gopls
            nerdfix
          ];
        };
      }
    );
}
