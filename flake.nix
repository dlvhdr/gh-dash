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
      packages.${system}.default = pkgs.callPackage ./nix/package.nix { };
      overlays.default = final: prev: {
        ghostty = self.packages.${prev.system}.default;
      };
    };
}
