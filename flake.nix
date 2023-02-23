{
  description = "Development environments on your infrastructure";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs-terraform.url = "github:nixos/nixpkgs/f44884060cb94240efbe55620f38a8ec8d9af601";
    flake-utils.url = "github:numtide/flake-utils";
    drpc.url = "github:storj/drpc/v0.0.32";
  };

  outputs = { self, nixpkgs, nixpkgs-terraform, flake-utils, drpc }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        formatter = pkgs.nixpkgs-fmt;
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            bash
            bat
            cairo
            drpc.defaultPackage.${system}
            exa
            getopt
            git
            go-migrate
            go_1_20
            golangci-lint
            gopls
            gotestsum
            jq
            nfpm
            nodePackages.typescript
            nodePackages.typescript-language-server
            nodejs
            openssh
            openssl
            pango
            pixman
            postgresql
            pkg-config
            protoc-gen-go
            ripgrep
            shellcheck
            shfmt
            sqlc
            nixpkgs-terraform.legacyPackages.${system}.terraform
            typos
            yarn
            yq
            zip
            zstd
          ];
        };
      }
    );
}
