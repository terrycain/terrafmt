# Terraform formatting tool

So `terraform fmt` doesn't give you any options at all, for example indent space count so made this for those whos formatting opinions differ from Hashicorp's :D

Running `terrafmt` without any args will format files in the current directory, it has `--check` and `--diff` for source validation in CI.

```
terrafmt --help
Usage: terrafmt [<path> ...]

Formats terraform files. If no path is specified, the current working directory is used.

Arguments:
  [<path> ...]    Paths or files to format

Flags:
      --help                         Show context-sensitive help.
      --indent-length=2              Indent size in spaces
      --recursive                    Search recurively for .tf files
      --check                        Check files dont require modification, returns 0 when no changes are required, 1 when changes are needed
      --diff                         Dont modify files but show diff of the changes
  -V, --version                      Displays version
      --line-up-assignment-blocks    Line up blocks of assignments
      --line-up-comment-blocks       Line up blocks of comments

```

## TODO

* Make diff output like `diff -u`
* Fix bug when searching for files (must be a better way to do this)
* Sign and notarize mac binaries as Catalina gets unhappy
