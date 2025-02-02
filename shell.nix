{ pkgs ? import <nixpkgs> {} }:
let
  unstable = import (fetchTarball "https://github.com/nixos/nixpkgs/archive/nixos-unstable.tar.gz") {
    inherit (pkgs) system;
  };
in
pkgs.mkShell {
  buildInputs = with pkgs; [ 
      unstable.sqlc 
      air
    ];
}
