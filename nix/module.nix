# NixOS module: regenerate dindoa calendars on a schedule.
#
# Every entry under services.dindoa.calendars.<name> becomes a oneshot
# dindoa-<name>.service with a dindoa-<name>.timer, writing
# <outputDir>/dindoa-<name>.ics. When `email` is set, a failed run sends the
# tail of the journal to that address via sendmail.
self:
{ config, lib, pkgs, ... }:

let
  cfg = config.services.dindoa;
  inherit (lib) mkOption mkIf types;

  calendarModule = { name, ... }: {
    options = {
      team = mkOption {
        type = types.str;
        example = "j4";
        description = "Team to generate the calendar for, as accepted by `dindoa --team`.";
      };

      email = mkOption {
        type = types.nullOr types.str;
        default = null;
        example = "ouder@example.org";
        description = "Address that receives a mail when a run fails. null disables mail.";
      };

      schedule = mkOption {
        type = types.str;
        default = "Mon,Thu 06:00";
        description = "systemd OnCalendar expression for this calendar.";
      };

      fileName = mkOption {
        type = types.str;
        default = "dindoa-${name}.ics";
        description = "File name of the calendar inside outputDir.";
      };
    };
  };

  generateScript = name: cal: pkgs.writeShellScript "dindoa-${name}" ''
    set -euo pipefail
    target=${lib.escapeShellArg "${cfg.outputDir}/${cal.fileName}"}

    # Write next to the target so the final mv is an atomic rename: a client
    # fetching the feed never sees a half-written file.
    tmp=$(mktemp -p ${lib.escapeShellArg cfg.outputDir} .dindoa-${name}.XXXXXX)
    trap 'rm -f "$tmp"' EXIT
    rm -f "$tmp"

    ${lib.getExe' cfg.package "dindoa"} --team ${lib.escapeShellArg cal.team} --output "$tmp"

    # dindoa exits 0 without writing a file when the programme has no matches
    # for the team yet (it is published in blocks). Keep the current feed.
    if [ ! -e "$tmp" ]; then
      echo "No published matches for ${cal.team}; keeping $target"
      exit 0
    fi

    if ! grep -q '^BEGIN:VEVENT' "$tmp"; then
      echo "Generated calendar contains no events; keeping $target" >&2
      exit 1
    fi

    chmod 0644 "$tmp"
    mv -f "$tmp" "$target"
    echo "Updated $target ($(grep -c '^BEGIN:VEVENT' "$target") matches)"
  '';

  failureScript = name: cal: pkgs.writeShellScript "dindoa-${name}-failure" ''
    set -euo pipefail
    {
      echo "From: dindoa@${config.networking.hostName}"
      echo "To: ${cal.email}"
      echo "Subject: dindoa-${name} failed on ${config.networking.hostName}"
      echo "Content-Type: text/plain; charset=UTF-8"
      echo
      echo "dindoa-${name}.service (team ${cal.team}) failed at $(date '+%F %T %Z')."
      echo "The previous calendar in ${cfg.outputDir} was left in place."
      echo
      ${pkgs.systemd}/bin/journalctl -u dindoa-${name}.service -n 40 --no-pager -o short-iso
    } | ${cfg.sendmail} -t
  '';

  forCalendars = f: lib.concatMapAttrs f cfg.calendars;
in
{
  options.services.dindoa = {
    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.dindoa;
      defaultText = lib.literalExpression "dindoa.packages.\${system}.dindoa";
      description = "The dindoa package to run.";
    };

    outputDir = mkOption {
      type = types.str;
      default = "/var/lib/dindoa/calendars";
      example = "/var/www/example.org/dindoa";
      description = ''
        Directory the calendars are written to. It is created if missing and
        owned by `user`:`group` with mode 2775, so members of `group` can
        still place files there by hand.
      '';
    };

    user = mkOption {
      type = types.str;
      default = "dindoa";
      description = "User the generator runs as. Created when left at the default.";
    };

    group = mkOption {
      type = types.str;
      default = "dindoa";
      description = "Group owning outputDir. Created when left at the default.";
    };

    sendmail = mkOption {
      type = types.str;
      default = "/run/wrappers/bin/sendmail";
      description = "sendmail-compatible binary used for failure mails.";
    };

    calendars = mkOption {
      type = types.attrsOf (types.submodule calendarModule);
      default = { };
      example = lib.literalExpression ''
        { boaz = { team = "j4"; email = "ouder@example.org"; }; }
      '';
      description = "Calendars to generate, keyed by name (used in unit and file names).";
    };
  };

  config = mkIf (cfg.calendars != { }) {
    users.users = mkIf (cfg.user == "dindoa") {
      dindoa = {
        isSystemUser = true;
        group = cfg.group;
        home = "/var/lib/dindoa";
      };
    };
    users.groups = mkIf (cfg.group == "dindoa") { dindoa = { }; };

    systemd.tmpfiles.rules = [ "d ${cfg.outputDir} 2775 ${cfg.user} ${cfg.group} -" ];

    systemd.services = forCalendars (name: cal: {
      "dindoa-${name}" = {
        description = "Generate dindoa calendar ${name} (team ${cal.team})";
        wants = [ "network-online.target" ];
        after = [ "network-online.target" ];
        onFailure = lib.optional (cal.email != null) "dindoa-${name}-failure.service";
        path = [ pkgs.coreutils pkgs.gnugrep ];
        environment.HOME = "/var/lib/dindoa";
        serviceConfig = {
          Type = "oneshot";
          ExecStart = generateScript name cal;
          User = cfg.user;
          Group = cfg.group;
          UMask = "0002";
          StateDirectory = "dindoa";
          ReadWritePaths = [ cfg.outputDir ];
          ProtectSystem = "strict";
          ProtectHome = true;
          PrivateTmp = true;
          PrivateDevices = true;
          NoNewPrivileges = true;
          ProtectKernelTunables = true;
          ProtectKernelModules = true;
          ProtectControlGroups = true;
          RestrictSUIDSGID = true;
          LockPersonality = true;
          RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
        };
      };
    } // lib.optionalAttrs (cal.email != null) {
      "dindoa-${name}-failure" = {
        description = "Mail about a failed dindoa-${name} run";
        serviceConfig = {
          Type = "oneshot";
          ExecStart = failureScript name cal;
          User = cfg.user;
          Group = cfg.group;
          # Reading another unit's journal needs this group; sendmail is a
          # setgid wrapper, so no NoNewPrivileges here.
          SupplementaryGroups = [ "systemd-journal" ];
        };
      };
    });

    systemd.timers = forCalendars (name: cal: {
      "dindoa-${name}" = {
        description = "Schedule for dindoa-${name}";
        wantedBy = [ "timers.target" ];
        timerConfig = {
          OnCalendar = cal.schedule;
          Persistent = true;
          RandomizedDelaySec = "10min";
        };
      };
    });
  };
}
