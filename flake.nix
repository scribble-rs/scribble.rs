{
  description = "Scribble.rs flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }: let
    lib = nixpkgs.lib;
  in {
    packages = lib.genAttrs lib.systems.flakeExposed (system: let
      pkgs = import nixpkgs { inherit system; };
      scribblers = pkgs.buildGoModule {
        pname = "scribblers";
        version = "0.9.14";
        src = lib.cleanSource ./.;
        vendorHash = "sha256-EsN/bfYyMkhXXeDCmTNRCxxSIjmqkNEBARbGy8UJ6xo=";
        subPackages = [ "cmd/scribblers" ];
      };
    in {
      scribblers = scribblers;
      default = scribblers;
    });
  };
}
