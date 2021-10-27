# Git Pull Command using Golang

This is my project built using golang for pulling git repository recursively. 
This project was inspired by [gitlabber project](https://github.com/ezbz/gitlabber). One of the purpose of making this because we need to install python to use it.
So we are dependent on python and it's version, some of our teams can't install it properly because of it's version. Because of that we are proposing to make it on other way that we don't have to depend on something and can build it on final executeable program.

## Arguments Information

Below are the command parameter that can be used to execute the programs

| Flag          | Default Value | Possible Value | Description | Mandatory |
|---------------|---------------|----------------|-------------|-----------|
| `-h`            | -  | - | this can be used to show all of available option | No |
| `-c`, `-action`   | - | `update`, `update-gitlab` | this flag are giving information about what action being executed | Yes |
| `-u`, `-url`      | `https://gitlab.com/` | Ex: `http://172.20.3.50/` | Set the default url of the repository. This flag is mandatory for `update-gitlab` action for definning your repo (for example if you are using your own local gitlab repo like at my place) | Optional |
| `-U`, `-username` | - | Ex: kevin | Set the username for authentication | Yes |
| `-P`, `-password` | - | your password | Set the password for authentication | Yes |
| `-t`, `-token`    | - | your token | Set your private token for authentication. If this field's not empty than you don't have to define username and password | Yes |
| `-path`         | `.` | `/path/to/dir` | Set root path for action performed. Default value is current directory | No |
| `-verbose`      | `false` | `true`, `false` | Set program output. If it's being set then all the log information would be printed. Default is false | No |
| `-hard-reset`   | `false` | `true`, `false` | Tell the program wheter a hard reset is required when updating repository. Becarefull when setting it to `true` because it's same as putting *--hard* while exec git reset | No |

## Dependency Used on this project 

Here list of dependency was used to make this project:

- [Go-Git](https://github.com/go-git/go-git)
- [Go-Gitlab](https://github.com/xanzy/go-gitlab)
- [Zap Logger](https://github.com/uber-go/zap)
- [Progress Bar](https://github.com/schollz/progressbar)

## Example 

Example for updating one of your repo folder

```
go-git-puller -c=update -verbose=true -path=D:/Developer/git/workplace -U=user -P=pass
```

Example for update using gitlab with token

```
go-git-puller -c=update-gitlab -verbose=true -path=D:/path/gitlab -u=http://172.20.5.20/ -t=5BevGkY-asdf
```
