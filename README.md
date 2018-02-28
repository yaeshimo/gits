gits
====
gits is command runner for git repositories

Usage:
------
1. Generate JSON format configuration file
	```sh
	cd /path/to/repository
	gits -template > "$HOME"/.gits.json
	```

2. Append some repositories
	```sh
	# append from pwd
	cd /path/to/repository && gits -add ./
	# or open with $EDITOR then edit
	gits -edit
	```

3. Can run some commands on all repositories
	```sh
	gits status
	gits diff
	gits fetch
	# ...etc

	# exchange executable
	gits -exec pwd

	# see "allowd_commands" on configuration file
	```

4. If need remove repository from configuration file
	```sh
	# show keys
	gits -list-keys
	# remove repository from configuration file
	gits -rm "$repokey"

	# or edit yourself
	vim /path/to/gits.json
	# or open with $EDITOR then edit
	gits -edit
	```

5. Other options and Examples
	```sh
	# show help
	gits -help
	```
	```sh
	# list candidate paths of configuration file
	gits -list-candidates
	```
	```sh
	# pick the repositories with regex RE2
	gits -match "^go-.*" status
	```
	```sh
	# set url
	# append "allow_commands" in configuration file
	# "sh": { "set-url": [ "-c", "git remote set-url origin git@github.com:$(git config user.name)/$(basename $(pwd)).git" ] }
	gits -exec sh set-url
	```

Requirements:
-------------
git

Install:
--------
```sh
go get -v -u github.com/kamisari/gits
```

License:
--------
MIT
