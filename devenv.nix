{ pkgs, ... }:

{
  packages = [
    pkgs.nodePackages.serverless
  ];

  languages.go.enable = true;
}
