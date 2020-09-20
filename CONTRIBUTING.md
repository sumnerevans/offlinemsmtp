# Contributing

Contributions are welcome! Please
[submit a ticket](https://todo.sr.ht/~sumner/offlinemsmtp) to the ticket
tracker, submit a patch to the
[~sumner/offlinemsmtp-devel](https://lists.sr.ht/~sumner/offlinemsmtp-devel)
mailing list, or discuss the project in general on the
[~sumner/offlinemsmtp-discuss](https://lists.sr.ht/~sumner/offlinemsmtp-discuss)
mailing list.

## Issue Reporting

You can report issues or propose features in the
[ticket tracker](https://todo.sr.ht/~sumner/offlinemsmtp). For longer-form
discourse, please use either
[~sumner/offlinemsmtp-devel](https://lists.sr.ht/~sumner/offlinemsmtp-devel) or
[~sumner/offlinemsmtp-discuss](https://lists.sr.ht/~sumner/offlinemsmtp-discuss).

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

If you want to propose a code change, please submit a patch to the
[~sumner/offlinemsmtp-devel](https://lists.sr.ht/~sumner/offlinemsmtp-devel)
mailing list. If it is good, I will merge it in.

### Installing Development Dependencies

This project uses [Poetry](https://python-poetry.org/) for dependency
management. Make sure that you have poetry (and
[pyenv](https://github.com/pyenv/pyenv) if necessary) set up properly, then run:

    $ poetry install

to install the development dependencies as well as install `offlinemsmtp` into
the virtual environment.

**Note:** a `.envrc` is provided for use with [direnv](https://direnv.net/) to
automatically run the `poetry install` and activate the virtual environment when
in the project directory.

### Running

It is recommended to activate the virtual environment using either `poetry
shell` or by activating it manually using:

    source $(poetry env info -p)/bin/activate

Once the virtual environment is activated, you can run `offlinemsmtp` as
described in the [README](./README.rst#L74).

### Code Style

This project follows [PEP-8](https://www.python.org/dev/peps/pep-0008/)
**strictly**. The *only* exception is maximum line length, which is 88 for this
project (in accordance with `black`'s defaults). Lines that contain a single
string literal are allowed to extend past the maximum line length limit.

This project uses flake8, mypy, and black to do static analysis of the code and
to enforce a consistent (and as deterministic as possible) code style.

Although you can technically do all of the formatting yourself, it is
recommended that you use the following tools (they are automatically installed
if you are using poetry). The CI process uses these to check all commits, so you
will probably want to run them locally so you don't have to wait for results of
the build before knowing if your code is the correct style.

* [`flake8`](https://flake8.pycqa.org/en/latest/) is used for linting. The
  following additional plugin is also used:

  * [`flake8-pep3101`](https://pypi.org/project/flake8-pep3101/): no `%` string
    formatting.

* [`mypy`](http://mypy-lang.org/) is used for type checking. All type errors
  must be resolved.

* [`black`](https://black.readthedocs.io/en/stable/) is used for
  auto-formatting. The CI process runs `black --check` to make sure that you've
  run `black` on all files (or are just good at manually formatting).

The CI process uses all three tools to analyse the Python code. You can run the
same checks that the lint job runs yourself with the following commands:

    $ python setup.py check -mrs
    $ flake8
    $ mypy offlinemsmtp
    $ black --check .
