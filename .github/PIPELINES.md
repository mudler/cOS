# Github pipelines for cos toolkit

Requirements:
- gomplate

Pipelines are generated from `build.yaml.gomplate`, which is a gomplate template. To generate the pipelines, run `./pipelines.sh`.

Each configuration in `config` is turned to its own pipeline, named `build-$NAME`.

## Github Secrets

The pipeline needs `QUAY_USERNAME` and `QUAY_PASSWORD` as Github Secret for logging in for pushing container images.