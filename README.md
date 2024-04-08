# ecsexec
`ecsexec` is a tool designed to access a shell session within a container running in an ECS task.

## Prerequisites
Before using ecsexec, ensure that you have the following prerequisites installed on your local machine:

1. **AWS CLI**: The AWS Command Line Interface (CLI) is required for managing AWS resources from the command line.
Installation instructions for the AWS CLI can be found in the AWS documentation: [Installing the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html).

2. **Session Manager Plugin** (`session-manager-plugin`):
Installation instructions for session-manager-plugin can be found in the AWS documentation: [Session Manager Plugin Installation](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).

## Installation
You can install this tool by running the following command:

### Homebrew
```bash
brew install Taiki130/ecsexec/ecsexec
```

### go
```bash
go install -a github.com/Taiki130/ecsexec/cmd/ecsexec@latest
```

## Usage
The `ecsexec` command can be used with the following syntax:

```bash
ecsexec [global options]
```

## Interactive Prompt
If any of the required information is not provided through environment variables or command-line options, `ecsexec` will prompt you interactively to select the necessary details.

## Global Options
- `--region value`: Specifies the AWS region name. Default is `$AWS_REGION`.
- `--profile value`: Specifies the AWS profile name. Default is `$AWS_PROFILE`.
- `--cluster value`: Specifies the ECS cluster name. Default is `$ECSEXEC_CLUSTER`.
- `--service value`: Specifies the ECS service name. Default is `$ECSEXEC_SERVICE`.
- `--container value`: Specifies the container name. Default is `$ECSEXEC_CONTAINER`.
- `--command value`: Specifies the login shell. Default is /bin/sh, and it can be overridden by `$ESCEXEC_COMMAND`.
- `--help, -h`: Displays the help message.

## Examples
### Simple Usage Example

```bash
ecsexec
```

### Specifying flags Example
```bash
ecsexec --region us-east-1 --cluster my-cluster --service my-service --container my-container
```

### Specifying a Login Shell Example
```bash
ecsexec --command /bin/bash
```

## Configuration
You can customize the behavior of the ecsexec command using environment variables. Here are the available environment variables:

- `AWS_REGION`: AWS region name
- `AWS_PROFILE`: AWS profile name
- `ECSEXEC_CLUSTER`: ECS cluster name
- `ECSEXEC_SERVICE`: ECS service name
- `ECSEXEC_CONTAINER`: Container name
- `ESCEXEC_COMMAND`: Login shell

## License
This project is licensed under the [MIT License](https://github.com/Taiki130/ecsexec?tab=MIT-1-ov-file#readme). See the LICENSE file for details.
