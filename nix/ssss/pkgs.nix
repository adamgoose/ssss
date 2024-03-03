{ inputs, cell }: {
  default = inputs.nixpkgs.appendOverlays [
    inputs.gomod2nix.overlays.default
  ];
}
