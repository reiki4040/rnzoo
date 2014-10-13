rnzoo
====

rnzoo is useful cli to use ec2.

## How to Install

- download archive
- homebrew (coming soon... maybe)

### download and set PATH

download tar.gz file and set PATH

    tar zxf rnzoo-darwin-amd64.tar.gz

    # set .bashrc etc...
    export PATH="PATH:$pathto/rnzoo/bin"

## dependency

### [peco](https://github.com/peco/peco)

    brew tap peco/peco
    brew install peco

if you want more detail, please reference [peco project page](https://github.com/peco/peco)

## Settings

- set AWS ENV variables
- ssh config (Optional but recommended)

### set AWS variables (.bashrc, .bash_profile etc...)

    export AWS_ACCESS_KEY_ID=
    export AWS_SECRET_ACCESS_KEY=
    
    # option: specify default region
    export AWS_REGION=


### ssh config

`vi ~/.ssh/config`

    Host X.X.X.X
      User your_user
      IdentityFile you_key_file

***More useful If you added your ec2 instances to ssh config before using rnssh by yourself.***

rnzoo is going to add way that generate ssh config from AWS.

## Command

| command | description |
|:-|:-|
| rnssh | you can select server easy from list and do ssh. |

## How to use

### run command

    rnssh -l ssh_user -i identity_file

you can run `rnssh` (without options `-l`,`-i`) if you added instances to ssh config.

show ec2 instances list. you can filter instances list by peco.

    Select ssh instance. You can do filtering>
    instance name1 X.X.X.X
    instance name2 X.X.X.Y
    
choose the instance, then start ssh to the instance.

    instanse $ 

## More useful

### cache

rnssh does create cache the instances list automatically.
if you update instances, you must be reload with `-f` option.
(launch, start, stop etc...)

without `-f`, rnssh does load from cache file. it is faster than connect to AWS(with `-f`).

### ssh config

if you created ssh config (ex ~/.ssh/config), rnssh can works without `-l`, `-i` options.

    Host <ec2_ipaddress>
         User <ssh_user>
         IdentityFile <to_identity_fie_path>

### filtering

rnssh can filter instances with using arguments 

    rnssh web server

already filtered and it is able to modify if you want.

    QUERY>web server
    web server1 X.X.X.X
    web server2 Y.Y.Y.Y


## Features in future

- rnzoo is going to add way that generate ssh config from AWS.

## TODO

- Test code

## Copyright and LICENSE

Copyright (c) 2014- reiki4040

MIT LICENSE