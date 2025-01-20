{
  mkShell,
  go_1_22,
  golangci-lint,
  goimports-reviser,
}:
mkShell {
  name = "gh-dash";

  packages = [
    go_1_22
    golangci-lint
    goimports-reviser
  ];
}
