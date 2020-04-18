# Terraform formatting tool

So `terraform fmt` doesn't give you any options at all, for example indent space count so made this for those whos formatting opinions differ from Hashicorp's :D

Running `terrafmt` without any args will format files in the current directory, it has `--check` and `--diff` for source validation in CI.

```
terrafmt --help
```
