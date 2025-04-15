# Upload Files to S3 Object Storage with Rate limit Parameters

Generate 5000 1MB files of mp4 format and upload to Object Storage with rate limit parameters - ex 200 concurrent uploads.
## Getting Started

1. Clone the repository

```bash
git clone https://github.com/yourusername/upload-files-to-s3.git
```

2. Generate 5000 1MB files of mp4 format

```bash
./generate_file.sh
```

3. Run go upload script

```bash
go run upload_file.go -rate 200 -dir load_test_files -bucket <BUCKET_NAME> -endpoint <OBJECT_STORE_ENDPOINT> -access-key <OBJECT_STORE_ACCESS_KEY> -secret-key <OBJECT_STORE_SECRET_KEY>
```

## Parameters

- `-rate`: Number of concurrent uploads
- `-dir`: Directory containing files to upload
- `-bucket`: Bucket name
- `-endpoint`: Object store endpoint
- `-access-key`: Object store access key
- `-secret-key`: Object store secret key
