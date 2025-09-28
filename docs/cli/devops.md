# devops CLI Documentation

**Version:** 0.0.3

**Description:** DevOps: Simplifying your CI/CD pipelines.

## Usage

```bash
devops [command] [flags] [arguments]
```

## Global Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --file | -f | string | Path to the project definition file |
| --verbose | -v | count | Increase verbosity (-v or -vv) |

## Commands

  ### build

  **Description:** Run the build operations

  Build the project according to the configuration..

  **Usage:**
  ```bash
  devops build
  ```


  ### completion

  **Description:** Generate the autocompletion script for the specified shell

  Generate the autocompletion script for devops for the specified shell.
See each sub-command's help for details on how to use the generated script.


  **Usage:**
  ```bash
  devops completion
  ```


    ### bash

    **Description:** Generate the autocompletion script for bash

    Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(devops completion bash)

To load completions for every new session, execute once:

#### Linux:

	devops completion bash > /etc/bash_completion.d/devops

#### macOS:

	devops completion bash > $(brew --prefix)/etc/bash_completion.d/devops

You will need to start a new shell for this setup to take effect.


    **Usage:**
    ```bash
    devops completion bash
    ```

    **Flags:**

    | Flag | Short | Type | Description |
    |------|-------|------|-------------|
    | --no-descriptions |  | bool | disable completion descriptions |


    ### fish

    **Description:** Generate the autocompletion script for fish

    Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	devops completion fish | source

To load completions for every new session, execute once:

	devops completion fish > ~/.config/fish/completions/devops.fish

You will need to start a new shell for this setup to take effect.


    **Usage:**
    ```bash
    devops completion fish [flags]
    ```

    **Flags:**

    | Flag | Short | Type | Description |
    |------|-------|------|-------------|
    | --no-descriptions |  | bool | disable completion descriptions |


    ### powershell

    **Description:** Generate the autocompletion script for powershell

    Generate the autocompletion script for powershell.

To load completions in your current shell session:

	devops completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


    **Usage:**
    ```bash
    devops completion powershell [flags]
    ```

    **Flags:**

    | Flag | Short | Type | Description |
    |------|-------|------|-------------|
    | --no-descriptions |  | bool | disable completion descriptions |


    ### zsh

    **Description:** Generate the autocompletion script for zsh

    Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(devops completion zsh)

To load completions for every new session, execute once:

#### Linux:

	devops completion zsh > "${fpath[1]}/_devops"

#### macOS:

	devops completion zsh > $(brew --prefix)/share/zsh/site-functions/_devops

You will need to start a new shell for this setup to take effect.


    **Usage:**
    ```bash
    devops completion zsh [flags]
    ```

    **Flags:**

    | Flag | Short | Type | Description |
    |------|-------|------|-------------|
    | --no-descriptions |  | bool | disable completion descriptions |


  ### doctor

  **Description:** Validate your configuration

  Run checks on your configuration file to ensure it is ready for use.

  **Usage:**
  ```bash
  devops doctor
  ```


  ### help

  **Description:** Help about any command

  Help provides help for any command in the application.
Simply type devops help [path to command] for full details.

  **Usage:**
  ```bash
  devops help [command]
  ```


  ### test

  **Description:** Run the test operations

  Run the designated test operations.

  **Usage:**
  ```bash
  devops test
  ```


