There is an [official API client](https://github.com/timeweb-cloud/sdk-go) but this client is completely broken:
- the code is generated and the module name is `github.com/GIT_USER_ID/GIT_REPO_ID`
- the code contains redeclared constants
- Even with fixes to the module name and the redeclared constants, the module doesn't compile.

https://github.com/timeweb-cloud/sdk-go/pull/1

So, for now, this API client is unusable.
