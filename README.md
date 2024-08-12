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

- `--repository` / `-r` - the name of the repository where the chart is located
- `--chart` / `-c` - the name of the chart
- `--keep` / `-k` - exclude specific values from the upgrade process. can be used multiple times. format: `--keep "key1.subkey" --keep "key2"`
- `--silent` / `-s` - suppress all output
- `--log-level` / `-l` - set the log level (debug, info, warn, error, fatal). default: info
- `--dry-run` / `-d` - print the result without writing to the output file
- `--ignore-missing` - ignore missing values in the old chart version. does not apply to user-specified changes
- `--help` / `-h` - display the help message

## usage

to use helm-valgrade, run:

```bash
helm valgrade [flags]
```

example:

```bash
helm valgrade -b 58.5.2 -t 58.7.0 -f values.yaml -r prometheus-community -c kube-prometheus-stack -o new-values.yaml
```

note: ensure that the repository (e.g., 'prometheus-community') is already added to your helm repositories. you can add a repository using `helm repo add prometheus-community https://prometheus-community.github.io/helm-charts`

## license

this project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
