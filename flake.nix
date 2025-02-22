{
  description = "gh-dash";

  inputs = {
    nur.url = "github:nix-community/NUR";

    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

    flake-utils.url = "github:numtide/flake-utils";

    caarlos0-nur.url = "github:caarlos0/nur";

  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      nur,
      caarlos0-nur,
    }:
    {
      overlays.default = final: prev: {
        gh-dash = self.packages.${prev.system}.default;
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let

        # overlays allow extending and changing nixpkgs with custom packages
        overlays = [
          (final: prev: {
            nur = import nur {
              nurpkgs = prev;
              pkgs = prev;
              # add caarlos0 and charmbracelet modules to nixpkgs
              repoOverrides = {
                caarlos0 = import caarlos0-nur { pkgs = prev; };
              };
            };
          })
        ];

        pkgs = nixpkgs.legacyPackages.${system};

        gh-dash = pkgs.buildGoModule {
          pname = "gh-dash";
          version = "v4.12.0";
          src = ./.;
          vendorHash = "sha256-7s+Lp8CHo1+h2TmbTOcAGZORK+/1wytk4nv9fgD2Mhw=";
        };
      in
      {
        pkgs.overlays = overlays;

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
            svu
            (callPackage ./docs { })
          ];
        };
      }
    );
}
