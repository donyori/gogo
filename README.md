# gogo

A Go (Golang) toolbox.

---

This library is not systematic and just contains some scattered tools.

The library is under development.
More codes and documents will be added in the future.

## Incompatibility

At the current stage, there is no guarantee that
this library will keep compatible with previous versions.

To avoid incompatibility issues,
please specify the version (e.g., `v0.3.0`) or
the revision number / commit hash (e.g., `309594fd`) explicitly
and do not use the `latest` version.

For information about how to specify the version,
see the [Go documentation](https://go.dev/doc/modules/managing-dependencies#getting_version "Getting a specific dependency version").

## Branch name change

The default branch has been renamed from `master` to `main` since April 4, 2022.

Please update your local clone if it still refers to `origin/master`.
To do so, from your local clone of the repository, run the following commands:

```bash
$ git branch -m OLD-BRANCH-NAME NEW-BRANCH-NAME
$ git fetch origin
$ git branch -u origin/main NEW-BRANCH-NAME
$ git remote set-head origin -a
$ git remote prune origin
```

where `OLD-BRANCH-NAME` is the current name of your local branch
referring to `origin/master`, usually named `master`;
`NEW-BRANCH-NAME` is the new name of that branch, usually named `main`.

For more information about renaming a branch,
see the [GitHub documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-branches-in-your-repository/renaming-a-branch "Renaming a branch").

## License

The GNU Affero General Public License 3.0 (AGPL-3.0) - [Yuan Gao](https://github.com/donyori/).
Please have a look at the [LICENSE](LICENSE).

## Contact

You can contact me by email: [<donyoridoyodoyo@outlook.com>](mailto:donyoridoyodoyo@outlook.com).
