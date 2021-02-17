# How to contribute to lego

Contributions in the form of patches and proposals are essential to keep lego great and to make it even better.
To ensure a great and easy experience for everyone, please review the few guidelines in this document.

## Bug reports

- Use the issue search to see if the issue has already been reported.
- Also look for closed issues to see if your issue has already been fixed.
- If both of the above do not apply create a new issue and include as much information as possible.

Bug reports should include all information a person could need to reproduce your problem without the need to
follow up for more information. If possible, provide detailed steps for us to reproduce it, the expected behaviour and the actual behaviour.

## Feature proposals and requests

Feature requests are welcome and should be discussed in an issue.
Please keep proposals focused on one thing at a time and be as detailed as possible.
It is up to you to make a strong point about your proposal and convince us of the merits and the added complexity of this feature.

## Pull requests

Patches, new features and improvements are a great way to help the project.
Please keep them focused on one thing and do not include unrelated commits.

All pull requests which alter the behaviour of the program, add new behaviour or somehow alter code in a non-trivial way should **always** include tests.

If you want to contribute a significant pull request (with a non-trivial workload for you) please **ask first**. We do not want you to spend
a lot of time on something the project's developers might not want to merge into the project.

**IMPORTANT**: By submitting a patch, you agree to allow the project
owners to license your work under the terms of the [MIT License](LICENSE).

### How to create a pull request

Requirements:

- `go` v1.15+
- environment variable: `GO111MODULE=on`

First, you have to install [GoLang](https://golang.org/doc/install) and [golangci-lint](https://github.com/golangci/golangci-lint#install).

```bash
# Create the root folder
mkdir -p $GOPATH/src/github.com/go-acme
cd $GOPATH/src/github.com/go-acme

# clone your fork
git clone git@github.com:YOUR_USERNAME/lego.git
cd lego

# Add the go-acme/lego remote
git remote add upstream git@github.com:go-acme/lego.git
git fetch upstream
```

```bash
# Create your branch
git checkout -b my-feature

## Create your code ##
```

```bash
# Format
make fmt
# Linters
make checks
# Tests
make test
# Compile
make build
```

```bash
# push your branch
git push -u origin my-feature

## create a pull request on GitHub ##
```
