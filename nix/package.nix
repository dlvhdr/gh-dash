{
  lib,
  buildGoModule,
}:
let
in
buildGoModule {
  pname = "gh-dash";
  version = "v4.9.1";

  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.intersection (lib.fileset.fromSource (lib.sources.cleanSource ../.)) (
      lib.fileset.unions [
        ../cmd
        ../config
        ../data
        ../docs
        ../git
        ../imposters
        ../nix
        ../ui
        ../utils
        ../gh-dash.go
        ../go.mod
        ../go.sum
      ]
    );
  };

  vendorHash = "sha256-lqmz+6Cr9U5IBoJ5OeSN6HKY/nKSAmszfvifzbxG7NE=";
}
