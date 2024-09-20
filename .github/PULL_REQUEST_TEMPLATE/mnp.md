PULL REQUEST TEMPLATE FOR MAINTAINERS ONLY.

https://github.com/go-acme/lego/compare/master...ldez:branch?quick_pull=1&title=Add+DNS+provider+for+&labels=enhancement,area/dnsprovider,state/need-user-tests&template=mnp.md

?quick_pull=1&title=Add+DNS+provider+for+&labels=enhancement,area/dnsprovider,state/need-user-tests&template=mnp.md

---

- [x] adds a description to your PR
- [x] have a homogeneous design with the other providers
- [ ] add tests (units)
- [ ] add tests ("live")
- [ ] add a provider descriptor
- [ ] generate CLI help, documentation, and readme.
- [ ] be able to do: _(and put the output of this command to a comment)_
  ```bash
  make build
  rm -rf .lego

  EXAMPLE_USERNAME=xxx \
  ./dist/lego -m your_email@example.com --dns EXAMPLE -d *.example.com -d example.com -s https://acme-staging-v02.api.letsencrypt.org/directory run
  ```
  Note the wildcard domain is important.
- [ ] pass the linter
- [ ] do `go mod tidy`

Ping @xxx, can you run the command (with your domain, email, credentials, etc.)?

Closes #

