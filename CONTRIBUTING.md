# Contributing

Contributions are welcome! Please open issues on the repository, or submit a PR.

## Issue Reporting

Please report any bugs and suggest features by creating an issue.

*Please note that as of right now, I (Sumner) am basically the only contributor
to this project, so my response time to your issue may be anywhere from instant
to infinite.*

**When reporting a bug**, please be as specific as possible, and include steps
to reproduce. Additionally, you can run offlinemsmtp with the `-m` flag to
enable logging at different levels. For the most verbose logging, run
offlinemsmtp with debug level logging:

```
offlinemsmtp -m debug
```

You can also send logs to a file using the `-l` flag.

## Code

If you want to propose a code change, please submit a PR. If it is good, I will
merge it in.

### Installing Development Dependencies

This project uses [pip-tools][] and [flit][] to manage dependences for both the
core package as well as for development. You only need to install `pip-tools` or
`flit` if you want to change any of the project's dependencies.

### Installation

It is recommended to develop within a virtual environment. See [the docs for
setting up a virtual
environment](https://docs.python.org/3/library/venv.html#creating-virtual-environments)

Then, after activating your virtual environment, run:
```
$ pip install -r requirements.txt -r dev-requirements.txt
$ pip install -e .
```
to install the development dependencies as well as install `offlinemsmtp` into
the virtual environment as editable.

### Running

Run:
```
$ offlinemsmtp
```
to launch the application.

### Code Style

This project follows [black][] strictly. The *only* exception is maximum line
length, which is 99 for this project (in accordance with `black`'s defaults).
Lines that contain a single string literal are allowed to extend past the
maximum line length limit.

This project uses [flake8][], [isort][], [mypy][], and [black][] to do static
analysis of the code and to enforce a consistent (and as deterministic as
possible) code style.

The linting checks are enforced at commit-time using [pre-commit][]. The
pre-commit hooks can be installed using:
```
$ pre-commit install --install-hooks
```

Although you can technically do all of the formatting yourself, it is
recommended that you use the following tools (they are included in
`all-requirements.txt`). The pre-commit hooks and CI process uses these to check
all commits, so you will probably want these so you don't have to wait for
results of the build before knowing if your code is the correct style.

* [`flake8`][flake8] is used for linting. The following
  additional plugins are also used:

  * `flake8-annotations`: enforce type annotations on function definitions.
  * `flake8-bugbear`: enforce a bunch of fairly opinionated styles.
  * `flake8-comprehensions`: enforce usage of comprehensions wherever possible.
  * `flake8-pep3101`: no `%` string formatting.
  * `flake8-print`: to prevent using the `print` function. The more powerful
    `logging` should be used instead. In the rare case that you actually want to
    print to the terminal (the `--version` flag for example), then just disable
    this check with a `# noqa` or a `# noqa: T001` comment.

* [`isort`][isort] is used to sort the imports consistently.

* [`mypy`][mypy] is used for type checking. All type errors must be resolved.

* [`black`][black] is used for auto-formatting. The CI process runs `black
  --check` to make sure that you've run `black` on all files (or are just good
  at manually formatting).

* `TODO` statements must include an associated issue number (in other words, if
  you want to check in a change with outstanding TODOs, there must be an issue
  associated with it to fix it).

The CI process runs all of the above checks on the code. You can run the same
checks that the lint job runs yourself with the following commands:
```
$ flake8
$ isort . --check --diff
$ mypy offlinemsmtp
$ black --check .
$ ./cicd/custom_style_check.py
```

### Commit Message Format

Commits should be reasonably self-contained, that is, each commit should make
sense in isolation. Amending and force pushing is encouraged to help maintain
this.

Commit messages should be formatted as follows:

```
{component}: {short description}

{long description}
```

### GitHub Actions Workflows

This project uses two GitHub Actions workflows for building, testing, and
deploying the application to PyPi. A brief description of each of the workflows
is as follows:

* `build.yaml` - lint, build, and (if a release) deploy the project to PyPi

[black]: https://github.com/psf/black
[flake8]: https://github.com/pycqa/flake8
[flit]: https://github.com/pypa/flit
[isort]: https://pycqa.github.io/isort/
[matrix]: https://matrix.to/#/!veTDkgvBExJGKIBYlU:matrix.org?via=matrix.org
[mypy]: http://mypy-lang.org/
[pip-tools]: https://github.com/jazzband/pip-tools
[pyenv]: https://github.com/pyenv/pyenv
[pre-commit]: https://pre-commit.com/
