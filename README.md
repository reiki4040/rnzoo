rnzoo
====

[![GitHub release](http://img.shields.io/github/release/reiki4040/rnzoo.svg?style=flat-square)][release]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[release]: https://github.com/reiki4040/rnzoo/releases
[license]: https://github.com/reiki4040/rnzoo/blob/master/LICENSE

rnzoo is useful cli to use ec2.

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
| ec2run, run | run new ec2 instances |
| ec2list, ls | listing ec2 instances |
| ec2start, start | start ec2 instances (it already created, not launch) |
| ec2stop, stop | stop ec2 instances |
| ec2type, type | modify ec2 instance type |
| attach-eip | allocate new EIP(allow reassociate) and associate it to the instance |
| detach-eip | disassociate EIP and release it |

## Copyright and LICENSE

Copyright (c) 2015- reiki4040

MIT LICENSE
