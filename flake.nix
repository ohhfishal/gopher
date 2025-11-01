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
        version = "0.1.0";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "gopher";
          inherit version;
          src = pkgs.fetchFromGitHub {
            owner = "ohhfishal";
            repo = "gopher";
            rev = "v${version}";
            hash = "sha256-nsc05TIQ0keojF0FMQyeRHhhM8oC5rUyH9UcFs2F1wo=";
            
          };
          vendorHash = "sha256-bb4R0cCKVPr5Rrh1l0kZmsijg8Ojsf3KUaO/szsHU6I=";
          meta = {
            description = "Improved go build output.";
          };
        };
      }
    );
}
