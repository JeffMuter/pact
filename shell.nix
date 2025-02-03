#! /usr/bin/env nix-shell
{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [ 
      sqlc 
      air
      sqlite
      tailwindcss
      go
    ];
}
