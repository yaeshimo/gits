gits
====
git wrapper


Usage:
------
1. **Generate watchlist**  
```sh
gits -template > /path/to/your/watchlist.json
```

2. **Append your git repositories to watchlist**  
  edit watchlist.json yourself, append your repository info  
  or  
  can use command  
```sh
cd /your/git/repository
gits -conf /path/to/your/watchlist.json -watch ./
gits -conf /path/to/your/watchlist.json -list
```

3. **After append**  
```sh
gits -conf /path/to/your/watchlist.json status
gits -conf /path/to/your/watchlist.json fetch
gits -conf /path/to/your/watchlist.json grep [word]
...etc
```

4. **Unwatch**  
```sh
cd /your/git/repository
gits -conf /path/to/your/watchlist.json -unwatch ./
gits -conf /path/to/your/watchlist.json -list
```

5. **Conf Path**  
  if you use default conf path, you can trim conf flag  
  show default conf path  
```sh
gits --show-conf
```


Requirements:
-------------
git


Install:
--------
```sh
go get github.com/kamisari/gits
```


License:
--------
MIT
