# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
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


[0.0.3]: https://github.com/applandinc/appland-cli/releases/tag/0.0.3
[0.0.2]: https://github.com/applandinc/appland-cli/releases/tag/0.0.2
[0.0.1]: https://github.com/applandinc/appland-cli/releases/tag/0.0.1