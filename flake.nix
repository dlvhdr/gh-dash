{
  description = "gh-dash";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
      ...
    }:
    let
      system = "aarch64-darwin";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShell.${system} = pkgs.callPackage ./nix/devShell.nix { };

      # this allows users to install gh-dash with nix
      # by including gh-dash.packages.aarch64-darwin.default in their systemPackages
      packages.${system}.default = pkgs.callPackage ./nix/package.nix { };

      # this allows?
      overlays.default = final: prev: {
        gh-dash = self.packages.${prev.system}.default;
      };
    };
}
