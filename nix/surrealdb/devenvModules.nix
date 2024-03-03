{ inputs, cell }: {

  surrealdb = { pkgs, lib, config, ... }:
    with lib;
    let
      cfg = config.services.surrealdb;
    in
    {
      options.services.surrealdb = {
        enable = mkEnableOption "SurrealDB";

        package = mkOption {
          type = types.package;
          default = pkgs.surrealdb;
          defaultText = literalExpression "pkgs.surrealdb";
          description = "Which package of SurrealDB to use.";
        };

        host = mkOption {
          type = types.str;
          default = "127.0.0.1";
          description = "The host that surrealdb will connect to.";
        };

        port = mkOption {
          type = types.port;
          default = 4222;
          description = "The TCP port to accept connections.";
        };

        dbPath = mkOption {
          type = types.str;
          default = "file://" + config.env.DEVENV_STATE + "/surrealdb";
          description = "Database path. Can be 'memory' or file://...";
        };

        extraFlags = mkOption {
          type = types.listOf types.str;
          default = [ ];
          description = "Extra flags to pass to the SurrealDB process.";
        };

      };

      config = mkIf cfg.enable {
        packages = [
          cfg.package
        ];

        processes.surrealdb = {
          exec = "${cfg.package}/bin/surreal start --bind ${cfg.host}:${toString cfg.port} ${escapeShellArgs cfg.extraFlags} -- ${cfg.dbPath}";

          process-compose = {
            readiness_probe = {
              exec.command = "${cfg.package}/bin/surreal is-ready -e ws://${cfg.host}:${toString cfg.port}";
              initial_delay_seconds = 2;
              period_seconds = 10;
              timeout_seconds = 4;
              success_threshold = 1;
              failure_threshold = 5;
            };

            # https://github.com/F1bonacc1/process-compose#-auto-restart-if-not-healthy
            availability.restart = "on_failure";
          };
        };
      };
    };

}
