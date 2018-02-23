gits
====
Management tool for git repositories

Usage:
------
1. Generate JSON format configuration file (is contain paths to git repositories)
	```sh
	# list candidate paths to the configuration file
	gits -list-candidates

	# change directory to repository
	cd /path/to/repository

	# output to stdout then check the template
	gits -template

	# on linux
	# write configuration file
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
	# append from PWD
	cd /path/to/repository && gits -add .
	# or
	gits -add /path/to/repository
	# or open with $EDITOR then edit
	gits -edit
	```

3. Can run some commands on all repositories
	```sh
	gits status
	gits diff
	gits fetch
	# ...etc

	# exchange executable from git(default)
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
	gits -rm $key
	# or edit yourself
	vim /path/to/conf
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
