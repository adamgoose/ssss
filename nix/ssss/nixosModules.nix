{ inputs, cell }:
{
  default = { lib, config, pkgs, ... }:
    let
      cfg = config.services.ssss;
      l = lib // builtins;
    in
    {

      options.services.ssss = {
        enable = l.mkEnableOption (l.mdDoc "Shamir's Secret Sharing Service");

        host = l.mkOption {
          default = "127.0.0.1";
          type = l.types.str;
          description = l.mdDoc ''
            The address to bind the SSSS server to.
          '';
        };

        port = l.mkOption {
          default = 23234;
          type = l.types.port;
          description = l.mdDoc ''
            The port to bind the SSSS server to.
          '';
        };

        hostKeyPath = l.mkOption {
          default = "/etc/ssh/ssh_host_ed25519_key";
          type = l.types.str;
          description = l.mdDoc ''
            The location of the host key file.
          '';
        };

        surrealdb = {

          address = l.mkOption {
            default = "ws://127.0.0.1:8000/rpc";
            type = l.types.str;
            description = l.mdDoc ''
              The address of the SurrealDB server.
            '';
          };

          user = l.mkOption {
            default = "root";
            type = l.types.str;
            description = l.mdDoc ''
              The username to use when connecting to the SurrealDB server.
            '';
          };

          pass = l.mkOption {
            default = "root";
            type = l.types.str;
            description = l.mdDoc ''
              The password to use when connecting to the SurrealDB server.
            '';
          };

          ns = l.mkOption {
            default = "ssss";
            type = l.types.str;
            description = l.mdDoc ''
              The namespace to use when connecting to the SurrealDB server.
            '';
          };

          db = l.mkOption {
            default = "ssss";
            type = l.types.str;
            description = l.mdDoc ''
              The namespace to use when connecting to the SurrealDB server.
            '';
          };

        };
      };

      config = l.mkIf cfg.enable {
        systemd.packages = [
          cell.apps.default
          pkgs.surrealdb-migrations
        ];

        # Run the SSSS server
        systemd.services.ssss = {
          description = "Shamir's Secret Sharing Service";
          after = [ "network.target" ];
          wantedBy = [ "multi-user.target" ];

          environment = {
            SSSS_HOST = cfg.host;
            SSSS_PORT = "${toString cfg.port}";
            SSSS_HOST_KEY_PATH = cfg.hostKeyPath;
            SSSS_SURREALDB_ADDRESS = cfg.surrealdb.address;
            SSSS_SURREALDB_USER = cfg.surrealdb.user;
            SSSS_SURREALDB_PASS = cfg.surrealdb.pass;
            SSSS_SURREALDB_NS = cfg.surrealdb.ns;
            SSSS_SURREALDB_DB = cfg.surrealdb.db;
          };

          serviceConfig = {
            Restart = "on-failure";
            SuccessExitStatus = "3 4";
            RestartForceExitStatus = "3 4";
            ExecStart = "${cell.apps.default}/bin/ssss";
          };
        };
      };

    };
}
