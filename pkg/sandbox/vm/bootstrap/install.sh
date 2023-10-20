#!/bin/sh

mkfs.ext4 /dev/vda
mount -t ext4 /dev/vda /mnt
mkdir -p /mnt/nix /mnt/boot /mnt/etc/nixos
mount -t virtiofs boot /mnt/boot
cat << 'EOF' > /mnt/etc/nixos/configuration.nix
{{ template "configuration.nix" . -}}
EOF
NIX_CONFIG="experimental-features = nix-command flakes" nixos-install --no-root-password --show-trace --root /mnt
