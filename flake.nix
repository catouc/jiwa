{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils/v1.0.0";
  };

  description = "CLI to make Jira less annoying to interact with";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem ( system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        build = pkgs.buildGoModule {
          pname = "jiwa";
          version = "v0.7.2";
          modSha256 = pkgs.lib.fakeSha256;
          vendorSha256 = null;
          src = ./.;

          meta = {
            description = "CLI to make Jira less annoying to interact with";
            homepage = "https://github.com/catouc/jiwa";
            license = pkgs.lib.licenses.mit;
            maintainers = [ "catouc" ];
            platforms = pkgs.lib.platforms.linux ++ pkgs.lib.platforms.darwin;
          };
        };
      in
        rec {
        packages = {
          jiwa = build;
          default = build;
        };

        devShells = {
          default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
            ];
          };
        };
      }
    );
}
