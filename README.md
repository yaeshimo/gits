gits
====
taskrunner for git repositories

Usage:
------
1. Generate JSON format configuration file (is contain paths to git repositories)
	```sh
	# list candidate paths
	gits -list-candidates
	```

	```sh
	# change directory to repository
	cd /path/to/repository

	# output template to stdout
	gits -template

	# after check the candidate paths and template
	# write configuration file

	# on linux
	gits -template > "$HOME"/.gits.json
	# or
	mkdir -p "$HOME"/.config/gits
	gits -template > "$HOME"/.config/gits/gits.json
	# or if you have $XDG_CONFIG_HOME
	mkdir -p "$XDG_CONFIG_HOME"/gits
	gits -template > "$XDG_CONFIG_HOME"/gits/gits.json
	```

2. Append some repositories
	```sh
	# append from pwd
	cd /path/to/repository && gits -add .
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
	# first check the list
	gits -list-keys
	# or show full list
	gits -list

	# after check the key then remove repository from configuration file
	gits -rm "$repokey"
	# or edit yourself
	vim /path/to/gits.json
	# or open with $EDITOR
	gits -edit
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
