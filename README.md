# Google Ads MCP Server

A Model Context Protocol (MCP) server for Google Ads API integration with OAuth 2.0 service account authentication.

## What is a Service Account?

A **service account** is a special type of Google account that represents an application or service, rather than a human user. Service accounts are used for:

- **Server-to-server authentication**: Applications can authenticate with Google APIs without user interaction
- **Automated processes**: Background services, scheduled tasks, and API integrations
- **Secure credential management**: Private keys are used instead of passwords
- **Fine-grained permissions**: Specific roles and permissions can be assigned

For Google Ads API, service accounts allow your MCP server to:
- Access Google Ads data programmatically
- Make API calls without user login prompts
- Maintain secure, long-running connections
- Scale to handle multiple requests efficiently

## Obtaining a Service Account

To get a service account for Google Ads API:

1. **Create a Google Cloud Project** (if you don't have one)
2. **Enable the Google Ads API** in your project
3. **Create a Service Account**:
   - Go to Google Cloud Console → IAM & Admin → Service Accounts
   - Click "Create Service Account"
   - Give it a name (e.g., "google-ads-api-mcp-service-acc")
   - Assign appropriate roles (e.g., "Google Ads API - Read Only")
4. **Generate a JSON Key**:
   - Click on your service account → Keys tab
   - Click "Add Key" → "Create new key" → JSON
   - Download the JSON file (this is your `your-service-account.json`)

## Setup Instructions

### Local Development Setup

1. **Create Unified Configuration File**:
   ```bash
   # Create the unified configuration file
   cp internal/app/configs/google-ads-config.json.example internal/app/configs/google-ads-config.json
   
   # Edit the file with your actual values:
   # - customer_id: Your Google Ads customer ID
   # - developer_token: Your Google Ads developer token
   # - service_account_json: Your complete service account JSON as a string
   ```

2. **Set Environment Variables** (Optional - Non-sensitive):
   ```bash
   # Optional server configuration (with defaults)
   export MCP_SERVER_HOST="0.0.0.0"
   export MCP_SERVER_PATH="/mcp"
   export PORT="8080"
   ```

3. **Run the Server**:
   ```bash
   go run main.go
   ```

### Production Setup

1. **Google Secret Manager Setup**:
   ```bash
   # Create the unified configuration secret
   gcloud secrets create GOOGLE_ADS_CONFIG \
     --data-file=google-ads-config.json \
     --project=your-project-id
   ```

2. **Configure Secret Environment Variable**:
   ```bash
   # In Google Cloud Run/Cloud Functions, configure GOOGLE_ADS_CONFIG as a secret environment variable
   # Google Cloud will automatically populate the environment variable with the secret value
   ```

3. **Environment Variables**:
   ```bash
   # Required for production
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   
   # Optional server configuration (with defaults)
   export MCP_SERVER_HOST="0.0.0.0"
   export MCP_SERVER_PATH="/mcp"
   export PORT="8080"
   ```

4. **Service Account Permissions**:
   Ensure your service account has the following permissions:
    - `roles/ads.readonly` or appropriate Google Ads API permissions
    - Access to the GOOGLE_ADS_CONFIG secret (automatically handled by Google Cloud)

5. **Deploy**:
   The server will automatically detect the environment and use the appropriate credential source.

## Architecture

- **configs/configs.go**: Hybrid configuration that reads from local file (dev) or Google Secret Manager (prod)
- **wire.go**: Dependency injection and service initialization
- **auth/token_manager.go**: OAuth 2.0 token management with automatic refresh
- **api/listadaccounts/**: Google Ads API integration
- **tools/listadaccounts/**: MCP tool implementation

## Environment Detection

The application automatically detects the environment:

1. **Local Development**:
    - Looks for `internal/app/configs/google-ads-config.json`
    - If found, reads the unified configuration from the local file
    - Contains all Google Ads API credentials (customer_id, developer_token, service_account_json)
    - No Google Cloud configuration required

2. **Production**:
    - If local config file doesn't exist, falls back to environment variable
    - Reads `GOOGLE_ADS_CONFIG` environment variable (automatically populated by Google Cloud)
    - Google Cloud Run/Cloud Functions automatically injects secret values into environment variables

## Security

- **Local Development**: All Google Ads credentials stored in local config file (excluded from git)
- **Production**: All Google Ads credentials stored securely in Google Secret Manager
- **Environment Variables**: Only non-sensitive server configuration (PORT, HOST, PATH)
- **Unified Configuration**: Single source of truth for all Google Ads API credentials
- Automatic token refresh with proper caching
- Thread-safe token management
