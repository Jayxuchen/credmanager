# Dynamic RDS Credential Example

This example demonstrates how to use the credential manager with AWS RDS IAM authentication tokens, suitable for EKS services with IAM roles.

## Prerequisites

1. **AWS IAM Role**: Your EKS service account must have an IAM role with the following permission:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "rds-db:connect"
         ],
         "Resource": [
           "arn:aws:rds-db:REGION:ACCOUNT:dbuser:DB_INSTANCE_ID/DB_USERNAME"
         ]
       }
     ]
   }
   ```

2. **RDS Instance**: Your RDS instance must have IAM database authentication enabled.

## Environment Variables

Set the following environment variables:

```bash
export DB_HOST="your-rds-instance.cluster-abc123.us-west-2.rds.amazonaws.com"
export DB_PORT="5432"
export DB_USER="your_iam_user"
export DB_NAME="your_database"
export AWS_REGION="us-west-2"
```

## Running the Example

```bash
# Install dependencies
go mod tidy

# Run the example
go run rds.go
```

## How It Works

1. **Token Generation**: Uses `rdsutils.BuildAuthToken()` to generate a presigned URL that serves as the password
2. **IAM Authentication**: The token is valid for 15 minutes and uses the IAM role attached to your EKS service
3. **Fallback**: If dynamic credential generation fails, it falls back to static credentials
4. **Automatic Refresh**: Each call to `GetCredentials()` generates a fresh token

## Expected Output

```
=== Dynamic RDS Credential Manager Example ===

This example demonstrates AWS RDS IAM authentication token generation
Suitable for EKS services with IAM roles attached

--- Attempt 1 ---
✓ Source: Dynamic RDS (IAM Auth Token)
  Credential Key: dynamic_rds_postgres
  Token Length: 456 characters
  Expires: 2025-08-01T23:23:47Z (in 14m59s)
  Metadata:
    host: your-rds-instance.cluster-abc123.us-west-2.rds.amazonaws.com
    port: 5432
    username: your_iam_user
    database: your_database
    type: dynamic
    auth_method: iam_token
    region: us-west-2
  Connection: postgres://your_iam_user@your-host:5432/your_db?authmethod=iam&token=your-token...
```

## Production Usage

In a production EKS environment:

1. Set up service account with IAM role annotations
2. Configure the RDS instance with IAM authentication
3. Use environment variables or AWS Parameter Store for configuration
4. Implement proper error handling and retry logic
5. Consider token caching to avoid unnecessary API calls
