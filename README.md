# Google Drive API Integration with Go

This project demonstrates how to integrate with the Google Drive API using Go. It provides functionalities to create folders and upload files to Google Drive via a web interface.

## Features

- OAuth2 authentication with Google Drive.
- Create folders inside existing folders.
- Upload files to specified folders in Google Drive.

## Prerequisites

- Go 1.16 or higher.
- Google Cloud project with Drive API enabled.
- OAuth2 credentials (Client ID and Client Secret).

## Setup

### 1. Create a Google Cloud Project and Enable Drive API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new project.
3. Enable the Google Drive API for the project.
4. Create OAuth2 consent screen in `External` and add redirect URL `http://localhost:8080/callback` 
5. Create OAuth2 credentials and download the `client_id.json` file.
6. Set the `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` environment variables from the downloaded credentials.

### 2. Clone the Repository

```sh
git clone https://github.com/rishad004/google-drive-api-go.git
cd google-drive-api-go
```

### 3. Set Environment Variables

Set the following environment variables:

```sh
export GOOGLE_CLIENT_ID=your-client-id
export GOOGLE_CLIENT_SECRET=your-client-secret
```

Alternatively, you can create a `.env` file in the project root with the following content:

```sh
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
```

### 4. Run the Application

```sh
go run main.go
```

### 5. Authenticate with Google

1. Open your browser and navigate to `http://localhost:8080`.
2. Click the authentication link to authenticate with your Google account.
3. Once authenticated, you can create folders and upload files using the web interface.

## Usage

### Creating a Folder

1. Enter the parent folder ID (leave empty for root) and the name of the new folder.
2. Click "Create Folder".
3. The new folder will be created, and its ID will be displayed.

### Uploading a File

1. Enter the folder ID where you want to upload the file.
2. Choose the file to upload.
3. Click "Upload File".
4. The file will be uploaded to the specified folder, and its ID will be displayed.

## Code Overview

- **main.go**: Main application file containing the routes and logic for Google Drive API operations.
- **templates/index.html**: HTML template for the web interface.

## Token Storage

The OAuth2 token is stored in `token.json`. This file is created automatically after the first authentication and is used for subsequent API requests.

### Common Errors

- **Access blocked: Google_API_Testing has not completed the Google verification process**: Add that email id as test user in OAuth consent screen on [Google Cloud Console](https://console.cloud.google.com/)
- **File not found (404)**: Ensure the folder ID is correct and accessible.
- **Invalid parent folder ID**: Verify the parent folder ID exists and is accessible.

## Contributing

Feel free to open issues or submit pull requests if you find any bugs or want to add new features.