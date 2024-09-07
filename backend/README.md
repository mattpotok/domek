# Domek/Backend

The back of the house where the all the magic happens.

## Requirements

Domek/Backend depends on two resources
1. an `.env` file for environment variables
2. an `.aws` directory for AWS credentials

### Environment variables

Create an `.env` file with the following contents

```
REGION="<region>"
SNS_TOPIC_ARN="<sns_topic_arn>"
```

### AWS credentials

Create an `.aws` directory with `config` and `credentials` files

The `config` file
```
[profile domek]
region = <region>
```

The `credentials` file
```
[domek]
aws_access_key_id = <aws_access_key_id>
aws_secret_access_key = <aws_secret_access_key>
```