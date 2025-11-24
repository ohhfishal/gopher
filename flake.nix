{
  description = "Improved go build output.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = builtins.readFile ./version.txt;
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "gopher";
          inherit version;
          src = self;
          vendorHash = "sha256-OpAqYDjVo5frP8Z+C7oBv8p+IEttsnjj/3TGccC7SAc=";
          meta = {
            description = "Improved go build output.";
          };
        };
      }
    );
}
