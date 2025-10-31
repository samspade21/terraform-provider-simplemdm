# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Deprecated

- **Assignment Group Resource**: Added deprecation warnings for `group_type` and `install_type` fields
  - `group_type`: This field is deprecated by the SimpleMDM API and may be ignored for accounts using the New Groups Experience
  - `install_type`: The SimpleMDM API recommends setting install_type per-app using the Assign App endpoint instead of at the group level
  - Both fields remain supported for backward compatibility but their behavior may vary by account type
  - See documentation for migration guidance and alternative approaches