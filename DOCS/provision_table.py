# Before running the code below, please follow these steps to setup your workspace if you have not
# set it up already:
#
# 1. Setup credentials for DynamoDB access. One of the ways to setup credentials is to add them to
#    ~/.aws/credentials file (C:\Users\USER_NAME\.aws\credentials file for Windows users) in
#    following format:
#
#    [<profile_name>]
#    aws_access_key_id = YOUR_ACCESS_KEY_ID
#    aws_secret_access_key = YOUR_SECRET_ACCESS_KEY
#
#    If <profile_name> is specified as "default" then AWS SDKs and CLI will be able to read the credentials
#    without any additional configuration. But if a different profile name is used then it needs to be
#    specified while initializing DynamoDB client via AWS SDKs or while configuring AWS CLI.

#    Please refer following guide for more details on credential configuration:
#    https://boto3.amazonaws.com/v1/documentation/api/latest/guide/quickstart.html#configuration
#
# 2. Install the latest Boto 3 release via pip:
#
#    pip install boto3
#
#    Please refer following guide for more details on Boto 3 installation:
#    https://boto3.amazonaws.com/v1/documentation/api/latest/guide/quickstart.html#installation
#    Please note that you may need to follow additional setup steps for using Boto 3 from an IDE. Refer
#    your IDE's documentation if you run into issues.


# Load the AWS SDK for Python
import boto3
from botocore.exceptions import ClientError

ERROR_HELP_STRINGS = {
    # Operation specific errors
    'LimitExceededException': 'Number of simultaneous table operations may exceed the limit. Up to 50 simultaneous table operations are allowed per account.' +
                              'You can have up to 25 such requests running at a time; however, if the table or index specifications are complex,' +
                              'DynamoDB might temporarily reduce the number of concurrent operations. Consider retry it later',
    'ResourceInUseException': 'Table already exists, verify table does not exist before retrying',
    # Common Errors
    'InternalServerError': 'Internal Server Error, generally safe to retry with exponential back-off',
    'ProvisionedThroughputExceededException': 'Request rate is too high. If you\'re using a custom retry strategy make sure to retry with exponential back-off.' +
                                              'Otherwise consider reducing frequency of requests or increasing provisioned capacity for your table or secondary index',
    'ResourceNotFoundException': 'One of the tables was not found, verify table exists before retrying',
    'ServiceUnavailable': 'Had trouble reaching DynamoDB. generally safe to retry with exponential back-off',
    'ThrottlingException': 'Request denied due to throttling, generally safe to retry with exponential back-off',
    'UnrecognizedClientException': 'The request signature is incorrect most likely due to an invalid AWS access key ID or secret key, fix before retrying',
    'ValidationException': 'The input fails to satisfy the constraints specified by DynamoDB, fix input before retrying',
    'RequestLimitExceeded': 'Throughput exceeds the current throughput limit for your account, increase account level throughput before retrying',
}

# Use the following function instead when using DynamoDB local
#def create_dynamodb_client(region):
#    return boto3.client("dynamodb", region_name="localhost", endpoint_url="http://localhost:8000", aws_access_key_id="access_key_id", aws_secret_access_key="secret_access_key")

def create_dynamodb_client(region="us-west-2"):
    return boto3.client("dynamodb", region_name=region)


def create_table_input():
    return {
    "TableName": "krabber",
    "KeySchema": [
        {
            "AttributeName": "PK",
            "KeyType": "HASH"
        },
        {
            "AttributeName": "SK",
            "KeyType": "RANGE"
        }
    ],
    "BillingMode": "PROVISIONED",
    "AttributeDefinitions": [
        {
            "AttributeName": "PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI1PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI1SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI2PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI2SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI3PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI3SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI4PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI4SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI5PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI5SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI6PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI6SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI7PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI7SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI8PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI8SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI9PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI9SK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI10PK",
            "AttributeType": "S"
        },
        {
            "AttributeName": "GSI10SK",
            "AttributeType": "S"
        }
    ],
    "ProvisionedThroughput": {
        "ReadCapacityUnits": 1,
        "WriteCapacityUnits": 1
    },
    "GlobalSecondaryIndexes": [
        {
            "IndexName": "GSI1",
            "KeySchema": [
                {
                    "AttributeName": "GSI1PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI1SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI2",
            "KeySchema": [
                {
                    "AttributeName": "GSI2PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI2SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI3",
            "KeySchema": [
                {
                    "AttributeName": "GSI3PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI3SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI4",
            "KeySchema": [
                {
                    "AttributeName": "GSI4PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI4SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI5",
            "KeySchema": [
                {
                    "AttributeName": "GSI5PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI5SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI6",
            "KeySchema": [
                {
                    "AttributeName": "GSI6PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI6SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI7",
            "KeySchema": [
                {
                    "AttributeName": "GSI7PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI7SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI8",
            "KeySchema": [
                {
                    "AttributeName": "GSI8PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI8SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI9",
            "KeySchema": [
                {
                    "AttributeName": "GSI9PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI9SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        },
        {
            "IndexName": "GSI10",
            "KeySchema": [
                {
                    "AttributeName": "GSI10PK",
                    "KeyType": "HASH"
                },
                {
                    "AttributeName": "GSI10SK",
                    "KeyType": "RANGE"
                }
            ],
            "Projection": {
                "ProjectionType": "ALL"
            },
            "ProvisionedThroughput": {
                "ReadCapacityUnits": 1,
                "WriteCapacityUnits": 1
            }
        }
    ]
}


def execute_create_table(dynamodb_client, input):
    try:
        response = dynamodb_client.create_table(**input)
        print("Successfully created table.")
        # Handle response
    except ClientError as error:
        handle_error(error)
    except BaseException as error:
        print("Unknown error while creating table: " + error.response['Error']['Message'])


def handle_error(error):
    error_code = error.response['Error']['Code']
    error_message = error.response['Error']['Message']

    error_help_string = ERROR_HELP_STRINGS[error_code]

    print('[{error_code}] {help_string}. Error message: {error_message}'
          .format(error_code=error_code,
                  help_string=error_help_string,
                  error_message=error_message))


def main():
    # Create the DynamoDB Client with the region you want
    dynamodb_client = create_dynamodb_client()

    # Create the dictionary containing arguments for create_table call
    create_table_request = create_table_input()

    # Call DynamoDB's create_table API
    execute_create_table(dynamodb_client, create_table_request)


if __name__ == "__main__":
    main()

# ===========================================================================
# Autoscaling Section
# ===========================================================================

ERROR_HELP_STRINGS_AUTOSCALING = {
    # Operation specific errors
    'ConcurrentUpdateException': 'There is already a pending update to an Auto Scaling resource for this table.',
    'FailedResourceAccessException': 'The operation could not be completed due to not having access to the resource due to permission restrictions.',
    'ObjectNotFoundException': 'Object not found. The operation could not be completed because the resource was not found.',

    # Common Errors
    'InternalServerError': 'Internal Server Error, generally safe to retry with exponential back-off',
    'ServiceUnavailable': 'Had trouble reaching DynamoDB. generally safe to retry with exponential back-off',
    'ThrottlingException': 'Request denied due to throttling, generally safe to retry with exponential back-off',
    'ValidationException': 'The input fails to satisfy the constraints specified by DynamoDB, fix input before retrying',
    'RequestLimitExceeded': 'Throughput exceeds the current throughput limit for your account, increase account level throughput before retrying',
}

def handle_error_autoscaling(error):
    error_code = error.response['Error']['Code']
    error_message = error.response['Error']['Message']

    error_help_string = ERROR_HELP_STRINGS_AUTOSCALING[error_code]

    print('[{error_code}] {help_string}. Error message: {error_message}'
          .format(error_code=error_code,
                  help_string=error_help_string,
                  error_message=error_message))

def createAutoScalingClient(region="us-west-2"):
  return boto3.client('autoscaling', region_name=region)

# Call this function if you want to apply autoscaling to your DynamoDB table
def applyAutoScalingMain():
  autoScalingClient = createAutoScalingClient()

def put_scaling_policy(dynamodb_scaling, table_name: str, target_value: int, isRead: bool, index_name: str = None):
    resourceId = "table/" + table_name + "/index/" + index_name if index_name else "table/" + table_name
    indexOrTable = "index" if index_name else "table"
    scalableDimension = "dynamodb:" + indexOrTable + ":ReadCapacityUnits" if isRead else "dynamodb:" + indexOrTable + ":WriteCapacityUnits"
    policyName = "ScaleDynamoDBReadCapacityUtilization" if isRead else "ScaleDynamoDBWriteCapacityUtilization"
    metricType = "DynamoDBReadCapacityUtilization" if isRead else "DynamoDBWriteCapacityUtilization"

    dynamodb_scaling.put_scaling_policy(ServiceNamespace='dynamodb',
                                        ResourceId=resourceId,
                                        PolicyType='TargetTrackingScaling',
                                        PolicyName=policyName,
                                        ScalableDimension=scalableDimension,
                                        TargetTrackingScalingPolicyConfiguration={
                                          'TargetValue': target_value,
                                          'PredefinedMetricSpecification': {
                                            'PredefinedMetricType': metricType
                                          },
                                          'ScaleOutCooldown': 60,
                                          'ScaleInCooldown': 60
                                        })

def register_scalable_target(dynamodb_scaling, table_name: str, min_capacity: int, max_capacity: int, isRead: bool, index_name: str = None):
    resourceId = "table/" + table_name + "/index/" + index_name if index_name else "table/" + table_name
    indexOrTable = "index" if index_name else "table"
    scalableDimension = "dynamodb:" + indexOrTable + ":ReadCapacityUnits" if isRead else "dynamodb:" + indexOrTable + ":WriteCapacityUnits"
    policyName = "ScaleDynamoDBReadCapacityUtilization" if isRead else "ScaleDynamoDBWriteCapacityUtilization"

    dynamodb_scaling.register_scalable_target(ServiceNamespace='dynamodb',
                                              ResourceId=resourceId,
                                              ScalableDimension=scalableDimension,
                                              MinCapacity=min_capacity,
                                              MaxCapacity=max_capacity)

def generatePolicyName(table_name: str, isRead, index_name: str = None):
    policyName = "read-capacity-scaling-policy" if isRead else "write-capacity-scaling-policy";
    if(index_name):
      policyName = table_name + "-index-" + index_name + "-" + policyName
    else:
      policyName = table_name + "-" + policyName
    return policyName

# ===========================================================================
# Autoscaling Execution Main
# ===========================================================================
def applyAutoScalingMain():
    try:
      dynamodb_scaling = createAutoScalingClient()


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False )
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False )

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True )
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True )


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI1")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI1")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI1")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI1")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI2")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI2")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI2")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI2")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI3")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI3")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI3")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI3")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI4")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI4")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI4")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI4")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI5")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI5")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI5")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI5")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI6")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI6")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI6")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI6")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI7")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI7")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI7")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI7")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI8")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI8")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI8")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI8")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI9")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI9")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI9")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI9")


      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, False , "GSI10")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, False , "GSI10")

      register_scalable_target(dynamodb_scaling, "krabber", 1, 10, True , "GSI10")
      put_scaling_policy(dynamodb_scaling, "krabber", 70, True , "GSI10")


      print("Autoscaling Policy successfully applied")
    except ClientError as error:
      handle_error_autoscaling(error)
    except BaseException as error:
      print("Unknown error while creating table: " + error.response['Error']['Message'])

# Uncomment this code out after the main function runs to apply necessary autoscaling policies
# applyAutoScalingMain()

