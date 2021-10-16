# Git Pull Command using Golang

This is my project built using golang for pulling git repository recursively. 
I got idea for doing this project for completting the gitlabber minor function that has not been doing 
and the project going down or no updates anymore. 

## How to use it

Below are the command parameter that can be used to execute the programs

| Flag          | Default Value | Possible Value | Description |
|---------------|---------------|----------------|-------------|
| -h            | -  | - | this can be used to show all of the option available  |
| -c, -action   | - | `update`, `update-gitlab` | this flag are giving information about what action it takes |
| -u, -url      | `https://gitlab.com/` | Ex: `http://172.20.3.50/` | Set the default url of the repository. This flag is mandatory for `update-gitlab` action for definning your repo (for example if you are using your own local gitlab repo like at my place) |
| -U, -username | - | Ex: kevin | Set the username for authentication |
| -P, -password | - | your password | Set the password for authentication |
| -t, -token    | - | your token | Set your private token for authentication. If this field not empty than you don't have to define username and password |
| -path         | `.` | `/path/to/dir` | Set root path for the action performed. Default value is current directory |
| -verbose      | `false` | `true`, `false` | Set the output of the program. If it's being set then all the log information would be printed. Default is false |
| -hard-reset   | `false` | `true`, `false` | Set the need of doing hard reset while updating repo. Becarefull when setting this to `true` because it same as setting *--hard* while executing git reset |

## Example 

Example for updating one of your repo folder

```
go-git-puller -c=update -verbose=true -path=D:/Developer/git/workplace -U=user -P=pass
```

Example for update using gitlab with token

```
go-git-puller -c=update-gitlab -verbose=true -path=D:/path/gitlab -u=http://172.20.5.20/ -t=5BevGkY-asdf
```
