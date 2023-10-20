{ config, pkgs, lib, ... }:

{
  boot = {
    consoleLogLevel = 0;
    kernelParams = [ "quiet" "udev.log_level=3" ];

    initrd = {
      enable = true;
      verbose = false;
    };

    loader = {
      grub.enable = false;
      generationsDir = {
        enable = true;
        copyKernels = true;
      };
    };
  };

  environment = {
    defaultPackages = [ ];
    systemPackages = with pkgs; [
      curl
      git
      vim
      ((builtins.getFlake "github:jetpack-io/devbox?ref=gcurtis/flake").packages."{{.System}}".default)
    ];
  };

  fileSystems = {
    "/" = {
      device = "/dev/vda";
      fsType = "ext4";
    };
    "/boot" = {
      device = "boot";
      fsType = "virtiofs";
    };
    "/home/{{.User.Username}}/devbox" = {
      device = "home";
      fsType = "virtiofs";
      options = [ "nofail" ];
    };
  };

  nix = {
    settings = {
      auto-optimise-store = true;
      experimental-features = [ "ca-derivations" "flakes" "nix-command" "repl-flake" ];
    };
  };

  nixpkgs = {
    config = {
      allowInsecure = true;
      allowUnfree = true;
    };
    hostPlatform = lib.mkDefault "{{.System}}";
  };

  programs.bash.promptInit = "PS1='dxvm\$ '";

  security.sudo = {
    extraConfig = "Defaults lecture = never";
    wheelNeedsPassword = false;
  };

  services.getty = {
    autologinUser = "{{.User.Username}}";
    greetingLine = lib.mkForce "";
    helpLine = lib.mkForce "";
    extraArgs = [ "--skip-login" "--nohostname" "--noissue" "--noclear" "--nonewline" "--8bits" ];
  };

  system.stateVersion = "23.05";

  time.timeZone = "{{.TimeZone}}";

  users.users = {
    root = {
      hashedPassword = "";
    };
    "{{.User.Username}}" = {
      isNormalUser = true;
      description = "{{.User.Name}}";
      hashedPassword = ""; # passwordless login
      extraGroups = [ "wheel" ];
    };
  };

  virtualisation.rosetta.enable = {{.Rosetta}};
}
