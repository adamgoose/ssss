{ inputs, cell }:
let
  inherit (inputs) devenv n2c cells;
  pkgs = cell.pkgs.default;
in
{
  default = devenv.lib.mkShell {
    inherit inputs pkgs;
    modules = [
      ({ pkgs, ... }: {

        imports = [
          cells.surrealdb.devenvModules.surrealdb
        ];

        languages.go.enable = true;

        packages = with pkgs; [
          air
          expect
          gomod2nix
          surrealdb
        ];

        services.surrealdb = {
          enable = true;
          dbPath = "memory";
        };

        processes.air.exec = "unbuffer air";

        pre-commit.hooks = {
          gomod2nix = {
            enable = true;
            entry = "${pkgs.gomod2nix}/bin/gomod2nix";
            files = "go.mod|go.sum";
            pass_filenames = false;
          };
        };

      })
    ];
  };
}
