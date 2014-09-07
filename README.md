rnzoo
====

rnzoo is useful cli to use ec2. it has some commands.

### ec2ssh

you can select server easy from list and do ssh.

### ec2list

you can show ec2 servers list.

### ltsv_filter

filtering ltsv

## How to Install

### install boto

    pip install boto

### set access key (.bashrc etc...)

    export AWS_ACCESS_KEY_ID=
    export AWS_SECRET_ACCESS_KEY=
    
    # option: specify region
    export AWS_REGION=

### install peco

please check project page.
[peco](https://github.com/peco/peco)

### clone and PATH

    git clone git@github.com:reiki4040/rnzoo.git

    export PATH="PATH:$pathto/rnzoo/bin"


## How to use

run command

    ec2ssh

show ec2 instances list. you can filter instances list by peco.

    Query>
    instance name1 X.X.X.X
    instance name2 X.X.X.Y
    
choose the instance, then start ssh to the instance.

    instanse $ 

## More useful

### cache

ec2ssh(ec2list) does create cache the instances list automatically.
if you update instances, you must be reload with -f option.
(launch, start, stop etc...)

    ec2ssh -f

without -f, ec2ssh does load from cache file. it is faster than connect to AWS(with -f).

### ssh config

if you created ssh config (ex ~/.ssh/config), ec2ssh can works without -l, -i options.

    Host <ec2_ipaddress>
         User <ssh_user>
         IdentityFile <to_identity_fie_path>

### filtering

ec2ssh can filter instances with using arguments 

    ec2ssh web server

already filtered and it is able to modify if you want.

    QUERY>web server
    web server1 X.X.X.X
    web server2 Y.Y.Y.Y

## TODO

- Test code

## Copyright and LICENSE

Copyright (c) 2014- reiki4040

MIT LICENSE