
# Environment Variables 
This document provides you with in-depth Details on Environment Variables in Your `tfvenv` Environment

The `tfvenv` environment enables you to manage Terraform and Terragrunt workflows by allowing you to configure environment variables that are dynamically injected into your configuration files. These environment variables help define things like provider credentials, environment-specific configurations, and region settings. This approach makes your infrastructure management consistent and environment-agnostic.
1. [Environment Variables in `tfvenv`](#environment-variables-in-tfvenv)
2. [Understanding Custom Environment Variables in `tfvenv`](#understanding-custom-environment-variables-in-tfvenv)
    1. [Common Use Cases in `tfvenv`](#common-use-cases-in-tfvenv)
3. [Example Custom Environment Variables in `tfvenv`](#example-custom-environment-variables-in-tfvenv)
    1. [TF_VAR_environment](#tf_var_environment)
    2. [TF_VAR_region](#tf_var_region)
4. [Using Environment Variables with Terraform in `tfvenv`](#using-environment-variables-with-terraform-in-tfvenv)
    1. [Example 1: Declaring Variables in Terraform Configuration](#example-1-declaring-variables-in-terraform-configuration)
    2. [Example 2: Passing Environment Variables to Terraform via `tfvenv`](#example-2-passing-environment-variables-to-terraform-via-tfvenv)
5. [Using Environment Variables with Terragrunt in `tfvenv`](#using-environment-variables-with-terragrunt-in-tfvenv)
    1. [Example 1: Using Custom Variables in `terragrunt.hcl`](#example-1-using-custom-variables-in-terragrunt-hcl)
6. [Deploying an Apache Server in `tfvenv`](#deploying-an-apache-server-in-tfvenv)
    1. [Variables in `.tfenvrc` for Apache Server Deployment](#variables-in-tfenvrc-for-apache-server-deployment)
    2. [Using `terragrunt.hcl` for Apache Server](#using-terragrunt-hcl-for-apache-server)
    3. [Running the Deployment](#running-the-deployment)
7. [Passing Environment Variables Inline in `tfvenv`](#passing-environment-variables-inline-in-tfvenv)
    1. [Inline Environment Variables with Terraform](#inline-environment-variables-with-terraform-in-tfvenv)
    2. [Inline Environment Variables with Terragrunt](#inline-environment-variables-with-terragrunt-in-tfvenv)
8. [Best Practices for Managing Environment Variables in `tfvenv`](#best-practices-for-managing-environment-variables-in-tfvenv)
    1. [Use Descriptive Names](#use-descriptive-names)
    2. [Secure Secrets](#secure-secrets)
    3. [Organize Variables by Environment](#organize-variables-by-environment)
    4. [Version Control Considerations](#version-control-considerations)
9. [Custom Environment Variables for Terraform in `tfvenv`](#custom-environment-variables-for-terraform-in-tfvenv)
    1. [Example of Passing Variables to Terraform via `tfvenv`](#example-of-passing-variables-to-terraform-via-tfvenv)
    2. [Matching Variables in Terraform Configuration](#matching-variables-in-terraform-configuration)
10. [Custom Environment Variables for Terragrunt in `tfvenv`](#custom-environment-variables-for-terragrunt-in-tfvenv)
    1. [Using Custom Variables in Terragrunt](#example-of-using-custom-variables-in-terragrunt)
11. [Examples of `tfvenvrc` Files](#examples-of-tfvenvrc-files)
    1. [General `tfvenvrc` File](#example-generic-tfvenvrc-file)
    2. [Apache Server-Specific `tfvenvrc` File](#example-tfvenvrc-file-specifically-for-the-examples-above)
12. [Securely Managing Environment Variables](#securely-managing-environment-variables)
    1. [1. HashiCorp Vault](#hashicorp-vault)
    2. [2. AWS Secrets Manager](#aws-secrets-manager)
    3. [3. Google Cloud Secret Manager](#google-cloud-secret-manager)

Below is a breakdown of how custom environment variables can be used effectively within your `tfvenv` environment, especially for Terraform and Terragrunt.

## Understanding Custom Environment Variables in `tfvenv`

In `tfvenv`, environment variables can be set globally in the `.tfenvrc` configuration file, passed inline during command execution, or even pulled from secret management tools like AWS Secrets Manager. These environment variables dynamically inject values into your Terraform or Terragrunt configurations.

### Common Use Cases in `tfvenv`:
- **Dynamic Configuration:** Injecting environment-specific settings, such as region, instance types, or project identifiers.
- **Secret Management:** Passing sensitive credentials like API keys or SSH keys as environment variables (via services like AWS Secrets Manager, Vault, etc.).
- **CI/CD Integrations:** Simplify your CI/CD pipelines by setting variables in `tfvenv` and ensuring they are consistently applied across environments.

## Example Custom Environment Variables in `tfvenv`

In `tfvenv`, custom environment variables are frequently used to pass values into Terraform and Terragrunt configurations. These variables are often prefixed with `TF_VAR_` for Terraform or can be accessed via the `getenv()` function in Terragrunt.

- **TF_VAR_environment**: Represents the environment type (e.g., `development`, `staging`, or `production`).
  - Example: `TF_VAR_environment=development`
  - Usage: Used to set up an environment-specific configuration in your Terraform code.

- **TF_VAR_region**: Dynamically specifies the region (e.g., `us-west-1`, `us-east-1`).
  - Example: `TF_VAR_region=us-west-1`
  - Usage: Defines the AWS region for resource provisioning.

## Using Environment Variables with Terraform in `tfvenv`

`tfvenv` provides seamless integration with Terraform by managing environment variables that Terraform will automatically map to configuration variables. These variables can be declared in your `.tfenvrc` file and passed during `terraform apply`.

### Example 1: Declaring Variables in Terraform Configuration

```hcl
# variables.tf
variable "environment" {
  description = "The environment where resources are deployed"
  type        = string
}

variable "region" {
  description = "The AWS region where resources are deployed"
  type        = string
}

# main.tf
provider "aws" {
  region = var.region
}

resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  tags = {
    Name = "Web Server - ${var.environment}"
  }
}
```

### Example 2: Passing Environment Variables to Terraform via `tfvenv`

In `tfvenv`, environment variables can be passed in via the `.tfenvrc` file or dynamically set before running the command.

```bash
# Set the environment variables in your shell
export TF_VAR_environment=development
export TF_VAR_region=us-west-1

# Run terraform within the tfvenv environment
terraform apply
```

With `tfvenv`, Terraform automatically maps the `TF_VAR_environment` and `TF_VAR_region` to their respective variables in the configuration.

---

## Using Environment Variables with Terragrunt in `tfvenv`

Terragrunt works seamlessly with `tfvenv` to pass environment variables, using the `getenv()` function to dynamically read values defined in your environment.

### Example 1: Using Custom Variables in `terragrunt.hcl`

```hcl
# terragrunt.hcl
terraform {
  source = "./path/to/module"
}

inputs = {
  environment = getenv("TF_VAR_environment")
  region      = getenv("TF_VAR_region")
}
```

In this case, the `tfvenv` environment is automatically sourcing the environment variables and passing them into Terragrunt for use in Terraform configurations.

---

## Deploying an Apache Server in `tfvenv`

Here’s how `tfvenv` simplifies deploying infrastructure, such as an Apache server, by managing environment variables dynamically.

1. **Variables in `.tfenvrc` for Apache Server Deployment**

```hcl
# variables.tf
variable "region" {
  type        = string
  description = "AWS Region"
}

variable "instance_type" {
  type        = string
  description = "EC2 Instance Type"
  default     = "t2.micro"
}

variable "environment" {
  type        = string
  description = "Environment type (dev, staging, production)"
}

variable "ami_id" {
  type        = string
  description = "AMI ID for Linux distribution"
}

# tfvenvrc.sample
TF_VAR_environment=production
TF_VAR_region=us-west-2
TF_VAR_ami_id=ami-0c55b159cbfafe1f0
```

---
2. **Using `terragrunt.hcl` for Apache Server**

```hcl
# terragrunt.hcl
terraform {
  source = "./path/to/module"
}

inputs = {
  region       = getenv("TF_VAR_region")
  environment  = getenv("TF_VAR_environment")
  ami_id       = "ami-0c55b159cbfafe1f0"  # AMI for Linux
}
```

### 3. Running the Deployment

```bash
# Set your environment variables
export TF_VAR_environment=production
export TF_VAR_region=us-west-2

# Run terragrunt within the tfvenv environment
terragrunt apply
```

This configuration will provision an EC2 instance in the `us-west-2` region, tagged with the name `Apache Server - production`, and set up an Apache server automatically via the `user_data` script. Using `tfvenv` makes it easy to switch environments without needing to modify your underlying infrastructure code.

---

## Passing Environment Variables Inline in `tfvenv`

In `tfvenv`, you can also pass environment variables directly via inline shell commands. This can be useful for one-off executions or in CI/CD pipelines where values are dynamically set at runtime.

### Example: Inline Environment Variables with Terraform in `tfvenv`

```bash
TF_VAR_environment=production TF_VAR_region=us-west-2 terraform apply
```

### Example: Inline Environment Variables with Terragrunt in `tfvenv`

```bash
TF_VAR_environment=staging TF_VAR_region=us-east-1 terragrunt apply
```

With `tfvenv`, these inline variables are immediately recognized and used in the respective Terraform or Terragrunt configuration, making deployments flexible and adaptable to different environments.

---

## Best Practices for Managing Environment Variables in `tfvenv`

To ensure consistent and secure infrastructure management, follow these best practices when using environment variables in your `tfvenv` setup:

1. **Use Descriptive Names:** Always use clear and descriptive names for environment variables. For example, instead of using `ENV_VARS_VAR1`, use `TF_VAR_environment` or `TF_VAR_region` to make the purpose of the variables explicit.

2. **Secure Secrets:** Avoid hardcoding sensitive information (e.g., API keys or passwords) in your `.tfenvrc` or other files. Instead, use secret management tools such as AWS Secrets Manager, HashiCorp Vault, or Google Cloud Secret Manager, and inject those values securely into your `tfvenv` environment.

3. **Organize Variables by Environment:** In a multi-environment setup (e.g., dev, staging, production), organize your variables accordingly. Use naming conventions like `TF_VAR_environment=production` to make it clear which environment the variables are associated with.

4. **Version Control Considerations:** When using `.tfenvrc` files or `.env` files to store environment variables locally, ensure they are excluded from version control (e.g., via `.gitignore`). This helps prevent accidental exposure of sensitive data.

---

## Custom Environment Variables for Terraform in `tfvenv`

To pass custom variables into Terraform, `tfvenv` uses the `TF_VAR_` prefix to automatically map environment variables to input variables defined in your `variables.tf` file.

### Example of Passing Variables to Terraform via `tfvenv`

```bash
export TF_VAR_environment=production
export TF_VAR_region=us-west-1
terraform apply
```

In this case, `tfvenv` ensures that the `TF_VAR_environment` and `TF_VAR_region` variables are automatically passed to Terraform, allowing them to be used in your configuration without needing explicit command-line arguments.

---

### Matching Variables in Terraform Configuration

```hcl
# variables.tf
variable "environment" {
  description = "The environment (production, staging, dev)"
  type        = string
}

variable "region" {
  description = "The AWS region"
  type        = string
}
```

By using `tfvenv`, Terraform automatically maps environment variables prefixed with `TF_VAR_` to the corresponding variables in your configuration.

---

## Custom Environment Variables for Terragrunt in `tfvenv`

In Terragrunt, environment variables can be accessed using the `getenv()` function. `tfvenv` makes it simple to manage these environment variables dynamically.

### Example of Using Custom Variables in Terragrunt

1. Define environment variables in `tfvenv`:

```bash
export MY_ENVIRONMENT=production
export MY_REGION=us-west-1
terragrunt apply
```

2. Reference the environment variables in your `terragrunt.hcl`:

```hcl
# terragrunt.hcl
inputs = {
  environment = getenv("MY_ENVIRONMENT")
  region      = getenv("MY_REGION")
}
```

In this example, `tfvenv` handles the passing of environment variables (`MY_ENVIRONMENT` and `MY_REGION`), and they are used within the Terragrunt configuration to inject the correct values dynamically.

---

## Wrapping up: Leveraging `tfvenv` for Environment Variables

By using `tfvenv`, you can simplify and secure the process of managing environment variables in your Terraform and Terragrunt workflows. Whether you’re managing multiple environments, securely passing credentials, or integrating with CI/CD pipelines, `tfvenv` ensures that the correct variables are passed automatically, making your infrastructure management consistent and scalable.

Remember to use the `TF_VAR_` prefix for Terraform variables and the `getenv()` function for accessing environment variables in Terragrunt. This approach will help you streamline your infrastructure deployments while maintaining flexibility and security.


## Examples

Example generic tfvenvrc file:  
  
```
# tfvenvrc.sample

# This file contains environment variables used by tfvenv to manage Terraform and Terragrunt environments.
# Each environment variable is automatically passed to Terraform or Terragrunt when tfvenv is active.

# Example of specifying the Terraform version to use
TF_VERSION=1.5.0

# Example of specifying the Terragrunt version to use
TG_VERSION=0.35.16

# AWS S3 Backend Configuration for Terraform state storage
# Define the S3 bucket where Terraform states are stored
S3_STATE_BUCKET=my-terraform-state

# Define the path inside the S3 bucket for state files
S3_STATE_PATH=states

# Define the S3 bucket where Terraform modules are stored
S3_MODULES_BUCKET=my-terraform-modules

# Define the path inside the S3 bucket for Terraform modules
S3_MODULES_PATH=modules

# The AWS region where the S3 buckets are located
REGION=us-west-2

# AWS Access Keys (optional if using IAM roles)
# Store your access and secret keys here, or better yet, use secret management for sensitive data
ACCESS_KEY=your-access-key
SECRET_KEY=your-secret-key

# Custom environment variables for use within Terraform or Terragrunt
# Example custom variable for environment (e.g., development, production)
TF_VAR_environment=development

# Example custom variable for region (overrides REGION if defined separately)
TF_VAR_region=us-west-1

# Optional: Additional environment variables for other services or tools
CUSTOM_VAR1=value1
CUSTOM_VAR2=value2
```

Example tfvenvrc file specifically for the examples above:  
```  
# tfvenvrc.sample - Linux/Apache Example

# This file is used to manage environment variables for a Terraform/Terragrunt deployment
# involving an AWS EC2 instance running Linux and Apache web server.

# Set the Terraform version to use
TF_VERSION=1.5.0

# Set the Terragrunt version to use
TG_VERSION=0.35.16

# AWS S3 Backend Configuration
S3_STATE_BUCKET=my-terraform-state
S3_STATE_PATH=states
S3_MODULES_BUCKET=my-terraform-modules
S3_MODULES_PATH=modules
REGION=us-west-2

# AWS Credentials (optional, you can also use IAM roles)
ACCESS_KEY=your-access-key
SECRET_KEY=your-secret-key

# Custom environment variables used in this deployment

# Define the environment (e.g., dev, staging, production)
TF_VAR_environment=production

# AWS region to deploy to
TF_VAR_region=us-west-2

# AMI ID for Linux instance
TF_VAR_ami_id=ami-0c55b159cbfafe1f0

# EC2 instance type for the Apache server
TF_VAR_instance_type=t2.micro

# Any additional custom variables
CUSTOM_VAR1=value1
```

## Securely Managing Environment Variables

Managing sensitive environment variables (such as access keys, secrets, and credentials) should always be handled securely. Below are examples of securely managing environment variables using HashiCorp Vault, AWS Secrets Manager, and Google Cloud Secret Manager.

## 1. HashiCorp Vault

HashiCorp Vault provides a secure way to store and access secrets such as API keys, passwords, and certificates.

### Example Usage:

- **Set Up Vault**:
    Install and configure HashiCorp Vault for your organization, ensuring access to your Kubernetes or AWS infrastructure.
    
    Example steps:
    ```bash
    vault operator init -key-shares=5 -key-threshold=3
    vault operator unseal <unseal_key_1>
    vault operator unseal <unseal_key_2>
    vault operator unseal <unseal_key_3>
    ```

- **Enable Secrets Engine**:
    Vault provides a key-value store for secrets.
    ```bash
    vault secrets enable -path=apps kv-v2
    vault kv put apps/test-app AWS_SECRET_ACCESS_KEY='your-access-key' AWS_ACCESS_KEY_ID='your-secret-key'
    ```

- **Inject Secrets into Applications**:
    Vault allows injection of secrets via annotations:
    ```yaml
    annotations:
      vault.hashicorp.com/agent-inject: 'true'
      vault.hashicorp.com/agent-inject-secret-AWS_SECRET_ACCESS_KEY: 'apps/data/test-app'
    ```

### Benefits:
- Centralized secrets management.
- Rotates and revokes secrets easily.
- Encrypted at rest and in transit.

---

## 2. AWS Secrets Manager

AWS Secrets Manager allows you to store and manage access to your secrets (API keys, passwords, etc.) securely.

### Example Usage:

- **Create a Secret**:
    Using the AWS CLI, you can create a secret:
    ```bash
    aws secretsmanager create-secret --name MyTestSecret --secret-string "{"username":"admin","password":"your-password"}"
    ```

- **Retrieve a Secret**:
    To retrieve your secret using AWS CLI:
    ```bash
    aws secretsmanager get-secret-value --secret-id MyTestSecret
    ```

### Benefits:
- Integrated with AWS IAM for access control.
- Supports automatic rotation of secrets.
- Securely store and retrieve secrets across AWS services.

---

## 3. Google Cloud Secret Manager

Google Cloud Secret Manager allows you to store, manage, and access secrets such as API keys, passwords, and certificates.

### Example Usage:

- **Create a Secret**:
    Using the `gcloud` CLI:
    ```bash
    gcloud secrets create my-secret --replication-policy="automatic"
    ```

- **Add a Secret Version**:
    You can add a secret version (the actual secret value):
    ```bash
    echo -n "my-secret-value" | gcloud secrets versions add my-secret --data-file=-
    ```

- **Access a Secret**:
    Access the secret value using `gcloud`:
    ```bash
    gcloud secrets versions access latest --secret=my-secret
    ```

### Benefits:
- Secure, centralized management of secrets.
- Integrated with Google Cloud IAM for access control.
- Automatic versioning and replication of secrets.

---