gits
====
Management tool for git repositories

Usage:
------
1. Generate JSON format configuration file (contain paths to git repositories)
	```sh
	# list candidate paths to the configuration file
	gits -list-config

	# change directory to repository
	cd /path/to/repository

	# output to stdout then check the template
	gits -template

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

3. Run commands on all repositories
	```sh
	gits status
	gits diff
	gits fetch
	# ...etc

	# see "allowd_commands" on configuration file
	```

4. If need remove repository from configuration file
	```sh
	# list key of repositories
	gits -list-keys
	# or show full list
	gits -list

	# after check the key then remove repository from configuration file
	gits -rm $key
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
