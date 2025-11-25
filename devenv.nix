{
  pkgs,
  ...
}:
{
  # https://devenv.sh/languages/
  languages.go = {
    enable = true;
  };

  enterShell = ''
    go version
  '';
}

