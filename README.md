# helm-valgrade

a helm plugin that carries over the changes made to the base values file to the new version of the values file.

inspired by: [databus23/helm-diff](https://github.com/databus23/helm-diff)

## installation

```bash
helm plugin install https://github.com/cstanislawski/helm-valgrade
```

## available flags

### required flags

- `--version-base` / `-b` - the version of the chart you are upgrading from
- `--version-target` / `-t` - the version of the chart you are upgrading to
- `--values` / `-f` - the path to the values file you are using

### output options (one required)

- `--output-file` / `-o` - the path to the output file
- `--in-place` / `-i` - update the values file in place

### optional flags

- `--repository` / `-r` - the repository where the chart is located
- `--chart` / `-c` - the name of the chart
- `--keep` / `-k` - exclude specific values from the upgrade process. can be used multiple times. format: `--keep "key1.subkey" --keep "key2"`
- `--silent` / `-s` - suppress all output
- `--log-level` / `-l` - set the log level (debug, info, warn, error, fatal). default: info
- `--dry-run` / `-d` - print the result without writing to the output file
- `--ignore-missing` - ignore missing values in the old chart version. does not apply to user-specified changes
- `--help` / `-h` - display the help message

## TODO

sections to be added:

- usage
- examples
- how it works
- local development
