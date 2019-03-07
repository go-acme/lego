.PHONY: default clean hugo hugo-build

default: hugo

clean:
	rm -rf public/


hugo-build: clean hugo-themes
	hugo --enableGitInfo --source .

hugo:
	hugo server --disableFastRender --enableGitInfo --watch --source .
	# hugo server -D

hugo-themes:
	rm -rf themes
	mkdir themes
	git clone --depth=1 https://github.com/matcornic/hugo-theme-learn.git themes/hugo-theme-learn
	rm -rf themes/hugo-theme-learn/.git
