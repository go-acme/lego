.PHONY: default clean hugo hugo-build

default: clean hugo

clean:
	rm -rf public/


hugo-build: clean
	hugo --enableGitInfo --source .

hugo:
	hugo server --disableFastRender --enableGitInfo --watch --source .
	# hugo server -D
