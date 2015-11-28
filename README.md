rnzoo
====

rnzoo is useful cli to use ec2.

rnzoo has been refactor...

## How to install and settings

- **homebrew** (recommend)
- download archive

### homebrew

```
brew tap reiki4040/rnzoo
brew install rnzoo
```

## Settings

- set AWS credentials
- set AWS default region

### set AWS credentials

* credential file (`~/.aws/credentials`)

```
[default]
aws_access_key_id=your_key_id
aws_secret_access_key=your_secret
```

* Environment variable (`~/.bashrc`, `~/.bash_profile`, etc...)

```
export AWS_ACCESS_KEY_ID=
export AWS_SECRET_ACCESS_KEY=
```

### set default AWS region

use init sub command. it shows AWS regions and store to config file (`~/.rnzoo/config`)

```
rnzoo init
```

## Sub Command

| sub command | description |
|-------------|-------------|
| init | start rnzoo config wizard |
| ec2start | start ec2 instances (it already created, not launch) |
| ec2stop | stop ec2 instances |
| ec2list | listing ec2 instances |

## Copyright and LICENSE

Copyright (c) 2015- reiki4040

MIT LICENSE
