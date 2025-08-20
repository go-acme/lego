.PHONY: default clean serve build

default: clean serve

clean:
	rm -rf public/


build: clean
	hugo --enableGitInfo --source .

serve:
	hugo server --disableFastRender --enableGitInfo --watch --source .
	# hugo server -D
