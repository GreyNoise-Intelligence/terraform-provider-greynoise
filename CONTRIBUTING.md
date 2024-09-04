# Contributing

Thank you for investing your time and energy by contributing to our project: please ensure you are familiar
with the [HashiCorp Code of Conduct](https://github.com/hashicorp/.github/blob/master/CODE_OF_CONDUCT.md).

This provider is a HashiCorp **provider**, which means any bug fix and feature
has to be considered in the context of the various permutations of configurations in which this provider is used.
This is great as your contribution can have a big positive impact, but we have to assess potential negative impact too
(e.g. breaking existing configurations). _Stability over features_.

To provide some safety to the wider provider ecosystem, we strictly follow
[semantic versioning](https://semver.org/) and HashiCorp's own
[versioning specification](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#versioning-specification).
Any changes that could be considered as breaking will only be included as part of a major release.
In case multiple breaking changes need to happen, we will group them in the next upcoming major release.

## Asking Questions

For questions, curiosity, or if still unsure what you are dealing with,
please see the HashiCorp [Terraform Providers Discuss](https://discuss.hashicorp.com/c/terraform-providers/31)
forum.

## Reporting Vulnerabilities

Please disclose security vulnerabilities responsibly by following the
[HashiCorp Vulnerability Reporting guidelines](https://www.hashicorp.com/security#vulnerability-reporting).

### Changelog

HashiCorp’s open-source projects have always maintained user-friendly, readable `CHANGELOG`s that allow
practitioners and developers to tell at a glance whether a release should have any effect on them,
and to gauge the risk of an upgrade.

We follow Terraform Plugin
[changelog specifications](https://www.terraform.io/plugin/sdkv2/best-practices/versioning#changelog-specification).

#### Changie Automation Tool

This provider uses the [Changie](https://changie.dev/) automation tool for changelog automation.
To add a new entry to the `CHANGELOG` install Changie using the
following [instructions](https://changie.dev/guide/installation/)
and run:

```bash
changie new
```

then choose a `kind` of change corresponding to the Terraform
Plugin [changelog categories](https://developer.hashicorp.com/terraform/plugin/sdkv2/best-practices/versioning#categorization)
and then fill out the body following the entry format. Changie will then prompt for a Github issue or pull request
number.
Repeat this process for any additional changes. The `.yaml` files created in the `.changes/unreleased` folder
should be pushed the repository along with any code changes.

#### Entry format

Change entries that are specific to _resources_ or _data sources_, they should look like:

```markdown
* resource/RESOURCE_NAME: ENTRY DESCRIPTION

* data-source/DATA-SOURCE_NAME: ENTRY DESCRIPTION
```

#### Pull Request Types to `CHANGELOG`

The `CHANGELOG` is intended to show developer-impacting changes to the codebase for a particular version.
If every change or commit to the code resulted in an entry, the `CHANGELOG` would become less useful for developers.
The lists below are general guidelines to decide whether a change should have an entry.

##### Changes that should not have a `CHANGELOG` entry

* Documentation updates
* Testing updates
* Code refactoring

##### Changes that may have a `CHANGELOG` entry

* Dependency updates: If the update contains relevant bug fixes or enhancements that affect developers,
  those should be called out.

##### Changes that should have a `CHANGELOG` entry

* Major features
* Bug fixes
* Enhancements
* Deprecations
* Breaking changes