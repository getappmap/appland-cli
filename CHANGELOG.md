# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0]
### Added
- `stats` subcommand to show some simple statistics for a set AppMaps.

## [0.6.1]
### Fixed
- Close each scenario file after it's uploaded.

## [0.6.0]
### Changed
- Use multipart/mixed type for the upload. This enables the server to process the data
  faster by removing the need for JSON parsing.
  Note this requires a recent server revision; the code won't work with older servers.

## [0.5.3]
### Fixed
- Segfault on login introduced in 0.5.2.

## [0.5.2] (yanked)
### Added
- `--bench` flag for the upload command allows printing detailed timings of the process.
  (This is mainly useful for development, hence the patch version bump.)

## [0.5.1]
### Fixed
- Errors unrelated to command line arguments will no longer print the command usage.
- Upload errors now report the file name being processed at the time the error occurred.

## [0.5.0]
### Changed
- By default, each uploaded file must be 2MB or less. When a larger file is encountered within a directory that is being uploaded, a warning is printed and the large file is skipped.
- Large files can be uploaded using the `--force/-f` option.

## [0.4.0]
### Changed
- Git metadata collection will now continue and retain information in the event
  of an error.
- When uploading AppMaps, `branch` and `commit` must both be present or ommitted.
  Presence of one without the other will result in failure.
- Branch override flag provided during upload will be propagated to scenario
  metadata.

### Added
- The `--app`/`-a` flag may be used on upload to override the application
  `name` property from `appmap.yml`.
- `upload` arguments can be files or directories. When a directory is specified,
  all `*.appmap.json` files in the directory will be loaded into the mapset.
- The `-f` flag may be used on upload to specify the path to an `appmap.yml` file

### Removed
- Alternate strategies to resolve a branch name have been removed in favor of
  specifying the `--branch` or `-b` flag upon upload.
- `--org/-o` flag has been removed. Organization can be specified by passing `<org-name>/<app-name>` as the `--app/-a` flag.

## [0.3.0]
### Changed
- Git metadata collection should be more resilient in cases where `HEAD` is not
  a branch
- Add a `--branch/-b` flag to `appland upload` which specifies a branch name
  fallback if the branch name cannot be resolved from Git.

## [0.2.0]
### Changed
- MapSet creation now supplies `branch`, `commit`, `environment`, `version`
  parameters.

## [0.1.0]
### Changed
- `appland login` accepts an API key as the username.
- The `.appland` configuration file will only be written when necessary.
- Updated instructions for running in CI/CD environments.

## [0.0.5]
### Added
- `APPLAND_API_KEY` environment variable overrides all configuration.
- `APPLAND_URL` environment variable overrides all configuration.

## [0.0.4]
### Added
- Testing git metadata collection

### Changed
- The browser will open by default after a successful upload. The `--no-open`
  flag disables this behavior.
- Progress bar rendering improved under edge cases

### Fixed
- Fixes an issue when reading an appmap.yml file from the current directory

## [0.0.3]
### Changed
- Git tag resolution should be more robust under strange conditions
- Provide more information on positional arguments
- Values preceded by `$` will be resolved as environment variables when reading
  the `.appland.yml` configuration file. For example:
  ```yml
  current_context: default
  contexts:
    default:
      url: https://app.land
      api_key: $APPLAND_API_KEY
  ```

## [0.0.2] - 2014-05-31
### Added
- Improved error messages if git metadata collection fails

### Changed
- Uploads no longer fail if git metadata cannot be collected. Instead, a warning
  is logged.


[0.0.5]: https://github.com/applandinc/appland-cli/releases/tag/0.0.5
[0.0.4]: https://github.com/applandinc/appland-cli/releases/tag/0.0.4
[0.0.3]: https://github.com/applandinc/appland-cli/releases/tag/0.0.3
[0.0.2]: https://github.com/applandinc/appland-cli/releases/tag/0.0.2
[0.0.1]: https://github.com/applandinc/appland-cli/releases/tag/0.0.1
