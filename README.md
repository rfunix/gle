# gle - The logentries log watcher written in Go.

gle is very simple script written in Go to analyze logentries logs.

## Getting Started

### Prerequisites

Go :D 

### Installing

```
git clone https://github.com/rfunix/gle
cd gle && make compile
```

### How to use

Help Message
```console
rfunix@rfunix:~$ ./gle -h
NAME:
   gle - logentries cli tool

USAGE:
   gle [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log value, -l value  Name of log in logentries
   --api-key value        The logentries api-key, its recomend export envvar with name X_API_KEY [$X_API_KEY]
   --start-date value     The start date period to search log
   --end-date value       The end date period to search log
   --query value          the query to search pattern
   --help, -h             show help
   --version, -v          print the version
```

search log by period and log name
```console
rfunix@rfunix:~$ export X_API_KEY=YOUR_API_KEY
rfunix@rfunix:~$ ./gle --log "log_name" --start-date "2019-10-10 00:00:00" --end-date "2019-10-11 00:00:00" --query "where(query_to_match)"
```


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/rfunix/gle/tags). 


## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

Pompem is free software, keeping the picture can USE AND ABUSE
