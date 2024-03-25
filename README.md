# ecsexec
`ecsexec` is a tool designed to access a shell session within a container running in an ECS task.

## Installation
You can install this tool by running the following command:

```bash
go get -u github.com/your_username/ecsexec/cmd/ecsexec
```

## Usage
The `ecsexec` command can be used with the following syntax:

```bash
ecsexec [global options] command [command options]
```

## Global Options
- --region value: Specifies the AWS region name. Default is $AWS_REGION.
- --profile value: Specifies the AWS profile name. Default is $AWS_PROFILE.
- --cluster value: Specifies the ECS cluster name. Default is $ECSEXEC_CLUSTER.
- --service value: Specifies the ECS service name. Default is $ECSEXEC_SERVICE.
- --container value: Specifies the container name. Default is $ECSEXEC_CONTAINER.
- --command value: Specifies the login shell. Default is /bin/sh, and it can be overridden by $ESCEXEC_COMMAND.
- --help, -h: Displays the help message.

## Examples
### Simple Usage Example

```bash
ecsexec --region us-east-1 --cluster my-cluster --service my-service --container my-container
```

### Specifying a Login Shell Example
```bash
ecsexec --region us-east-1 --cluster my-cluster --service my-service --container my-container --command /bin/bash
```

## Configuration
You can customize the behavior of the ecsexec command using environment variables. Here are the available environment variables:

- `AWS_REGION`: AWS region name
- `AWS_PROFILE`: AWS profile name
- `ECSEXEC_CLUSTER`: ECS cluster name
- `ECSEXEC_SERVICE`: ECS service name
- `ECSEXEC_CONTAINER`: Container name
- `ESCEXEC_COMMAND`: Login shell
