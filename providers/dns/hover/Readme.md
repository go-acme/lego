# Hover DNS Provider

Hover has no formal API, so this mimicks a http client.  Therefore, it uses a plaintext password,
which is generally a very bad idea.  I would recommend where possible, and not protected by some
sort of configmap or such, to put those plaintext creds in a file, and reference that file from the
lego process

```bash
HOVER_PASSFILE=/private/hover.passwd lego ...
```

## Parameters

Config is done using either one or two environment variables (if you see different text in
`go run ./cmd/lego/ -- dnshelp -c hover`, this document needs update, `go run` output is correct)

### Authenticate using username and Plaintext password in environment

I really don't recommend using this, but if your username is "scott" and password "tiger", you should use:
```bash
export HOVER_USERNAME="scott"
export HOVER_PASSWORD="tiger"
lego -a -m chickenandpork@github.com -d example.com --dns hover run
```

Really, if you cannot protect visibility of the environment, use a file to further restrict access,
per next example.

### Authenticate using username and Plaintext password in file

This is the safer approach.  Using a temporary file in a directory restricted to just the necessary
processes, or mapped into a container, this can give a higher level of protection on plaintext
passwords.

The parser for this auth is currently just JSON, but intends to fallback to other formats for
compatibility.  There are so many formats to choose from, so there's intention to expand to 2 or 3
common formats for the most versatility.  I get discouraged when using a tool that says "in order
to use this product, you need to learn an arcane object-oriented COBOL variant" or such.  This
initial release is JSON, however.

```bash
export HOVER_PASSFILE=<some-temp-file>
echo '{"username": "scott", "plaintextpassword": "tiger"}' > "${HOVER_PASSFILE}"
lego -a -m chickenandpork@github.com -d example.com --dns hover run
```

This initial version of the underlying library does spam a lot of logs; this is intended to offer
debug details in case things go poorly.  I expect the logs will silence over time.

## Testing

Testing is done similar to the recommended usage above:

```bash
HOVER_PASSFILE=$(mktemp)
echo '{"username": "scott", "plaintextpassword": "tiger"}' > "${HOVER_PASSFILE}" && \
	go run ./cmd/lego/ -a -m chickenandpork@github.com -d example.com --dns hover run
rm -f "${HOVER_PASSFILE}"
```
