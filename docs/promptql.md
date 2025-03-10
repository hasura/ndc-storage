# Use the Storage connector with PromptQL

The Storage connector can help PromptQL:

- Manage buckets and objects.
- Read file objects on popular storage services.
- Analyze text, CSV and JSON file content.

## Prompt Examples

### List and analyze CSV files

This example list objects in the bucket, download a CSV file to decode and find items that have the first name Shelia.

```
List all objects recursively.
```

```
Download the people-1000.csv file as text, decode CSV.
```

```
Find rows which have the first name Shelia in the CSV file.
```

### Analyze JSON files

This example download a JSON file, decode and analyze movies in 1990s.

```
Download the movies-1990s.json file as text, decode json and return the first 10 results
```

```
Find and count number of items that have the Comedy genre in in the 1990s movies file
```

### Dynamic Credentials

If the storage connector enables [dynamic credentials](dynamic-credentials.md) you can input custom storage accounts to access your files. PromptQL can smart enough to remember your credential context so you don't need to input again in next prompts.

```
Remember the following credentials. List all objects in the bucket recursively:
client_type: s3
endpoint: http://minio:9000
access_key_id: test-key
secret_access_key: randomsecret
bucket: default
```

```
Download the xxx file as text...
```
